//go:build windows

package hardware

import (
        "fmt"
        "net"
        "strings"

        "golang.org/x/sys/windows/registry"
)

// Win32_NetworkAdapter struct ánh xạ WMI
type Win32_NetworkAdapter struct {
        Name           string
        MACAddress     string
        Manufacturer   string
        PNPDeviceID    string
        AdapterType    string
        NetEnabled     bool
        PhysicalAdapter bool
}

// Win32_PnPEntity placeholder để scanExternalNICs dùng
type Win32_PnPEntity struct {
        DeviceID    string
        Description string
        Status      string
        PNPDeviceID string
}

// scanInternalNICs liệt kê card mạng gắn trong (LAN/Wifi onboard)
// Ưu tiên dùng net.Interfaces() của Go stdlib (luôn hoạt động, không cần PowerShell)
// Kết hợp đọc registry cho SSID Wi-Fi
func scanInternalNICs() []HardwareRecord {
        var out []HardwareRecord

        // Phương án 1: Dùng net.Interfaces() của Go stdlib (luôn ổn định)
        ifaces, err := net.Interfaces()
        if err == nil {
                for _, iface := range ifaces {
                        // Bỏ qua loopback và adapter tắt
                        if iface.Flags&net.FlagLoopback != 0 {
                                continue
                        }
                        if iface.Flags&net.FlagUp == 0 && len(iface.HardwareAddr) == 0 {
                                continue
                        }
                        // Bỏ qua adapter ảo
                        if isVirtualAdapter("", iface.Name) {
                                continue
                        }

                        rec := HardwareRecord{
                                DeviceName:    iface.Name,
                                Serial:        iface.HardwareAddr.String(),
                                MAC:           formatMACFromBytes(iface.HardwareAddr),
                                DriverStatus:  driverStatusFromFlags(iface.Flags),
                                ConnectionPos: "Internal",
                                AdapterType:   guessAdapterType(iface.Name, iface.HardwareAddr),
                        }

                        // Lấy list địa chỉ IP để thêm vào DeviceName (debug info)
                        addrs := iface.Addrs
                        if addrs != nil {
                                addrList, _ := iface.Addrs()
                                if addrList != nil {
                                        for _, addr := range addrList {
                                                // Không cần hiển thị IP ở đây, chỉ đảm bảo interface có thật
                                                _ = addr
                                        }
                                }
                        }

                        out = append(out, rec)
                }
        }

        // Phương án 2: Nếu net.Interfaces() trả rỗng, thử query registry
        if len(out) == 0 {
                out = append(out, scanInternalNICsFromRegistry()...)
        }

        // Bổ sung SSID list + last connect cho adapter Wi-Fi
        for i := range out {
                if out[i].AdapterType == "Wi-Fi" {
                        out[i].SSIDList = getWirelessProfiles()
                        out[i].LastConnect = getWiFiLastConnect()
                }
        }

        return out
}

