package license

import (
	"embed"
	"io/fs"
	"strings"
)

// CrackSignature là một chữ ký công cụ crack cần khớp
type CrackSignature struct {
	Pattern  string // chuỗi con cần khớp trong tên file (lowercase)
	ToolName string // KMSpico, KMSAuto, MAS, HWIDGen, AdobeZii...
	Version   string // phiên bản phổ biến (nếu biết)
}

// CrackHit là kết quả phát hiện được trên đĩa
type CrackHit struct {
	Path     string
	Name     string
	ToolName string
	Version  string
}

// loadCrackSignatures nạp danh sách chữ ký từ assets/crack_signatures.txt
// Mỗi dòng format: pattern|tool_name|version (version có thể rỗng)
func loadCrackSignatures(fsys embed.FS) []CrackSignature {
	b, err := fs.ReadFile(fsys, "assets/crack_signatures.txt")
	if err != nil {
		return defaultSignatures()
	}
	return parseSignatures(string(b))
}

func parseSignatures(s string) []CrackSignature {
	var out []CrackSignature
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 2 {
			continue
		}
		sig := CrackSignature{
			Pattern:  strings.ToLower(strings.TrimSpace(parts[0])),
			ToolName: strings.TrimSpace(parts[1]),
		}
		if len(parts) == 3 {
			sig.Version = strings.TrimSpace(parts[2])
		}
		out = append(out, sig)
	}
	if len(out) == 0 {
		return defaultSignatures()
	}
	return out
}

// defaultSignatures được dùng nếu không load được file
func defaultSignatures() []CrackSignature {
	return []CrackSignature{
		{"kmspico", "KMSpico", ""},
		{"kmsauto", "KMSAuto", ""},
		{"mas_aio", "Microsoft Activation Scripts (MAS)", ""},
		{"mas-aio", "Microsoft Activation Scripts (MAS)", ""},
		{"hwidgen", "HWID Gen", ""},
		{"hwid-gen", "HWID Gen", ""},
		{"adobezii", "Adobe Zii", ""},
		{"adobe-zii", "Adobe Zii", ""},
		{"amtlib.dll", "AMT Emulator (Adobe)", ""},
		{"universal adobe patcher", "Universal Adobe Patcher", ""},
	}
}

// ScanCrackTools hàm API chính mà scanner.go gọi
func ScanCrackTools(fsys embed.FS) []CrackHit {
	sigs := loadCrackSignatures(fsys)
	return ScanCrackToolsOnWindows(sigs)
}
