//go:build !windows

package hardware

type Win32_NetworkAdapter struct {
	Name         string
	MACAddress   string
	Manufacturer string
	PNPDeviceID  string
	AdapterType  string
}

type Win32_PnPEntity struct {
	DeviceID    string
	Description string
	Status      string
	PNPDeviceID string
}

func scanInternalNICs() []HardwareRecord {
	return []HardwareRecord{
		{
			AdapterType:   "LAN",
			ConnectionPos: "Internal",
			DeviceName:    "Intel I219-V Gigabit",
			Serial:        "PCI\\VEN_8086&DEV_15BB",
			MAC:           "AA:BB:CC:11:22:33",
			DriverStatus:  "OK",
		},
		{
			AdapterType:   "Wi-Fi",
			ConnectionPos: "Internal (M.2)",
			DeviceName:    "Intel Wi-Fi 6E AX211",
			Serial:        "PCI\\VEN_8086&DEV_51F0",
			MAC:           "BB:CC:DD:22:33:44",
			DriverStatus:  "OK",
		},
	}
}

func scanExternalNICs() []HardwareRecord {
	return []HardwareRecord{
		{
			AdapterType:   "Wi-Fi",
			ConnectionPos: "External USB",
			DeviceName:    "TP-Link TL-WN725N",
			Serial:        "USB\\VID_2357&PID_0124",
			MAC:           "CC:DD:EE:33:44:55",
			DriverStatus:  "OK",
		},
	}
}
