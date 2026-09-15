package hardware

import (
        "encoding/xml"
        "sort"
        "strings"
        "time"
)

// timeLayout là định dạng thời gian chuẩn toàn bộ app sử dụng
const timeLayout = "2006-01-02 15:04:05"

// PnPEvent là MỘT sự kiện PnP của Windows (Kernel-PnP Configuration log
// hoặc SetupAPI log), đã chuẩn hoá.
//   - EventID 400/410 : thiết bị được cấu hình / khởi động  -> LẦN CẮM
//   - EventID 420     : thiết bị bị xoá khỏi hệ thống       -> LẦN RÚT
type PnPEvent struct {
        Time     time.Time
        EventID  int
        DeviceID string // Device Instance ID đã normalize (lowercase, # -> \)
}

// DeviceSessions gom toàn bộ session của một thiết bị (theo DeviceID)
type DeviceSessions struct {
        DeviceID string           // Device Instance ID đã normalize
        Sessions []ConnectSession // sắp xếp theo thời gian tăng dần
        Arrivals int              // tổng số lần cắm đếm được
}

// pnpEventXML ánh xạ fragment XML output của wevtutil /f:xml
type pnpEventXML struct {
        System struct {
                EventID     int `xml:"EventID"`
                TimeCreated struct {
                        SystemTime string `xml:"SystemTime,attr"`
                } `xml:"TimeCreated"`
        } `xml:"System"`
        EventData struct {
                Data []struct {
                        Name  string `xml:"Name,attr"`
                        Value string `xml:",chardata"`
                } `xml:"Data"`
        } `xml:"EventData"`
}

// parseKernelPnPXMLEvents phân tích output XML thô của wevtutil thành []PnPEvent.
// wevtutil xuất nhiều phần tử <Event> liên tiếp (không phải XML document hợp lệ),
// nên ta tách theo thẻ đóng </Event> rồi unmarshal từng fragment.
func parseKernelPnPXMLEvents(raw string) []PnPEvent {
        var out []PnPEvent
        if strings.TrimSpace(raw) == "" {
                return out
        }
        parts := strings.Split(raw, "</Event>")
        for _, p := range parts {
                p = strings.TrimSpace(p)
                if p == "" {
                        continue
                }
                idx := strings.Index(p, "<Event")
                if idx < 0 {
                        continue
                }
                frag := p[idx:] + "</Event>"
                var e pnpEventXML
                if err := xml.Unmarshal([]byte(frag), &e); err != nil {
                        continue
                }
                dev := ""
                for _, d := range e.EventData.Data {
                        n := strings.ToLower(strings.TrimSpace(d.Name))
                        if n == "deviceinstanceid" || n == "deviceid" || n == "deviceinstance" {
                                dev = d.Value
                                break
                        }
                }
                if dev == "" || e.System.EventID == 0 {
                        continue
                }
                t, err := time.Parse(time.RFC3339Nano, e.System.TimeCreated.SystemTime)
                if err != nil {
                        // fallback bỏ nano
                        t, err = time.Parse(time.RFC3339, e.System.TimeCreated.SystemTime)
                        if err != nil {
                                continue
                        }
                }
                out = append(out, PnPEvent{
                        Time:     t,
                        EventID:  e.System.EventID,
                        DeviceID: normalizePnpID(dev),
                })
        }
        return out
}

