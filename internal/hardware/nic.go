// Package hardware chứa logic kiểm kê phần cứng mạng, Wifi, USB ngoại vi,
// và dựng lại lịch sử kết nối.
//
// File chính:
//   - nic.go                : card mạng gắn trong + gắn ngoài (NIC)
//   - wifi.go               : wireless profiles (SSID đã từng kết nối)
//   - driverstore.go        : ghosted devices (driver đã từng cắm)
//   - usb.go                : thiết bị ngoại vi + BadUSB detection
//   - usbsessions*.go       : lịch sử TỪNG LẦN kết nối (Event Log + SetupAPI)
//   - recentfiles_*.go      : dấu vết tệp liên quan tới thiết bị USB
//
// API chính:
//   ScanNIC()              -> []HardwareRecord (Bảng 3)
//   ScanPeripherals(fs)    -> []PeripheralRec  (Bảng 4)
package hardware

import (
        "embed"
        "strings"
)

// HardwareRecord tương ứng BẢNG 3 trong đặc tả đầu ra.
type HardwareRecord struct {
        AdapterType   string `json:"adapter_type"`    // LAN | Wi-Fi | Bluetooth
        ConnectionPos string `json:"connection_pos"`  // Internal | External USB
        DeviceName    string `json:"device_name"`
        Serial        string `json:"serial"`
        MAC           string `json:"mac"`
        DriverStatus  string `json:"driver_status"`
        SSIDList      string `json:"ssid_list"`
        LastConnect   string `json:"last_connect"`
}

// ConnectSession mô tả MỘT LẦN kết nối của thiết bị ngoại vi:
// thời điểm cắm (Arrival), thời điểm rút (Removal) và thời lượng cắm.
// Nếu Removal rỗng -> thiết bị đang cắm trên máy (on-going session).
type ConnectSession struct {
        Arrival  string `json:"arrival"`            // "2006-01-02 15:04:05"
        Removal  string `json:"removal,omitempty"`  // rỗng = đang cắm
        Duration string `json:"duration,omitempty"` // vd "41 phút", "2 giờ 13 phút"
}

// PeripheralRec tương ứng BẢNG 4 trong đặc tả đầu ra.
type PeripheralRec struct {
        DeviceType          string           `json:"device_type"`      // USB Flash / External HDD / Printer / Phone
        VendorModel         string           `json:"vendor_model"`
        HardwareID          string           `json:"hardware_id"`
        VIDPID              string           `json:"vid_pid"`
        DriveLetter         string           `json:"drive_letter"`
        FirstPlug           string           `json:"first_plug"`
        LastPlug            string           `json:"last_plug"`
        PlugCount           int              `json:"plug_count"`
        BadUSBWarning       bool             `json:"badusb_warning"`
        Sessions            []ConnectSession `json:"sessions,omitempty"`           // lịch sử TỪNG lần cắm/rút
        RecentFilesSummary  string           `json:"recent_files_summary"`         // dấu vết tệp (Recent/Jump Lists)
}

// ScanNIC là entry point để lấy Bảng 3.
func ScanNIC() ([]HardwareRecord, error) {
        internalNICs := scanInternalNICs()
        externalNICs := scanExternalNICs()
        out := append(internalNICs, externalNICs...)

        // Bổ sung SSID list + last connect cho mỗi adapter Wi-Fi
        for i := range out {
                if out[i].AdapterType == "Wi-Fi" {
                        out[i].SSIDList = getWirelessProfiles()
                        out[i].LastConnect = getWiFiLastConnect()
                }
        }
        return out, nil
}

// ScanPeripherals là entry point để lấy Bảng 4.
// Pipeline:
//  1. scanCurrentPeripherals   - thiết bị đang cắm (registry USB/USBSTOR/MountedDevices)
//  2. scanPeripheralHistory    - lịch sử từng lần cắm/rút (Event Log Kernel-PnP,
//                                SetupAPI log, Properties LastArrival/LastRemoval)
//  3. mergePeripheralRecords   - gộp theo HardwareID / serial / VID-PID
//  4. Đối chiếu BadUSB VID/PID
//  5. Bổ sung RecentFilesSummary + đồng bộ đầy đủ các field thời gian
func ScanPeripherals(fs embed.FS) ([]PeripheralRec, error) {
        current := scanCurrentPeripherals()
        history := scanPeripheralHistory()

        // Gộp history vào current (nếu đã có thì cập nhật FirstPlug, PlugCount,
        // nếu không có thì thêm vào làm record lịch sử).
        merged := mergePeripheralRecords(current, history)

        // Đánh dấu BadUSB
        badUSBList := loadBadUSBList(fs)
        for i := range merged {
                if isBadUSB(merged[i].VIDPID, badUSBList) {
                        merged[i].BadUSBWarning = true
                }
        }

        // Hoàn thiện số liệu cuối: PlugCount/FirstPlug/LastPlug đồng bộ từ Sessions,
        // dựng dấu vết tệp cho thiết bị có ký tự ổ đĩa.
        for i := range merged {
                m := &merged[i]
                if n := len(m.Sessions); n > 0 {
                        if m.PlugCount < n {
                                m.PlugCount = n
                        }
                        if m.FirstPlug == "" {
                                m.FirstPlug = m.Sessions[0].Arrival
                        }
                        if m.LastPlug == "" {
                                m.LastPlug = m.Sessions[n-1].Arrival
                        }
                }
                if m.PlugCount == 0 && m.FirstPlug != "" && m.LastPlug != "" {
                        m.PlugCount = 1
                }
                if m.DriveLetter != "" {
                        m.RecentFilesSummary = joinNotes(m.RecentFilesSummary, buildRecentFilesSummary(m.DriveLetter))
                }
                if m.PlugCount == 0 && m.FirstPlug == "" {
                        m.RecentFilesSummary = joinNotes(m.RecentFilesSummary,
                                "Không truy vết được lịch sử cắm (Event Log Kernel-PnP/SetupAPI có thể đã bị tắt hoặc xoá)")
                }
        }
        return merged, nil
}

