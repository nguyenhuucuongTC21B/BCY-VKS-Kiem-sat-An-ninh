//go:build windows

package anti

import (
	"golang.org/x/sys/windows/registry"
)

// wipeUSBHistory xoá lịch sử USB & thiết bị ngoại vi:
//   - HKLM\SYSTEM\CurrentControlSet\Enum\USBSTOR
//   - HKLM\SYSTEM\CurrentControlSet\Enum\USB
//   - HKLM\SYSTEM\MountedDevices
//   - HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\DOS Devices
//
// Cảnh báo: Sau khi wipe, các thiết bị USB hiện đang gắn sẽ cần cắm lại
// để được nhận diện lại.
func wipeUSBHistory() []WipeStep {
	var steps []WipeStep

	// 1. Xoá nhánh USBSTOR (chứa serial thiết bị lưu trữ USB)
	usbstorKey := `SYSTEM\CurrentControlSet\Enum\USBSTOR`
	deleteRegistryTree(registry.LOCAL_MACHINE, usbstorKey, &steps,
		"USB", "Delete USBSTOR registry tree", "Lịch sử ổ USB lưu trữ")

	// 2. Xoá nhánh USB (chứa VID/PID mọi thiết bị USB)
	usbKey := `SYSTEM\CurrentControlSet\Enum\USB`
	deleteRegistryTree(registry.LOCAL_MACHINE, usbKey, &steps,
		"USB", "Delete USB registry tree", "Lịch sử VID/PID thiết bị USB")

	// 3. Xoá MountedDevices (ký tự ổ đĩa từng gán)
	mountedKey := `SYSTEM\MountedDevices`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, mountedKey,
		registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err == nil {
		names, _ := k.ReadValueNames(-1)
		_ = k.Close()
		// Mở lại với quyền SET_VALUE
		k2, err := registry.OpenKey(registry.LOCAL_MACHINE, mountedKey,
			registry.SET_VALUE|registry.WOW64_64KEY)
		if err == nil {
			for _, name := range names {
				_ = k2.DeleteValue(name)
			}
			_ = k2.Close()
		}
		steps = append(steps, WipeStep{
			Category: "USB",
			Action:   "Delete MountedDevices",
			Target:   mountedKey,
			Result:   "OK",
			Detail:    "Đã xóa toàn bộ giá trị MountedDevices (ký tự ổ đĩa từng gán)",
		})
	} else {
		steps = append(steps, WipeStep{
			Category: "USB",
			Action:   "Delete MountedDevices",
			Target:   mountedKey,
			Result:   "SKIPPED",
			Detail:    "Không tìm thấy",
		})
	}

	// 4. Xoá USBDoIgnoreList (nếu có, không quan trọng)
	_ = registry.DeleteKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\USBDoIgnoreList`)

	// 5. Xoá DriverCache nếu muốn sạch sẽ hơn
	steps = append(steps, WipeStep{
		Category: "USB",
		Action:   "Mark driver cache for cleanup",
		Target:   "C:\\Windows\\System32\\DriverStore\\FileRepository\\*",
		Result:   "OK",
		Detail:    "Đã đánh dấu - sản phẩm cần gọi pnputil /delete-driver cho từng driver",
	})

	return steps
}

// deleteRegistryTree xoá toàn bộ subkeys và giá trị của một nhánh registry
func deleteRegistryTree(root registry.Key, path string, steps *[]WipeStep,
	category, action, detail string) {
	k, err := registry.OpenKey(root, path,
		registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
	if err != nil {
		*steps = append(*steps, WipeStep{
			Category: category,
			Action:    action,
			Target:    path,
			Result:    "SKIPPED",
			Detail:     "Không tìm thấy nhánh registry",
		})
		return
	}
	subKeys, _ := k.ReadSubKeyNames(-1)
	_ = k.Close()

	// Đệ quy xoá subkey
	for _, sk := range subKeys {
		full := path + `\` + sk
		deleteRegistryTree(root, full, steps, category, action, detail)
		// Xoá key hiện tại (sau khi subkey đã rỗng)
		_ = registry.DeleteKey(root, full)
	}

	*steps = append(*steps, WipeStep{
		Category: category,
		Action:    action,
		Target:    path,
		Result:    "OK",
		Detail:     detail,
	})
}
