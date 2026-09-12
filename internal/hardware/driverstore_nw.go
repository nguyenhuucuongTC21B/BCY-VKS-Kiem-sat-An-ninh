//go:build !windows

package hardware

type GhostedDevice struct {
	PNPDeviceID  string
	FriendlyName string
	LastArrived  string
}

func scanGhostedDevices() []GhostedDevice { return nil }
