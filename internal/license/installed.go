//go:build windows

package license

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

// InstalledSoftware chứa thông tin một phần mềm đã cài qua Uninstall registry
type InstalledSoftware struct {
	DisplayName    string
	DisplayVersion string
	Publisher      string
	ProductID      string
	InstallDate    string
	InstallLocation string
	UninstallString string
}

// scanInstalledSoftware liệt kê toàn bộ phần mềm đã cài từ 3 nhánh registry:
//   - HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall
//   - HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall
//   - HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall
func scanInstalledSoftware() []InstalledSoftware {
	var out []InstalledSoftware
	roots := []struct {
		Root registry.Key
		Path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
	}
	seen := map[string]bool{} // để khử trùng lặp

	for _, r := range roots {
		out = append(out, enumerateUninstallKeys(r.Root, r.Path, seen)...)
	}
	return out
}

// enumerateUninstallKeys mở nhánh Uninstall và duyệt qua các subkey
func enumerateUninstallKeys(root registry.Key, path string, seen map[string]bool) []InstalledSoftware {
	var out []InstalledSoftware

	k, err := registry.OpenKey(root, path,
		registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return out
	}
	defer k.Close()

	subKeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return out
	}

	for _, sub := range subKeys {
		// Bỏ qua các entry rỗng hoặc Microsoft system hidden
		if strings.HasPrefix(sub, "Microsoft .NET") {
			continue
		}

		sk, err := registry.OpenKey(root, path+`\`+sub,
			registry.QUERY_VALUE|registry.WOW64_64KEY)
		if err != nil {
			continue
		}

		// Đọc DisplayName - nếu không có thì bỏ qua (không phải app entry thật)
		displayName, _, err := sk.GetStringValue("DisplayName")
		if err != nil || strings.TrimSpace(displayName) == "" {
			sk.Close()
			continue
		}

		// Khử trùng lặp theo DisplayName (case insensitive)
		key := strings.ToLower(strings.TrimSpace(displayName))
		if seen[key] {
			sk.Close()
			continue
		}
		seen[key] = true

		sw := InstalledSoftware{
			DisplayName: strings.TrimSpace(displayName),
		}
		if v, _, err := sk.GetStringValue("DisplayVersion"); err == nil {
			sw.DisplayVersion = strings.TrimSpace(v)
		}
		if v, _, err := sk.GetStringValue("Publisher"); err == nil {
			sw.Publisher = strings.TrimSpace(v)
		}
		if v, _, err := sk.GetStringValue("ProductID"); err == nil {
			sw.ProductID = strings.TrimSpace(v)
		}
		if v, _, err := sk.GetStringValue("InstallDate"); err == nil {
			sw.InstallDate = strings.TrimSpace(v)
		}
		if v, _, err := sk.GetStringValue("InstallLocation"); err == nil {
			sw.InstallLocation = strings.TrimSpace(v)
		}
		if v, _, err := sk.GetStringValue("UninstallString"); err == nil {
			sw.UninstallString = strings.TrimSpace(v)
		}

		sk.Close()
		out = append(out, sw)
	}
	return out
}

// isAuditWorthy kiểm tra phần mềm có đáng để đưa vào bảng audit không
// (để lọc ra các phần mềm thương mại hay bị crack, không show toàn bộ phần mềm thừa)
func isAuditWorthy(name, publisher string) bool {
	n := strings.ToLower(name)
	p := strings.ToLower(publisher)

	// Microsoft products
	if strings.Contains(n, "microsoft office") ||
		strings.Contains(n, "microsoft windows") ||
		strings.Contains(n, "microsoft visio") ||
		strings.Contains(n, "microsoft project") ||
		strings.Contains(n, "microsoft 365") ||
		strings.Contains(n, "microsoft sql server") ||
		strings.Contains(n, "microsoft visual studio") {
		return true
	}

	// Adobe products
	if strings.Contains(p, "adobe") ||
		strings.Contains(n, "adobe acrobat") ||
		strings.Contains(n, "adobe reader") ||
		strings.Contains(n, "adobe photoshop") ||
		strings.Contains(n, "adobe illustrator") ||
		strings.Contains(n, "adobe lightroom") ||
		strings.Contains(n, "adobe premiere") ||
		strings.Contains(n, "adobe creative cloud") {
		return true
	}

	// ABBYY
	if strings.Contains(n, "abbyy") || strings.Contains(p, "abbyy") {
		return true
	}

	// XMind
	if strings.Contains(n, "xmind") || strings.Contains(p, "xmind") {
		return true
	}

	// Foxit
	if strings.Contains(n, "foxit") || strings.Contains(p, "foxit") {
		return true
	}

	// Other commonly pirated software
	keywords := []string{
		"winrar", "winzip", "7-zip",
		"idm", "internet download manager",
		"teamviewer", "anydesk",
		"sublime text", "notepad++",
		"jetbrains", "intellij", "pycharm", "webstorm", "clion",
		"autocad", "autodesk",
		"coreldraw", "corel",
		"vmware", "virtualbox",
		"camtasia", "snagit",
		"faststone", "irfanview",
		"office tab", "kutools",
		"filezilla", "winscp",
		"netflix", "spotify",
		"skype", "zoom",
		"nero", "poweriso", "ultraiso",
		"ccleaner", "glary",
		"davinci", "wondershare",
		" EaseUS ",
		"acronis",
		"paragon",
		"matlab", "mathematica",
		"stata",
		"origin",
		"tableau", "power bi",
		"sql server management studio",
		"navicat",
		"powerdirector",
		"filmora",
		"4k video",
		"screencast",
		"bandicam",
		"fraps",
		"obs",
		"reaper",
		"fl studio",
		"ableton",
		"sony vegas",
		"premiere",
	}
	for _, kw := range keywords {
		if strings.Contains(n, kw) {
			return true
		}
	}

	return false
}

// isLikelyLegalByDefault kiểm tra phần mềm có vẻ hợp pháp mặc định
// (giả định hợp pháp đến khi phát hiện bằng chứng ngược)
func isLikelyLegalByDefault(name, publisher string) bool {
	// Phần mềm miễn phí / open source thường hợp pháp
	freeSoftware := []string{
		"7-zip", "notepad++", "filezilla", "winscp",
		"irfanview", "faststone image viewer",
		"obs studio", "obs",
		"firefox", "chrome", "edge",
		"thunderbird",
		"libreoffice", "openoffice",
		"virtualbox",
		"vlc", "audacity", "handbrake",
		"python", "node.js", "java",
		"git", "docker",
		"vscode", "visual studio code",
	}
	n := strings.ToLower(name)
	for _, sw := range freeSoftware {
		if strings.Contains(n, sw) {
			return true
		}
	}
	// Nếu không khớp free list, giả định hợp pháp (sẽ bị false nếu phát hiện crack)
	return true
}
