//go:build windows

package license

import (
	"bytes"
	"os/exec"
	"strings"
)

// ScheduledTask chứa thông tin cơ bản của một tác vụ ngầm đáng ngờ
type ScheduledTask struct {
	Name       string
	Status     string
	Command    string
	Suspicious bool
	Reason     string
}

// scanScheduledTasks liệt kê tất cả tác vụ trong Task Scheduler và
// lọc ra các task có tên/command liên quan tới KMS/AutoActivation.
func scanScheduledTasks() []ScheduledTask {
	cmd := exec.Command("schtasks", "/query", "/fo", "CSV", "/v")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil
	}
	return parseScheduledTasksCSV(stdout.String())
}

// parseScheduledTasksCSV tách CSV output của schtasks và đánh dấu task khả nghi
func parseScheduledTasksCSV(s string) []ScheduledTask {
	lines := splitLines(s)
	if len(lines) < 2 {
		return nil
	}
	// Header: "TaskName","Next Run Time","Status","Last Run Time","Last Result",
	//         "Author","Task To Run","..."  (thứ tự có thể khác nhau theo locale)
	header := parseCSVRow(lines[0])
	colName := indexOf(header, "TaskName")
	colStatus := indexOf(header, "Status")
	colCommand := indexOf(header, "Task To Run")
	if colName < 0 || colCommand < 0 {
		return nil
	}

	var out []ScheduledTask
	for i := 1; i < len(lines); i++ {
		row := parseCSVRow(lines[i])
		if len(row) <= maxOf(colName, colStatus, colCommand) {
			continue
		}
		name := row[colName]
		status := ""
		if colStatus >= 0 {
			status = row[colStatus]
		}
		command := row[colCommand]
		if isSuspiciousTask(name, command) {
			out = append(out, ScheduledTask{
				Name:       name,
				Status:     status,
				Command:    command,
				Suspicious: true,
				Reason:     reasonFor(name, command),
			})
		}
	}
	return out
}

func isSuspiciousTask(name, command string) bool {
	combined := strings.ToLower(name + " " + command)
	keywords := []string{
		"kms",
		"autoactivation",
		"windowsactivation",
		"officeactivation",
		"rearm",
		"slmgr",
		"softwareprotection",
		"act 1year",
		"act_1year",
		"kmsauto",
		"kmspico",
		"mas_aio",
		"1clickactivation",
	}
	for _, kw := range keywords {
		if strings.Contains(combined, kw) {
			return true
		}
	}
	// Kiểm tra đường dẫn chạy file trong thư mục Downloads/Public cũng đáng ngờ
	if strings.Contains(strings.ToLower(command), `c:\users\public\`) ||
		strings.Contains(strings.ToLower(command), `c:\users\public\downloads\`) ||
		strings.Contains(strings.ToLower(command), `%temp%\`) {
		// Chỉ đánh dấu nếu tên task có vẻ liên quan activation
		if strings.Contains(combined, "act") || strings.Contains(combined, "kms") {
			return true
		}
	}
	return false
}

func reasonFor(name, command string) string {
	if strings.Contains(strings.ToLower(name), "kms") ||
		strings.Contains(strings.ToLower(command), "kms") {
		return "Chứa chuỗi 'KMS' - đặc trưng công cụ kích hoạt lậu"
	}
	if strings.Contains(strings.ToLower(name), "activation") ||
		strings.Contains(strings.ToLower(command), "activation") {
		return "Tên task gợi ý tự động gia hạn bản quyền"
	}
	return "Command trỏ tới vị trí thư mục Downloads/Public - bất thường"
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n")
}

func parseCSVRow(row string) []string {
	// schtasks output dùng dấu phẩy và quote theo RFC4180 đơn giản
	var out []string
	var cur strings.Builder
	inQuote := false
	for i := 0; i < len(row); i++ {
		c := row[i]
		switch {
		case c == '"':
			if inQuote && i+1 < len(row) && row[i+1] == '"' {
				cur.WriteByte('"')
				i++
			} else {
				inQuote = !inQuote
			}
		case c == ',' && !inQuote:
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	out = append(out, cur.String())
	return out
}

func indexOf(arr []string, key string) int {
	for i, v := range arr {
		if strings.EqualFold(v, key) {
			return i
		}
	}
	return -1
}

func maxOf(nums ...int) int {
	m := 0
	for _, n := range nums {
		if n > m {
			m = n
		}
	}
	return m
}
