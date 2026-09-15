//go:build windows

package hardware

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RecentFilesResult chứa thông tin Recent Files và Jump Lists
type RecentFilesResult struct {
	RecentFilesCount   int
	SuspiciousRecent   []string // Tài liệu có từ khóa "bí mật"
	ExternalDriveRefs  []string // Path tham chiếu ổ ngoài (D:, E:...)
	UNCPathRefs        []string // Path UNC (mạng)
	Summary            string
}

// Từ khóa nhận diện tài liệu bí mật
var secretKeywords = []string{
	"mat", "bi-mat", "bimat", "quyet-dinh", "quyetdinh",
	"cao-trang", "caotrang", "an-thu", "anthu", "noi-bo", "noibo",
	"tuyet-mat", "tuyetmat", "bang-ke", "bangke",
	"van-ban", "vanban", "lenh", "lenhcv",
	"chi-thi", "chithi", "quy-trinh", "quytrinh",
}

// scanRecentFilesAndJumpLists rà quét Recent Files và Jump Lists
func scanRecentFilesAndJumpLists() RecentFilesResult {
	out := RecentFilesResult{}

	up := os.Getenv("USERPROFILE")
	if up == "" {
		return out
	}

	recent := filepath.Join(up, `AppData\Roaming\Microsoft\Windows\Recent`)
	automatic := filepath.Join(recent, "AutomaticDestinations")
	custom := filepath.Join(recent, "CustomDestinations")

	// 1. Quét Recent Files (Shortcuts .lnk)
	if entries, err := os.ReadDir(recent); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			out.RecentFilesCount++
			name := strings.ToLower(e.Name())
			for _, kw := range secretKeywords {
				if strings.Contains(name, kw) {
					out.SuspiciousRecent = append(out.SuspiciousRecent, e.Name())
					break
				}
			}
		}
	}

	// 2. Quét Jump Lists tìm path ổ ngoài + UNC
	driveRe := regexp.MustCompile(`[A-Z]:\\[^\x00-\x1f"]+`)
	uncRe := regexp.MustCompile(`\\\\[A-Za-z0-9._-]+\\[^\x00-\x1f"]+`)
	seen := map[string]bool{}

	listJumpLists := func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			p := filepath.Join(dir, e.Name())
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			for _, m := range driveRe.FindAllString(string(data), -1) {
				// Chỉ lấy ổ ngoài (D-Z)
				if regexp.MustCompile(`^[D-Z]:\\`).MatchString(m) && !seen[m] {
					seen[m] = true
					out.ExternalDriveRefs = append(out.ExternalDriveRefs, m)
				}
			}
			for _, m := range uncRe.FindAllString(string(data), -1) {
				if !seen[m] {
					seen[m] = true
					out.UNCPathRefs = append(out.UNCPathRefs, m)
				}
			}
		}
	}

	listJumpLists(automatic)
	listJumpLists(custom)

	// Tổng hợp summary
	var sb strings.Builder
	sb.WriteString("Recent: " + itoaHelper(out.RecentFilesCount) + " mục")
	if len(out.SuspiciousRecent) > 0 {
		sb.WriteString(" | ⚠ " + itoaHelper(len(out.SuspiciousRecent)) + " tài liệu BÍ MẬT")
	}
	if len(out.ExternalDriveRefs) > 0 {
		sb.WriteString(" | " + itoaHelper(len(out.ExternalDriveRefs)) + " ref ổ ngoài")
	}
	if len(out.UNCPathRefs) > 0 {
		sb.WriteString(" | " + itoaHelper(len(out.UNCPathRefs)) + " ref UNC mạng")
	}
	out.Summary = sb.String()
	return out
}
