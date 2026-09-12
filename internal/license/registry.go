//go:build windows

package license

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// getOSLicenseMetadata đọc Product ID, Licensing Channel, BIOS OEM Key
// từ các nhánh Registry Windows.
func getOSLicenseMetadata() LicenseRecord {
	rec := LicenseRecord{
		SoftwareName: "Microsoft Windows",
		Legal:        true, // mặc định hợp pháp đến khi phát hiện bằng chứng ngược
	}

	// 1. Tên + phiên bản hệ điều hành
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE|registry.WOW64_64KEY); err == nil {
		if v, _, err := k.GetStringValue("ProductName"); err == nil {
			rec.SoftwareName = v
		}
		if v, _, err := k.GetStringValue("DisplayVersion"); err == nil {
			rec.Version = v
		} else if v, _, err := k.GetStringValue("ReleaseId"); err == nil {
			rec.Version = v
		}
		if v, _, err := k.GetStringValue("ProductId"); err == nil {
			rec.ProductID = v
		}
		if v, _, err := k.GetStringValue("BuildBranch"); err == nil {
			rec.Version += " build " + v
		}
		_ = k.Close()
	}

	// 2. Licensing Channel dựa vào ProductKeyChannel / DigitalProductId
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE|registry.WOW64_64KEY); err == nil {
		if v, _, err := k.GetStringValue("ProductKeyChannel"); err == nil && v != "" {
			rec.LicensingChannel = v // "OEM" | "Retail" | "Volume"
		} else if v, _, err := k.GetStringValue("EditionID"); err == nil {
			// Volume:KMSClient / Enterprise thường là Volume
			if strings.Contains(strings.ToLower(v), "volume") ||
				strings.Contains(strings.ToLower(v), "enterprise") {
				rec.LicensingChannel = "Volume"
			} else if strings.Contains(strings.ToLower(v), "oem") {
				rec.LicensingChannel = "OEM"
			} else {
				rec.LicensingChannel = "Retail"
			}
		}
		_ = k.Close()
	}

	// 3. BIOS OEM Key (DefaultProductKey)
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE|registry.WOW64_64KEY); err == nil {
		if v, _, err := k.GetStringValue("DefaultProductKey"); err == nil {
			rec.BIOSOEMKey = v
		} else if v, _, err := k.GetStringValue("InstallProductKey"); err == nil {
			rec.BIOSOEMKey = v
		}
		_ = k.Close()
	}

	// 4. Đọc thêm từ ProductOptions để xác nhận kênh
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\ProductOptions`,
		registry.QUERY_VALUE|registry.WOW64_64KEY); err == nil {
		if _, _, err := k.GetStringValue("ProductType"); err == nil {
			// OK - chứng tỏ đây là nhánh Windows có thật
		}
		_ = k.Close()
	}

	return rec
}

// getOfficeLicenseMetadata đọc siêu dữ liệu bản quyền Office (nếu cài)
func getOfficeLicenseMetadata() []LicenseRecord {
	var out []LicenseRecord
	// Office có thể cài ở nhiều vị trí registry tuỳ phiên bản (15.0, 16.0...)
	for _, ver := range []string{"16.0", "15.0", "14.0"} {
		key := `SOFTWARE\Microsoft\Office\` + ver + `\Common\ProductVersion`
		if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
			key, registry.QUERY_VALUE|registry.WOW64_64KEY); err == nil {
			rec := LicenseRecord{
				SoftwareName: "Microsoft Office",
				Version:      ver,
				Legal:        true,
			}
			if v, _, err := k.GetStringValue(""); err == nil {
				rec.Version = ver + " (" + v + ")"
			}
			_ = k.Close()
			out = append(out, rec)
		}
	}
	return out
}

// readKMSRegistryServer đọc KeyManagementServiceName để phát hiện server KMS lậu
func readKMSRegistryServer() string {
	paths := []string{
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion\SoftwareProtectionPlatform`,
		`SOFTWARE\Microsoft\OfficeSoftwareProtectionPlatform`,
	}
	for _, p := range paths {
		if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
			p, registry.QUERY_VALUE|registry.WOW64_64KEY); err == nil {
			if v, _, err := k.GetStringValue("KeyManagementServiceName"); err == nil && v != "" {
				_ = k.Close()
				return v
			}
			_ = k.Close()
		}
	}
	return ""
}