// joinNotes nối 2 ghi chú bằng dấu " | " (bỏ phần rỗng)
func joinNotes(a, b string) string {
        switch {
        case a == "":
                return b
        case b == "":
                return a
        default:
                return a + " | " + b
        }
}

// mergePeripheralRecords gộp record hiện tại với lịch sử.
// Matching theo thứ tự ưu tiên: HardwareID chính xác -> serial tail -> VID/PID.
func mergePeripheralRecords(current, history []PeripheralRec) []PeripheralRec {
        out := []PeripheralRec{}
        for i := range current {
                out = append(out, current[i])
        }
        for i := range history {
                h := history[i]
                matched := false
                for j := range out {
                        if !sameDevice(out[j], h) {
                                continue
                        }
                        if h.FirstPlug != "" && (out[j].FirstPlug == "" || h.FirstPlug < out[j].FirstPlug) {
                                out[j].FirstPlug = h.FirstPlug
                        }
                        if h.LastPlug != "" && (out[j].LastPlug == "" || h.LastPlug > out[j].LastPlug) {
                                out[j].LastPlug = h.LastPlug
                        }
                        if h.PlugCount > out[j].PlugCount {
                                out[j].PlugCount = h.PlugCount
                        }
                        if len(h.Sessions) > 0 {
                                out[j].Sessions = h.Sessions
                        }
                        if out[j].VIDPID == "" {
                                out[j].VIDPID = h.VIDPID
                        }
                        if out[j].VendorModel == "" {
                                out[j].VendorModel = h.VendorModel
                        }
                        matched = true
                        break
                }
                if !matched {
                        out = append(out, h)
                }
        }
        return out
}

// sameDevice kiểm tra 2 record có cùng một thiết bị vật lý không.
// Windows lưu Device Instance ID ở nhiều dạng (USBSTOR\...\SERIAL&0,
// USB\VID_xxxx&PID_xxxx\<serial>), nên so sánh theo serial tail + VID/PID.
func sameDevice(a, b PeripheralRec) bool {
        if a.HardwareID != "" && strings.EqualFold(a.HardwareID, b.HardwareID) {
                return true
        }
        ta, tb := pnpSerialTail(a.HardwareID), pnpSerialTail(b.HardwareID)
        if len(ta) >= 6 && strings.EqualFold(ta, tb) {
                return true
        }
        // Cùng VID/PID chỉ được coi là cùng thiết bị khi một bên không có serial riêng
        if a.VIDPID != "" && b.VIDPID != "" && strings.EqualFold(a.VIDPID, b.VIDPID) {
                if ta == "" || tb == "" {
                        return true
                }
        }
        return false
}

// pnpSerialTail lấy phần serial cuối của Device Instance ID
// "USBSTOR\Disk&Ven_X\07815FB8&0" -> "07815FB8&0"
func pnpSerialTail(pnp string) string {
        if pnp == "" {
                return ""
        }
        idx := strings.LastIndexAny(pnp, "\\/")
        if idx < 0 {
                return pnp
        }
        return pnp[idx+1:]
}

// isBadUSB kiểm tra VID/PID có trong danh sách thiết bị tấn công
func isBadUSB(vidPID string, list []BadUSBVIDPID) bool {
        for _, b := range list {
                if matchVIDPID(vidPID, b.VID, b.PID) {
                        return true
                }
        }
        // Ngoài ra, các thiết bị USB Generic Keyboard với VID đặc thù (Atmel/Teensy/Arduino)
        // cũng có thể là BadUSB - kiểm tra VID riêng
        badVIDs := []string{"03EB", "16C0", "2341", "F00D"}
        for _, vid := range badVIDs {
                if matchVIDOnly(vidPID, vid) {
                        return true
                }
        }
        return false
}

func matchVIDPID(haystack, vid, pid string) bool {
        return contains(haystack, "VID_"+vid) && (pid == "" || contains(haystack, "PID_"+pid))
}

func matchVIDOnly(haystack, vid string) bool {
        return contains(haystack, "VID_"+vid)
}

func contains(s, sub string) bool {
        if len(sub) == 0 {
                return true
        }
        for i := 0; i <= len(s)-len(sub); i++ {
                if s[i:i+len(sub)] == sub {
                        return true
                }
        }
        return false
}
