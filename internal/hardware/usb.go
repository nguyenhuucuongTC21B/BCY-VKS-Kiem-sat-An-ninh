//go:build windows

package hardware

import (
        "encoding/binary"
        "strings"
        "time"

        "golang.org/x/sys/windows/registry"
)

// DEVPKEY Properties: {83da6326-97a6-4088-9453-a1923f573b29}
//   - 0064 = DEVPKEY_Device_LastArrival (thời điểm cắm gần nhất)
//   - 0066 = DEVPKEY_Device_LastRemoval (thời điểm rút gần nhất)
const (
        propLastArrival = "0064"
        propLastRemoval = "0066"
        devPropsGUID    = "{83da6326-97a6-4088-9453-a1923f573b29}"
)

// MountedVolume là một entry trong SYSTEM\MountedDevices
type MountedVolume struct {
        Letter  string // "E:"
        Decoded string // chuỗi UTF-16 đã decode, vd "USBSTOR\Disk&Ven_...\07815...&0#{...}"
}

// scanCurrentPeripherals liệt kê thiết bị ngoại vi đang cắm
// Kết hợp 3 nhánh registry: USB, USBSTOR, MountedDevices
func scanCurrentPeripherals() []PeripheralRec {
        var out []PeripheralRec

        // 1. USBSTOR - thiết bị lưu trữ USB (Flash, External HDD)
        out = append(out, scanUSBSTOR()...)

        // 2. USB - các thiết bị USB khác (printer, keyboard, mouse)
        out = append(out, scanUSBDevices()...)

        // 3. MountedDevices - lấy ký tự ổ đĩa cho thiết bị lưu trữ
        mounted := scanMountedDevices()
        for i := range out {
                if letter, ok := matchDriveLetter(mounted, out[i]); ok {
                        out[i].DriveLetter = letter
                }
        }

        return out
}

// matchDriveLetter đối chiếu record với danh sách MountedDevices.
// Giá trị MountedDevices chứa Device Instance ID dạng UTF-16 với "#" làm separator
// (vd "_??_USBSTOR#Disk&Ven_SanDisk...#07815FB8&0#{53f56307-...}"),
// nên ta match theo serial tail / HardwareID substring (case-insensitive).
func matchDriveLetter(mounted []MountedVolume, rec PeripheralRec) (string, bool) {
        recLower := strings.ToLower(rec.HardwareID)
        serialTail := strings.ToLower(pnpSerialTail(rec.HardwareID))
        for _, mv := range mounted {
                d := strings.ToLower(mv.Decoded)
                if recLower != "" && strings.Contains(d, recLower) {
                        return mv.Letter, true
                }
                if len(serialTail) >= 6 && strings.Contains(d, serialTail) {
                        return mv.Letter, true
                }
                if rec.VIDPID != "" && strings.Contains(d, strings.ToLower(rec.VIDPID)) {
                        return mv.Letter, true
                }
        }
        return "", false
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
                // class có dạng "Disk&Ven_SanDisk&Prod_Ultra_Flair&Rev_1.00"
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
                                friendly = stripDevDescPrefix(v)
                        }
                        _ = dev.Close()

                        rec := PeripheralRec{
                                DeviceType:  "USB Storage",
                                VendorModel: friendly,
                                HardwareID:  class + "\\" + serial,
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
// Format: "Disk&Ven_SanDisk&Prod_Ultra_Flair&Rev_1.00"
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

// scanMountedDevices trả về danh sách MountedDevices đã decode.
// HKLM\SYSTEM\MountedDevices: mỗi giá trị "\DosDevices\E:" là REG_BINARY
// chứa chuỗi UTF-16LE. Cấu trúc byte:
//   - 8 byte header (signature + offset)
//   - từ offset 8: chuỗi UTF-16LE ví dụ "_??_USBSTOR#Disk&Ven_...#serial#{guid}"
//
// Code cũ decode từ offset 4 -> ký tự rác đầu chuỗi và lệch alignment,
// làm việc match HardwareID thất bại. Ở đây decode từ offset 8, và nếu
// chuỗi không hợp lệ thì thử quét tìm cụm "USBSTOR"/"Volume" trong binary.
func scanMountedDevices() []MountedVolume {
        var out []MountedVolume
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
                if err != nil || len(v) < 24 {
                        continue
                }
                decoded := bestDecodeMountedBinary(v)
                if decoded != "" {
                        out = append(out, MountedVolume{Letter: letter, Decoded: decoded})
                }
        }
        return out
}