// normalizePnpID chuẩn hoá Device Instance ID:
//   - "#" -> "\" (dạng registry đếm ngược \\?\USBSTOR#Disk...)
//   - bỏ tiền tố \\?\ hoặc \??\
//   - lowercase để so khớp case-insensitive
func normalizePnpID(s string) string {
        s = strings.ReplaceAll(s, "#", "\\")
        s = strings.TrimPrefix(s, `\\?\`)
        s = strings.TrimPrefix(s, `\??\`)
        s = strings.TrimPrefix(s, `\?\?`)
        return strings.ToLower(strings.TrimSpace(s))
}

// buildDeviceSessions nhóm events theo thiết bị và dựng từng session
// (cắm -> rút). Trả về map[deviceID]*DeviceSessions.
func buildDeviceSessions(events []PnPEvent) map[string]*DeviceSessions {
        if len(events) == 0 {
                return map[string]*DeviceSessions{}
        }
        evs := make([]PnPEvent, len(events))
        copy(evs, events)
        sort.Slice(evs, func(i, j int) bool { return evs[i].Time.Before(evs[j].Time) })

        // Nếu log có EventID 410 (Device started) thì ưu tiên 410 làm mốc "cắm",
        // vì 400 (Device configured) còn bắn khi cài lại driver trong cùng 1 lần cắm.
        has410 := false
        for _, e := range evs {
                if e.EventID == 410 {
                        has410 = true
                        break
                }
        }
        arrivalID := 400
        if has410 {
                arrivalID = 410
        }

        m := map[string]*DeviceSessions{}
        for _, e := range evs {
                ds := m[e.DeviceID]
                if ds == nil {
                        ds = &DeviceSessions{DeviceID: e.DeviceID}
                        m[e.DeviceID] = ds
                }
                switch e.EventID {
                case arrivalID:
                        n := len(ds.Sessions)
                        if n > 0 && ds.Sessions[n-1].Removal == "" {
                                // Session trước vẫn đang mở -> arrival trùng lặp, bỏ qua
                                continue
                        }
                        ds.Sessions = append(ds.Sessions, ConnectSession{Arrival: e.Time.Format(timeLayout)})
                        ds.Arrivals++
                case 420:
                        n := len(ds.Sessions)
                        if n > 0 && ds.Sessions[n-1].Removal == "" {
                                ds.Sessions[n-1].Removal = e.Time.Format(timeLayout)
                                ds.Sessions[n-1].Duration = vnDuration(e.Time.Sub(
                                        mustParseLayout(ds.Sessions[n-1].Arrival)))
                        }
                }
        }
        return m
}

// mustParseLayout parse theo timeLayout, trả zero time nếu lỗi (không panic)
func mustParseLayout(s string) time.Time {
        t, err := time.Parse(timeLayout, s)
        if err != nil {
                return time.Time{}
        }
        return t
}

// vnDuration định dạng khoảng thời gian kiểu tiếng Việt:
// "45 giây" | "41 phút" | "2 giờ 13 phút" | "3 ngày 4 giờ"
func vnDuration(d time.Duration) string {
        if d <= 0 {
                return "—"
        }
        m := int(d / time.Minute)
        s := int(d/time.Second) % 60
        if m < 1 {
                return itoaSmall(s) + " giây"
        }
        if m < 60 {
                return itoaSmall(m) + " phút"
        }
        h := m / 60
        mm := m % 60
        if h < 24 {
                out := itoaSmall(h) + " giờ"
                if mm > 0 {
                        out += " " + itoaSmall(mm) + " phút"
                }
                return out
        }
        days := h / 24
        hh := h % 24
        out := itoaSmall(days) + " ngày"
        if hh > 0 {
                out += " " + itoaSmall(hh) + " giờ"
        }
        return out
}

// itoaSmall chuyển int nhỏ sang chuỗi (tránh phụ thuộc strconv cho đồng bộ style)
func itoaSmall(n int) string {
        if n == 0 {
                return "0"
        }
        neg := n < 0
        if neg {
                n = -n
        }
        buf := [8]byte{}
        i := len(buf)
        for n > 0 {
                i--
                buf[i] = byte('0' + n%10)
                n /= 10
        }
        if neg {
                i--
                buf[i] = '-'
        }
        return string(buf[i:])
}

// parseSetupAPILogs phân tích nội dung file setupapi.dev.log (đã đọc sẵn)
// thành []PnPEvent (chỉ có mốc cắm - device install).
// Format 1 section:
//
//      >>>  [Device Install (Hardware initiated) - USBSTOR\DISK&VEN_...&PROD_...\7&2f3a&0]
//      >>>  Section start 2024/08/12 10:15:22.123
func parseSetupAPILogs(content string) []PnPEvent {
        var out []PnPEvent
        lines := strings.Split(content, "\n")
        currentDev := ""
        for _, line := range lines {
                l := strings.TrimRight(line, "\r")
                if strings.HasPrefix(l, ">>>  [") && strings.HasSuffix(l, "]") {
                        header := l[len(">>>  [") : len(l)-1]
                        // Header dạng "Device Install (Hardware initiated) - DEVICE_ID"
                        // hoặc "Device Install (DiInstallDriver) - DEVICE_ID"
                        if idx := strings.LastIndex(header, " - "); idx >= 0 {
                                dev := strings.TrimSpace(header[idx+3:])
                                // bỏ phần " (giấu tên)" nếu có
                                if strings.Contains(dev, " ") && !strings.Contains(dev, "\\") {
                                        dev = ""
                                }
                                currentDev = normalizePnpID(dev)
                        }
                        continue
                }
                if currentDev != "" && strings.HasPrefix(l, ">>>  Section start ") {
                        stamp := strings.TrimSpace(strings.TrimPrefix(l, ">>>  Section start "))
                        if t, err := time.Parse("2006/01/02 15:04:05.000", stamp); err == nil {
                                out = append(out, PnPEvent{Time: t, EventID: 400, DeviceID: currentDev})
                        } else if t, err2 := time.Parse("2006/01/02 15:04:05", stamp); err2 == nil {
                                out = append(out, PnPEvent{Time: t, EventID: 400, DeviceID: currentDev})
                        }
                        currentDev = ""
                }
        }
        return out
}

// extractVIDPID tách VID và PID từ DeviceID — hỗ trợ cả dạng ngắn
// "VID_0781&PID_5590" lẫn dạng PNP đầy đủ "USB\VID_045E&PID_07C0\6&2f3a&0&3"
// (pure string parser - đặt ở file đa nền tảng để dùng chung)
func extractVIDPID(s string) (vid, pid string) {
        u := strings.ToUpper(s)
        if i := strings.Index(u, "VID_"); i >= 0 {
                vid = extractHexToken(u, i+4, 6)
        }
        if i := strings.Index(u, "PID_"); i >= 0 {
                pid = extractHexToken(u, i+4, 6)
        }
        return
}

// extractHexToken lấy tối đa max ký tự hex bắt đầu từ vị trí start
func extractHexToken(u string, start, max int) string {
        var b strings.Builder
        for i := start; i < len(u) && b.Len() < max; i++ {
                c := u[i]
                if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') {
                        b.WriteByte(c)
                } else {
                        break
                }
        }
        return b.String()
}

// extractVIDPIDStr dựng chuỗi "VID_XXXX&PID_YYYY" từ Device Instance ID
func extractVIDPIDStr(pnp string) string {
        vid, pid := extractVIDPID(pnp)
        if vid == "" {
                return ""
        }
        return "VID_" + vid + "&PID_" + pid
}

// sessionsToHistRec chuyển DeviceSessions thành PeripheralRec lịch sử
// (dùng chung cho cả windows và non-windows demo)
func sessionsToHistRec(devID string, ds *DeviceSessions) PeripheralRec {
        rec := PeripheralRec{
                HardwareID:  devID,
                VIDPID:      extractVIDPIDStr(devID),
                VendorModel: friendlyNameForPNP(devID),
                PlugCount:   ds.Arrivals,
                Sessions:    ds.Sessions,
        }
        if n := len(ds.Sessions); n > 0 {
                rec.FirstPlug = ds.Sessions[0].Arrival
                rec.LastPlug = ds.Sessions[n-1].Arrival
        }
        return rec
}
