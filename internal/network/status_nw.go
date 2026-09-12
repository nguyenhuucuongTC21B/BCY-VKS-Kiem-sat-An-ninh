//go:build !windows

package network

import (
	"net"
	"time"
)

type IPInfo struct {
	IP  string
	MAC string
	ISP string
}

func getInternetStatus() string {
	conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 3*time.Second)
	if err != nil {
		return "Disconnected (demo)"
	}
	_ = conn.Close()
	return "Connected (demo)"
}

func getIPInfo() IPInfo {
	return IPInfo{
		IP:  "127.0.0.1",
		MAC: "00:00:00:00:00:00",
		ISP: "Demo ISP",
	}
}

func getConnectionHistory() string {
	return "2026-09-10 08:23 | 192.168.1.105 | DHCP\n"
}
