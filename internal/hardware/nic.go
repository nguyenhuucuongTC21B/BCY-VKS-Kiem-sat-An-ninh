// Package hardware chứa logic kiểm kê phần cứng mạng, Wifi, USB ngoại vi,
// và dựng lại lịch sử kết nối.
//
// File chính:
//   - nic.go         : card mạng gắn trong + gắn ngoài (NIC)
//   - wifi.go        : wireless profiles (SSID đã từng kết nối)
//   - driverstore.go : ghosted devices (driver đã từng cắm)
//   - usb.go         : thiết bị ngoại vi + BadUSB detection
//
// API chính:
//   ScanNIC()              -> []HardwareRecord (Bảng 3)
//   ScanPeripherals(fs)    -> []PeripheralRec  (Bảng 4)
package hardware

import "embed"

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

// PeripheralRec tương ứng BẢNG 4 trong đặc tả đầu ra.
type PeripheralRec struct {
	DeviceType     string `json:"device_type"`     // USB Flash / External HDD / Printer / Phone
	VendorModel    string `json:"vendor_model"`
	HardwareID     string `json:"hardware_id"`
	VIDPID         string `json:"vid_pid"`
	DriveLetter    string `json:"drive_letter"`
	FirstPlug      string `json:"first_plug"`
	LastPlug       string `json:"last_plug"`
	PlugCount      int    `json:"plug_count"`
	BadUSBWarning  bool   `json:"badusb_warning"`
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
	return merged, nil
}

// mergePeripheralRecords gộp record hiện tại với lịch sử
func mergePeripheralRecords(current, history []PeripheralRec) []PeripheralRec {
	// Map theo HardwareID
	byID := map[string]*PeripheralRec{}
	out := []PeripheralRec{}
	for i := range current {
		rec := current[i]
		byID[rec.HardwareID] = &rec
		out = append(out, rec)
	}
	for i := range history {
		rec := history[i]
		if existing, ok := byID[rec.HardwareID]; ok {
			// Cập nhật FirstPlug nếu history cũ hơn
			if rec.FirstPlug != "" && (existing.FirstPlug == "" || rec.FirstPlug < existing.FirstPlug) {
				existing.FirstPlug = rec.FirstPlug
			}
			existing.PlugCount += rec.PlugCount
		} else {
			out = append(out, rec)
		}
	}
	return out
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
