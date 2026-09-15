// Package hardware chứa logic kiểm kê phần cứng mạng, Wifi, USB ngoại vi,
// và dựng lại lịch sử kết nối.
//
// File chính:
//   - nic.go            : entry point + struct ConnectSession, merge logic
//   - wifi_*.go         : wireless profiles (SSID đã từng kết nối)
//   - driverstore_*.go  : ghosted devices (driver đã từng cắm)
//   - usb_*.go          : thiết bị ngoại vi + BadUSB detection
//   - vidpid.go         : hàm pure-string đa nền tảng (extractVIDPID, ...)
//   - recentfiles_*.go  : Recent Files & Jump Lists audit
//   - session_history_*.go : truy vết lịch sử từng lần kết nối (pure-string)
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

// ConnectSession mô tả 1 lần kết nối của thiết bị ngoại vi.
// Mỗi thiết bị có thể có nhiều session (cắm/rút nhiều lần).
type ConnectSession struct {
	StartTime  string `json:"start_time"`  // Lần cắm (ISO 8601 hoặc yyyy-mm-dd hh:mm:ss)
	EndTime    string `json:"end_time"`    // Lần rút (rỗng nếu vẫn đang cắm)
	DriveLetter string `json:"drive_letter"` // Ký tự ổ đĩa gán trong session đó (E:, F:...)
	Source     string `json:"source"`      // "Registry Properties" | "Event Log" | "SetupAPI"
}

// PeripheralRec tương ứng BẢNG 4 trong đặc tả đầu ra.
type PeripheralRec struct {
	DeviceType         string           `json:"device_type"`           // USB Flash / External HDD / Printer / Phone
	VendorModel        string           `json:"vendor_model"`
	HardwareID         string           `json:"hardware_id"`
	VIDPID             string           `json:"vid_pid"`
	DriveLetter        string           `json:"drive_letter"`
	FirstPlug          string           `json:"first_plug"`
	LastPlug           string           `json:"last_plug"`
	PlugCount          int              `json:"plug_count"`
	BadUSBWarning      bool             `json:"badusb_warning"`
	RecentFilesSummary string           `json:"recent_files_summary"` // Recent Files + Jump Lists audit
	Sessions           []ConnectSession `json:"sessions"`             // Lịch sử từng lần kết nối
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
	sessions := scanPeripheralSessions()

	// Gộp history vào current (theo serial/VID-PID thông minh hơn)
	merged := mergePeripheralRecords(current, history)

	// Gộp sessions vào merged (theo serial/VID-PID)
	merged = mergeSessionsIntoRecords(merged, sessions)

	// Đánh dấu BadUSB
	badUSBList := loadBadUSBList(fs)
	for i := range merged {
		if isBadUSB(merged[i].VIDPID, badUSBList) {
			merged[i].BadUSBWarning = true
		}
	}

	// Bổ sung Recent Files & Jump Lists audit vào record đầu tiên
	if len(merged) > 0 {
		recentResult := scanRecentFilesAndJumpLists()
		// Ưu tiên gán cho record USB Storage nếu có
		assigned := false
		for i := range merged {
			if merged[i].DeviceType == "USB Storage" {
				merged[i].RecentFilesSummary = recentResult.Summary
				assigned = true
				break
			}
		}
		if !assigned {
			merged[0].RecentFilesSummary = recentResult.Summary
		}
	}
	return merged, nil
}

