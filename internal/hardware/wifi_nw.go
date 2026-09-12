//go:build !windows

package hardware

func getWirelessProfiles() string {
	return "BCY-Office,Wifi-Cafe,Guest-NET"
}

func getWiFiLastConnect() string {
	return "2026-09-11T19:42:00Z"
}
