//go:build !windows

package hardware

// friendlyNameForPNP stub cho non-Windows (dev mode dùng demo data)
func friendlyNameForPNP(devID string) string { return "" }

// scanCurrentPeripherals trả về dữ liệu demo trên môi trường non-Windows
// (có đầy đủ Sessions + RecentFilesSummary để dev thấy được tính năng Bảng 4)
func scanCurrentPeripherals() []PeripheralRec {
        return []PeripheralRec{
                {
                        DeviceType:  "USB Flash Drive",
                        VendorModel: "SanDisk Ultra Flair 64GB",
                        HardwareID:  "disk&ven_sandisk&prod_ultra_flair&rev_1.00\\4c530001234567890123&0",
                        VIDPID:      "VID_0781&PID_5590",
                        DriveLetter: "E:",
                        FirstPlug:   "2026-08-12 10:15:22",
                        LastPlug:    "2026-09-11 15:08:01",
                        PlugCount:   6,
                        RecentFilesSummary: "8 shortcut Recent: bao-cao-q3.docx, danh-sach-can-bo.xlsx, anh-hoi-nghi.zip | " +
                                "Gốc ổ E:: 18 mục, mới nhất: BAO-CAO-Q3.docx (11/09/2026 14:55)",
                        Sessions: []ConnectSession{
                                {Arrival: "2026-09-11 15:08:01", Removal: "", Duration: "—"},
                                {Arrival: "2026-09-10 08:23:44", Removal: "2026-09-10 09:05:12", Duration: "41 phút"},
                                {Arrival: "2026-09-05 14:02:10", Removal: "2026-09-05 14:47:39", Duration: "45 phút"},
                                {Arrival: "2026-08-29 09:11:00", Removal: "2026-08-29 11:32:15", Duration: "2 giờ 21 phút"},
                                {Arrival: "2026-08-20 16:40:33", Removal: "2026-08-20 16:44:58", Duration: "4 phút"},
                                {Arrival: "2026-08-12 10:15:22", Removal: "2026-08-12 12:02:41", Duration: "1 giờ 47 phút"},
                        },
                },
                {
                        DeviceType:  "USB Mouse",
                        VendorModel: "Logitech USB Optical Mouse",
                        HardwareID:  "5&1f2e3d4c&0&3",
                        VIDPID:      "VID_046D&PID_C077",
                        FirstPlug:   "2026-08-12 10:14:55",
                        LastPlug:    "2026-09-11 15:08:00",
                        PlugCount:   3,
                        Sessions: []ConnectSession{
                                {Arrival: "2026-09-11 15:08:00", Removal: "", Duration: "—"},
                                {Arrival: "2026-08-12 10:14:55", Removal: "2026-08-12 18:22:10", Duration: "8 giờ 7 phút"},
                                {Arrival: "2026-08-12 09:50:01", Removal: "2026-08-12 10:01:23", Duration: "11 phút"},
                        },
                },
        }
}

// scanPeripheralHistory trên non-Windows đã gộp sẵn trong demo ở trên
func scanPeripheralHistory() []PeripheralRec { return nil }

// buildRecentFilesSummary stub cho non-Windows
func buildRecentFilesSummary(driveLetter string) string { return "" }