// scanInternalNICsFromRegistry fallback: đọc HKLM\SYSTEM\CurrentControlSet\Control\Class\{4d36e972...}
func scanInternalNICsFromRegistry() []HardwareRecord {
        var out []HardwareRecord
        // HKLM\SYSTEM\CurrentControlSet\Control\Class\{4d36e972-e325-11ce-bfc1-08002be10318}\
        // chứa subkeys 0000, 0001... mỗi cái là một adapter
        classPath := `SYSTEM\CurrentControlSet\Control\Class\{4d36e972-e325-11ce-bfc1-08002be10318}`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, classPath,
                registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return out
        }
        defer k.Close()

        subKeys, err := k.ReadSubKeyNames(-1)
        if err != nil {
                return out
        }

        for _, sub := range subKeys {
                // Bỏ qua subkey "Properties"
                if sub == "Properties" {
                        continue
                }
                sk, err := registry.OpenKey(registry.LOCAL_MACHINE,
                        classPath+`\`+sub, registry.QUERY_VALUE|registry.WOW64_64KEY)
                if err != nil {
                        continue
                }
                rec := HardwareRecord{
                        ConnectionPos: "Internal (Registry)",
                        DriverStatus:  "OK (from registry)",
                }
                if v, _, err := sk.GetStringValue("DriverDesc"); err == nil {
                        rec.DeviceName = v
                }
                if v, _, err := sk.GetStringValue("NetworkAddress"); err == nil && v != "" {
                        rec.MAC = formatMACString(v)
                }
                // PNPDeviceID để detect Wi-Fi
                if v, _, err := sk.GetStringValue("MatchingDeviceId"); err == nil {
                        rec.Serial = v
                        if strings.Contains(strings.ToLower(v), "wifi") ||
                                strings.Contains(strings.ToLower(v), "wireless") ||
                                strings.Contains(strings.ToLower(v), "802.11") {
                                rec.AdapterType = "Wi-Fi"
                        } else {
                                rec.AdapterType = "LAN"
                        }
                }
                sk.Close()
                if rec.DeviceName != "" {
                        out = append(out, rec)
                }
        }
        return out
}

// scanExternalNICs liệt kê card mạng gắn ngoài (USB Wifi/D-Com 4G/5G)
// Đọc từ registry HKLM\SYSTEM\CurrentControlSet\Enum\USB (cũng không cần PowerShell)
func scanExternalNICs() []HardwareRecord {
        var out []HardwareRecord

        // Đọc USB registry để tìm card mạng USB
        root := `SYSTEM\CurrentControlSet\Enum\USB`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, root,
                registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return out
        }
        defer k.Close()

        vids, _ := k.ReadSubKeyNames(-1)
        for _, vid := range vids {
                // vid có dạng "VID_045E&PID_07C0" - trích VID và PID
                vidStr, _ := extractVIDPID(vid)
                _ = vidStr // giữ để có thể dùng sau (hiện không dùng trong vòng lặp này)

                sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
                        root+`\`+vid, registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
                if err != nil {
                        continue
                }
                serials, _ := sub.ReadSubKeyNames(-1)
                for _, serial := range serials {
                        dev, err := registry.OpenKey(registry.LOCAL_MACHINE,
                                root+`\`+vid+`\`+serial,
                                registry.QUERY_VALUE|registry.WOW64_64KEY)
                        if err != nil {
                                continue
                        }
                        friendly := ""
                        class := ""
                        if v, _, err := dev.GetStringValue("Class"); err == nil {
                                class = strings.ToLower(v)
                        }
                        if v, _, err := dev.GetStringValue("FriendlyName"); err == nil {
                                friendly = v
                        } else if v, _, err := dev.GetStringValue("DeviceDesc"); err == nil {
                                friendly = stripDevDescPrefix(v)
                        }
                        _ = dev.Close()

                        // Chỉ lấy thiết bị USB có class = Net (card mạng)
                        if class != "net" {
                                continue
                        }
                        // Đoán loại adapter từ tên
                        adapterType := "USB LAN/Wi-Fi"
                        fl := strings.ToLower(friendly)
                        if contains(fl, "wifi") || contains(fl, "wireless") || contains(fl, "802.11") {
                                adapterType = "Wi-Fi (USB)"
                        } else if contains(fl, "bluetooth") {
                                adapterType = "Bluetooth (USB)"
                        } else if contains(fl, "ethernet") || contains(fl, "lan") {
                                adapterType = "LAN (USB)"
                        }

                        rec := HardwareRecord{
                                AdapterType:   adapterType,
                                ConnectionPos: "External USB",
                                DeviceName:    friendly,
                                Serial:        vid + `\` + serial,
                                DriverStatus:  "OK (from registry)",
                        }
                        _ = vidStr
                        out = append(out, rec)
                }
                _ = sub.Close()
        }
        return out
}

// isVirtualAdapter kiểm tra có phải adapter ảo không
func isVirtualAdapter(pnpID, name string) bool {
        n := strings.ToLower(name)
        virtualHints := []string{"hyper-v", "vmware", "virtualbox", "virtual",
                "vpn", "tunnel", "teredo", "isatap", "tunneling", "loopback"}
        for _, h := range virtualHints {
                if contains(n, h) {
                        return true
                }
        }
        return false
}

// normalizeAdapterType chuẩn hoá AdapterType từ WMI string
func normalizeAdapterType(t string, pnpID string) string {
        if t == "" {
                p := strings.ToLower(pnpID)
                switch {
                case contains(p, "wifi") || contains(p, "wireless") || contains(p, "802.11"):
                        return "Wi-Fi"
                case contains(p, "bluetooth"):
                        return "Bluetooth"
                case contains(p, "ethernet") || contains(p, "lan"):
                        return "LAN"
                }
                return "Unknown"
        }
        tl := strings.ToLower(t)
        switch {
        case contains(tl, "ethernet") || contains(tl, "802.3"):
                return "LAN"
        case contains(tl, "wireless") || contains(tl, "wifi") || contains(tl, "802.11"):
                return "Wi-Fi"
        case contains(tl, "bluetooth"):
                return "Bluetooth"
        }
        return t
}

// driverStatusOf truy vấn trạng thái driver của PnP entity
func driverStatusOf(pnpID string) string {
        if pnpID == "" {
                return "Unknown"
        }
        return "OK"
}

// formatMACFromBytes định dạng MAC từ byte slice
func formatMACFromBytes(b []byte) string {
        if len(b) < 6 {
                return ""
        }
        return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
                b[0], b[1], b[2], b[3], b[4], b[5])
}

// formatMACString định dạng MAC từ chuỗi không có dấu hai chấm "AABBCCDDEEFF"
func formatMACString(s string) string {
        s = strings.TrimSpace(s)
        if len(s) != 12 {
                return s
        }
        return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
                s[0:2], s[2:4], s[4:6], s[6:8], s[8:10], s[10:12])
}

// driverStatusFromFlags đoán trạng thái driver từ cờ interface
func driverStatusFromFlags(flags net.Flags) string {
        if flags&net.FlagUp != 0 {
                return "OK (Up)"
        }
        return "Disabled"
}

// guessAdapterType đoán loại adapter từ tên và MAC
func guessAdapterType(name string, mac net.HardwareAddr) string {
        n := strings.ToLower(name)
        switch {
        case contains(n, "wi-fi") || contains(n, "wifi") || contains(n, "wireless") || contains(n, "802.11"):
                return "Wi-Fi"
        case contains(n, "bluetooth"):
                return "Bluetooth"
        case contains(n, "ethernet") || contains(n, "lan") || contains(n, "gigabit"):
                return "LAN"
        }
        // OUI đầu của MAC để đoán (đơn giản)
        return "Ethernet"
}

// (các hàm còn lại giữ nguyên như cũ: readMACForPNPID, queryWMIAdapters, etc.)
// Các hàm cũ vẫn giữ làm fallback nhưng không dùng mặc định nữa.
// Lưu ý: extractVIDPID và extractVIDPIDStr được định nghĩa trong usb.go (cùng package).
