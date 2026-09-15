// Package output chứa logic sinh báo cáo đề xuất khắc phục.
//
// 3 chế độ:
//   - popup : trả về chuỗi HTML hiển thị nhanh trong cửa sổ Pop-up
//   - docx  : sinh file .docx hoàn chỉnh
//   - html  : sinh file .html hoàn chỉnh có in được
//
// API chính: GenerateRemediation(result, format, assets) (string, error)
package output

import (
        "embed"
        "encoding/json"
        "fmt"
        "os"
        "path/filepath"
        "strings"
        "time"

        "bcy-vks/internal/hardware"
        "bcy-vks/internal/license"
        "bcy-vks/internal/malware"
        "bcy-vks/internal/network"
)

// remediationContext gói dữ liệu đầu vào cho generator
type remediationContext struct {
        Result *ScanResultWrapper
        Assets embed.FS
}

// ScanResultWrapper là alias của struct ScanResult trong main.go
// Do package output không thể import package main (vòng), ta định nghĩa lại.
type ScanResultWrapper struct {
        Bang1 []license.LicenseRecord
        Bang2 []network.NetworkRecord
        Bang3 []hardware.HardwareRecord
        Bang4 []hardware.PeripheralRec
        Bang5 []malware.MalwareRecord
}

// GenerateRemediation là API chính cho nút "Đề xuất, xử lý khắc phục"
// Trả về:
//   - format="popup": chuỗi HTML để hiển thị nhanh
//   - format="docx": đường dẫn file .docx
//   - format="html": đường dẫn file .html
func GenerateRemediation(result interface{}, format string, fs embed.FS) (string, error) {
        // Ép kiểu từ interface{} sang struct ScanResult của main
        // (main.ScanResult có cùng field tag JSON)
        wrapper := convertFromMain(result)
        ctx := &remiationContextWrapper{
                Result: wrapper,
                Assets: fs,
        }

        switch strings.ToLower(format) {
        case "popup":
                return generatePopupHTML(ctx), nil
        case "docx":
                path, err := writeDOCXReport(ctx)
                if err != nil {
                        return "", err
                }
                return path, nil
        case "html":
                path, err := writeHTMLReport(ctx)
                if err != nil {
                        return "", err
                }
                return path, nil
        default:
                return "", fmt.Errorf("format không hỗ trợ: %s", format)
        }
}

// remediationContext alias
type remediationContextOld = remediationContext

// remiationContextWrapper chỉ là alias để không trùng tên
type remiationContextWrapper = remediationContext

// convertFromMain ép kiểu interface{} -> ScanResultWrapper
// Vì main.ScanResult có JSON tag giống nhau nhưng là kiểu khác,
// ta dùng JSON marshal để copy field an toàn (reflection tự nhiên).
func convertFromMain(result interface{}) *ScanResultWrapper {
        // Trường hợp nhanh: người gọi đã pass đúng kiểu
        if r, ok := result.(*ScanResultWrapper); ok {
                return r
        }
        // Đường dẫn chung: marshal -> unmarshal
        b, err := json.Marshal(result)
        if err != nil {
                return &ScanResultWrapper{}
        }
        out := &ScanResultWrapper{}
        if err := json.Unmarshal(b, out); err != nil {
                return &ScanResultWrapper{}
        }
        return out
}

// timestampFileName sinh tên file an toàn có timestamp
func timestampFileName(prefix, ext string) string {
        home, _ := os.UserHomeDir()
        dir := filepath.Join(home, "Documents", "BCY-VKS-Reports")
        _ = os.MkdirAll(dir, 0o755)
        stamp := time.Now().Format("20060102-150405")
        return filepath.Join(dir, fmt.Sprintf("%s-%s.%s", prefix, stamp, ext))
}

// orDash thay chuỗi rỗng bằng dấu gạch ngang để hiển thị
func orDash(s string) string {
        if strings.TrimSpace(s) == "" {
                return "—"
        }
        return s
}

// htmlEscape escape các ký tự HTML nguy hiểm khi nhúng dữ liệu vào báo cáo
func htmlEscape(s string) string {
        r := strings.NewReplacer(
                "&", "&amp;",
                "<", "&lt;",
                ">", "&gt;",
                `"`, "&quot;",
                "'", "&#39;",
        )
        return r.Replace(s)
}
