//go:build !windows

package hardware

func scanCurrentPeripherals() []PeripheralRec {
	return []PeripheralRec{
		{
			DeviceType:  "USB Flash Drive",
			VendorModel: "SanDisk Ultra Flair 64GB",
			HardwareID:  "045E07C0123456789ABCDEF",
			VIDPID:      "VID_0781&PID_5590",
			DriveLetter: "E:",
			FirstPlug:    "2026-08-12T10:15:22Z",
			LastPlug:     "2026-09-11T15:08:01Z",
			PlugCount:    14,
		},
	}
}

func scanPeripheralHistory() []PeripheralRec { return nil }
