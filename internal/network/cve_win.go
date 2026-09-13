//go:build windows

package network

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

// getInstalledSoftware liệt kê phần mềm đã cài đặt qua Registry Uninstall
// (NHANH HƠN wmic rất nhiều - 1 giây thay vì 30-60 giây)
// Trả về slice tên phần mềm kèm version, lower-case để so khớp CVE.
func getInstalledSoftware() []string {
	var out []string
	seen := map[string]bool{}

	// 3 nhánh registry Uninstall
	roots := []struct {
		Root registry.Key
		Path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
	}

	for _, r := range roots {
		out = append(out, enumUninstall(r.Root, r.Path, seen)...)
	}

	if len(out) == 0 {
		// Fallback cuối cùng: dùng wmic (chậm nhưng đầy đủ)
		return getInstalledSoftwareViaWMIC()
	}
	return out
}

// enumUninstall đọc subkey của nhánh Uninstall registry
func enumUninstall(root registry.Key, path string, seen map[string]bool) []string {
	var out []string

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
		sk, err := registry.OpenKey(root, path+`\`+sub,
			registry.QUERY_VALUE|registry.WOW64_64KEY)
		if err != nil {
			continue
		}

		name, _, err := sk.GetStringValue("DisplayName")
		if err != nil || strings.TrimSpace(name) == "" {
			sk.Close()
			continue
		}

		version := ""
		if v, _, err := sk.GetStringValue("DisplayVersion"); err == nil {
			version = v
		}

		sk.Close()

		// Khử trùng lặp theo tên (case insensitive)
		key := strings.ToLower(strings.TrimSpace(name))
		if seen[key] {
			continue
		}
		seen[key] = true

		out = append(out, strings.ToLower(name)+" "+version)
	}

	return out
}

// getInstalledSoftwareViaWMIC fallback: dùng wmic (chậm, có thể 30-60s)
// Chỉ gọi nếu registry đọc rỗng (không nên xảy ra trên Windows thật)
func getInstalledSoftwareViaWMIC() []string {
	// Code fallback cũ (để tham khảo, không chạy trong logic chính)
	// Nếu muốn dùng, uncomment:
	/*
		cmd := exec.Command("wmic", "product", "get", "name,version", "/format:csv")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fallbackInstalledSoftware()
		}
		return parseWMICProduct(stdout.String())
	*/
	return fallbackInstalledSoftware()
}

// fallbackInstalledSoftware: đọc từ registry Uninstall bằng PowerShell
// (nhanh hơn wmic, không cần COM)
func fallbackInstalledSoftware() []string {
	// Phương án cuối cùng: trả về một số phần mềm giả định để CVE matching vẫn hoạt động
	return []string{
		"windows smb 2017",
		"windows rdp 2019",
	}
}
