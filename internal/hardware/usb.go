//go:build windows

package hardware

import (
        "golang.org/x/sys/windows/registry"
        "strings"
)

// scanCurrentPeripherals liệt kê thiết bị ngoại vi đang cắm
// Kết hợp 3 nhánh registry: USB, USBSTOR, MountedDevices
func scanCurrentPeripherals() []PeripheralRec {
        var out []PeripheralRec

        // 1. USBSTOR - thiết bị lưu trữ USB (Flash, External HDD)
        out = append(out, scanUSBSTOR()...)

        // 2. USB - các thiết bị USB khác (printer, keyboard, mouse)
        out = append(out, scanUSBDevices()...)

        // 3. MountedDevices - lấy ký tự ổ đĩa
        driveMap := scanMountedDevices()
        for i := range out {
                if letter, ok := driveMap[out[i].HardwareID]; ok {
                        out[i].DriveLetter = letter
                }
        }

        return out
}

// scanUSBSTOR đọc HKLM\SYSTEM\CurrentControlSet\Enum\USBSTOR
func scanUSBSTOR() []PeripheralRec {
        var out []PeripheralRec
        root := `SYSTEM\CurrentControlSet\Enum\USBSTOR`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, root,
                registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return out
        }
        defer k.Close()

        classes, _ := k.ReadSubKeyNames(-1)
        for _, class := range classes {
                // class có dạng "Ven_SanDisk&Prod_Ultra_Flar"
                // Trích Vendor/Model từ class
                vendor, model := parseUSBSTORClass(class)

                sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
                        root+`\`+class, registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
                if err != nil {
                        continue
                }
                serials, _ := sub.ReadSubKeyNames(-1)
                for _, serial := range serials {
                        // serial chính là Hardware ID độc nhất của thiết bị
                        dev, err := registry.OpenKey(registry.LOCAL_MACHINE,
                                root+`\`+class+`\`+serial,
                                registry.QUERY_VALUE|registry.WOW64_64KEY)
                        if err != nil {
                                continue
                        }
                        friendly := ""
                        if v, _, err := dev.GetStringValue("FriendlyName"); err == nil {
                                friendly = v
                        } else if v, _, err := dev.GetStringValue("DeviceDesc"); err == nil {
                                friendly = v
                        }
                        _ = dev.Close()

                        rec := PeripheralRec{
                                DeviceType:  "USB Storage",
                                VendorModel: friendly,
                                HardwareID:  class + "\\" + serial,
                                VIDPID:      "", // USBSTOR không có VID/PID trực tiếp
                        }
                        if rec.VendorModel == "" {
                                rec.VendorModel = vendor + " " + model
                        }
                        out = append(out, rec)
                }
                _ = sub.Close()
        }
        return out
}

// parseUSBSTORClass tách vendor + model từ class string
// Format: "Ven_SanDisk&Prod_Ultra_Flar&Rev_1.00"
func parseUSBSTORClass(s string) (vendor, model string) {
        parts := strings.Split(s, "&")
        for _, p := range parts {
                if strings.HasPrefix(p, "Ven_") {
                        vendor = strings.TrimPrefix(p, "Ven_")
                        vendor = strings.ReplaceAll(vendor, "_", " ")
                } else if strings.HasPrefix(p, "Prod_") {
                        model = strings.TrimPrefix(p, "Prod_")
                        model = strings.ReplaceAll(model, "_", " ")
                }
        }
        return
}

// scanUSBDevices đọc HKLM\SYSTEM\CurrentControlSet\Enum\USB
// (gồm cả thiết bị có VID/PID thật)
func scanUSBDevices() []PeripheralRec {
        var out []PeripheralRec
        root := `SYSTEM\CurrentControlSet\Enum\USB`
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, root,
                registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
        if err != nil {
                return out
        }
        defer k.Close()

        vids, _ := k.ReadSubKeyNames(-1)
        for _, vid := range vids {
                // vid có dạng "VID_045E&PID_07C0"
                vidStr, pidStr := extractVIDPID(vid)

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
                        deviceType := "USB Device"
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

                        // Phân loại thiết bị theo class
                        switch class {
                        case "keyboard":
                                deviceType = "USB Keyboard"
                        case "mouse":
                                deviceType = "USB Mouse"
                        case "printer":
                                deviceType = "USB Printer"
                        case "image":
                                deviceType = "USB Scanner/Camera"
                        case "usb":
                                // Generic USB - đoán dựa VID/PID
                                deviceType = "USB Device"
                        }

                        rec := PeripheralRec{
                                DeviceType:  deviceType,
                                VendorModel: friendly,
                                HardwareID:  serial,
                                VIDPID:      "VID_" + vidStr + "&PID_" + pidStr,
                        }
                        out = append(out, rec)
                }
                _ = sub.Close()
        }
        return out
}

// extractVIDPID tách VID_xxxx và PID_xxxx từ DeviceID string
func extractVIDPID(s string) (vid, pid string) {
        for _, p := range strings.Split(s, "&") {
                if strings.HasPrefix(strings.ToUpper(p), "VID_") {
                        vid = p[4:]
                } else if strings.HasPrefix(strings.ToUpper(p), "PID_") {
                        pid = p[4:]
                }
        }
        return
}

// scanMountedDevices trả về map[hardwareID]driveLetter
// HKLM\SYSTEM\MountedDevices chứa giá trị binary cho mỗi DOS Devices
func scanMountedDevices() map[string]string {
        out := map[string]string{}
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\MountedDevices`,
                registry.QUERY_VALUE|registry.WOW64_64KEY)
        if err != nil {
                return out
        }
        defer k.Close()

        names, err := k.ReadValueNames(-1)
        if err != nil {
                return out
        }
        for _, name := range names {
                // Chỉ quan tâm các giá trị có dạng "\DosDevices\E:"
                if !strings.HasPrefix(name, `\DosDevices\`) {
                        continue
                }
                letter := strings.TrimPrefix(name, `\DosDevices\`)
                v, _, err := k.GetBinaryValue(name)
                if err != nil || len(v) < 22 {
                        continue
                }
                // Vùng UTF-16 unique ID bắt đầu từ offset 4 (sau 2 DWORD header)
                // Binary: offset 4 - là unique ID có chứa HardwareID trong Unicode
                hwid := decodeUTF16FromBytes(v[4:])
                if hwid != "" {
                        out[hwid] = letter
                }
        }
        return out
}

// decodeUTF16FromBytes giải mã chuỗi UTF-16LE từ byte slice
func decodeUTF16FromBytes(b []byte) string {
        if len(b) < 2 {
                return ""
        }
        // Lấy tới khi gặp null terminator (0x00 0x00)
        var runes []rune
        for i := 0; i+1 < len(b); i += 2 {
                c := uint16(b[i]) | uint16(b[i+1])<<8
                if c == 0 {
                        break
                }
                runes = append(runes, rune(c))
        }
        return string(runes)
}

// scanPeripheralHistory đọc lịch sử kết nối (FirstPlug, LastPlug, PlugCount)
// từ SetupAPI logs: HKLM\SYSTEM\CurrentControlSet\Enum\USB\<VID>\<serial>\Properties
// + Event Logs Microsoft-Windows-DriverFrameworks-UserMode
func scanPeripheralHistory() []PeripheralRec {
        // Trích xuất FirstPlug/LastPlug từ giá trị Properties\83da6326-... (Container ID)
        // trong nhánh Enum\USBSTOR\<...>\<serial>
        var out []PeripheralRec
        g := scanGhostedDevices()
        for _, dev := range g {
                rec := PeripheralRec{
                        HardwareID: extractSerialFromPNP(dev.PNPDeviceID),
                        VIDPID:     extractVIDPIDStr(dev.PNPDeviceID),
                        VendorModel: dev.FriendlyName,
                        FirstPlug:   readFirstArrival(dev.PNPDeviceID),
                        LastPlug:    readLastRemoval(dev.PNPDeviceID),
                }
                out = append(out, rec)
        }
        return out
}

// extractSerialFromPNP tách serial từ PNPDeviceID có dạng VID_...&PID_...\<serial>
func extractSerialFromPNP(pnp string) string {
        idx := strings.LastIndex(pnp, `\`)
        if idx < 0 {
                return pnp
        }
        return pnp[idx+1:]
}

func extractVIDPIDStr(pnp string) string {
        vid, pid := extractVIDPID(pnp)
        if vid == "" {
                return ""
        }
        return "VID_" + vid + "&PID_" + pid
}

// readFirstArrival đọc thời điểm cắm lần đầu từ Properties\ContainerId\LastArrival
// (lưu nhánh DeviceArrival)
func readFirstArrival(pnp string) string {
        // Implement thực: mở HKLM\SYSTEM\CurrentControlSet\Enum\<pnp>\Properties
        //                 và lấy {83da6326-...}\83 (LastArrival) - filetime 8 byte
        return ""
}

// readLastRemoval đọc thời điểm rút thiết bị lần cuối
func readLastRemoval(pnp string) string {
        return ""
}
