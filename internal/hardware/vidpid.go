// Package hardware - vidpid.go
// File pure-string ĐA NỀN TẢNG - không có build tag windows/non-windows
// Chứa các hàm xử lý VID/PID dùng cho cả 2 hệ, có thể test bằng Go test
package hardware

import "strings"

// extractVIDPID tách VID_xxxx và PID_xxxx từ DeviceID string
// Ví dụ input: "USB\VID_045E&PID_07C0\6&1a2b3c4d&0&1"
// → returns ("045E", "07C0")
func extractVIDPID(s string) (vid, pid string) {
        parts := strings.Split(s, "&")
        for _, p := range parts {
                upper := strings.ToUpper(p)
                if strings.HasPrefix(upper, "VID_") {
                        vid = p[4:]
                } else if strings.HasPrefix(upper, "PID_") {
                        pid = p[4:]
                }
        }
        return
}

// extractVIDPIDStr trả về chuỗi "VID_xxxx&PID_yyyy" hoặc "" nếu không tìm thấy
func extractVIDPIDStr(pnp string) string {
        vid, pid := extractVIDPID(pnp)
        if vid == "" {
                return ""
        }
        return "VID_" + vid + "&PID_" + pid
}

// extractSerialFromPNP tách serial từ PNPDeviceID có dạng VID_...&PID_...\<serial>
func extractSerialFromPNP(pnp string) string {
        idx := strings.LastIndex(pnp, `\`)
        if idx < 0 {
                return pnp
        }
        return pnp[idx+1:]
}

// friendlyNameForPNP đoán tên thân thiện từ PNPDeviceID
// (dùng khi registry không có FriendlyName)
func friendlyNameForPNP(pnp string) string {
        if pnp == "" {
                return ""
        }
        // Loại bỏ prefix USB\ hoặc USBSTOR\
        parts := strings.Split(pnp, `\`)
        if len(parts) < 2 {
                return pnp
        }
        class := strings.ToUpper(parts[0])
        rest := strings.Join(parts[1:], `\`)

        switch class {
        case "USBSTOR":
                // USBSTOR\Disk&Ven_SanDisk&Prod_Ultra_Flar&Rev_1.00\...
                // → "SanDisk Ultra Flar"
                subParts := strings.Split(rest, `&`)
                vendor, product := "", ""
                for _, sp := range subParts {
                        upper := strings.ToUpper(sp)
                        if strings.HasPrefix(upper, "VEN_") {
                                vendor = strings.TrimPrefix(sp, "Ven_")
                                vendor = strings.TrimPrefix(vendor, "VEN_")
                                vendor = strings.ReplaceAll(vendor, "_", " ")
                        } else if strings.HasPrefix(upper, "PROD_") {
                                product = strings.TrimPrefix(sp, "Prod_")
                                product = strings.TrimPrefix(product, "PROD_")
                                product = strings.ReplaceAll(product, "_", " ")
                        }
                }
                if vendor != "" && product != "" {
                        return vendor + " " + product
                }
                if product != "" {
                        return product
                }
                return rest
        case "USB":
                // USB\VID_045E&PID_07C0\...
                // → đoán từ VID
                vid, _ := extractVIDPID(rest)
                return guessVendorFromVID(vid)
        default:
                return pnp
        }
}

// guessVendorFromVID đoán tên hãng từ VID (bảng OUI phổ biến)
func guessVendorFromVID(vid string) string {
        knownVIDs := map[string]string{
                "045E": "Microsoft",
                "046D": "Logitech",
                "0461": "Primax",
                "05AC": "Apple",
                "05DC": "Lexar",
                "0781": "SanDisk",
                "0951": "Kingston",
                "0DC3": "ADATA",
                "13FE": "Kingston",
                "1F75": "Toshiba",
                "152D": "JMicron",
                "1E3D": "Silicon Power",
                "090C": "Solid State System",
                "0BDA": "Realtek",
                "8087": "Intel",
                "2357": "TP-Link",
                "148F": "Ralink",
                "14B2": "Hawking",
                "050D": "Belkin",
                "0846": "Netgear",
                "1690": "Asus",
                "1044": "Cherry",
                "04A9": "Canon",
                "04F9": "Brother",
                "04B8": "Epson",
                "0699": "Yamaha",
                "07A1": "Genius",
                "04E8": "Samsung",
                "0B05": "Asus",
                "04CC": "STMicroelectronics",
                "0C45": "Microdia",
                "05E3": "Genesys Logic",
                "174C": "ASMedia",
                "1B6C": "Corsair",
                "1C34": "Goodix",
                "0FD9": "Elgato",
        }
        if name, ok := knownVIDs[vid]; ok {
                return name
        }
        return "USB Device"
}

// itoaHelper chuyển int sang string (tránh import strconv cho file pure-string)
func itoaHelper(n int) string {
        if n == 0 {
                return "0"
        }
        neg := false
        if n < 0 {
                neg = true
                n = -n
        }
        buf := [16]byte{}
        i := len(buf)
        for n > 0 {
                i--
                buf[i] = byte('0' + n%10)
                n /= 10
        }
        if neg {
                i--
                buf[i] = '-'
        }
        return string(buf[i:])
}
