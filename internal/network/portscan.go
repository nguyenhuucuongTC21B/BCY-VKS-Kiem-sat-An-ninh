//go:build windows

package network

import (
        "fmt"
        "net"
        "sync"
        "time"

        "golang.org/x/sys/windows/registry"
)

// dangerousPorts là các cổng có rủi ro cao khi mở
var dangerousPorts = []int{21, 22, 23, 25, 53, 80, 110, 135, 139, 143, 443, 445,
        587, 993, 995, 1433, 1521, 3306, 3389, 5432, 5900, 6379, 8080, 8443, 9200}

// scanDangerousPorts quét song song các cổng nguy hiểm trên localhost
// (chỉ quét danh sách nóng thay vì toàn bộ 65535 để tránh nặng)
// Production: chạy thêm full-range scan với timeout ngắn nếu người dùng bật tuỳ chọn.
func scanDangerousPorts() []int {
        host := "127.0.0.1"
        openCh := make(chan int, len(dangerousPorts))
        var wg sync.WaitGroup

        for _, p := range dangerousPorts {
                wg.Add(1)
                go func(port int) {
                        defer wg.Done()
                        if scanPort(host, port, 200*time.Millisecond) {
                                openCh <- port
                        }
                }(p)
        }
        wg.Wait()
        close(openCh)

        var opens []int
        for p := range openCh {
                opens = append(opens, p)
        }
        return opens
}

// scanPort thử TCP connect tới host:port với timeout
// Dùng net.JoinHostPort để xử lý đúng cả IPv4 và IPv6
// (IPv6 cần format "[::1]:80" thay vì "::1:80")
func scanPort(host string, port int, timeout time.Duration) bool {
        target := net.JoinHostPort(host, fmt.Sprintf("%d", port))
        conn, err := net.DialTimeout("tcp", target, timeout)
        if err != nil {
                return false
        }
        _ = conn.Close()
        return true
}

// SMB1Enabled kiểm tra SMBv1 có đang bật không
func SMB1Enabled() bool {
        k, err := registry.OpenKey(registry.LOCAL_MACHINE,
                `SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters`,
                registry.QUERY_VALUE|registry.WOW64_64KEY)
        if err != nil {
                return false
        }
        defer k.Close()
        v, _, err := k.GetIntegerValue("SMB1")
        if err != nil {
                return true // mặc định Windows bật nếu không có giá trị
        }
        return v == 1
}

// RDPEnabled kiểm tra RDP có đang bật không
// fDenyTSConnections = 1 nghĩa là RDP ĐÓNG, 0 là BẬT
func RDPEnabled() bool {
        k, err := registry.OpenKey(registry.LOCAL_MACHINE,
                `SYSTEM\CurrentControlSet\Control\Terminal Server`,
                registry.QUERY_VALUE|registry.WOW64_64KEY)
        if err != nil {
                return false
        }
        defer k.Close()
        v, _, err := k.GetIntegerValue("fDenyTSConnections")
        if err != nil {
                return false
        }
        return v == 0
}
