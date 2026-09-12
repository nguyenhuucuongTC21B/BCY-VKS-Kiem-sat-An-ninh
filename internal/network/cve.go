package network

import (
	"embed"
	"encoding/csv"
	"io"
	"io/fs"
	"strings"
)

// CVEMatch lưu kết quả đối chiếu phần mềm với CSDL CVE
type CVEMatch struct {
	CVEID   string
	CVSS    float64
	Summary string
}

// cveRow là một dòng trong cve_db.csv
type cveRow struct {
	Software string
	Version  string
	CVEID    string
	CVSS     float64
	Summary  string
}

// matchCVE đối chiếu phần mềm đã cài với CSDL CVE offline
// CSDL được nhúng trong assets/cve_db.csv
func matchCVE(fs embed.FS) CVEMatch {
	rows := loadCVEDB(fs)
	if len(rows) == 0 {
		return CVEMatch{}
	}

	// Lấy danh sách phần mềm đã cài
	installed := getInstalledSoftware()

	best := CVEMatch{}
	for _, sw := range installed {
		for _, row := range rows {
			if matchSoftware(sw, row.Software) && row.CVSS > best.CVSS {
				best = CVEMatch{
					CVEID:   row.CVEID,
					CVSS:    row.CVSS,
					Summary: row.Summary,
				}
			}
		}
	}
	return best
}

// loadCVEDB nạp file assets/cve_db.csv
func loadCVEDB(fsys embed.FS) []cveRow {
	b, err := fs.ReadFile(fsys, "assets/cve_db.csv")
	if err != nil {
		return defaultCVE()
	}
	return parseCSV(b)
}

func parseCSV(b []byte) []cveRow {
	r := csv.NewReader(strings.NewReader(string(b)))
	r.Comma = ','
	r.LazyQuotes = true
	var out []cveRow
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(rec) < 5 {
			continue
		}
		cvss := parseFloat(rec[3])
		out = append(out, cveRow{
			Software: strings.ToLower(rec[0]),
			Version:  rec[1],
			CVEID:    rec[2],
			CVSS:     cvss,
			Summary:  rec[4],
		})
	}
	if len(out) == 0 {
		return defaultCVE()
	}
	return out
}

// defaultCVE fallback khi không đọc được file
func defaultCVE() []cveRow {
	return []cveRow{
		{"windows smb", "2017", "CVE-2017-0144", 8.1, "EternalBlue - RCE qua SMBv1"},
		{"windows rdp", "2019", "CVE-2019-0708", 9.8, "BlueKeep - RCE qua RDP"},
		{"windows smb", "2020", "CVE-2020-0796", 10.0, "SMBGhost - RCE qua SMBv3"},
		{"openssh", "8.7", "CVE-2016-10009", 7.5, "Bypass xác thực SSH"},
		{"apache log4j", "2.14", "CVE-2021-44228", 10.0, "Log4Shell - RCE qua JNDI"},
	}
}

func matchSoftware(installed, target string) bool {
	return strings.Contains(installed, target)
}

// parseFloat đơn giản - chỉ hỗ trợ định dạng cơ bản
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	intPart := 0.0
	fracPart := 0.0
	fracMul := 0.1
	seenDot := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			seenDot = true
			continue
		}
		if c < '0' || c > '9' {
			break
		}
		if !seenDot {
			intPart = intPart*10 + float64(c-'0')
		} else {
			fracPart += float64(c-'0') * fracMul
			fracMul *= 0.1
		}
	}
	v := intPart + fracPart
	if neg {
		v = -v
	}
	return v
}
