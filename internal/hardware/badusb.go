package hardware

import (
        "embed"
        "io/fs"
        "strings"
)

// BadUSBVIDPID là cấu trúc cho file assets/badusb_vid_pid.txt
type BadUSBVIDPID struct {
        VID string
        PID string // có thể rỗng để match mọi PID
}

// embeddedFS alias để các file build tag windows và non-windows cùng dùng được
type embeddedFS = embed.FS

// loadBadUSBList nạp file assets/badusb_vid_pid.txt
// Mỗi dòng: VID,PID  (PID có thể rỗng)
func loadBadUSBList(fsys embeddedFS) []BadUSBVIDPID {
        b, err := fs.ReadFile(fsys, "assets/badusb_vid_pid.txt")
        if err != nil {
                return defaultBadUSB()
        }
        return parseBadUSBText(string(b))
}

// parseBadUSBText tách text file thành list VID/PID
func parseBadUSBText(s string) []BadUSBVIDPID {
        var out []BadUSBVIDPID
        for _, line := range strings.Split(s, "\n") {
                line = strings.TrimSpace(line)
                if line == "" || strings.HasPrefix(line, "#") {
                        continue
                }
                parts := strings.SplitN(line, ",", 2)
                if len(parts) < 1 {
                        continue
                }
                b := BadUSBVIDPID{
                        VID: strings.ToUpper(strings.TrimSpace(parts[0])),
                }
                if len(parts) == 2 {
                        b.PID = strings.ToUpper(strings.TrimSpace(parts[1]))
                }
                out = append(out, b)
        }
        if len(out) == 0 {
                return defaultBadUSB()
        }
        return out
}

// defaultBadUSB fallback
func defaultBadUSB() []BadUSBVIDPID {
        return []BadUSBVIDPID{
                {"03EB", ""},     // Atmel (USB Rubber Ducky gốc)
                {"16C0", ""},     // Teensy (BadUSB phổ biến)
                {"2341", ""},     // Arduino
                {"F00D", ""},     // Hak5
                {"1B3F", ""},     // Hobby Electronics
                {"20A0", ""},     // PJRC
        }
}
