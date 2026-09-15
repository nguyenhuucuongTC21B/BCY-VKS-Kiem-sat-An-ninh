//go:build windows

package hardware

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
)

// scanPeripheralSessions truy vấn Event Log Kernel-PnP để lấy lịch sử từng lần cắm/rút thiết bị
// Trả về map[pnpDeviceID][]ConnectSession
//
// Cơ chế: dùng wevtutil để query event ID 20001 (Device Arrived) và 20002 (Device Removed)
// từ Microsoft-Windows-Kernel-PnP/Operational log
func scanPeripheralSessions() map[string][]ConnectSession {
	out := map[string][]ConnectSession{}

	// Phương án 1: Query Event Log Kernel-PnP
	sessions := queryKernelPnPSessions()
	if len(sessions) > 0 {
		for pnp, sess := range sessions {
			out[pnp] = sess
		}
	}

	// Phương án 2: Fallback SetupAPI log (setupapi.dev.log)
	if len(out) == 0 {
		setupSessions := parseSetupAPILog()
		for pnp, sess := range setupSessions {
			out[pnp] = sess
		}
	}

	return out
}

// queryKernelPnPSessions query event log Microsoft-Windows-Kernel-PnP/Operational
// Event ID 20001 = Device Arrived, Event ID 20002 = Device Removed
func queryKernelPnPSessions() map[string][]ConnectSession {
	out := map[string][]ConnectSession{}

	// wevtutil qe Microsoft-Windows-Kernel-PnP/Operational /q:"*[System[(EventID=20001 or EventID=20002)]]" /c:100 /rd:true /f:text
	cmd := exec.Command("wevtutil", "qe", "Microsoft-Windows-Kernel-PnP/Operational",
		"/q:*[System[(EventID=20001 or EventID=20002)]]",
		"/c:200", "/rd:true", "/f:text")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return out
	}

	// Parse output: tìm Date, EventID, và DeviceID
	content := stdout.String()
	return parseKernelPnPEventLog(content)
}

// parseKernelPnPEventLog parse output của wevtutil cho Kernel-PnP events
func parseKernelPnPEventLog(content string) map[string][]ConnectSession {
	out := map[string][]ConnectSession{}

	// Format wevtutil text output:
	//   Event[0]:
	//     Log Name: Microsoft-Windows-Kernel-PnP/Operational
	//     Source: Microsoft-Windows-Kernel-PnP
	//     Date: 2024-08-12T10:15:22.000Z
	//     Event ID: 20001
	//     ...
	//     Device: USB\VID_045E&PID_07C0\6&1a2b3c4d&0&1

	dateRe := regexp.MustCompile(`(?i)Date:\s*(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})`)
	eventIDRe := regexp.MustCompile(`(?i)Event ID:\s*(\d+)`)
	deviceRe := regexp.MustCompile(`(?i)USB\\VID_[A-F0-9&PID_]+\\\S+`)

	// Tách theo "Event[N]:"
	events := strings.Split(content, "Event[")
	var currentTS, currentEventID, currentDevice string
	for i, ev := range events {
		if i == 0 {
			continue // skip preamble
		}
		currentTS = ""
		currentEventID = ""
		currentDevice = ""

		if m := dateRe.FindStringSubmatch(ev); len(m) > 1 {
			currentTS = m[1]
		}
		if m := eventIDRe.FindStringSubmatch(ev); len(m) > 1 {
			currentEventID = m[1]
		}
		if m := deviceRe.FindStringSubmatch(ev); len(m) > 0 {
			currentDevice = m[0]
		}

		if currentDevice == "" || currentTS == "" {
			continue
		}

		// Event ID 20001 = Arrival, 20002 = Removal
		isArrival := currentEventID == "20001"
		sessions := out[currentDevice]
		if isArrival {
			sessions = append(sessions, ConnectSession{
				StartTime: currentTS,
				Source:    "Event Log Kernel-PnP",
			})
		} else {
			// Cập nhật EndTime cho session cuối chưa có EndTime
			if len(sessions) > 0 && sessions[len(sessions)-1].EndTime == "" {
				sessions[len(sessions)-1].EndTime = currentTS
			}
		}
		out[currentDevice] = sessions
	}
	return out
}

// parseSetupAPILog fallback: đọc C:\Windows\inf\setupapi.dev.log
// File log này ghi lại mọi thao tác cài driver thiết bị, có timestamp
func parseSetupAPILog() map[string][]ConnectSession {
	out := map[string][]ConnectSession{}

	// Đọc setupapi.dev.log
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-Content 'C:\Windows\inf\setupapi.dev.log' -Tail 2000 -ErrorAction SilentlyContinue`)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return out
	}

	content := stdout.String()
	// Parse các dòng có chứa "USB\" và timestamp
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		// Tìm dòng có "USB\VID_" + "Device Install"
		if !strings.Contains(lower, "usb\\vid_") {
			continue
		}
		// Trích PNP device ID
		usbIdx := strings.Index(line, "USB\\")
		if usbIdx < 0 {
			continue
		}
		rest := line[usbIdx:]
		parts := strings.Fields(rest)
		if len(parts) == 0 {
			continue
		}
		pnp := parts[0]
		// Tìm timestamp (định dạng yyyy/mm/dd hh:mm:ss)
		tsRe := regexp.MustCompile(`(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})`)
		var ts string
		if m := tsRe.FindStringSubmatch(line); len(m) > 1 {
			ts = strings.ReplaceAll(m[1], "/", "-")
		}
		if ts == "" {
			continue
		}
		// SetupAPI chỉ có "Device Install" (= arrival, không có removal)
		sessions := out[pnp]
		sessions = append(sessions, ConnectSession{
			StartTime: ts,
			Source:    "SetupAPI log",
		})
		out[pnp] = sessions
	}
	return out
}
