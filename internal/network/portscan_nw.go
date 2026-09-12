//go:build !windows

package network

func scanDangerousPorts() []int { return nil }

func SMB1Enabled() bool { return false }
func RDPEnabled() bool { return false }
