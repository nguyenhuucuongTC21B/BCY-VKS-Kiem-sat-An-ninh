package network

import (
	"strings"
)

// simulateExploit đánh giá rủi ro dựa trên cổng mở + CVE
// KHÔNG thực hiện khai thác thực tế - chỉ kiểm tra an toàn
// để trả về "SUCCESS" nếu hệ thống có thể bị khai thác
// hoặc "BLOCKED" nếu đã vá/cổng đóng.
func simulateExploit(openPorts []int, cve CVEMatch) string {
	if cve.CVEID == "" {
		return "BLOCKED - không có CVE phát hiện"
	}
	if cve.CVSS < 7.0 {
		return "PARTIAL - rủi ro thấp"
	}
	// Kiểm tra xem cổng liên quan có mở không
	cveTag := strings.ToLower(cve.CVEID)
	for _, port := range openPorts {
		switch {
		case isSMBExploit(cveTag) && (port == 445 || port == 139):
			return "SUCCESS - SMBv1 mở + " + cve.CVEID + " (RCE)"
		case isRDPExploit(cveTag) && port == 3389:
			return "SUCCESS - RDP mở + " + cve.CVEID + " (RCE)"
		case isSSHExploit(cveTag) && port == 22:
			return "SUCCESS - SSH mở + " + cve.CVEID
		case isFTPExploit(cveTag) && port == 21:
			return "SUCCESS - FTP mở + " + cve.CVEID
		}
	}
	return "BLOCKED - cổng liên quan đã đóng"
}

func isSMBExploit(s string) bool {
	return contains(s, "smb") || contains(s, "eternalblue") || contains(s, "smbghost")
}
func isRDPExploit(s string) bool {
	return contains(s, "rdp") || contains(s, "bluekeep")
}
func isSSHExploit(s string) bool {
	return contains(s, "ssh") || contains(s, "openssh")
}
func isFTPExploit(s string) bool {
	return contains(s, "ftp") || contains(s, "proftpd") || contains(s, "vsftpd")
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
