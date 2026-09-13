// Package output - legal_basis.go
// Chứa dữ liệu căn cứ pháp lý để chèn vào báo cáo Word/HTML
// Mỗi loại vi phạm được ánh xạ tới:
//   - Mức độ nghiêm trọng (CRITICAL/HIGH/MEDIUM/LOW)
//   - Mô tả nguy cơ
//   - Căn cứ pháp lý (Luật Cơ yếu, Luật An ninh mạng, Luật Bảo vệ bí mật nhà nước, BLHS)
//   - Hình thức xử lý (kỷ luật / hành chính / hình sự)
package output

import (
	"fmt"
	"strings"
)

// LegalAssessment chứa đánh giá pháp lý cho 1 vi phạm
type LegalAssessment struct {
	Severity    string // "CRITICAL", "HIGH", "MEDIUM", "LOW"
	Description string
	LegalBasis  string // Các Luật, Điều khoản áp dụng
	Liability   string // Hình thức xử lý (kỷ luật/hành chính/hình sự)
}

// AssessLicenseLegal đánh giá pháp lý cho bản quyền phần mềm
func AssessLicenseLegal(software, crackTool string, legal bool) LegalAssessment {
	if legal {
		return LegalAssessment{
			Severity:    "AN TOÀN",
			Description: fmt.Sprintf("Phần mềm %s có bản quyền hợp pháp, không phát hiện dấu hiệu vi phạm.", software),
			LegalBasis:  "Tuân thủ Luật Sở hữu trí tuệ 2005 (sửa đổi 2009).",
			Liability:   "Không áp dụng xử lý.",
		}
	}
	if crackTool == "" {
		return LegalAssessment{
			Severity:    "CẦN KIỂM TRA",
			Description: fmt.Sprintf("Phần mềm %s có dấu hiệu bất hợp pháp, cần kiểm tra giấy phép.", software),
			LegalBasis:  "Luật Sở hữu trí tuệ 2005; Nghị định 22/2018/NĐ-CP (quyền tác giả, quyền liên quan).",
			Liability:   "HÀNH CHÍNH: Phạt tiền 5-50 triệu đồng (Nghị định 22/2018/NĐ-CP).",
		}
	}
	// Có crack tool
	switch {
	case strings.Contains(strings.ToLower(crackTool), "kms"):
		return LegalAssessment{
			Severity:    "CRITICAL",
			Description: fmt.Sprintf("Phát hiện công cụ crack KMS '%s' để kích hoạt bản quyền lậu cho %s. Vi phạm nghiêm trọng quyền sở hữu trí tuệ, có nguy cơ bị hacker lợi dụng KMS server để phát tán mã độc.", crackTool, software),
			LegalBasis:  "Luật Sở hữu trí tuệ 2005 Điều 7, 18, 22; Nghị định 22/2018/NĐ-CP; Luật An ninh mạng 2018 Điều 28; Quy định của Ban Cơ yếu về sử dụng phần mềm.",
			Liability:   "KỶ LUẬT: Khiển trách, cảnh cáo, cách chức. HÀNH CHÍNH: Phạt tiền 50-100 triệu đồng. HÌNH SỰ: BLHS 2015 Điều 225 (Tội xâm phạm quyền sở hữu trí tuệ) nếu gây thiệt hại nghiêm trọng, phạt tù 06 tháng - 03 năm.",
		}
	case strings.Contains(strings.ToLower(crackTool), "adobe"):
		return LegalAssessment{
			Severity:    "CRITICAL",
			Description: fmt.Sprintf("Phát hiện công cụ crack Adobe '%s' vi phạm bản quyền phần mềm Adobe.", crackTool),
			LegalBasis:  "Luật Sở hữu trí tuệ 2005; Nghị định 22/2018/NĐ-CP; BLHS 2015 Điều 225.",
			Liability:   "HÀNH CHÍNH: Phạt tiền 30-100 triệu đồng. HÌNH SỰ: BLHS Điều 225 phạt tù 06 tháng - 03 năm. KỶ LUẬT: Cảnh cáo, cách chức.",
		}
	default:
		return LegalAssessment{
			Severity:    "HIGH",
			Description: fmt.Sprintf("Phát hiện công cụ crack '%s' cho phần mềm %s. Vi phạm bản quyền phần mềm.", crackTool, software),
			LegalBasis:  "Luật Sở hữu trí tuệ 2005; Nghị định 22/2018/NĐ-CP; BLHS 2015 Điều 225.",
			Liability:   "HÀNH CHÍNH: Phạt tiền 20-50 triệu đồng. KỶ LUẬT: Khiển trách, cảnh cáo.",
		}
	}
}

