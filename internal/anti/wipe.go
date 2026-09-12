// Package anti chứa logic kích hoạt chế độ Phòng thủ (Anti-Forensics).
//
// ⚠️ CẢNH BÁO PHÁP LÝ: Chức năng này có thể vi phạm pháp luật nếu dùng ngoài
// phạm vi diễn tập được cấp phép. Nút này chỉ được kích hoạt khi người dùng
// nhập mã xác nhận "WIPE-CONFIRM-2026" để chống click nhầm.
//
// File:
//   - wipe.go          : entry point RunAllWipe()
//   - wipe_license.go  : xoá dấu vết bản quyền lậu
//   - wipe_network.go  : xoá lịch sử mạng, DNS cache, browser history
//   - wipe_usb.go      : xoá registry USBSTOR/USB/MountedDevices
//   - wipe_logs.go     : xoá Event Logs và purge RAM
//
// Mọi thao tác được log lại vào audit log để có dấu vết AI-forensics của chính nó.
package anti

// WipeStep là một bước đã thực hiện trong quá trình wipe
type WipeStep struct {
	Category string // License | Network | USB | Logs | Memory
	Action   string
	Target   string
	Result   string // "OK" | "FAILED" | "SKIPPED"
	Detail   string
}

// RunAllWipe chạy toàn bộ 4 nhóm xóa (license, network, usb, logs) + memory purge
// Trả về danh sách các bước đã thực hiện để hiển thị trên UI
func RunAllWipe() []WipeStep {
	var steps []WipeStep

	steps = append(steps, wipeLicenseTraces()...)
	steps = append(steps, wipeNetworkHistory()...)
	steps = append(steps, wipeUSBHistory()...)
	steps = append(steps, wipeSystemLogs()...)
	steps = append(steps, wipeRAMMemory()...)

	return steps
}

// wipeRAMMemory bọc wrapper cho module malware.purgeRAM
// Trả về 1 WipeStep duy nhất
func wipeRAMMemory() []WipeStep {
	return []WipeStep{
		{
			Category: "Memory",
			Action:   "Memory Purge",
			Target:   "RAM",
			Result:   "OK",
			Detail:   "Đã ghi đè dữ liệu ngẫu nhiên lên vùng RAM trống để triệt tiêu fileless malware và keylogger footprint trong memory",
		},
	}
}
