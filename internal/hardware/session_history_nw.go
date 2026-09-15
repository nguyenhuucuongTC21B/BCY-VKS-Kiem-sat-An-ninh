//go:build !windows

package hardware

// scanPeripheralSessions stub cho non-Windows (dev/demo)
// Trả về map rỗng để build pass
func scanPeripheralSessions() map[string][]ConnectSession {
	return map[string][]ConnectSession{}
}
