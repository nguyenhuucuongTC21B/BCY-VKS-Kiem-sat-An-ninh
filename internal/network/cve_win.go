//go:build windows

package network

import (
	"bytes"
	"os/exec"
	"strings"
)

// getInstalledSoftware liệt kê phần mềm đã cài đặt qua WMI/CIM
// Trả về slice tên phần mềm kèm version, lower-case để so khớp CVE.
func getInstalledSoftware() []string {
	cmd := exec.Command("wmic", "product", "get", "name,version", "/format:csv")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fallbackInstalledSoftware()
	}
	return parseWMICProduct(stdout.String())
}

// parseWMICProduct tách CSV output của wmic product
func parseWMICProduct(s string) []string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Node,") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		name := parts[1]
		version := parts[2]
		out = append(out, strings.ToLower(name)+" "+version)
	}
	return out
}

// fallbackInstalledSoftware đọc từ registry Uninstall
func fallbackInstalledSoftware() []string {
	var out []string
	// Sản phẩm cài đặt thường ở 2 nhánh
	// HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*
	// HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*
	// Đọc bằng Go registry sẽ phức tạp - ở đây dùng powershell làm cầu phương án 2
	cmd := exec.Command("powershell", "-NoProfile",
		"-Command", "Get-ItemProperty 'HKLM:\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\*' | "+
			"Where-Object {$_.DisplayName} | Select-Object -ExpandProperty DisplayName")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil
	}
	for _, line := range strings.Split(stdout.String(), "\n") {
		l := strings.TrimSpace(line)
		if l != "" {
			out = append(out, strings.ToLower(l))
		}
	}
	return out
}