// ScanCrackToolsOnWindows duyệt CÁC THƯ MỤC NGHI VẤN và so khớp tên file với chữ ký crack.
// KHÔNG duyệt toàn bộ ổ đĩa (chậm 5-30 phút).
// Chỉ quét: Downloads, Public, Temp, Program Files, Program Files (x86),
// AppData\Local\Temp, AppData\Roaming, Desktop, Documents.
func ScanCrackToolsOnWindows(signatures []CrackSignature) []CrackHit {
	var hits []CrackHit

	// Danh sách thư mục nghi vấn chứa crack (theo profile user)
	candidateDirs := listSuspiciousDirs()

	// Danh sách thư mục con bị loại trừ (làm chậm mà không có crack)
	excludeDirs := map[string]bool{
		`c:\windows\system32`:              true,
		`c:\windows\syswow64`:              true,
		`c:\windows\winsxs`:                true,
		`c:\windows\assembly`:              true,
		`c:\windows\servicing`:             true,
		`c:\windows\installer`:            true,
		`c:\windows\softwaredistribution`:  true,
		`c:\windows\drivercache`:           true,
		`c:\windows\microsoft.net`:         true,
		`c:\$recycle.bin`:                  true,
		`c:\windows\temp`:                  true,
		`c:\windows\prefetch`:              true,
		`c:\windows\logs`:                   true,
		`c:\windows\debug`:                 true,
		`c:\windows\minidump`:               true,
	}

	for _, root := range candidateDirs {
		// Bỏ qua thư mục không tồn tại
		if _, err := os.Stat(root); err != nil {
			continue
		}

		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // bỏ qua file không có quyền đọc
			}
			if d == nil {
				return nil
			}

			// Nếu là thư mục, kiểm tra excludeDirs
			if d.IsDir() {
				lowerPath := strings.ToLower(path)
				if excludeDirs[lowerPath] {
					return filepath.SkipDir
				}
				return nil
			}

			// Bỏ qua file quá lớn (>100MB) để tránh chậm
			if info, err := d.Info(); err == nil {
				if info.Size() > 100*1024*1024 {
					return nil
				}
			}

			// Chỉ xét file thực thi và file nén
			name := strings.ToLower(d.Name())
			if !isExecutableOrArchive(name) {
				return nil
			}

			// So khớp với danh sách chữ ký
			for _, sig := range signatures {
				if matchCrack(name, sig) {
					hits = append(hits, CrackHit{
						Path:     path,
						Name:     d.Name(),
						ToolName: sig.ToolName,
						Version:  sig.Version,
					})
					break
				}
			}
			return nil
		})
	}
	return hits
}

// isExecutableOrArchive kiểm tra phần mở rộng có khả năng chứa crack không
func isExecutableOrArchive(name string) bool {
	extensions := []string{
		".exe", ".bat", ".cmd", ".ps1", ".vbs", ".msi", ".scr",
		".zip", ".rar", ".7z", ".iso", ".img",
		".dll", // amtlib.dll và các DLL crack phổ biến
	}
	for _, ext := range extensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// listSuspiciousDirs trả về danh sách thư mục nghi vấn (tương đối với user profile)
func listSuspiciousDirs() []string {
	userProfile := os.Getenv("USERPROFILE")
	public := os.Getenv("PUBLIC")
	if userProfile == "" {
		userProfile = `C:\Users\Default`
	}
	if public == "" {
		public = `C:\Users\Public`
	}

	dirs := []string{
		// Downloads của user
		filepath.Join(userProfile, "Downloads"),
		// Desktop của user
		filepath.Join(userProfile, "Desktop"),
		// Public Downloads (hay giấu crack)
		filepath.Join(public, "Downloads"),
		// Public Documents
		filepath.Join(public, "Documents"),
		// Temp của user
		filepath.Join(userProfile, "AppData", "Local", "Temp"),
		// Roaming (nơi keylogger hay cài)
		filepath.Join(userProfile, "AppData", "Roaming"),
		// Program Files
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		// ProgramData (hay cài chung)
		`C:\ProgramData`,
	}

	return dirs
}

// listDrives trả về danh sách gốc của các ổ đĩa (C:\, D:\...)
// (giữ lại để dùng cho chức năng quét toàn bộ ổ nếu user bật)
func listDrives() []string {
	var drives []string
	for c := 'A'; c <= 'Z'; c++ {
		p := string(c) + `:\`
		if _, err := os.Stat(p); err == nil {
			drives = append(drives, p)
		}
	}
	return drives
}

// matchCrack kiểm tra tên file có khớp với signature không
func matchCrack(name string, sig CrackSignature) bool {
	return contains(name, strings.ToLower(sig.Pattern))
}
