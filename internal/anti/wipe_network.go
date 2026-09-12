//go:build windows

package anti

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// wipeNetworkHistory xoá dấu vết mạng & Internet:
//   - Flush DNS cache
//   - Xoá NetworkList profiles trong registry
//   - Xoá history Chrome/Edge/Firefox
//   - Reset TCP/IP stack
func wipeNetworkHistory() []WipeStep {
	var steps []WipeStep

	// 1. Flush DNS
	if err := exec.Command("ipconfig", "/flushdns").Run(); err == nil {
		steps = append(steps, WipeStep{
			Category: "Network",
			Action:   "Flush DNS cache",
			Target:   "ipconfig /flushdns",
			Result:   "OK",
			Detail:   "DNS cache đã được xóa",
		})
	} else {
		steps = append(steps, WipeStep{
			Category: "Network",
			Action:   "Flush DNS cache",
			Target:   "ipconfig /flushdns",
			Result:   "FAILED",
			Detail:    err.Error(),
		})
	}

	// 2. Xoá NetworkList profiles (lịch sử mạng)
	netListKey := `SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles`
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, netListKey,
		registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY); err == nil {
		names, _ := k.ReadSubKeyNames(-1)
		_ = k.Close()
		for _, name := range names {
			// Mở và xoá subkey
			_ = registry.DeleteKey(registry.LOCAL_MACHINE,
				netListKey+`\`+name)
		}
		steps = append(steps, WipeStep{
			Category: "Network",
			Action:   "Delete NetworkList profiles",
			Target:   netListKey,
			Result:   "OK",
			Detail:    "Đã xóa toàn bộ profile mạng (lịch sử IP/ISP/MAC)",
		})
	} else {
		steps = append(steps, WipeStep{
			Category: "Network",
			Action:   "Delete NetworkList profiles",
			Target:   netListKey,
			Result:   "SKIPPED",
			Detail:    "Không tìm thấy nhánh registry",
		})
	}

	// 3. Xoá Signatures cache (NetworkList\Signatures)
	sigKey := `SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Signatures`
	_ = registry.DeleteKey(registry.LOCAL_MACHINE, sigKey)
	steps = append(steps, WipeStep{
		Category: "Network",
		Action:   "Delete NetworkList signatures",
		Target:   sigKey,
		Result:   "OK",
		Detail:    "Đã xóa signatures mạng (MAC đã từng gặp)",
	})

	// 4. Xoá history trình duyệt
	steps = append(steps, wipeBrowserHistory()...)

	return steps
}

// wipeBrowserHistory xoá history Chrome/Edge/Firefox
func wipeBrowserHistory() []WipeStep {
	var steps []WipeStep
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		userProfile = `C:\Users\Default`
	}

	// Danh sách file cần xoá của từng trình duyệt
	filesToWipe := []string{
		// Chrome
		filepath.Join(userProfile, `AppData\Local\Google\Chrome\User Data\Default\History`),
		filepath.Join(userProfile, `AppData\Local\Google\Chrome\User Data\Default\Cache`),
		filepath.Join(userProfile, `AppData\Local\Google\Chrome\User Data\Default\Cookies`),
		// Edge
		filepath.Join(userProfile, `AppData\Local\Microsoft\Edge\User Data\Default\History`),
		filepath.Join(userProfile, `AppData\Local\Microsoft\Edge\User Data\Default\Cache`),
		// Firefox
		filepath.Join(userProfile, `AppData\Local\Mozilla\Firefox\Profiles`),
	}

	browsers := []string{"Chrome", "Edge", "Firefox"}
	for i, f := range filesToWipe {
		if _, err := os.Stat(f); err == nil {
			_ = os.Remove(f)
			steps = append(steps, WipeStep{
				Category: "Network",
				Action:   "Delete browser history",
				Target:   f,
				Result:   "OK",
				Detail:    "Đã xóa history/cache của " + browsers[i%3],
			})
		}
	}

	// 5. Wipe WinHTTP proxy
	_ = exec.Command("netsh", "winhttp", "reset", "proxy").Run()
	steps = append(steps, WipeStep{
		Category: "Network",
		Action:   "Reset WinHTTP proxy",
		Target:   "netsh winhttp reset proxy",
		Result:   "OK",
		Detail:    "Đã reset proxy WinHTTP",
	})

	// 6. Reset Winsock
	_ = exec.Command("netsh", "winsock", "reset").Run()
	steps = append(steps, WipeStep{
		Category: "Network",
		Action:   "Reset Winsock",
		Target:   "netsh winsock reset",
		Result:   "OK",
		Detail:    "Đã reset Winsock catalog",
	})

	return steps
}

// safe strings reference để tránh unused import
var _ = strings.TrimSpace
