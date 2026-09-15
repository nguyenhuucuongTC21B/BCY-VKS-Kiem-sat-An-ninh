//go:build windows

package hardware

import (
        "bytes"
        "fmt"
        "os"
        "path/filepath"
        "sort"
        "strings"
        "time"
)

// recentShortcutMaxScan giới hạn số file .lnk quét trong thư mục Recent
const recentShortcutMaxScan = 500

// buildRecentFilesSummary dựng tóm tắt dấu vết tệp liên quan tới một
// thiết bị USB có ký tự ổ đĩa (driveLetter dạng "E:"):
//
//  1. Shortcut trong %APPDATA%\Microsoft\Windows\Recent trỏ tới ổ đĩa
//     (Recent Documents là nguồn jump-list forensics kinh điển)
//  2. File/folder mới nhất ở thư mục gốc của ổ đĩa
//  3. Cảnh báo autorun.inf nếu tồn tại ở gốc ổ (dấu hiệu lây lan qua USB)
func buildRecentFilesSummary(driveLetter string) string {
        if driveLetter == "" {
                return ""
        }
        var parts []string

        // ---- 1. Recent Documents (.lnk) trỏ tới ổ đĩa ----
        lnkDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Recent")
        if entries, err := os.ReadDir(lnkDir); err == nil {
                drivePrefix := strings.ToUpper(strings.TrimSuffix(driveLetter, ":")) + `:\`
                type lnkHit struct {
                        name string
                        mod  time.Time
                }
                var hits []lnkHit
                total := 0
                scanned := 0
                for _, e := range entries {
                        if scanned >= recentShortcutMaxScan {
                                break
                        }
                        if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".lnk") {
                                continue
                        }
                        scanned++
                        full := filepath.Join(lnkDir, e.Name())
                        data, err := os.ReadFile(full)
                        if err != nil {
                                continue
                        }
                        if lnkTargetsDrive(data, drivePrefix) {
                                total++
                                info, err := e.Info()
                                mod := time.Time{}
                                if err == nil {
                                        mod = info.ModTime()
                                }
                                hits = append(hits, lnkHit{
                                        name: strings.TrimSuffix(e.Name(), ".lnk"),
                                        mod:  mod,
                                })
                        }
                }
                if total > 0 {
                        sort.Slice(hits, func(i, j int) bool { return hits[i].mod.After(hits[j].mod) })
                        names := make([]string, 0, 3)
                        for k, h := range hits {
                                if k >= 3 {
                                        break
                                }
                                names = append(names, h.name)
                        }
                        parts = append(parts, fmt.Sprintf("%d shortcut Recent: %s",
                                total, strings.Join(names, ", ")))
                }
        }

        root := strings.ToUpper(strings.TrimSuffix(driveLetter, ":")) + `:\`

        // ---- 2. Cảnh báo autorun.inf (dấu hiệu malware lây qua USB) ----
        if _, err := os.Stat(filepath.Join(root, "autorun.inf")); err == nil {
                parts = append(parts, "PHÁT HIỆN autorun.inf ở gốc ổ (dấu hiệu malware lây lan qua USB)")
        }

        // ---- 3. File/folder mới nhất ở thư mục gốc ----
        if entries, err := os.ReadDir(root); err == nil {
                type fe struct {
                        name string
                        mod  time.Time
                }
                var files []fe
                for _, e := range entries {
                        if isRecycleOrSystemEntry(e.Name()) {
                                continue
                        }
                        info, err := e.Info()
                        mod := time.Time{}
                        if err == nil {
                                mod = info.ModTime()
                        }
                        files = append(files, fe{e.Name(), mod})
                }
                if len(files) > 0 {
                        sort.Slice(files, func(i, j int) bool { return files[i].mod.After(files[j].mod) })
                        names := make([]string, 0, 4)
                        for k, f := range files {
                                if k >= 4 {
                                        break
                                }
                                names = append(names, f.name)
                        }
                        modStr := ""
                        if !files[0].mod.IsZero() {
                                modStr = " (" + files[0].mod.Format("02/01/2006 15:04") + ")"
                        }
                        parts = append(parts, fmt.Sprintf("Gốc ổ %s: %d mục, mới nhất: %s%s",
                                driveLetter, len(files), strings.Join(names, ", "), modStr))
                }
        }

        return strings.Join(parts, " | ")
}

// lnkTargetsDrive kiểm tra thô nội dung file .lnk có tham chiếu tới ổ đĩa không.
// Format .lnk là binary phức tạp (Shell Link), nhưng đường drive xuất hiện
// dưới dạng ASCII hoặc UTF-16LE nên kiểm tra 2 pattern là đủ cho mục đích
// thống kê nhanh (không cần parse đầy đủ).
func lnkTargetsDrive(data []byte, drivePrefix string) bool {
        if len(data) == 0 {
                return false
        }
        // ASCII
        if bytes.Contains(data, []byte(drivePrefix)) {
                return true
        }
        // UTF-16LE
        u := make([]byte, 0, len(drivePrefix)*2)
        for _, ch := range []byte(drivePrefix) {
                u = append(u, ch, 0)
        }
        return bytes.Contains(data, u)
}

// isRecycleOrSystemEntry bỏ qua các thư mục hệ thống khi liệt kê gốc ổ
func isRecycleOrSystemEntry(name string) bool {
        n := strings.ToLower(name)
        switch n {
        case "system volume information", "recycler", "recycled", "$recycle.bin":
                return true
        }
        return false
}
