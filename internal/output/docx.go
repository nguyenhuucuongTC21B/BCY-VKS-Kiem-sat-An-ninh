package output

import (
        "archive/zip"
        "bytes"
        "fmt"
        "os"
        "strings"
        "time"
)

// writeDOCXReport tạo file .docx THẬT theo chuẩn Office Open XML (OOXML)
// File .docx thực chất là ZIP archive chứa XML files:
//   - [Content_Types].xml
//   - _rels/.rels
//   - word/document.xml
//   - word/_rels/document.xml.rels
//   - word/styles.xml
//
// Word 2016+ từ chối mở file .docx giả (HTML đổi đuôi) vì lý do bảo mật.
// Cần tạo .docx đúng chuẩn để mở được.
func writeDOCXReport(ctx *remediationContext) (string, error) {
        path := timestampFileName("BCY-VKS-Remediation", "docx")

        // Tạo nội dung document.xml với escaping XML
        var doc strings.Builder
        doc.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
        doc.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
        doc.WriteString(`<w:body>`)

        // === Tiêu đề ===
        addPara(&doc, "BCY-VKS - BÁO CÁO ĐỀ XUẤT KHẮC PHỤC", true, 28, "8B0000")
        addPara(&doc, "Ban Cơ Yếu - Viện Kiểm Sát An Ninh", false, 14, "1A1F2C")
        addPara(&doc, "Thời điểm sinh: "+time.Now().Format("02/01/2006 15:04:05"), false, 11, "6B6B6B")
        addPara(&doc, "", false, 11, "")

        // === Tóm tắt phát hiện ===
        addPara(&doc, "TÓM TẮT PHÁT HIỆN", true, 16, "8B0000")
        addPara(&doc, fmt.Sprintf("- Bản quyền lậu: %d mục", len(ctx.Result.Bang1)), false, 12, "")
        addPara(&doc, fmt.Sprintf("- Record mạng: %d mục", len(ctx.Result.Bang2)), false, 12, "")
        addPara(&doc, fmt.Sprintf("- Card mạng: %d mục", len(ctx.Result.Bang3)), false, 12, "")
        addPara(&doc, fmt.Sprintf("- Thiết bị ngoại vi: %d mục", len(ctx.Result.Bang4)), false, 12, "")
        addPara(&doc, fmt.Sprintf("- Tiến trình mã độc: %d mục", len(ctx.Result.Bang5)), false, 12, "")
        addPara(&doc, "", false, 11, "")

        // === PHẦN 1: BẢN QUYỀN & CĂN CỨ PHÁP LÝ ===
        addPara(&doc, "1. BẢN QUYỀN PHẦN MỀM & CĂN CỨ PHÁP LÝ", true, 14, "8B0000")
        for _, r := range ctx.Result.Bang1 {
                addPara(&doc, "── "+r.SoftwareName+" ──", true, 12, "1A1F2C")
                addPara(&doc, fmt.Sprintf("   Phiên bản: %s", r.Version), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Product ID: %s", r.ProductID), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Kênh cấp phép: %s", r.LicensingChannel), false, 11, "")
                addPara(&doc, fmt.Sprintf("   BIOS OEM Key: %s", r.BIOSOEMKey), false, 11, "")
                if r.CrackTool != "" {
                        addPara(&doc, fmt.Sprintf("   ⚠ Công cụ crack: %s", r.CrackTool), false, 11, "8B0000")
                }
                if r.CrackPath != "" {
                        addPara(&doc, fmt.Sprintf("   Đường dẫn: %s", r.CrackPath), false, 11, "")
                }
                if r.Legal {
                        addPara(&doc, "   Trạng thái: HỢP PHÁP", false, 11, "2D7A3E")
                } else {
                        addPara(&doc, "   Trạng thái: BẤT HỢP PHÁP", false, 11, "8B0000")
                }

                // === CĂN CỨ PHÁP LÝ (BỔ SUNG) ===
                assessment := AssessLicenseLegal(r.SoftwareName, r.CrackTool, r.Legal)
                addPara(&doc, "   [ĐÁNH GIÁ PHÁP LÝ]", true, 11, "8B0000")
                addPara(&doc, fmt.Sprintf("   Mức độ nghiêm trọng: %s", assessment.Severity), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Đánh giá nguy cơ: %s", assessment.Description), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Căn cứ pháp lý: %s", assessment.LegalBasis), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Hình thức xử lý: %s", assessment.Liability), false, 11, "")

                // Lệnh khắc phục
                if !r.Legal {
                        addPara(&doc, "   Lệnh khắc phục:", true, 11, "")
                        addPara(&doc, "     slmgr /upk", false, 11, "")
                        addPara(&doc, "     slmgr /ckms", false, 11, "")
                        addPara(&doc, "     slmgr /cpky", false, 11, "")
                        addPara(&doc, fmt.Sprintf("     del /F \"%s\"", r.CrackPath), false, 11, "")
                }
                addPara(&doc, "", false, 11, "")
        }

        // === PHẦN 2: MẠNG & LỖ HỔNG + CĂN CỨ PHÁP LÝ ===
        addPara(&doc, "2. MẠNG & LỖ HỔNG + CĂN CỨ PHÁP LÝ", true, 14, "8B0000")
        for _, r := range ctx.Result.Bang2 {
                addPara(&doc, fmt.Sprintf("   Trạng thái Internet: %s", r.InternetStatus), false, 11, "")
                addPara(&doc, fmt.Sprintf("   IP: %s | MAC: %s | ISP: %s", r.CurrentIP, r.CurrentMAC, r.ISP), false, 11, "")
                if r.OpenPorts != "" {
                        addPara(&doc, fmt.Sprintf("   Cổng mở: %s", r.OpenPorts), false, 11, "8B0000")
                        // Đánh giá pháp lý cho từng cổng
                        ports := strings.Split(r.OpenPorts, ",")
                        for _, p := range ports {
                                p = strings.TrimSpace(p)
                                if p == "" {
                                        continue
                                }
                                portNum := parseIntSafe(p)
                                portAssessment := AssessPortLegal(portNum, p)
                                addPara(&doc, fmt.Sprintf("   ── Cổng %s ──", p), true, 11, "1A1F2C")
                                addPara(&doc, fmt.Sprintf("   Mức độ: %s", portAssessment.Severity), false, 11, "")
                                addPara(&doc, fmt.Sprintf("   Đánh giá: %s", portAssessment.Description), false, 11, "")
                                addPara(&doc, fmt.Sprintf("   Căn cứ: %s", portAssessment.LegalBasis), false, 11, "")
                                addPara(&doc, fmt.Sprintf("   Xử lý: %s", portAssessment.Liability), false, 11, "")
                        }
                        // Lệnh khắc phục
                        addPara(&doc, "   Lệnh khắc phục:", true, 11, "")
                        addPara(&doc, "     netsh advfirewall firewall add rule name=\"Block-SMB-445\" dir=in action=block protocol=TCP localport=445", false, 11, "")
                        addPara(&doc, "     netsh advfirewall firewall add rule name=\"Block-RDP-3389\" dir=in action=block protocol=TCP localport=3389", false, 11, "")
                }
                if r.CVEID != "" {
                        addPara(&doc, fmt.Sprintf("   CVE: %s (CVSS %.1f) - %s", r.CVEID, r.CVSSScore, r.Notes), false, 11, "8B0000")
                }
                if r.ExploitResult != "" {
                        addPara(&doc, fmt.Sprintf("   Kết quả pentest: %s", r.ExploitResult), false, 11, "")
                }
                addPara(&doc, "", false, 11, "")
        }

        // === PHẦN 3: CARD MẠNG ===
        addPara(&doc, "3. KIỂM KÊ CARD MẠNG", true, 14, "8B0000")
        for _, r := range ctx.Result.Bang3 {
                addPara(&doc, fmt.Sprintf("   %s (%s) - MAC: %s, Driver: %s", r.DeviceName, r.AdapterType, r.MAC, r.DriverStatus), false, 11, "")
                if r.SSIDList != "" {
                        addPara(&doc, fmt.Sprintf("   SSID đã kết nối: %s", r.SSIDList), false, 11, "")
                }
        }
        addPara(&doc, "", false, 11, "")

        // === PHẦN 4: THIẾT BỊ NGOẠI VI + CĂN CỨ PHÁP LÝ ===
        addPara(&doc, "4. THIẾT BỊ NGOẠI VI & CĂN CỨ PHÁP LÝ", true, 14, "8B0000")
        for _, r := range ctx.Result.Bang4 {
                addPara(&doc, fmt.Sprintf("   %s (%s) - VID/PID: %s", r.VendorModel, r.DeviceType, r.VIDPID), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Hardware ID: %s | Ổ đĩa: %s | Số lần cắm: %d", r.HardwareID, r.DriveLetter, r.PlugCount), false, 11, "")
                if r.BadUSBWarning {
                        addPara(&doc, "   ⚠ CẢNH BÁO: BADUSB!", false, 11, "8B0000")
                        // Đánh giá pháp lý cho BadUSB
                        badUSBAssessment := AssessBadUSBLegal(r.VendorModel, r.VIDPID)
                        addPara(&doc, "   [ĐÁNH GIÁ PHÁP LÝ BADUSB]", true, 11, "8B0000")
                        addPara(&doc, fmt.Sprintf("   Mức độ: %s", badUSBAssessment.Severity), false, 11, "")
                        addPara(&doc, fmt.Sprintf("   Đánh giá: %s", badUSBAssessment.Description), false, 11, "")
                        addPara(&doc, fmt.Sprintf("   Căn cứ: %s", badUSBAssessment.LegalBasis), false, 11, "")
                        addPara(&doc, fmt.Sprintf("   Xử lý: %s", badUSBAssessment.Liability), false, 11, "")
                }
        }
        // Lệnh khắc phục USB
        addPara(&doc, "   Lệnh khắc phục (chặn USB qua GPO):", true, 11, "")
        addPara(&doc, "     reg add \"HKLM\\SYSTEM\\CurrentControlSet\\Services\\USBSTOR\" /v Start /t REG_DWORD /d 4 /f", false, 11, "")
        addPara(&doc, "     reg add \"HKLM\\SOFTWARE\\Policies\\Microsoft\\Windows\\RemovableStorage\" /v Deny_All /t REG_DWORD /d 1 /f", false, 11, "")
        addPara(&doc, "", false, 11, "")

        // === PHẦN 5: MÃ ĐỘC + CĂN CỨ PHÁP LÝ ===
        addPara(&doc, "5. GIÁM ĐỊNH MÃ ĐỘC & CĂN CỨ PHÁP LÝ", true, 14, "8B0000")
        for _, r := range ctx.Result.Bang5 {
                addPara(&doc, fmt.Sprintf("   %s (PID %d, loại: %s)", r.ProcessName, r.PID, r.Type), false, 11, "8B0000")
                addPara(&doc, fmt.Sprintf("   Đường dẫn: %s", r.FilePath), false, 11, "")
                if r.C2Server != "" {
                        addPara(&doc, fmt.Sprintf("   C2 Server: %s", r.C2Server), false, 11, "8B0000")
                }
                if r.LogWipeEvidence != "" {
                        addPara(&doc, fmt.Sprintf("   Dấu vết xóa log: %s", r.LogWipeEvidence), false, 11, "")
                }
                if r.DangerLevel != "" {
                        addPara(&doc, fmt.Sprintf("   Mức độ nguy hiểm: %s", r.DangerLevel), false, 11, "")
                }

                // Đánh giá pháp lý cho mã độc
                malwareAssessment := AssessMalwareLegal(r.ProcessName, r.C2Server != "", r.RunningInRAM, r.DangerLevel)
                addPara(&doc, "   [ĐÁNH GIÁ PHÁP LÝ MÃ ĐỘC]", true, 11, "8B0000")
                addPara(&doc, fmt.Sprintf("   Mức độ: %s", malwareAssessment.Severity), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Đánh giá: %s", malwareAssessment.Description), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Căn cứ: %s", malwareAssessment.LegalBasis), false, 11, "")
                addPara(&doc, fmt.Sprintf("   Xử lý: %s", malwareAssessment.Liability), false, 11, "")

                // Lệnh khắc phục
                addPara(&doc, fmt.Sprintf("   Lệnh khắc phục:"), true, 11, "")
                addPara(&doc, fmt.Sprintf("     taskkill /F /PID %d", r.PID), false, 11, "")
                addPara(&doc, fmt.Sprintf("     del /F \"%s\"", r.FilePath), false, 11, "")
                if r.C2Server != "" {
                        addPara(&doc, fmt.Sprintf("     netsh advfirewall firewall add rule name=\"Block C2 Outbound\" dir=out action=block remoteip=%s", extractIP(r.C2Server)), false, 11, "")
                }
                addPara(&doc, "", false, 11, "")
        }

        // === FOOTER: LƯU Ý PHÁP LÝ ===
        addPara(&doc, "", false, 11, "")
        addPara(&doc, "LƯU Ý PHÁP LÝ", true, 12, "8B0000")
        addPara(&doc, "Báo cáo này được lập theo quy định của:", false, 11, "")
        addPara(&doc, "- Luật An ninh mạng 2018 (Điều 28, 29 - bảo đảm an toàn thông tin)", false, 11, "")
        addPara(&doc, "- Luật Bảo vệ bí mật nhà nước 2018 (Điều 8 - quản lý bí mật nhà nước)", false, 11, "")
        addPara(&doc, "- Bộ luật Hình sự 2015 (sửa đổi 2017) - Điều 225, 288, 289, 290", false, 11, "")
        addPara(&doc, "- Luật Sở hữu trí tuệ 2005 + Nghị định 22/2018/NĐ-CP (bản quyền phần mềm)", false, 11, "")
        addPara(&doc, "- Quy định của Ban Cơ Yếu về an toàn thông tin và bảo mật", false, 11, "")
        addPara(&doc, "", false, 11, "")
        addPara(&doc, "Hình thức xử lý:", false, 11, "")
        addPara(&doc, "- KỶ LUẬT: Khiển trách, cảnh cáo, cách chức, sa thải (theo quy định nội bộ)", false, 11, "")
        addPara(&doc, "- HÀNH CHÍNH: Phạt tiền theo Nghị định 15/2020/NĐ-CP, 22/2018/NĐ-CP", false, 11, "")
        addPara(&doc, "- HÌNH SỰ: Phạt tù theo BLHS 2015 nếu vi phạm nghiêm trọng", false, 11, "")
        addPara(&doc, "", false, 11, "")
        addPara(&doc, "Ngày lập báo cáo: "+time.Now().Format("02/01/2006"), false, 11, "6B6B6B")

        doc.WriteString(`</w:body>`)
        doc.WriteString(`</w:document>`)

        // Tạo file .docx thật (ZIP với cấu trúc OOXML)
        if err := writeRealDOCX(path, doc.String()); err != nil {
                return "", err
        }
        return path, nil
}

