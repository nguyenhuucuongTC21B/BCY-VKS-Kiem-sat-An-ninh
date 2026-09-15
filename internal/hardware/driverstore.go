//go:build windows

package hardware

import (
        "golang.org/x/sys/windows/registry"
        "strings"
)

// GhostedDevice là thiết bị đã từng cắm và cài driver, hiện đã rút
type GhostedDevice struct {
        PNPDeviceID string
        FriendlyName string
        LastArrived string
}

// scanGhostedDevices đọc HKLM\SYSTEM\CurrentControlSet\Enum\USB và
// lọc ra các thiết bị có cờ "Ghosted" trong thuộc tính (phan hiện đã rút)
func scanGhostedDevices() []GhostedDevice {
        var out []GhostedDevice
        root := `SYSTEM\CurrentControlSet\Enum\USB`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, root, registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return out
        }
        defer k.Close()

        vids, _ := k.ReadSubKeyNames(-1)
        for _, vid := range vids {
                // vid có dạng "VID_045E&PID_07C0"
                sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
                        root+`\`+vid, registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
                if err != nil {
                        continue
                }
                serials, _ := sub.ReadSubKeyNames(-1)
                for _, serial := range serials {
                        // Mỗi serial là một instance của thiết bị
                        dev, err := registry.OpenKey(registry.LOCAL_MACHINE,
                                root+`\`+vid+`\`+serial, registry.QUERY_VALUE|registry.WOW64_64KEY)
                        if err != nil {
                                continue
                        }
                        friendly := ""
                        if v, _, err := dev.GetStringValue("FriendlyName"); err == nil {
                                friendly = v
                        } else if v, _, err := dev.GetStringValue("DeviceDesc"); err == nil {
                                friendly = v
                        }
                        out = append(out, GhostedDevice{
                                // PNPDeviceID chuẩn: USB\VID_xxxx&PID_xxxx\<serial>
                                // (bị thiếu tiền tố "USB\" - đã sửa để tra Properties được)
                                PNPDeviceID:  `USB\` + vid + `\` + serial,
                                FriendlyName: stripDevDescPrefix(friendly),
                        })
                        _ = dev.Close()
                }
                _ = sub.Close()
        }
        return out
}

// stripDevDescPrefix bỏ tiền tố "@xxx.dll,-1234;" trong DeviceDesc
func stripDevDescPrefix(s string) string {
        if s == "" {
                return s
        }
        if idx := strings.LastIndex(s, ";"); idx >= 0 && idx+1 < len(s) {
                return strings.TrimSpace(s[idx+1:])
        }
        return s
}