// bestDecodeMountedBinary decode chuỗi UTF-16LE từ binary MountedDevices
// một cách bền bỉ: thử offset chuẩn (8), rồi offset khác, rồi quét pattern.
func bestDecodeMountedBinary(v []byte) string {
        // 1. Offset chuẩn 8
        if s := decodeUTF16FromBytes(v[8:]); strings.Contains(strings.ToLower(s), "\\") {
                return s
        }
        // 2. Thử một số offset khác (một số build Windows dùng header 4/12 byte)
        for _, off := range []int{4, 12, 0} {
                if off >= len(v) {
                        continue
                }
                if s := decodeUTF16FromBytes(v[off:]); s != "" && looksLikeMountPath(s) {
                        return s
                }
        }
        // 3. Quét tìm cụm "usbstor" hoặc "volume" dạng UTF-16LE trong binary
        for _, pat := range []string{"usbstor", "volume"} {
                if idx := indexUTF16LECaseInsensitive(v, pat); idx > 0 {
                        // lùi lại để lấy cả prefix "\??\" nếu có
                        start := idx - 8 // "\??\" = 4 ký tự UTF-16 = 8 byte
                        if start < 0 {
                                start = 0
                        }
                        if s := decodeUTF16FromBytes(v[start:]); s != "" {
                                return s
                        }
                }
        }
        return ""
}

// looksLikeMountPath kiểm tra chuỗi decode có giống đường dẫn mount không
func looksLikeMountPath(s string) bool {
        l := strings.ToLower(s)
        return strings.HasPrefix(l, "\\??\\") || strings.HasPrefix(l, "_??_") ||
                strings.Contains(l, "usbstor") || strings.Contains(l, "volume{") ||
                strings.Contains(l, "storage")
}

// indexUTF16LECaseInsensitive tìm vị trí byte của pattern ASCII trong chuỗi UTF-16LE
func indexUTF16LECaseInsensitive(data []byte, pattern string) int {
        // Chuẩn bị pattern dạng UTF-16LE thường + HOA
        lower := []byte(pattern)
        upper := []byte(strings.ToUpper(pattern))
        var patL, patU []byte
        for _, ch := range lower {
                patL = append(patL, ch, 0)
        }
        for _, ch := range upper {
                patU = append(patU, ch, 0)
        }
        for i := 0; i+1 < len(data); i += 2 {
                if matchAt(data, i, patL) || matchAt(data, i, patU) {
                        return i
                }
        }
        return -1
}

func matchAt(data []byte, off int, pat []byte) bool {
        if off+len(pat) > len(data) {
                return false
        }
        for i := 0; i < len(pat); i++ {
                if data[off+i] != pat[i] {
                        return false
                }
        }
        return true
}

// decodeUTF16FromBytes giải mã chuỗi UTF-16LE từ byte slice tới null terminator
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
                // Chuỗi mount path thường là ASCII; bỏ qua ký tự không in được
                // (dấu hiệu decode sai offset)
                if c < 0x20 || c > 0x7E {
                        return ""
                }
                runes = append(runes, rune(c))
        }
        return string(runes)
}

// fileTimeToTime chuyển FILETIME 8-byte (100ns từ 1601-01-01) sang time.Time
func fileTimeToTime(b []byte) time.Time {
        if len(b) < 8 {
                return time.Time{}
        }
        ft := binary.LittleEndian.Uint64(b)
        if ft == 0 {
                return time.Time{}
        }
        // Khoảng 1601 -> 1970 là 11644473600 giây
        const epochDiff = uint64(116444736000000000)
        if ft < epochDiff {
                return time.Time{}
        }
        ns := int64(ft-epochDiff) * 100
        return time.Unix(0, ns)
}

