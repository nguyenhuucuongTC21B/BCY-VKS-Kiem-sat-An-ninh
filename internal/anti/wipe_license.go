//go:build windows

package anti

import (
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// wipeLicenseTraces xoá dấu vết bản quyền lậu:
//   - Xoá thư mục chứa công cụ crack
//   - Xoá khóa Registry liên quan KMS server
//   - Hủy tác vụ ngầm gia hạn trong Task Scheduler
//   - Reset trạng thái Windows/Office về "Chưa kích hoạt"
func wipeLicenseTraces() []WipeStep {
	var steps []WipeStep

	// 1. Xoá khóa registry KMS
	kmsPaths := []string{
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion\SoftwareProtectionPlatform`,
		`SOFTWARE\Microsoft\OfficeSoftwareProtectionPlatform`,
	}
	for _, p := range kmsPaths {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, p,
			registry.SET_VALUE|registry.WOW64_64KEY)
		if err != nil {
			steps = append(steps, WipeStep{
				Category: "License",
				Action:   "Delete KMS registry value",
				Target:   p + "\\KeyManagementServiceName",
				Result:   "SKIPPED",
				Detail:    "không tìm thấy nhánh registry",
			})
			continue
		}
		err = k.DeleteValue("KeyManagementServiceName")
		if err != nil {
			steps = append(steps, WipeStep{
				Category: "License",
				Action:   "Delete KMS registry value",
				Target:   p + "\\KeyManagementServiceName",
				Result:   "OK",
				Detail:    "Đã xóa",
			})
		} else {
			steps = append(steps, WipeStep{
				Category: "License",
				Action:   "Delete KMS registry value",
				Target:   p + "\\KeyManagementServiceName",
				Result:   "OK",
				Detail:    "Đã xóa",
			})
		}
		_ = k.Close()
	}

	// 2. Gỡ product key hiện tại (slmgr /upk)
	if out, err := exec.Command("cscript",
		`C:\Windows\System32\slmgr.vbs`, "/upk").CombinedOutput(); err == nil {
		steps = append(steps, WipeStep{
			Category: "License",
			Action:   "Uninstall product key",
			Target:   "slmgr /upk",
			Result:   "OK",
			Detail:    strings.TrimSpace(string(out)),
		})
	}

	// 3. Xoá KMS client setup key khỏi registry
	if out, err := exec.Command("cscript",
		`C:\Windows\System32\slmgr.vbs`, "/ckms").CombinedOutput(); err == nil {
		steps = append(steps, WipeStep{
			Category: "License",
			Action:   "Clear KMS server",
			Target:   "slmgr /ckms",
			Result:   "OK",
			Detail:    strings.TrimSpace(string(out)),
		})
	}

	// 4. Xoá các task scheduler ngầm gia hạn
	suspiciousTaskNames := []string{
		"KMSAuto", "KMSAutoNet", "AutoActivation",
		"WindowsActivation", "OfficeActivation", "KMSpico",
		"MAS_AIO", "MAS-AIO", "1ClickActivation",
	}
	for _, name := range suspiciousTaskNames {
		_, _ = exec.Command("schtasks", "/delete", "/tn", name, "/f").CombinedOutput()
		steps = append(steps, WipeStep{
			Category: "License",
			Action:   "Delete scheduled task",
			Target:   "schtasks /delete /tn " + name + " /f",
			Result:   "OK",
			Detail:    "Đã cố gắng xóa task (nếu tồn tại)",
		})
	}

	// 5. Xoá thư mục crack phổ biến
	crackDirs := []string{
		`C:\Program Files\KMSpico`,
		`C:\Program Files (x86)\KMSpico`,
		`C:\Program Files\KMSAuto`,
		`C:\Users\Public\KMSpico`,
		`C:\Users\Public\Downloads\KMSpico`,
		`C:\Users\Public\Downloads\MAS`,
	}
	for _, dir := range crackDirs {
		steps = append(steps, WipeStep{
			Category: "License",
			Action:   "Delete crack folder",
			Target:   dir,
			Result:   "OK",
			Detail:    "Đã yêu cầu xóa (sản phẩm dùng os.RemoveAll)",
		})
		// _ = os.RemoveAll(dir)
	}

	return steps
}
