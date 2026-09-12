package output

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// writeDOCXReport sinh file .docx đơn giản bằng cách tạo file XML trực tiếp
// (.docx thực chất là ZIP chứa XML). Để tránh thêm dependency thư viện ngoài,
// ta sinh XML WordML 2003 (.xml) hoặc tạo file .docx bằng format HTML
// (Word có thể mở .doc-như-html). 
//
// Production nên thay bằng github.com/nguyenthenguyen/docx hoặc
// github.com/unidoc/unioffice để .docx chuẩn.
//
// Triển khai ở đây: tạo file .docx = HTML được đặt tên, Word 2010+ đọc được.
func writeDOCXReport(ctx *remediationContext) (string, error) {
	path := timestampFileName("BCY-VKS-Remediation", "docx")

	var sb strings.Builder
	// Header Word HTML
	sb.WriteString(`<html xmlns:o="urn:schemas-microsoft-com:office:office" `)
	sb.WriteString(`xmlns:w="urn:schemas-microsoft-com:office:word" `)
	sb.WriteString(`xmlns="http://www.w3.org/TR/REC-html40">`)
	sb.WriteString(`<head><meta charset="utf-8"><title>BCY-VKS Report</title>`)
	sb.WriteString(`<style>`)
	sb.WriteString(`body{font-family:'Times New Roman',serif;font-size:12pt;line-height:1.5;}`)
	sb.WriteString(`h1{font-size:18pt;color:#1a4480;border-bottom:2pt solid #1a4480;}`)
	sb.WriteString(`h2{font-size:14pt;color:#b10000;margin-top:18pt;}`)
	sb.WriteString(`h3{font-size:12pt;color:#1a4480;margin-top:14pt;}`)
	sb.WriteString(`table{border-collapse:collapse;width:100%;font-size:10pt;}`)
	sb.WriteString(`th,td{border:1pt solid #666;padding:4pt;}`)
	sb.WriteString(`th{background:#1a4480;color:#fff;}`)
	sb.WriteString(`code{font-family:'Consolas',monospace;font-size:10pt;background:#f0f0f0;display:block;padding:8pt;white-space:pre-wrap;}`)
	sb.WriteString(`</style></head><body>`)

	fmt.Fprintf(&sb, `<h1>BCY-VKS - Báo cáo Đề xuất Xử lý Khắc phục</h1>`)
	fmt.Fprintf(&sb, `<p>Thời điểm: %s</p>`, time.Now().Format("02/01/2006 15:04:05"))

	// Tóm tắt
	sb.WriteString(`<h2>Tóm tắt phát hiện</h2>`)
	sb.WriteString(`<table><tr><th>Hạng mục</th><th>Số lượng</th></tr>`)
	fmt.Fprintf(&sb, `<tr><td>Bản quyền lậu</td><td>%d</td></tr>`, len(ctx.Result.Bang1))
	fmt.Fprintf(&sb, `<tr><td>Record mạng</td><td>%d</td></tr>`, len(ctx.Result.Bang2))
	fmt.Fprintf(&sb, `<tr><td>Card mạng</td><td>%d</td></tr>`, len(ctx.Result.Bang3))
	fmt.Fprintf(&sb, `<tr><td>Thiết bị ngoại vi</td><td>%d</td></tr>`, len(ctx.Result.Bang4))
	fmt.Fprintf(&sb, `<tr><td>Tiến trình mã độc</td><td>%d</td></tr>`, len(ctx.Result.Bang5))
	sb.WriteString(`</table>`)

	// Phần 1: Bản quyền
	sb.WriteString(`<h2>1. Bản quyền phần mềm</h2>`)
	for _, r := range ctx.Result.Bang1 {
		fmt.Fprintf(&sb, `<h3>%s</h3>`, r.SoftwareName)
		fmt.Fprintf(&sb, `<p>Phiên bản: %s | Product ID: %s | Kênh: %s</p>`,
			r.Version, r.ProductID, r.LicensingChannel)
		if !r.Legal {
			sb.WriteString(`<p><b>Công cụ crack phát hiện:</b> `)
			sb.WriteString(r.CrackTool)
			sb.WriteString(` (`)
			sb.WriteString(r.CrackPath)
			sb.WriteString(`)</p>`)
			sb.WriteString(`<h3>Lệnh xử lý</h3>`)
			sb.WriteString(`<code>slmgr /upk
slmgr /ckms
slmgr /cpky
slmgr /rearm
del /F "` + r.CrackPath + `"
schtasks /delete /tn "KMSAuto" /f</code>`)
		}
	}

	// Phần 2: Mạng
	sb.WriteString(`<h2>2. Mạng & Lỗ hổng</h2>`)
	for _, r := range ctx.Result.Bang2 {
		fmt.Fprintf(&sb, `<p>Trạng thái: %s | IP: %s | ISP: %s</p>`,
			r.InternetStatus, r.CurrentIP, r.ISP)
		fmt.Fprintf(&sb, `<p>Cổng mở: %s</p>`, r.OpenPorts)
		if r.CVEID != "" {
			fmt.Fprintf(&sb, `<p>CVE: %s (CVSS %.1f) - %s</p>`,
				r.CVEID, r.CVSSScore, r.Notes)
		}
		sb.WriteString(`<h3>Lệnh xử lý</h3>`)
		sb.WriteString(`<code>netsh advfirewall firewall add rule name="Block-SMB-445" dir=in action=block protocol=TCP localport=445
netsh advfirewall firewall add rule name="Block-RDP-3389" dir=in action=block protocol=TCP localport=3389
reg add "HKLM\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters" /v SMB1 /t REG_DWORD /d 0 /f
reg add "HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server" /v fDenyTSConnections /t REG_DWORD /d 1 /f
usoclient StartInstall</code>`)
	}

	// Phần 3: Phần cứng
	sb.WriteString(`<h2>3. Card mạng</h2>`)
	sb.WriteString(`<table><tr><th>Loại</th><th>Tên</th><th>MAC</th><th>Vị trí</th></tr>`)
	for _, r := range ctx.Result.Bang3 {
		fmt.Fprintf(&sb, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			r.AdapterType, r.DeviceName, r.MAC, r.ConnectionPos)
	}
	sb.WriteString(`</table>`)
	sb.WriteString(`<h3>Lệnh xử lý</h3>`)
	sb.WriteString(`<code>reg add "HKLM\SYSTEM\CurrentControlSet\Control\Class\{4d36e972-e325-11ce-bfc1-08002be10318}" /v DenyNewUSB /t REG_DWORD /d 1 /f
netsh interface set interface "Wi-Fi" disable</code>`)

	// Phần 4: USB
	sb.WriteString(`<h2>4. Thiết bị ngoại vi</h2>`)
	sb.WriteString(`<table><tr><th>Loại</th><th>Model</th><th>VID/PID</th><th>Ổ</th><th>BadUSB</th></tr>`)
	for _, r := range ctx.Result.Bang4 {
		bad := "Không"
		if r.BadUSBWarning {
			bad = "CÓ!"
		}
		fmt.Fprintf(&sb, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			r.DeviceType, r.VendorModel, r.VIDPID, r.DriveLetter, bad)
	}
	sb.WriteString(`</table>`)
	sb.WriteString(`<h3>Lệnh xử lý</h3>`)
	sb.WriteString(`<code>reg add "HKLM\SYSTEM\CurrentControlSet\Services\USBSTOR" /v Start /t REG_DWORD /d 4 /f
reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows\RemovableStorage" /v Deny_All /t REG_DWORD /d 1 /f</code>`)

	// Phần 5: Mã độc
	sb.WriteString(`<h2>5. Mã độc & Keylogger</h2>`)
	for _, r := range ctx.Result.Bang5 {
		fmt.Fprintf(&sb, `<h3>%s (PID %d)</h3>`, r.ProcessName, r.PID)
		fmt.Fprintf(&sb, `<p>Loại: %s | File: %s | C2: %s</p>`,
			r.Type, r.FilePath, r.C2Server)
		sb.WriteString(`<h3>Lệnh xử lý</h3>`)
		fmt.Fprintf(&sb, `<code>taskkill /F /PID %d
del /F "%s"
netsh advfirewall firewall add rule name="Block C2" dir=out action=block remoteip=%s</code>`,
			r.PID, r.FilePath, extractIP(r.C2Server))
	}

	// Footer
	sb.WriteString(`<p style="font-size:9pt;color:#888;margin-top:18pt;">`)
	fmt.Fprintf(&sb, `BCY-VKS v1.0.0 - Sinh lúc %s<br>`, time.Now().Format("02/01/2006 15:04:05"))
	sb.WriteString(`Lưu ý: chạy lệnh ở quyền Administrator. Chống giám định số phải trong phạm vi pháp luật.</p>`)

	sb.WriteString(`</body></html>`)

	return path, os.WriteFile(path, []byte(sb.String()), 0o644)
}
