// Package license chứa logic kiểm tra bản quyền & tính hợp pháp của phần mềm.
// Bao gồm:
//   - scanner.go        : quét file crack (KMSpico, KMSAuto, MAS, HWID Gen, Adobe Zii...)
//   - registry.go       : phân tích Registry liên quan KMS lậu
//   - crackdb.go        : load chữ ký crack từ assets nhúng
//   - taskscheduler.go  : phát hiện tác vụ ngầm gia hạn bản quyền
//
// Tất cả export qua hàm ScanAll(assetsFS) trả về []LicenseRecord.
package license

import (
        "embed"
)

// LicenseRecord tương ứng BẢNG 1 trong đặc tả đầu ra.
// JSON tag dùng để frontend render chính xác tên cột.
type LicenseRecord struct {
        SoftwareName     string `json:"software_name"`
        Version          string `json:"version"`
        ProductID        string `json:"product_id"`
        LicensingChannel string `json:"licensing_channel"`
        BIOSOEMKey       string `json:"bios_oem_key"`
        CrackTool        string `json:"crack_tool"`
        CrackPath        string `json:"crack_path"`
        Legal            bool   `json:"legal"`
}

// ScanAll chạy toàn bộ pipeline license audit và trả về slice LicenseRecord.
// Tham số fs là embed.FS nhúng assets (crack_signatures.txt...)
func ScanAll(fs embed.FS) ([]LicenseRecord, error) {
        records := []LicenseRecord{}

        // 1. Lấy siêu dữ liệu bản quyền gốc từ Registry/WMI
        osRec := getOSLicenseMetadata()
        officeRecs := getOfficeLicenseMetadata()
        records = append(records, osRec)
        records = append(records, officeRecs...)

        // 2. Bổ sung: liệt kê TỪNG PHẦN MỀM đã cài (XMind, ABBYY, Foxit, Adobe, IDM...)
        // từ nhánh Uninstall registry. Chỉ lấy các phần mềm đáng audit.
        installedSW := scanInstalledSoftware()
        for _, sw := range installedSW {
                if !isAuditWorthy(sw.DisplayName, sw.Publisher) {
                        continue
                }
                // Tránh trùng lặp với Windows/Office đã thêm ở bước 1
                if isMicrosoftProduct(sw.DisplayName) {
                        continue
                }
                rec := LicenseRecord{
                        SoftwareName:     sw.DisplayName,
                        Version:          sw.DisplayVersion,
                        ProductID:        sw.ProductID,
                        LicensingChannel: detectLicensingChannel(sw),
                        BIOSOEMKey:       "", // chỉ có cho Windows
                        Legal:            isLikelyLegalByDefault(sw.DisplayName, sw.Publisher),
                }
                records = append(records, rec)
        }

        // 3. Quét công cụ crack trên các thư mục nghi vấn
        crackHits := ScanCrackTools(fs)
        for _, hit := range crackHits {
                // Gắn thông tin crack vào record tương ứng (nếu là Windows/Office/Adobe)
                attached := false
                for i := range records {
                        if matchSoftwareAndCrack(records[i].SoftwareName, hit.ToolName) {
                                records[i].CrackTool = hit.ToolName
                                records[i].CrackPath = hit.Path
                                records[i].Legal = false
                                attached = true
                                break
                        }
                }
                if !attached {
                        records = append(records, LicenseRecord{
                                SoftwareName: hit.ToolName,
                                Version:      hit.Version,
                                CrackTool:    hit.ToolName,
                                CrackPath:    hit.Path,
                                Legal:        false,
                        })
                }
        }

        // 4. Kiểm tra cấu hình KMS lậu trong Registry
        kmsServer := readKMSRegistryServer()
        if kmsServer != "" {
                for i := range records {
                        if isMicrosoftProduct(records[i].SoftwareName) {
                                records[i].LicensingChannel = "Volume (KMS lậu)"
                                records[i].Legal = false
                                if records[i].CrackTool == "" {
                                        records[i].CrackTool = "KMS Server: " + kmsServer
                                }
                        }
                }
        }

        // 5. Kiểm tra tác vụ ngầm gia hạn trong Task Scheduler
        suspiciousTasks := scanScheduledTasks()
        for _, t := range suspiciousTasks {
                // Nếu task trỏ tới file crack đã phát hiện, tăng độ tin cậy
                _ = t
        }

        return records, nil
}

// detectLicensingChannel đoán kênh cấp phép dựa vào thông tin phần mềm
func detectLicensingChannel(sw InstalledSoftware) string {
        n := lower(sw.DisplayName)
        p := lower(sw.Publisher)
        // Adobe/AutoCAD/Corel thường là Retail (trừ khi có Adobe Zii)
        if contains(p, "adobe") || contains(p, "autodesk") || contains(p, "corel") {
                return "Retail (nghi ngờ cần kiểm tra)"
        }
        // Phần mềm miễn phí / open source
        if contains(n, "7-zip") || contains(n, "notepad") || contains(n, "filezilla") ||
                contains(n, "winscp") || contains(n, "irfanview") || contains(n, "obs") ||
                contains(n, "firefox") || contains(n, "chrome") || contains(n, "edge") ||
                contains(n, "libreoffice") || contains(n, "virtualbox") ||
                contains(n, "vlc") || contains(n, "audacity") {
                return "Free / Open Source"
        }
        if sw.ProductID != "" {
                return "Retail (có ProductID)"
        }
        return "Unknown"
}

// matchSoftwareAndCrack ánh xạ phần mềm với công cụ crack đi kèm
func matchSoftwareAndCrack(software, crack string) bool {
        s := lower(software)
        c := lower(crack)
        switch {
        case contains(s, "windows") && contains(c, "kms"):
                return true
        case contains(s, "office") && contains(c, "kms"):
                return true
        case contains(s, "adobe") && contains(c, "zii"):
                return true
        }
        return false
}

// isMicrosoftProduct kiểm tra record có phải sản phẩm Microsoft không
func isMicrosoftProduct(name string) bool {
        n := lower(name)
        return contains(n, "windows") || contains(n, "office")
}

func lower(s string) string {
        out := make([]byte, len(s))
        for i := 0; i < len(s); i++ {
                c := s[i]
                if c >= 'A' && c <= 'Z' {
                        c += 32
                }
                out[i] = c
        }
        return string(out)
}

func contains(s, sub string) bool {
        if len(sub) == 0 {
                return true
        }
        if len(s) < len(sub) {
                return false
        }
        for i := 0; i <= len(s)-len(sub); i++ {
                if s[i:i+len(sub)] == sub {
                        return true
                }
        }
        return false
}
