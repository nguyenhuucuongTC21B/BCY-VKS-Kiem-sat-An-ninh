//go:build windows

package hardware

import (
	"bytes"
	"os/exec"
	"strings"
)

// getWirelessProfiles chạy netsh wlan show profiles để lấy SSID đã từng kết nối
// Trả về chuỗi phân cách bằng phẩy
func getWirelessProfiles() string {
	cmd := exec.Command("netsh", "wlan", "show", "profiles")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return readWirelessProfilesFromRegistry()
	}
	return parseWirelessProfiles(stdout.String())
}

// parseWirelessProfiles tách output của netsh wlan show profiles
// Tìm các dòng có dạng: "    User profiles    :  SSID_NAME"
// hoặc "    All User Profile     : SSID_NAME"  (locale EN)
// hoặc locale VN: "    Tất cả người dùng     : SSID_NAME"
func parseWirelessProfiles(s string) string {
	var ssids []string
	for _, line := range strings.Split(s, "\n") {
		l := strings.TrimSpace(line)
		// Tìm dấu ":" rồi lấy phần sau
		idx := strings.Index(l, ":")
		if idx < 0 {
			continue
		}
		// Phải chứa chữ "profile" hoặc "hồ sơ" để đảm bảo đúng dòng
		ll := strings.ToLower(l)
		if !strings.Contains(ll, "profile") && !strings.Contains(ll, "hồ sơ") {
			continue
		}
		ssid := strings.TrimSpace(l[idx+1:])
		if ssid != "" {
			ssids = append(ssids, ssid)
		}
	}
	return strings.Join(ssids, ",")
}

// readWirelessProfilesFromRegistry backup nếu netsh không chạy được
func readWirelessProfilesFromRegistry() string {
	// HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles\<...>
	// Đọc từng subkey và lấy giá trị ProfileName
	// Code tóm lược - implement bằng registry package
	return ""
}

// getWiFiLastConnect trả về thời gian kết nối Wi-Fi gần nhất (chuỗi RFC3339)
func getWiFiLastConnect() string {
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}
	// Tìm dòng "    State            : connected"
	// và dòng "    SSID             : ..."
	// (chỉ ra Wi-Fi hiện đang kết nối)
	return parseLastConnect(stdout.String())
}

func parseLastConnect(s string) string {
	// Phức tạp hơn - cần đối chiếu với NetworkList
	// Tạm trả về rỗng để frontend hiển thị "không rõ"
	return ""
}