// writeRealDOCX tạo file .docx thật bằng ZIP + XML theo chuẩn OOXML
// Cấu trúc ZIP:
//   [Content_Types].xml
//   _rels/.rels
//   word/document.xml
//   word/_rels/document.xml.rels
//   word/styles.xml
func writeRealDOCX(path string, documentXML string) error {
        // Tạo buffer cho ZIP
        buf := new(bytes.Buffer)
        zipWriter := zip.NewWriter(buf)

        // File 1: [Content_Types].xml
        contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`
        w1, err := zipWriter.Create("[Content_Types].xml")
        if err != nil {
                return err
        }
        w1.Write([]byte(contentTypes))

        // File 2: _rels/.rels
        rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
        w2, err := zipWriter.Create("_rels/.rels")
        if err != nil {
                return err
        }
        w2.Write([]byte(rels))

        // File 3: word/document.xml
        w3, err := zipWriter.Create("word/document.xml")
        if err != nil {
                return err
        }
        w3.Write([]byte(documentXML))

        // File 4: word/_rels/document.xml.rels
        docRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
        w4, err := zipWriter.Create("word/_rels/document.xml.rels")
        if err != nil {
                return err
        }
        w4.Write([]byte(docRels))

        // File 5: word/styles.xml - minimal styles để Word mở được
        styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri" w:cs="Calibri"/>
        <w:sz w:val="22"/>
      </w:rPr>
    </w:rPrDefault>
    <w:pPrDefault>
      <w:pPr>
        <w:spacing w:after="160" w:line="276" w:lineRule="auto"/>
      </w:pPr>
    </w:pPrDefault>
  </w:docDefaults>
</w:styles>`
        w5, err := zipWriter.Create("word/styles.xml")
        if err != nil {
                return err
        }
        w5.Write([]byte(styles))

        // Đóng ZIP writer
        if err := zipWriter.Close(); err != nil {
                return err
        }

        // Ghi file ra disk
        return os.WriteFile(path, buf.Bytes(), 0o644)
}