// mergePeripheralRecords gộp record hiện tại với lịch sử
// theo HardwareID và fallback theo VID/PID nếu HardwareID không khớp
func mergePeripheralRecords(current, history []PeripheralRec) []PeripheralRec {
	byID := map[string]int{}      // index trong out
	byVIDPID := map[string]int{}  // index theo VID/PID
	out := []PeripheralRec{}

	for i := range current {
		rec := current[i]
		out = append(out, rec)
		idx := len(out) - 1
		if rec.HardwareID != "" {
			byID[rec.HardwareID] = idx
		}
		if rec.VIDPID != "" {
			byVIDPID[normalizeVIDPID(rec.VIDPID)] = idx
		}
	}

	for i := range history {
		rec := history[i]
		// Tìm theo HardwareID trước
		if idx, ok := byID[rec.HardwareID]; ok && rec.HardwareID != "" {
			mergeIntoExisting(&out[idx], rec)
			continue
		}
		// Fallback: tìm theo VID/PID
		if rec.VIDPID != "" {
			key := normalizeVIDPID(rec.VIDPID)
			if idx, ok := byVIDPID[key]; ok {
				mergeIntoExisting(&out[idx], rec)
				continue
			}
		}
		// Không khớp → thêm mới
		out = append(out, rec)
		if rec.HardwareID != "" {
			byID[rec.HardwareID] = len(out) - 1
		}
		if rec.VIDPID != "" {
			byVIDPID[normalizeVIDPID(rec.VIDPID)] = len(out) - 1
		}
	}
	return out
}

// mergeIntoExisting cập nhật thông tin từ src vào dst (không ghi đè dữ liệu có sẵn)
func mergeIntoExisting(dst *PeripheralRec, src PeripheralRec) {
	if src.FirstPlug != "" && (dst.FirstPlug == "" || src.FirstPlug < dst.FirstPlug) {
		dst.FirstPlug = src.FirstPlug
	}
	if src.LastPlug != "" && (dst.LastPlug == "" || src.LastPlug < src.LastPlug) {
		// LastPlug nên là giá trị mới nhất
		if dst.LastPlug == "" || src.LastPlug > dst.LastPlug {
			dst.LastPlug = src.LastPlug
		}
	}
	if src.PlugCount > 0 {
		dst.PlugCount += src.PlugCount
	}
	if src.VendorModel != "" && dst.VendorModel == "" {
		dst.VendorModel = src.VendorModel
	}
	if src.DeviceType != "" && dst.DeviceType == "" {
		dst.DeviceType = src.DeviceType
	}
	if src.DriveLetter != "" && dst.DriveLetter == "" {
		dst.DriveLetter = src.DriveLetter
	}
}

// mergeSessionsIntoRecords gộp lịch sử session vào các record
// (theo HardwareID hoặc VID/PID)
func mergeSessionsIntoRecords(records []PeripheralRec, sessions map[string][]ConnectSession) []PeripheralRec {
	if len(sessions) == 0 {
		return records
	}
	for i := range records {
		// Tìm session theo HardwareID
		if records[i].HardwareID != "" {
			if sess, ok := sessions[records[i].HardwareID]; ok {
				records[i].Sessions = append(records[i].Sessions, sess...)
				continue
			}
		}
		// Fallback: tìm theo VID/PID
		if records[i].VIDPID != "" {
			key := normalizeVIDPID(records[i].VIDPID)
			if sess, ok := sessions[key]; ok {
				records[i].Sessions = append(records[i].Sessions, sess...)
			}
		}
	}
	// Nếu có session không khớp với record nào, thêm mới
	for id, sess := range sessions {
		matched := false
		for i := range records {
			if records[i].HardwareID == id || normalizeVIDPID(records[i].VIDPID) == id {
				matched = true
				break
			}
		}
		if !matched && len(sess) > 0 {
			newRec := PeripheralRec{
				DeviceType: "USB Storage (history)",
				Sessions:   sess,
				PlugCount:  len(sess),
			}
			if sess[0].StartTime != "" {
				newRec.FirstPlug = sess[0].StartTime
			}
			if sess[len(sess)-1].EndTime != "" {
				newRec.LastPlug = sess[len(sess)-1].EndTime
			} else if sess[len(sess)-1].StartTime != "" {
				newRec.LastPlug = sess[len(sess)-1].StartTime + " (đang cắm)"
			}
			records = append(records, newRec)
		}
	}
	return records
}

// normalizeVIDPID chuẩn hoá VID/PID để match (uppercase, no spaces)
func normalizeVIDPID(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Bỏ spaces
		if c == ' ' || c == '\t' {
			continue
		}
		// Uppercase
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		out = append(out, c)
	}
	return string(out)
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
