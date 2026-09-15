//go:build !windows

package main

import (
        "encoding/json"
        "strings"
        "testing"
)

// TestDemoScanAllJSONShape kiểm tra toàn bộ pipeline (demo mode) trả về
// JSON chứa đầy đủ các field mới mà frontend Bảng IV + V phụ thuộc:
//   - sessions + plug_count (lịch sử từng lần kết nối)
//   - recent_files_summary (Recent Files & Jump Lists)
//   - reason (lý do đánh dấu nghi của module mã độc)
func TestDemoScanAllJSONShape(t *testing.T) {
        app := NewApp()
        res, err := app.ScanAll()
        if err != nil {
                t.Fatalf("ScanAll lỗi: %v", err)
        }
        b, err := json.Marshal(res)
        if err != nil {
                t.Fatalf("Marshal lỗi: %v", err)
        }
        s := string(b)
        for _, key := range []string{
                `"sessions"`, `"plug_count"`, `"recent_files_summary"`,
                `"first_plug"`, `"last_plug"`, `"reason"`, `"stats"`,
        } {
                if !strings.Contains(s, key) {
                        t.Errorf("JSON thiếu field %s", key)
                }
        }

        // Bảng IV phải có ít nhất 1 thiết bị với sessions đồng bộ PlugCount
        if len(res.Bang4) == 0 {
                t.Fatal("Bang4 demo phải có dữ liệu")
        }
        for _, p := range res.Bang4 {
                if len(p.Sessions) > 0 && p.PlugCount < len(p.Sessions) {
                        t.Errorf("thiết bị %s: PlugCount=%d < số sessions=%d",
                                p.VendorModel, p.PlugCount, len(p.Sessions))
                }
        }
}