// addPara thêm 1 paragraph vào document XML
// bold=true: in đậm
// size: font size (half-points, vd 22 = 11pt)
// color: hex color (vd "8B0000"), "" = default black
func addPara(sb *strings.Builder, text string, bold bool, size int, color string) {
        sb.WriteString(`<w:p>`)
        sb.WriteString(`<w:r>`)
        rPr := `<w:rPr>`
        if bold {
                rPr += `<w:b/>`
        }
        if size > 0 {
                rPr += fmt.Sprintf(`<w:sz w:val="%d"/>`, size)
        }
        if color != "" {
                rPr += fmt.Sprintf(`<w:color w:val="%s"/>`, color)
        }
        rPr += `</w:rPr>`
        sb.WriteString(rPr)
        // Escape XML characters
        text = xmlEscape(text)
        sb.WriteString(fmt.Sprintf(`<w:t xml:space="preserve">%s</w:t>`, text))
        sb.WriteString(`</w:r>`)
        sb.WriteString(`</w:p>`)
}

// xmlEscape escape các ký tự đặc biệt trong XML
func xmlEscape(s string) string {
        s = strings.ReplaceAll(s, "&", "&amp;")
        s = strings.ReplaceAll(s, "<", "&lt;")
        s = strings.ReplaceAll(s, ">", "&gt;")
        s = strings.ReplaceAll(s, `"`, "&quot;")
        s = strings.ReplaceAll(s, `'`, "&apos;")
        return s
}
