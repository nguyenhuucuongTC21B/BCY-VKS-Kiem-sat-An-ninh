//go:build windows

package network

import (
        "fmt"
        "net"
        "os/exec"
        "strings"
        "time"

        "golang.org/x/sys/windows/registry"
)

// IPInfo tóm tắt thông tin IP/MAC/ISP hiện tại
type IPInfo struct {
        IP  string
        MAC string
        ISP string
}

// getInternetStatus kiểm tra máy có kết nối Internet không
// bằng cách thử TCP tới 8.8.8.8:53 (Google DNS) với timeout 3s
func getInternetStatus() string {
        conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 3*time.Second)
        if err != nil {
                return "Disconnected"
        }
        _ = conn.Close()
        return "Connected"
}

// getIPInfo trích xuất IP nội bộ, MAC, và nhà mạng ISP
func getIPInfo() IPInfo {
        out := IPInfo{}

        // IP nội bộ qua iface
        if ifaces, err := net.Interfaces(); err == nil {
                for _, iface := range ifaces {
                        if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
                                continue
                        }
                        if len(iface.HardwareAddr) >= 6 {
                                out.MAC = formatMAC(iface.HardwareAddr)
                        }
                        addrs, err := iface.Addrs()
                        if err != nil {
                                continue
                        }
                        for _, addr := range addrs {
                                if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
                                        out.IP = ipnet.IP.String()
                                        break
                                }
                        }
                        if out.IP != "" {
                                break
                        }
                }
        }

        // ISP: đọc từ NetworkList Profile nếu có
        out.ISP = readISPFromNetworkList()
        if out.ISP == "" {
                out.ISP = detectISPByMAC()
        }
        return out
}

// formatMAC định dạng MAC từ byte slice thành AA:BB:CC:DD:EE:FF
func formatMAC(b []byte) string {
        if len(b) < 6 {
                return ""
        }
        return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
                b[0], b[1], b[2], b[3], b[4], b[5])
}

// readISPFromNetworkList đọc nhà mạng từ nhánh NetworkList Profiles
func readISPFromNetworkList() string {
        key := `SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return ""
        }
        defer k.Close()
        names, err := k.ReadSubKeyNames(-1)
        if err != nil || len(names) == 0 {
                return ""
        }
        // Lấy profile đầu tiên - thường là profile mạng active
        sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
                key+`\`+names[0], registry.QUERY_VALUE|registry.WOW64_64KEY)
        if err != nil {
                return ""
        }
        defer sub.Close()
        if v, _, err := sub.GetStringValue("Description"); err == nil && v != "" {
                return v
        }
        return ""
}

// detectISPByBackup: dùng OUI (3 byte đầu MAC) để đoán nhà sản xuất NIC
// (chỉ là gợi ý, không chính xác tuyệt đối)
func detectISPByMAC() string {
        return "Unknown ISP"
}

// getConnectionHistory trích xuất lịch sử kết nối Internet từ NetworkList
// Trả về chuỗi nhiều dòng, mỗi dòng: "timestamp | ip | network_name"
func getConnectionHistory() string {
        out := &strings.Builder{}
        key := `SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return ""
        }
        defer k.Close()
        names, _ := k.ReadSubKeyNames(-1)
        for i, n := range names {
                if i >= 20 {
                        break
                }
                sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
                        key+`\`+n, registry.QUERY_VALUE|registry.WOW64_64KEY)
                if err != nil {
                        continue
                }
                desc := ""
                if v, _, err := sub.GetStringValue("Description"); err == nil {
                        desc = v
                }
                // DateLastConnected là FILETIME - cần decode
                last := readProfileLastConnected(sub)
                _ = sub.Close()
                fmt.Fprintf(out, "%s | %s | %s\n", last, n, desc)
        }
        return out.String()
}

// readProfileLastConnected giải mã FILETIME trong nhánh Profile
func readProfileLastConnected(k registry.Key) string {
        // Profile có giá trị binary "DateLastConnected" dạng FILETIME 8 byte
        v, _, err := k.GetBinaryValue("DateLastConnected")
        if err != nil || len(v) < 8 {
                return "unknown"
        }
        ft := uint64(v[0]) | uint64(v[1])<<8 | uint64(v[2])<<16 | uint64(v[3])<<24 |
                uint64(v[4])<<32 | uint64(v[5])<<40 | uint64(v[6])<<48 | uint64(v[7])<<56
        if ft == 0 {
                return "unknown"
        }
        // FILETIME = 100ns intervals since 1601-01-01
        t := time.Unix(0, int64(ft-116444736000000000)*100)
        return t.Format("2006-01-02 15:04:05")
}

// detectISPByMAC: dự phòng - trả về chuỗi rỗng nếu không đọc được NetworkList
func detectISPByMAC_() string { return "" }

// runIPConfig chạy lệnh ipconfig /all và trả về output
// (dùng để đối chiếu thêm DNS, Gateway...)
func runIPConfig() string {
        out, err := exec.Command("ipconfig", "/all").CombinedOutput()
        if err != nil {
                return ""
        }
        return string(out)
}