// AssessPortLegal đánh giá pháp lý cho cổng mạng mở
func AssessPortLegal(port int, name string) LegalAssessment {
	switch port {
	case 445:
		return LegalAssessment{
			Severity:    "CRITICAL",
			Description: fmt.Sprintf("Cổng %d (%s) đang MỞ. Cổng này có thể bị khai thác EternalBlue (MS17-010) → WannaCry ransomware → MẤT DỮ LIỆU VĨNH VIỄN. Đây là lỗ hổng cực kỳ nghiêm trọng đã được Microsoft vá từ 2017 nhưng vẫn còn xuất hiện.", port, name),
			LegalBasis:  "Luật An ninh mạng 2018 Điều 28, 29 (bảo đảm an toàn thông tin); Luật Bảo vệ bí mật nhà nước 2018 Điều 8; Quy định của Ban Cơ yếu về an toàn thông tin.",
			Liability:   "KỶ LUẬT: Cảnh cáo, cách chức. HÀNH CHÍNH: Nghị định 15/2020/NĐ-CP phạt tiền 10-20 triệu đồng. HÌNH SỰ: BLHS Điều 288, 290 nếu bị tấn công và gây thiệt hại, phạt tù 01-07 năm.",
		}
	case 3389:
		return LegalAssessment{
			Severity:    "CRITICAL",
			Description: fmt.Sprintf("Cổng %d (RDP - %s) đang MỞ. RDP có thể bị BlueKeep (CVE-2019-0708) RCE → chiếm quyền điều khiển máy từ xa → mất toàn bộ quyền kiểm soát. Đây là vector tấn công chính vào hệ thống nội bộ.", port, name),
			LegalBasis:  "Luật An ninh mạng 2018 Điều 28, 29; BLHS 2015 Điều 288 (Tội vi phạm quy định về bảo mật thông tin), Điều 290 (Tội phá rối hoạt động máy tính).",
			Liability:   "KỶ LUẬT: Cảnh cáo. HÀNH CHÍNH: Phạt tiền 10-20 triệu đồng. HÌNH SỰ: Phạt tù 01-07 năm nếu bị khai thác gây hậu quả.",
		}
	case 21, 23, 25, 135, 139, 80:
		return LegalAssessment{
			Severity:    "HIGH",
			Description: fmt.Sprintf("Cổng %d (%s) đang MỞ. Đây là cổng dịch vụ cũ, truyền dữ liệu clear text, dễ bị sniff hoặc khai thác.", port, name),
			LegalBasis:  "Luật An ninh mạng 2018 Điều 28 (bảo đảm an toàn thông tin); Nghị định 15/2020/NĐ-CP.",
			Liability:   "HÀNH CHÍNH: Phạt tiền 5-10 triệu đồng. KỶ LUẬT: Khiển trách.",
		}
	case 1433, 3306:
		return LegalAssessment{
			Severity:    "HIGH",
			Description: fmt.Sprintf("Cổng %d (%s) đang MỞ. Cơ sở dữ liệu mở ra ngoài có thể bị dump toàn bộ dữ liệu, rò rỉ thông tin.", port, name),
			LegalBasis:  "Luật An ninh mạng 2018 Điều 28; Luật Bảo vệ bí mật nhà nước 2018 Điều 8.",
			Liability:   "HÀNH CHÍNH: Phạt tiền 10-20 triệu đồng. KỶ LUẬT: Cảnh cáo. HÌNH SỰ: BLHS Điều 289 (Tội đánh cắp thông tin) nếu rò rỉ bí mật, phạt tù 01-12 năm.",
		}
	default:
		return LegalAssessment{
			Severity:    "MEDIUM",
			Description: fmt.Sprintf("Cổng %d (%s) đang MỞ. Cần đánh giá thêm mức độ an toàn.", port, name),
			LegalBasis:  "Luật An ninh mạng 2018 Điều 28.",
			Liability:   "HÀNH CHÍNH: Phạt tiền 5-10 triệu đồng.",
		}
	}
}

