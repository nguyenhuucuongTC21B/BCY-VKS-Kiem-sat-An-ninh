//go:build windows

package hardware

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

// getWirelessProfiles chạy netsh wlan show profiles để lấy SSID đã từng kết nối
// Trả về chuỗi phân cách bằng phẩy
// Fallback: đọc registry NetworkList\Profiles nếu netsh fail
func getWirelessProfiles() string {
	cmd := exec.Command("netsh", "wlan", "show", "profiles")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return readWirelessProfilesFromRegistry()
	}
	result := parseWirelessProfiles(stdout.String())
	if result != "" {
		return result
	}
	// Fallback nếu netsh không trả SSID
	return readWirelessProfilesFromRegistry()
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

// readWirelessProfilesFromRegistry fallback: đọc HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles
// Lấy ProfileName cho các profile có Type = 71 (Wi-Fi)
func readWirelessProfilesFromRegistry() string {
	var ssids []string
	key := `SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key,
		registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
	if err != nil {
		return ""
	}
	defer k.Close()

	names, err := k.ReadSubKeyNames(-1)
	if err != nil || len(names) == 0 {
		return ""
	}

	for _, n := range names {
		sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
			key+`\`+n, registry.QUERY_VALUE|registry.WOW64_64KEY)
		if err != nil {
			continue
		}
		// Đọc ProfileName
		profileName := ""
		if v, _, err := sub.GetStringValue("ProfileName"); err == nil {
			profileName = v
		}
		// Đọc Category - 0 = Wireless, 1 = Wired, 2 = Mobile Broadband
		// Nếu muốn chỉ lấy Wi-Fi thì check Category == 0
		// (nhưng để đầy đủ, lấy tất cả)
		_ = sub.Close()
		if profileName != "" {
			ssids = append(ssids, profileName)
		}
	}
	return strings.Join(ssids, ",")
}

// getWiFiLastConnect trả về thời gian kết nối Wi-Fi gần nhất (chuỗi RFC3339)
// Đọc từ NetworkList\Profiles\DateLastConnected
func getWiFiLastConnect() string {
	// Phương án 1: Đọc DateLastConnected từ NetworkList registry
	lastConnect := readLastConnectFromNetworkList()
	if lastConnect != "" {
		return lastConnect
	}

	// Phương án 2: Fallback netsh wlan show interfaces
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}
	return parseLastConnect(stdout.String())
}

// readLastConnectFromNetworkList đọc DateLastConnected từ NetworkList\Profiles
// Trả về thời gian gần nhất (RFC3339)
func readLastConnectFromNetworkList() string {
	key := `SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key,
		registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
	if err != nil {
		return ""
	}
	defer k.Close()

	names, _ := k.ReadSubKeyNames(-1)
	var latestTime time.Time

	for _, n := range names {
		sub, err := registry.OpenKey(registry.LOCAL_MACHINE,
			key+`\`+n, registry.QUERY_VALUE|registry.WOW64_64KEY)
		if err != nil {
			continue
		}
		// DateLastConnected là FILETIME 8 bytes
		v, _, err := sub.GetBinaryValue("DateLastConnected")
		_ = sub.Close()
		if err != nil || len(v) < 8 {
			continue
		}
		// Parse FILETIME (little-endian 8 bytes)
		ft := uint64(v[0]) | uint64(v[1])<<8 | uint64(v[2])<<16 | uint64(v[3])<<24 |
			uint64(v[4])<<32 | uint64(v[5])<<40 | uint64(v[6])<<48 | uint64(v[7])<<56
		if ft == 0 {
			continue
		}
		// Convert FILETIME (100ns từ 1601) → Unix time (ns từ 1970)
		// 116444736000000000 = số 100ns từ 1601 đến 1970
		unixNanos := int64(ft-116444736000000000) * 100
		if unixNanos < 0 {
			continue
		}
		t := time.Unix(0, unixNanos)
		if t.After(latestTime) {
			latestTime = t
		}
	}

	if latestTime.IsZero() {
		return ""
	}
	return latestTime.UTC().Format(time.RFC3339)
}

// parseLastConnect fallback nếu registry không có DateLastConnected
// Đọc netsh wlan show interfaces → tìm "Connection mode" + "Radio state"
func parseLastConnect(s string) string {
	// netsh wlan show interfaces không trực tiếp trả về thời gian kết nối
	// Trả về rỗng để frontend hiển thị "không rõ"
	// (Đã đọc từ registry ở readLastConnectFromNetworkList)
	_ = s
	return ""
}
