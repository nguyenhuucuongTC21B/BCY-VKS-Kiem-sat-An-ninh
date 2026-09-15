package output

import (
        "fmt"
        "os"
        "strings"
        "time"
)

// writeHTMLReport sinh file .html hoàn chỉnh
// Tương tự popup nhưng đầy đủ hơn, có in được
func writeHTMLReport(ctx *remediationContext) (string, error) {
        path := timestampFileName("BCY-VKS-Remediation", "html")

        var sb strings.Builder
        sb.WriteString(`<!DOCTYPE html><html lang="vi"><head><meta charset="utf-8">`)
        sb.WriteString(`<title>BCY-VKS - Báo cáo Đề xuất Khắc phục</title>`)
        sb.WriteString(`<style>`)
        sb.WriteString(`@page{size:A4;margin:1.5cm;}`)
        sb.WriteString(`body{font-family:'Segoe UI',Arial,sans-serif;color:#222;line-height:1.5;font-size:13px;}`)
        sb.WriteString(`h1{color:#1a4480;font-size:22px;border-bottom:3px solid #1a4480;padding-bottom:6px;}`)
        sb.WriteString(`h2{color:#b10000;font-size:16px;margin-top:24px;border-left:4px solid #b10000;padding-left:8px;}`)
        sb.WriteString(`h3{color:#1a4480;font-size:14px;margin-top:18px;}`)
        sb.WriteString(`table{border-collapse:collapse;width:100%;margin:12px 0;font-size:12px;}`)
        sb.WriteString(`th{background:#1a4480;color:#fff;padding:6px;text-align:left;}`)
        sb.WriteString(`td{border:1px solid #ccc;padding:6px;vertical-align:top;}`)
        sb.WriteString(`tr:nth-child(even){background:#f6f6f6;}`)
        sb.WriteString(`code{display:block;background:#f0f0f0;padding:10px;border:1px solid #ddd;border-radius:3px;font-family:'Consolas',monospace;font-size:11px;white-space:pre-wrap;word-wrap:break-word;}`)
        sb.WriteString(`.footer{margin-top:30px;padding-top:12px;border-top:1px solid #ccc;font-size:10px;color:#888;}`)
        sb.WriteString(`</style></head><body>`)

        // Tiêu đề
        fmt.Fprintf(&sb, `<h1>BCY-VKS KIỂM SÁT AN NINH<br>Báo cáo Đề xuất Xử lý Khắc phục</h1>`)
        fmt.Fprintf(&sb, `<p>Thời điểm sinh báo cáo: %s</p>`, time.Now().Format("02/01/2006 15:04:05"))
        fmt.Fprintf(&sb, `<p>Phiên bản phần mềm: BCY-VKS v1.0.0</p>`)

        // Tóm tắt số liệu
        sb.WriteString(`<h2>Tóm tắt phát hiện</h2>`)
        sb.WriteString(`<table><tr><th>Hạng mục</th><th>Số lượng</th></tr>`)
        fmt.Fprintf(&sb, `<tr><td>Bản quyền lậu phát hiện</td><td>%d</td></tr>`, len(ctx.Result.Bang1))
        fmt.Fprintf(&sb, `<tr><td>Cổng mạng nguy hiểm</td><td>%d (trong %d record)</td></tr>`,
                countOpenPortsInResult(ctx), len(ctx.Result.Bang2))
        fmt.Fprintf(&sb, `<tr><td>Card mạng</td><td>%d</td></tr>`, len(ctx.Result.Bang3))
        fmt.Fprintf(&sb, `<tr><td>Thiết bị ngoại vi</td><td>%d (BadUSB: %d)</td></tr>`,
                len(ctx.Result.Bang4), countBadUSBInResult(ctx))
        fmt.Fprintf(&sb, `<tr><td>Tiến trình mã độc</td><td>%d</td></tr>`, len(ctx.Result.Bang5))
        sb.WriteString(`</table>`)

        // Phần 1: Bản quyền
        sb.WriteString(`<h2>1. Bản quyền & Tính hợp pháp phần mềm</h2>`)
        for _, r := range ctx.Result.Bang1 {
                fmt.Fprintf(&sb, `<h3>Phần mềm: %s</h3>`, r.SoftwareName)
                fmt.Fprintf(&sb, `<table><tr><th>Thuộc tính</th><th>Giá trị</th></tr>`)
                fmt.Fprintf(&sb, `<tr><td>Phiên bản</td><td>%s</td></tr>`, r.Version)
                fmt.Fprintf(&sb, `<tr><td>Product ID</td><td>%s</td></tr>`, r.ProductID)
                fmt.Fprintf(&sb, `<tr><td>Kênh cấp phép</td><td>%s</td></tr>`, r.LicensingChannel)
                fmt.Fprintf(&sb, `<tr><td>BIOS OEM Key</td><td>%s</td></tr>`, r.BIOSOEMKey)
                fmt.Fprintf(&sb, `<tr><td>Công cụ crack</td><td>%s</td></tr>`, r.CrackTool)
                fmt.Fprintf(&sb, `<tr><td>Đường dẫn file crack</td><td>%s</td></tr>`, r.CrackPath)
                fmt.Fprintf(&sb, `<tr><td>Trạng thái hợp pháp</td><td>%s</td></tr>`, legalStatus(r.Legal))
                sb.WriteString(`</table>`)

                if !r.Legal {
                        sb.WriteString(`<h3>Lệnh xử lý</h3>`)
                        sb.WriteString(`<code>REM 1. Gỡ product key hiện tại
slmgr /upk

REM 2. Xoá server KMS lậu khỏi registry
slmgr /ckms

REM 3. Xoá product key khỏi registry
slmgr /cpky

REM 4. Reset trạng thái về bản gốc (cần key hợp pháp)
slmgr /rearm

REM 5. Xoá file crack
del /F "`)
                        sb.WriteString(r.CrackPath)
                        sb.WriteString(`"

REM 6. Xoá tác vụ ngầm gia hạn
schtasks /delete /tn "KMSAuto" /f
schtasks /delete /tn "AutoActivation" /f
schtasks /delete /tn "WindowsActivation" /f</code>`)
                }
        }

        // Phần 2: Mạng
        sb.WriteString(`<h2>2. Mạng & Lỗ hổng bảo mật</h2>`)
        for _, r := range ctx.Result.Bang2 {
                fmt.Fprintf(&sb, `<h3>Trạng thái mạng</h3>`)
                fmt.Fprintf(&sb, `<p>Internet: %s | IP: %s | MAC: %s | ISP: %s</p>`,
                        r.InternetStatus, r.CurrentIP, r.CurrentMAC, r.ISP)
                if r.OpenPorts != "" {
                        fmt.Fprintf(&sb, `<h3>Cổng mở</h3><p>%s</p>`, r.OpenPorts)
                }
                if r.CVEID != "" {
                        fmt.Fprintf(&sb, `<h3>Lỗ hổng CVE</h3><p>%s - CVSS: %.1f</p><p>%s</p>`,
                                r.CVEID, r.CVSSScore, r.Notes)
                }

                sb.WriteString(`<h3>Lệnh xử lý</h3>`)
                sb.WriteString(`<code>REM 1. Đóng cổng SMB 445 (chống EternalBlue)
netsh advfirewall firewall add rule name="Block-SMB-445" dir=in action=block protocol=TCP localport=445

REM 2. Đóng cổng RDP 3389 (chống BlueKeep)
netsh advfirewall firewall add rule name="Block-RDP-3389" dir=in action=block protocol=TCP localport=3389

REM 3. Tắt SMBv1
reg add "HKLM\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters" /v SMB1 /t REG_DWORD /d 0 /f

REM 4. Vô hiệu hoá RDP (nếu không cần)
reg add "HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server" /v fDenyTSConnections /t REG_DWORD /d 1 /f

REM 5. Cập nhật Windows
usoclient StartScan
usoclient StartInstall</code>`)
        }

        // Phần 3: Phần cứng
        sb.WriteString(`<h2>3. Card mạng & Wi-Fi</h2>`)
        fmt.Fprintf(&sb, `<p>Phát hiện %d card mạng:</p>`, len(ctx.Result.Bang3))
        sb.WriteString(`<table><tr><th>Loại</th><th>Tên</th><th>MAC</th><th>Vị trí</th><th>Driver</th></tr>`)
        for _, r := range ctx.Result.Bang3 {
                fmt.Fprintf(&sb, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
                        r.AdapterType, r.DeviceName, r.MAC, r.ConnectionPos, r.DriverStatus)
        }
        sb.WriteString(`</table>`)
        sb.WriteString(`<h3>Lệnh xử lý</h3>`)
        sb.WriteString(`<code>REM Khoá quyền cắm card mạng USB mới
reg add "HKLM\SYSTEM\CurrentControlSet\Control\Class\{4d36e972-e325-11ce-bfc1-08002be10318}" /v DenyNewUSB /t REG_DWORD /d 1 /f

REM Tuỳ chọn: tắt Wi-Fi qua netsh
netsh interface set interface "Wi-Fi" disable</code>`)

        // Phần 4: USB
        sb.WriteString(`<h2>4. Thiết bị ngoại vi</h2>`)
        fmt.Fprintf(&sb, `<p>Phát hiện %d thiết bị (BadUSB: %d):</p>`,
                len(ctx.Result.Bang4), countBadUSBInResult(ctx))
        sb.WriteString(`<table><tr><th>Loại</th><th>Model</th><th>VID/PID</th><th>Ổ đĩa</th><th>Số lần cắm</th><th>Lần đầu</th><th>Lần cuối</th><th>BadUSB</th></tr>`)
        for _, r := range ctx.Result.Bang4 {
                badStr := "Không"
                if r.BadUSBWarning {
                        badStr = "CÓ!"
                }
                fmt.Fprintf(&sb, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td><td>%s</td><td style="color:#b10000">%s</td></tr>`,
                        r.DeviceType, r.VendorModel, r.VIDPID, r.DriveLetter,
                        r.PlugCount, orDash(r.FirstPlug), orDash(r.LastPlug), badStr)
        }
        sb.WriteString(`</table>`)

        // Chi tiết lịch sử TỪNG LẦN kết nối mỗi thiết bị
        for _, r := range ctx.Result.Bang4 {
                if len(r.Sessions) == 0 {
                        continue
                }
                fmt.Fprintf(&sb, `<h3>Lịch sử kết nối từng lần — %s (tổng %d lần cắm)</h3>`,
                        htmlEscape(r.VendorModel), r.PlugCount)
                sb.WriteString(`<table><tr><th>#</th><th>Thời điểm cắm</th><th>Thời điểm rút</th><th>Thời lượng</th></tr>`)
                max := len(r.Sessions)
                if max > 30 {
                        max = 30
                }
                for k := 0; k < max; k++ {
                        s := r.Sessions[k]
                        rem := s.Removal
                        if rem == "" {
                                rem = `<strong style="color:#1F6B36">ĐANG CẮM</strong>`
                        }
                        fmt.Fprintf(&sb, `<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
                                k+1, htmlEscape(s.Arrival), rem, htmlEscape(s.Duration))
                }
                sb.WriteString(`</table>`)
                if len(r.Sessions) > 30 {
                        fmt.Fprintf(&sb, `<p style="font-size:11px;color:#888;">... và %d lần nữa (hiển thị 30 lần gần nhất).</p>`, len(r.Sessions)-30)
                }
        }
        sb.WriteString(`<h3>Lệnh xử lý</h3>`)
        sb.WriteString(`<code>REM Khoá toàn bộ USB storage (GPO)
reg add "HKLM\SYSTEM\CurrentControlSet\Services\USBSTOR" /v Start /t REG_DWORD /d 4 /f

REM Áp dụng Deny_All qua Removable Storage policy
reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows\RemovableStorage" /v Deny_All /t REG_DWORD /d 1 /f

REM Xoá lịch sử thiết bị đã từng cắm
reg delete "HKLM\SYSTEM\CurrentControlSet\Enum\USBSTOR" /f
reg delete "HKLM\SYSTEM\CurrentControlSet\Enum\USB" /f
reg delete "HKLM\SYSTEM\MountedDevices" /f</code>`)

        // Phần 5: Mã độc
        sb.WriteString(`<h2>5. Mã độc & Keylogger</h2>`)
        for _, r := range ctx.Result.Bang5 {
                fmt.Fprintf(&sb, `<h3>Tiến trình: %s (PID %d)</h3>`, r.ProcessName, r.PID)
                fmt.Fprintf(&sb, `<table><tr><th>Thuộc tính</th><th>Giá trị</th></tr>`)
                fmt.Fprintf(&sb, `<tr><td>Loại</td><td>%s</td></tr>`, r.Type)
                fmt.Fprintf(&sb, `<tr><td>Chạy trong RAM</td><td>%v</td></tr>`, r.RunningInRAM)
                fmt.Fprintf(&sb, `<tr><td>Đường dẫn file</td><td>%s</td></tr>`, r.FilePath)
                fmt.Fprintf(&sb, `<tr><td>C2 Server</td><td>%s</td></tr>`, r.C2Server)
                fmt.Fprintf(&sb, `<tr><td>Dấu hiệu xoá log</td><td>%s</td></tr>`, r.LogWipeEvidence)
                sb.WriteString(`</table>`)

                sb.WriteString(`<h3>Lệnh xử lý</h3>`)
                fmt.Fprintf(&sb, `<code>REM 1. Kill tiến trình
taskkill /F /PID %d

REM 2. Xoá file độc hại
del /F "%s"

REM 3. Block C2 server đi
netsh advfirewall firewall add rule name="Block C2 Outbound" dir=out action=block remoteip=%s

REM 4. Xoá entry autorun trong registry
reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run" /v "%s" /f

REM 5. Gỡ hook keylogger (PowerShell Admin)
Get-Process | Where-Object { $_.Modules -ne $null -and $_.Modules.FileName -match "SetWindowsHookEx" } | Stop-Process -Force

REM 6. Quét lại RAM sau khi kill để chắc chắn fileless đã sạch
REM Dùng WinDbg hoặc process memory dump tool để kiểm tra dump</code>`,
                        r.PID, r.FilePath, extractIP(r.C2Server), r.ProcessName)
        }

        // Footer
        sb.WriteString(`<div class="footer">`)
        sb.WriteString(`<p>BCY-VKS v1.0.0 - Báo cáo sinh tự động`)
        fmt.Fprintf(&sb, ` lúc %s</p>`, time.Now().Format("02/01/2006 15:04:05"))
        sb.WriteString(`<p>Lưu ý: Tất cả lệnh trong báo cáo cần chạy ở quyền Administrator (Win+X &gt; Command Prompt (Admin))</p>`)
        sb.WriteString(`<p>Cảnh báo pháp lý: Chỉ áp dụng trên hệ thống bạn được quyền kiểm soát. Việc chống giám định số có thể vi phạm pháp luật nếu không trong phạm vi diễn tập được cấp phép.</p>`)
        sb.WriteString(`</div>`)

        sb.WriteString(`</body></html>`)

        return path, os.WriteFile(path, []byte(sb.String()), 0o644)
}

func legalStatus(legal bool) string {
        if legal {
                return "Hợp pháp"
        }
        return "Bất hợp pháp"
}

func countOpenPortsInResult(ctx *remediationContext) int {
        n := 0
        for _, r := range ctx.Result.Bang2 {
                if r.OpenPorts != "" {
                        n += countCSVHelper(r.OpenPorts)
                }
        }
        return n
}

func countBadUSBInResult(ctx *remediationContext) int {
        n := 0
        for _, r := range ctx.Result.Bang4 {
                if r.BadUSBWarning {
                        n++
                }
        }
        return n
}

func countCSVHelper(s string) int {
        c := 1
        for i := 0; i < len(s); i++ {
                if s[i] == ',' {
                        c++
                }
        }
        return c
}