// readDevicePropertyFileTime đọc giá trị FILETIME từ
// HKLM\SYSTEM\CurrentControlSet\Enum\<pnp>\Properties\{guid}\<prop>\00000000
func readDevicePropertyFileTime(pnp, prop string) string {
        if pnp == "" {
                return ""
        }
        root := `SYSTEM\CurrentControlSet\Enum\` + pnp + `\Properties\` + devPropsGUID + `\` + prop
        k, err := registry.OpenKey(registry.LOCAL_MACHINE, root,
                registry.QUERY_VALUE|registry.WOW64_64KEY)
        if err != nil {
                return ""
        }
        defer k.Close()

        // Value name chuẩn là "00000000", nhưng đọc value đầu tiên cho chắc
        if v, _, err := k.GetBinaryValue("00000000"); err == nil {
                if t := fileTimeToTime(v); !t.IsZero() {
                        return t.Local().Format(timeLayout)
                }
        }
        names, _ := k.ReadValueNames(-1)
        for _, n := range names {
                v, _, err := k.GetBinaryValue(n)
                if err != nil {
                        continue
                }
                if t := fileTimeToTime(v); !t.IsZero() {
                        return t.Local().Format(timeLayout)
                }
        }
        return ""
}

// readFirstArrival đọc thời điểm cắm gần nhất từ Properties\LastArrival (0064)
func readFirstArrival(pnp string) string {
        return readDevicePropertyFileTime(pnp, propLastArrival)
}

// readLastRemoval đọc thời điểm rút gần nhất từ Properties\LastRemoval (0066)
func readLastRemoval(pnp string) string {
        return readDevicePropertyFileTime(pnp, propLastRemoval)
}

// friendlyNameForPNP tra tên thân thiện của thiết bị trong
// HKLM\SYSTEM\CurrentControlSet\Enum\<pnp> (pnp đã normalize cũng được vì
// registry path là case-insensitive).
func friendlyNameForPNP(devID string) string {
        if devID == "" {
                return ""
        }
        k, err := registry.OpenKey(registry.LOCAL_MACHINE,
                `SYSTEM\CurrentControlSet\Enum\`+devID,
                registry.QUERY_VALUE|registry.WOW64_64KEY)
        if err != nil {
                return ""
        }
        defer k.Close()
        if v, _, err := k.GetStringValue("FriendlyName"); err == nil && v != "" {
                return v
        }
        if v, _, err := k.GetStringValue("DeviceDesc"); err == nil {
                return stripDevDescPrefix(v)
        }
        return ""
}

// scanPeripheralHistory dựng lịch sử kết nối thiết bị ngoại vi:
//
//	Nguồn 1: Event Log Kernel-PnP / SetupAPI (scanDeviceSessions)
//	         -> lịch sử TỪNG LẦN cắm/rút + PlugCount chính xác
//	Nguồn 2: Ghosted devices + Properties LastArrival/LastRemoval
//	         -> bổ sung cho thiết bị mà Event Log không còn giữ
func scanPeripheralHistory() []PeripheralRec {
        var out []PeripheralRec
        seen := map[string]bool{}

        // Nguồn 1: sessions từ Event Log / SetupAPI
        devSessions := scanDeviceSessions()
        for devID, ds := range devSessions {
                if ds == nil || len(ds.Sessions) == 0 {
                        continue
                }
                rec := sessionsToHistRec(devID, ds)
                out = append(out, rec)
                seen[devID] = true
        }

        // Nguồn 2: ghosted devices + Properties LastArrival/LastRemoval
        for _, dev := range scanGhostedDevices() {
                key := normalizePnpID(dev.PNPDeviceID)
                if key == "" || seen[key] {
                        continue
                }
                first := readFirstArrival(dev.PNPDeviceID)
                last := readLastRemoval(dev.PNPDeviceID)
                if first == "" && last == "" {
                        // Không có dấu vết thời gian nào -> bỏ qua để tránh nhiễu
                        continue
                }
                rec := PeripheralRec{
                        HardwareID:  key,
                        VIDPID:      extractVIDPIDStr(dev.PNPDeviceID),
                        VendorModel: dev.FriendlyName,
                        FirstPlug:   first,
                        LastPlug:    last,
                        PlugCount:   1,
                }
                // Ghosted (đã rút) -> session cuối cùng đã đóng
                if first != "" {
                        rec.Sessions = []ConnectSession{{
                                Arrival: first,
                                Removal: last,
                                Duration: func() string {
                                        fa, la := mustParseLayout(first), mustParseLayout(last)
                                        if fa.IsZero() || la.IsZero() {
                                                return ""
                                        }
                                        return vnDuration(la.Sub(fa))
                                    }(),
                        }}
                }
                out = append(out, rec)
                seen[key] = true
        }
        return out
}
