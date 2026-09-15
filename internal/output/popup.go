package output

import (
        "fmt"
        "strings"
        "time"
)

// generatePopupHTML sinh HTML tóm tắt nhanh cho pop-up
// Trả về chuỗi HTML hoàn chỉnh (chứa style inline)
func generatePopupHTML(ctx *remediationContext) string {
        var sb strings.Builder
        sb.WriteString(`<!DOCTYPE html><html lang="vi"><head><meta charset="utf-8">`)
        sb.WriteString(`<meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline';">`)
        sb.WriteString(`<title>BCY-VKS - Đề xuất khắc phục</title>`)
        sb.WriteString(`<style>`)
        sb.WriteString(`body{font-family:'Segoe UI',sans-serif;background:#0F1419;color:#E0E0E0;padding:24px;margin:0;}`)
        sb.WriteString(`h1{color:#FFD93D;font-size:18px;border-bottom:1px solid #FFD93D;padding-bottom:8px;}`)
        sb.WriteString(`h2{color:#FF6B6B;font-size:14px;margin-top:16px;}`)
        sb.WriteString(`.section{margin-bottom:16px;padding:12px;border:1px solid #333;border-radius:4px;background:#1A1F2B;}`)
        sb.WriteString(`.critical{border-color:#FF4444;background:rgba(255,68,68,0.1);}`)
        sb.WriteString(`.high{border-color:#FFAA00;background:rgba(255,170,0,0.1);}`)
        sb.WriteString(`.low{border-color:#00AA00;background:rgba(0,170,0,0.1);}`)
        sb.WriteString(`code{display:block;background:#000;padding:8px;margin:6px 0;border-radius:3px;color:#7FDBFF;font-family:'Cascadia Mono',monospace;font-size:12px;word-wrap:break-word;}`)
        sb.WriteString(`.footer{margin-top:24px;font-size:11px;color:#666;text-align:center;}`)
        sb.WriteString(`</style></head><body>`)

        fmt.Fprintf(&sb, `<h1>BCY-VKS - Đề xuất Xử lý Khắc phục</h1>`)
        fmt.Fprintf(&sb, `<p>Thời điểm sinh: %s</p>`, time.Now().Format("02/01/2006 15:04:05"))

        // Phần 1: Bản quyền
        sb.WriteString(`<div class="section critical"><h2>1. Bản quyền & Tính hợp pháp</h2>`)
        for _, r := range ctx.Result.Bang1 {
                if !r.Legal {
                        fmt.Fprintf(&sb, `<p>Phát hiện: <strong>%s</strong> (%s)</p>`, r.CrackTool, r.SoftwareName)
                        sb.WriteString(`<p>Lệnh gỡ KMS lậu và bản quyền hệ điều hành:</p>`)
                        sb.WriteString(`<code>slmgr /upk   # gỡ product key hiện tại<br>`)
                        sb.WriteString(`slmgr /ckms  # xoá server KMS lậu<br>`)
                        sb.WriteString(`slmgr /cpky  # xoá product key khỏi registry<br>`)
                        sb.WriteString(`slmgr /rearm # reset trạng thái về bản gốc</code>`)
                        sb.WriteString(`<p>Nhiệm vụ: xoá task scheduler ngầm:</p>`)
                        sb.WriteString(`<code>schtasks /delete /tn "KMSAuto" /f<br>`)
                        sb.WriteString(`schtasks /delete /tn "AutoActivation" /f</code>`)
                }
        }
        sb.WriteString(`</div>`)

        // Phần 2: Mạng
        sb.WriteString(`<div class="section high"><h2>2. Mạng & Lỗ hổng</h2>`)
        for _, r := range ctx.Result.Bang2 {
                if r.OpenPorts != "" {
                        fmt.Fprintf(&sb, `<p>Cổng nguy hiểm đang mở: %s</p>`, r.OpenPorts)
                        sb.WriteString(`<p>Đóng cổng 445 (SMB) và 3389 (RDP):</p>`)
                        sb.WriteString(`<code>netsh advfirewall firewall add rule name="Block-SMB-445" dir=in action=block protocol=TCP localport=445<br>`)
                        sb.WriteString(`netsh advfirewall firewall add rule name="Block-RDP-3389" dir=in action=block protocol=TCP localport=3389<br>`)
                        sb.WriteString(`reg add "HKLM\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters" /v SMB1 /t REG_DWORD /d 0 /f<br>`)
                        sb.WriteString(`reg add "HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server" /v fDenyTSConnections /t REG_DWORD /d 1 /f</code>`)
                }
                if r.CVEID != "" {
                        fmt.Fprintf(&sb, `<p>CVE phát hiện: <strong>%s</strong> (CVSS %.1f) - %s</p>`,
                                r.CVEID, r.CVSSScore, r.Notes)
                        sb.WriteString(`<p>Hành động: cài đặt bản cập nhật Windows mới nhất:</p>`)
                        sb.WriteString(`<code>usoclient StartScan</code>`)
                }
        }
        sb.WriteString(`</div>`)

        // Phần 3: Phần cứng
        sb.WriteString(`<div class="section low"><h2>3. Card mạng & Wi-Fi</h2>`)
        sb.WriteString(`<p>Vô hiệu hoá card mạng gắn ngoài chưa được phê duyệt qua GPO:</p>`)
        sb.WriteString(`<code>reg add "HKLM\SYSTEM\CurrentControlSet\Control\Class\{4d36e972-e325-11ce-bfc1-08002be10318}" /v *DisableAdapter /t REG_DWORD /d 1 /f</code>`)
        sb.WriteString(`</div>`)

        // Phần 4: USB
        sb.WriteString(`<div class="section high"><h2>4. Thiết bị ngoại vi</h2>`)
        badUSBCount := 0
        for _, r := range ctx.Result.Bang4 {
                if r.BadUSBWarning {
                        badUSBCount++
                }
        }
        if badUSBCount > 0 {
                fmt.Fprintf(&sb, `<p style="color:#FF4444">Phát hiện %d thiết bị BadUSB!</p>`, badUSBCount)
                sb.WriteString(`<p>Khóa toàn bộ USB storage qua GPO:</p>`)
                sb.WriteString(`<code>reg add "HKLM\SYSTEM\CurrentControlSet\Services\USBSTOR" /v Start /t REG_DWORD /d 4 /f<br>`)
                sb.WriteString(`reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows\RemovableStorage" /v Deny_All /t REG_DWORD /d 1 /f</code>`)
        } else {
                sb.WriteString(`<p>Không phát hiện BadUSB. Khuyến nghị bật GPO khóa USB ngoài whitelist:</p>`)
                sb.WriteString(`<code>reg add "HKLM\SYSTEM\CurrentControlSet\Services\USBSTOR" /v Start /t REG_DWORD /d 4 /f</code>`)
        }
        sb.WriteString(`</div>`)

        // Phần 5: Mã độc
        sb.WriteString(`<div class="section critical"><h2>5. Mã độc & Keylogger</h2>`)
        for _, r := range ctx.Result.Bang5 {
                fmt.Fprintf(&sb, `<p>Tiến trình: <strong>%s</strong> (PID %d, loại: %s)</p>`,
                        r.ProcessName, r.PID, r.Type)
                fmt.Fprintf(&sb, `<p>Đường dẫn file: %s</p>`, r.FilePath)
                if r.C2Server != "" {
                        fmt.Fprintf(&sb, `<p>C2 Server: <span style="color:#FF6B6B">%s</span></p>`, r.C2Server)
                }
                sb.WriteString(`<p>Cô lập và tiêu diệt:</p>`)
                fmt.Fprintf(&sb, `<code>taskkill /F /PID %d<br>`, r.PID)
                fmt.Fprintf(&sb, `del /F "%s"<br>`, r.FilePath)
                sb.WriteString(`netsh advfirewall firewall add rule name="Block C2 Outbound" dir=out action=block remoteip=` + extractIP(r.C2Server) + `</code>`)
                sb.WriteString(`<p>Gỡ hook keylogger (chạy PowerShell Admin):</p>`)
                sb.WriteString(`<code>Get-Process | Where-Object {$_.Modules -ne $null -and $_.Modules.FileName -match "SetWindowsHookEx"} | Stop-Process -Force</code>`)
        }
        sb.WriteString(`</div>`)

        sb.WriteString(`<div class="footer">BCY-VKS v1.0.0 - Báo cáo sinh tự động. Vui lòng chạy ở quyền Administrator để áp dụng lệnh.</div>`)
        sb.WriteString(`</body></html>`)
        return sb.String()
}

// extractIP tách IP từ chuỗi "IP:port"
func extractIP(addr string) string {
        for i := 0; i < len(addr); i++ {
                if addr[i] == ':' {
                        return addr[:i]
                }
        }
        return addr
}

// parseIntSafe parse int an toàn (không dùng strconv)
func parseIntSafe(s string) int {
        n := 0
        for i := 0; i < len(s); i++ {
                c := s[i]
                if c < '0' || c > '9' {
                        break
                }
                n = n*10 + int(c-'0')
        }
        return n
}