// AssessMalwareLegal đánh giá pháp lý cho mã độc/keylogger
func AssessMalwareLegal(name string, hasC2, running bool, dangerLevel string) LegalAssessment {
	if hasC2 && running {
		return LegalAssessment{
			Severity:    "CRITICAL",
			Description: fmt.Sprintf("Tiến trình '%s' ĐANG CHẠY trong RAM và có kết nối tới C2 server (Command & Control). Đây là mã độc gián điệp đang hoạt động, thu thập và gửi dữ liệu ra ngoài. Có thể làm rò rỉ bí mật nhà nước, mật khẩu, văn bản mật.", name),
			LegalBasis:  "BLHS 2015 (sửa đổi 2017) Điều 288 (Tội vi phạm quy định về bảo mật thông tin - phạt tù 01-07 năm), Điều 289 (Tội đánh cắp thông tin - phạt tù 01-12 năm), Điều 290 (Tội phá rối hoạt động máy tính); Luật An ninh mạng 2018 Điều 18, 28; Luật Bảo vệ bí mật nhà nước 2018.",
			Liability:   "HÌNH SỰ: Phạt tù 01-07 năm (Điều 288) hoặc 01-12 năm (Điều 289) nếu rò rỉ bí mật nhà nước. KỶ LUẬT: Sa thải. Cách chức. Bắt giữ và báo cơ quan an ninh.",
		}
	}
	if running {
		return LegalAssessment{
			Severity:    "HIGH",
			Description: fmt.Sprintf("Tiến trình '%s' ĐANG CHẠY trong RAM. Có thể đang ghi lại thao tác người dùng hoặc chờ lệnh từ C2 server.", name),
			LegalBasis:  "BLHS 2015 Điều 288, 290; Luật An ninh mạng 2018 Điều 28.",
			Liability:   "HÌNH SỰ: Phạt tù 01-07 năm (Điều 288). KỶ LUẬT: Cảnh cáo, cách chức nếu để lọt mã độc vào máy cơ quan.",
		}
	}
	return LegalAssessment{
		Severity:    "MEDIUM",
		Description: fmt.Sprintf("Phát hiện file mã độc '%s' nhưng tiến trình không chạy. Cần xoá file và kiểm tra dấu vết hoạt động trước đó.", name),
		LegalBasis:  "BLHS 2015 Điều 288 (chuẩn bị phạm tội); Luật An ninh mạng 2018 Điều 28.",
		Liability:   "HÀNH CHÍNH: Phạt tiền 20-50 triệu đồng (Nghị định 15/2020/NĐ-CP). KỶ LUẬT: Cảnh cáo.",
	}
}

// AssessBadUSBLegal đánh giá pháp lý cho BadUSB
func AssessBadUSBLegal(vendorModel, vidpid string) LegalAssessment {
	return LegalAssessment{
		Severity:    "CRITICAL",
		Description: fmt.Sprintf("Thiết bị BadUSB '%s' (VID/PID: %s) có khả năng giả lập bàn phím tấn công tự động (Keystroke Injection). Khi cắm vào máy, nó có thể tự gõ lệnh PowerShell, cài backdoor, mở cổng reverse shell, download mã độc. Đây là phương pháp tấn công vật lý cực kỳ nguy hiểm.", vendorModel, vidpid),
		LegalBasis:  "BLHS 2015 Điều 290 (Tội phá rối hoạt động máy tính - phạt tù 01-07 năm); Luật An ninh mạng 2018 Điều 18, 28; Quy định của Ban Cơ yếu về quản lý thiết bị ngoại vi trong cơ quan nhà nước; Luật Cơ yếu về bảo mật.",
		Liability:   "HÌNH SỰ: Phạt tù 01-07 năm (Điều 290 BLHS). Thu giữ thiết bị. Báo cáo Ban Cơ yếu. Khởi tố nếu là đối tượng ngoại giao hoặc do người ngoài cắm.",
	}
}

// LegalSectionHTML trả về HTML cho phần đánh giá pháp lý trong báo cáo
func LegalSectionHTML(assessment LegalAssessment) string {
	severityColor := "#888"
	switch assessment.Severity {
	case "CRITICAL":
		severityColor = "#8B0000"
	case "HIGH":
		severityColor = "#dc143c"
	case "MEDIUM":
		severityColor = "#b8860b"
	case "AN TOÀN":
		severityColor = "#2d7a3e"
	}
	return fmt.Sprintf(`
		<div style="margin-top:8px;padding:8px;background:#FAF5E6;border-left:4px solid %s;">
			<p><strong style="color:%s;">Mức độ nghiêm trọng: %s</strong></p>
			<p style="margin:4px 0;"><strong>Đánh giá nguy cơ:</strong> %s</p>
			<p style="margin:4px 0;"><strong>Căn cứ pháp lý:</strong> %s</p>
			<p style="margin:4px 0;"><strong>Hình thức xử lý:</strong> %s</p>
		</div>
	`, severityColor, severityColor, assessment.Severity, assessment.Description, assessment.LegalBasis, assessment.Liability)
}

// LegalSectionPlain trả về văn bản plain text cho .docx (Word)
func LegalSectionPlain(assessment LegalAssessment) string {
	return fmt.Sprintf("Mức độ nghiêm trọng: %s\nĐánh giá nguy cơ: %s\nCăn cứ pháp lý: %s\nHình thức xử lý: %s",
		assessment.Severity, assessment.Description, assessment.LegalBasis, assessment.Liability)
}
