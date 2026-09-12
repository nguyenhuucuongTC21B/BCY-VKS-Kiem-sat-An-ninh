// Package network chứa logic dò quét lỗ hổng mạng, trạng thái Internet
// và tấn công kiểm thử (pentest) mô phỏng.
//
// Các file:
//   - status.go    : trạng thái Internet, DNS, Proxy, lịch sử kết nối
//   - portscan.go  : quét 1..65535 cổng TCP
//   - cve.go       : đối chiếu phần mềm đã cài với CSDL CVE offline
//   - logs.go      : trích xuất network logs từ Registry NetworkList
//
// API chính: ScanAll(embed.FS) ([]NetworkRecord, error)
package network

import "embed"

// NetworkRecord tương ứng BẢNG 2 trong đặc tả đầu ra.
type NetworkRecord struct {
        InternetStatus    string  `json:"internet_status"`
        CurrentIP         string  `json:"current_ip"`
        CurrentMAC        string  `json:"current_mac"`
        ISP               string  `json:"isp"`
        ConnectionHistory string  `json:"connection_history"`
        OpenPorts         string  `json:"open_ports"`
        CVEID             string  `json:"cve_id"`
        CVSSScore         float64 `json:"cvss_score"`
        ExploitResult     string  `json:"exploit_result"`
        Notes             string  `json:"notes"`
}

// ScanAll chạy toàn bộ nhóm 2 và trả về slice NetworkRecord.
// Một record duy nhất chứa toàn bộ thông tin (lịch sử, cổng mở, CVE, pentest).
// Mỗi bước có try/recover riêng để đảm bảo luôn trả về record (không bị trắng).
func ScanAll(fs embed.FS) ([]NetworkRecord, error) {
        rec := NetworkRecord{}

        // 1. Trạng thái Internet (luôn chạy được, dùng net stdlib)
        func() {
                defer func() { recover() }()
                rec.InternetStatus = getInternetStatus()
        }()

        // 2. IP/MAC/ISP hiện tại
        func() {
                defer func() { recover() }()
                ipInfo := getIPInfo()
                rec.CurrentIP = ipInfo.IP
                rec.CurrentMAC = ipInfo.MAC
                rec.ISP = ipInfo.ISP
        }()

        // 3. Lịch sử kết nối mạng (từ NetworkList registry)
        func() {
                defer func() { recover() }()
                rec.ConnectionHistory = getConnectionHistory()
        }()

        // 4. Quét cổng nguy hiểm
        var ports []int
        func() {
                defer func() { recover() }()
                ports = scanDangerousPorts()
        }()
        rec.OpenPorts = joinInts(ports, ",")

        // 5. Đối chiếu CVE
        var cveMatches CVEMatch
        func() {
                defer func() { recover() }()
                cveMatches = matchCVE(fs)
        }()
        rec.CVEID = cveMatches.CVEID
        rec.CVSSScore = cveMatches.CVSS
        rec.Notes = cveMatches.Summary

        // 6. Giả lập pentest
        func() {
                defer func() { recover() }()
                rec.ExploitResult = simulateExploit(ports, cveMatches)
        }()

        // Nếu toàn bộ đều rỗng (do scan không chạy được), ghi rõ
        if rec.InternetStatus == "" && rec.CurrentIP == "" && rec.OpenPorts == "" {
                rec.InternetStatus = "Không xác định được"
                rec.Notes = "Không thu thập được thông tin mạng - có thể do không có quyền Admin hoặc Windows shell không respond"
        }

        return []NetworkRecord{rec}, nil
}

// joinInts nối []int thành chuỗi phân cách bằng sep
func joinInts(nums []int, sep string) string {
        if len(nums) == 0 {
                return ""
        }
        out := ""
        for i, n := range nums {
                if i > 0 {
                        out += sep
                }
                out += itoa(n)
        }
        return out
}

// itoa chuyển int sang string (tránh import strconv để giữ cho file đơn giản)
func itoa(n int) string {
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
