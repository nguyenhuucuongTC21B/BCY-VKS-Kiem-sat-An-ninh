//go:build windows

package hardware

import (
        "bytes"
        "os/exec"
        "strings"
        "time"

        "golang.org/x/sys/windows/registry"
)

// getWirelessProfiles chạy netsh wlan show profiles để lấy SSID đã từng kết nối
// Trả về chuỗi phân cách bằng phẩy
func getWirelessProfiles() string {
        cmd := exec.Command("netsh", "wlan", "show", "profiles")
        var stdout bytes.Buffer
        cmd.Stdout = &stdout
        if err := cmd.Run(); err != nil {
                return readWirelessProfilesFromRegistry()
        }
        out := parseWirelessProfiles(stdout.String())
        if out == "" {
                // netsh chạy được nhưng locale lạ / không parse được -> dùng registry
                out = readWirelessProfilesFromRegistry()
        }
        return out
}

// parseWirelessProfiles tách output của netsh wlan show profiles
// Tìm các dòng có dạng: "    All User Profile     : SSID_NAME"
// hoặc locale VN: "    Tất cả người dùng     : SSID_NAME"
func parseWirelessProfiles(s string) string {
        var ssids []string
        for _, line := range strings.Split(s, "\n") {
                l := strings.TrimSpace(line)
                // Tìm dấu ":" rồi lấy phần sau
                idx := strings.Index(l, ":")
                if idx < 0 {
                        continue
                }
                // Phải chứa chữ "profile" hoặc "hồ sơ" để đảm bảo đúng dòng
                ll := strings.ToLower(l)
                if !strings.Contains(ll, "profile") && !strings.Contains(ll, "hồ sơ") {
                        continue
                }
                ssid := strings.TrimSpace(l[idx+1:])
                if ssid != "" {
                        ssids = append(ssids, ssid)
                }
        }
        return strings.Join(ssids, ",")
}

// readWirelessProfilesFromRegistry fallback khi netsh không chạy được:
// đọc ProfileName từ HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles
// (NetworkList ghi lại mọi mạng Wi-Fi/LAN máy từng kết nối)
func readWirelessProfilesFromRegistry() string {
        key := `SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, key,
                registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return ""
        }
        defer k.Close()

        names, _ := k.ReadSubKeyNames(-1)
        if len(names) > 30 {
                names = names[:30]
        }
        seen := map[string]bool{}
        var ssids []string
        for _, n := range names {
                sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
                        key+`\`+n, registry.QUERY_VALUE|registry.WOW64_64KEY)
                if err != nil {
                        continue
                }
                name := ""
                if v, _, err := sub.GetStringValue("ProfileName"); err == nil {
                        name = strings.TrimSpace(v)
                }
                _ = sub.Close()
                if name == "" || seen[strings.ToLower(name)] {
                        continue
                }
                seen[strings.ToLower(name)] = true
                ssids = append(ssids, name)
        }
        return strings.Join(ssids, ",")
}

// getWiFiLastConnect trả về thời điểm kết nối Wi-Fi gần nhất.
//   - Nếu Wi-Fi đang connected (netsh show interfaces) -> "Đang kết nối (mới nhất)"
//   - Ngược lại: đọc max(DateLastConnected) từ NetworkList\Profiles (FILETIME)
func getWiFiLastConnect() string {
        if isWiFiConnectedNow() {
                return "Đang kết nối (mới nhất)"
        }
        return latestDateLastConnected()
}

// isWiFiConnectedNow kiểm tra netsh wlan show interfaces có trạng thái connected
// không (hỗ trợ cả locale EN và VN)
func isWiFiConnectedNow() bool {
        cmd := exec.Command("netsh", "wlan", "show", "interfaces")
        var stdout bytes.Buffer
        cmd.Stdout = &stdout
        if err := cmd.Run(); err != nil {
                return false
        }
        for _, line := range strings.Split(stdout.String(), "\n") {
                idx := strings.Index(line, ":")
                if idx < 0 {
                        continue
                }
                left := strings.ToLower(strings.TrimSpace(line[:idx]))
                right := strings.ToLower(strings.TrimSpace(line[idx+1:]))
                isStateLine := strings.Contains(left, "state") || strings.Contains(left, "trạng thái")
                if !isStateLine {
                        continue
                }
                // Rà "disconnected" TRƯỚC vì nó chứa cả "connect"
                if strings.Contains(right, "disconnected") || strings.Contains(right, "ngắt kết nối") ||
                        strings.Contains(right, "không kết nối") {
                        continue
                }
                if strings.Contains(right, "connected") || strings.Contains(right, "kết nối") {
                        return true
                }
        }
        return false
}

// latestDateLastConnected duyệt NetworkList\Profiles, trả về thời điểm
// DateLastConnected mới nhất (định dạng "2006-01-02 15:04").
// Giá trị DateLastConnected là REG_BINARY 8-byte FILETIME.
func latestDateLastConnected() string {
        key := `SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, key,
                registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return ""
        }
        defer k.Close()

        names, _ := k.ReadSubKeyNames(-1)
        var latest uint64
        for _, n := range names {
                sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
                        key+`\`+n, registry.QUERY_VALUE|registry.WOW64_64KEY)
                if err != nil {
                        continue
                }
                v, _, err := sub.GetBinaryValue("DateLastConnected")
                _ = sub.Close()
                if err != nil || len(v) < 8 {
                        continue
                }
                ft := uint64(v[0]) | uint64(v[1])<<8 | uint64(v[2])<<16 | uint64(v[3])<<24 |
                        uint64(v[4])<<32 | uint64(v[5])<<40 | uint64(v[6])<<48 | uint64(v[7])<<56
                if ft > latest {
                        latest = ft
                }
        }
        if latest == 0 {
                return ""
        }
        const epochDiff = uint64(116444736000000000)
        if latest < epochDiff {
                return ""
        }
        t := time.Unix(0, int64(latest-epochDiff)*100)
        return t.Local().Format("2006-01-02 15:04")
}
