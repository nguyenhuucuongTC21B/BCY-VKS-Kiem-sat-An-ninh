//go:build windows

package hardware

import (
        "bytes"
        "context"
        "io"
        "os"
        "os/exec"
        "time"
)

// scanDeviceSessions truy vết lịch sử TỪNG LẦN kết nối thiết bị ngoại vi.
//
// Nguồn dữ liệu (theo thứ tự ưu tiên):
//  1. Event Log "Microsoft-Windows-Kernel-PnP/Configuration"
//     - EventID 400 (Device configured) / 410 (Device started) -> lần cắm
//     - EventID 420 (Device deleted/removed)                   -> lần rút
//     (Log này mặc định bật trên Windows 10/11, giữ ~1024 event gần nhất)
//  2. Fallback: C:\Windows\INF\setupapi.dev.log (nhật ký cài đặt thiết bị)
//  3. Fallback cuối: Properties\LastArrival/LastRemoval trong registry
//     (xử lý ở scanPeripheralHistory trong usb.go)
//
// Trả về map[deviceID]*DeviceSessions với deviceID đã normalize.
func scanDeviceSessions() map[string]*DeviceSessions {
        events := queryKernelPnPEvents()
        if len(events) == 0 {
                events = readSetupAPILog()
        }
        if len(events) == 0 {
                return map[string]*DeviceSessions{}
        }
        return buildDeviceSessions(events)
}

// queryKernelPnPEvents gọi wevtutil truy vấn Kernel-PnP Configuration log.
// Giới hạn 1000 event gần nhất (rd:true = newest first) và timeout 20s
// để không treo nhóm quét "peripheral" (tổng timeout của nhóm là 60s).
func queryKernelPnPEvents() []PnPEvent {
        ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
        defer cancel()
        cmd := exec.CommandContext(ctx, "wevtutil", "qe",
                "Microsoft-Windows-Kernel-PnP/Configuration",
                "/q:*[System[(EventID=400) or (EventID=410) or (EventID=420)]]",
                "/c:1000", "/rd:true", "/f:xml")
        var stdout bytes.Buffer
        cmd.Stdout = &stdout
        if err := cmd.Run(); err != nil {
                return nil
        }
        return parseKernelPnPXMLEvents(stdout.String())
}

// setupAPILogMaxRead giới hạn dung lượng đọc (file có thể rất lớn)
const setupAPILogMaxRead = 1 << 19 // 512 KB cuối file

// readSetupAPILog đọc 512KB cuối của C:\Windows\INF\setupapi.dev.log
// và parse các section cài đặt thiết bị thành mốc "cắm".
func readSetupAPILog() []PnPEvent {
        f, err := os.Open(`C:\Windows\INF\setupapi.dev.log`)
        if err != nil {
                return nil
        }
        defer f.Close()

        st, err := f.Stat()
        if err != nil {
                return nil
        }
        offset := int64(0)
        if st.Size() > setupAPILogMaxRead {
                offset = st.Size() - setupAPILogMaxRead
        }
        if _, err = f.Seek(offset, io.SeekStart); err != nil {
                return nil
        }
        buf := make([]byte, setupAPILogMaxRead)
        n, err := f.Read(buf)
        if err != nil && n == 0 {
                return nil
        }
        // Nếu đọc giữa chừng thì bỏ nửa dòng đầu để tránh section hỏng
        content := string(buf[:n])
        if offset > 0 {
                if idx := indexNewline(content); idx >= 0 {
                        content = content[idx+1:]
                }
        }
        return parseSetupAPILogs(content)
}

func indexNewline(s string) int {
        for i := 0; i < len(s); i++ {
                if s[i] == '\n' {
                        return i
                }
        }
        return -1
}
