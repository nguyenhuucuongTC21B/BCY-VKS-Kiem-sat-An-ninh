//go:build !windows

package anti

func wipeLicenseTraces() []WipeStep {
	return []WipeStep{
		{
			Category: "License",
			Action:   "License wipe (demo)",
			Target:   "n/a",
			Result:   "SKIPPED",
			Detail:   "Chạy trên Windows để kích hoạt xóa bản quyền lậu thực sự",
		},
	}
}

func wipeNetworkHistory() []WipeStep {
	return []WipeStep{
		{
			Category: "Network",
			Action:   "Network wipe (demo)",
			Target:   "n/a",
			Result:   "SKIPPED",
			Detail:   "Chạy trên Windows để kích hoạt xóa lịch sử mạng thực sự",
		},
	}
}

func wipeUSBHistory() []WipeStep {
	return []WipeStep{
		{
			Category: "USB",
			Action:   "USB wipe (demo)",
			Target:   "n/a",
			Result:   "SKIPPED",
			Detail:   "Chạy trên Windows để kích hoạt xóa registry USBSTOR thực sự",
		},
	}
}

func wipeSystemLogs() []WipeStep {
	return []WipeStep{
		{
			Category: "Logs",
			Action:   "Event log wipe (demo)",
			Target:   "n/a",
			Result:   "SKIPPED",
			Detail:   "Chạy trên Windows để xóa wevtutil Security/System/Application",
		},
	}
}
