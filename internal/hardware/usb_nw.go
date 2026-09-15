//go:build !windows

package hardware

// scanCurrentPeripherals demo data cho non-Windows
// Bao gồm Sessions để demo UI hiển thị lịch sử kết nối
func scanCurrentPeripherals() []PeripheralRec {
	return []PeripheralRec{
		{
			DeviceType:  "USB Flash Drive",
			VendorModel: "SanDisk Ultra Flair 64GB",
			HardwareID:  "045E07C0123456789ABCDEF",
			VIDPID:      "VID_0781&PID_5590",
			DriveLetter: "E:",
			FirstPlug:   "2026-08-12T10:15:22Z",
			LastPlug:    "2026-09-11T15:08:01Z",
			PlugCount:   14,
			Sessions: []ConnectSession{
				{StartTime: "2026-08-12T10:15:22Z", EndTime: "2026-08-12T11:30:00Z", DriveLetter: "E:", Source: "Demo"},
				{StartTime: "2026-08-25T09:00:00Z", EndTime: "2026-08-25T09:45:00Z", DriveLetter: "F:", Source: "Demo"},
				{StartTime: "2026-09-01T14:00:00Z", EndTime: "2026-09-01T16:30:00Z", DriveLetter: "E:", Source: "Demo"},
				{StartTime: "2026-09-11T15:08:01Z", EndTime: "", DriveLetter: "E:", Source: "Demo"},
			},
		},
		{
			DeviceType:  "External HDD",
			VendorModel: "Kingston DataTraveler 32GB",
			HardwareID:  "09511234567890ABCDEF",
			VIDPID:      "VID_0951&PID_1666",
			DriveLetter: "F:",
			FirstPlug:   "2026-07-01T08:00:00Z",
			LastPlug:    "2026-09-10T20:15:00Z",
			PlugCount:   8,
			Sessions: []ConnectSession{
				{StartTime: "2026-07-01T08:00:00Z", EndTime: "2026-07-01T10:00:00Z", DriveLetter: "F:", Source: "Demo"},
				{StartTime: "2026-09-10T20:15:00Z", EndTime: "", DriveLetter: "F:", Source: "Demo"},
			},
		},
	}
}

// scanPeripheralHistory stub cho non-Windows
func scanPeripheralHistory() []PeripheralRec { return nil }

// friendlyNameForPNP stub cho non-Windows (đã có trong vidpid.go đa nền tảng)
// Không cần định nghĩa lại - vidpid.go đã có
