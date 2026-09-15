//go:build !windows

package hardware

// scanDeviceSessions trên non-Windows không truy vết được Event Log -
// lịch sử session đã được nhúng sẵn trong demo data của usb_nw.go.
func scanDeviceSessions() map[string]*DeviceSessions {
        return map[string]*DeviceSessions{}
}
