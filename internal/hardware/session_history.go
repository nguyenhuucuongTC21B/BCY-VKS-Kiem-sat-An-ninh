// Package hardware - session_history.go
// Pure-string parser cho session history - ĐA NỀN TẢNG (build được cả windows và non-windows)
// Cung cấp hàm parseSessionLogLine để parse log line từ Event Log hoặc SetupAPI log
// Hàm này có thể test bằng Go test trên cả Linux/macOS
package hardware

import (
	"strings"
	"time"
)

// ParseSessionLogLine parse 1 dòng log Event Log / SetupAPI
// Trả về (pnpDeviceID, startTime, isArrival) hoặc ("", "", false) nếu không match
//
// Format hỗ trợ:
//   - "2024-08-12 10:15:22 | USB\VID_045E&PID_07C0\6&1a2b3c4d&0&1 | Arrival"
//   - "2024-08-12 11:30:00 | USB\VID_045E&PID_07C0\6&1a2b3c4d&0&1 | Removal"
//   - "[10:15:22] Device USB\VID_045E&PID_07C0\6&1a2b3c4d&0&1 arrived"
//   - "[11:30:00] Device USB\VID_045E&PID_07C0\6&1a2b3c4d&0&1 removed"
func ParseSessionLogLine(line string) (pnpDeviceID string, timestamp string, isArrival bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}

	// Format 1: "timestamp | pnp | Arrival|Removal"
	if strings.Contains(line, " | ") {
		parts := strings.SplitN(line, " | ", 3)
		if len(parts) >= 3 {
			timestamp = strings.TrimSpace(parts[0])
			pnpDeviceID = strings.TrimSpace(parts[1])
			eventType := strings.ToLower(strings.TrimSpace(parts[2]))
			isArrival = strings.Contains(eventType, "arrival") || strings.Contains(eventType, "cắm") || strings.Contains(eventType, "cam")
			// Validate timestamp (định dạng yyyy-mm-dd hh:mm:ss)
			if _, err := time.Parse("2006-01-02 15:04:05", timestamp); err != nil {
				return "", "", false
			}
			return pnpDeviceID, timestamp, isArrival
		}
	}

	// Format 2: "[hh:mm:ss] Device USB\... arrived|removed"
	if strings.HasPrefix(line, "[") {
		endBracket := strings.Index(line, "]")
		if endBracket > 0 {
			timestamp = strings.TrimSpace(line[1:endBracket])
			rest := strings.TrimSpace(line[endBracket+1:])
			if strings.Contains(rest, "USB\\") {
				// Trích PNP device ID
				usbIdx := strings.Index(rest, "USB\\")
				if usbIdx >= 0 {
					// Lấy phần còn lại cho tới space tiếp theo
					rest = rest[usbIdx:]
					parts := strings.Fields(rest)
					if len(parts) > 0 {
						pnpDeviceID = parts[0]
					}
				}
				isArrival = strings.Contains(strings.ToLower(line), "arrived")
				// Validate timestamp (hh:mm:ss)
				if _, err := time.Parse("15:04:05", timestamp); err != nil {
					return "", "", false
				}
				return pnpDeviceID, timestamp, isArrival
			}
		}
	}

	return "", "", false
}

// ParseSessionLogEntries parse toàn bộ log (nhiều dòng) → map[pnpDeviceID][]ConnectSession
// Dùng cho cả Event Log và SetupAPI log
func ParseSessionLogEntries(logContent string) map[string][]ConnectSession {
	out := map[string][]ConnectSession{}
	lines := strings.Split(logContent, "\n")

	for _, line := range lines {
		pnp, ts, isArrival := ParseSessionLogLine(line)
		if pnp == "" {
			continue
		}

		sessions := out[pnp]
		if isArrival {
			// Tạo session mới
			sessions = append(sessions, ConnectSession{
				StartTime: ts,
				Source:    "Event Log",
			})
		} else {
			// Cập nhật EndTime cho session cuối cùng chưa có EndTime
			if len(sessions) > 0 && sessions[len(sessions)-1].EndTime == "" {
				sessions[len(sessions)-1].EndTime = ts
			}
		}
		out[pnp] = sessions
	}
	return out
}

// SortSessionsByStartTime sắp xếp sessions theo StartTime tăng dần
// (pure-string, có thể test trên cả Linux/macOS)
func SortSessionsByStartTime(sessions []ConnectSession) []ConnectSession {
	if len(sessions) <= 1 {
		return sessions
	}
	// Bubble sort đơn giản (danh sách session thường <100 nên OK)
	out := make([]ConnectSession, len(sessions))
	copy(out, sessions)
	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[i].StartTime > out[j].StartTime {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// FormatSessionsSummary trả về chuỗi tóm tắt cho hiển thị
// Ví dụ: "3 sessions: 2024-08-12 10:15 → 11:30 | 2024-08-15 09:00 → 09:45 | 2024-09-01 14:00 → (đang cắm)"
func FormatSessionsSummary(sessions []ConnectSession) string {
	if len(sessions) == 0 {
		return ""
	}
	out := []string{}
	for _, s := range sessions {
		if s.EndTime != "" {
			out = append(out, s.StartTime+" → "+s.EndTime)
		} else {
			out = append(out, s.StartTime+" → (đang cắm)")
		}
	}
	return "Số lần: " + itoaHelper(len(sessions)) + " | " + strings.Join(out, " | ")
}
