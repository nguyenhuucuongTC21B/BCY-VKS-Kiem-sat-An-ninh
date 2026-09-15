# Changelog

Tất cả thay đổi đáng chú ý của dự án **BCY-VKS Kiểm Sát An Ninh** sẽ được
ghi chép trong file này.

Format dựa trên [Keep a Changelog](https://keepachangelog.com/vi/1.1.0/),
phiên bản tuân thủ [Semantic Versioning](https://semver.org/lang/vi/spec/v2.0.0.html).

---

## [Unreleased]

### Planned
- Tối ưu tốc độ quét ổ đĩa lớn (>500GB)
- Hỗ trợ i18n (VN/EN)
- Export XML/JSON raw data cho SIEM
- Dark/Light theme toggle
- Auto-update signatures từ GitHub releases

---

## [1.1.0] - 2026-09-15

### 🎯 Trọng tâm: Hoàn thiện rà quét Thiết bị Ngoại vi (Bảng IV)

Phiên bản này khắc phục các tính năng chưa hoàn thiện được phát hiện khi
rà soát code, đặc biệt là việc **truy vết lịch sử từng lần kết nối thiết bị ngoại vi**.

### ✨ Added - Bảng IV (Ngoại vi)

- **Lịch sử TỪNG LẦN kết nối**: mỗi thiết bị giờ hiển thị danh sách session
  (thời điểm cắm, thời điểm rút, thời lượng từng lần) qua nút "📅 Xem N lần"
  trên Bảng IV và modal "Lịch sử kết nối từng lần"
  (`internal/hardware/usbsessions_win.go`, `usbsessions_parse.go`).
  Nguồn dữ liệu: Event Log `Microsoft-Windows-Kernel-PnP/Configuration`
  (EventID 400/410 = cắm, 420 = rút) → fallback `C:\Windows\INF\setupapi.dev.log`
  → fallback Properties registry LastArrival/LastRemoval.
- **PlugCount thật**: số lần cắm được đếm từ Event Log/SetupAPI thay vì luôn 0.
- **Recent Files & Jump Lists**: field `recent_files_summary` (frontend đã render
  từ trước nhưng backend chưa có) — đếm shortcut Recent trỏ tới ổ USB, liệt kê
  file/folder mới nhất ở gốc ổ, cảnh báo `autorun.inf`
  (`internal/hardware/recentfiles_win.go`).
- **Cột "Lịch sử kết nối"** mới trên Bảng IV + 9 class CSS còn thiếu
  (badge-red/green/orange/yellow/gray, cell-muted/mono/danger/warn).
- Báo cáo Popup/HTML/DOCX: phần 4 nay nêu rõ từng thiết bị dùng bao nhiêu lần,
  lần đầu, lần cuối và danh sách từng lần cắm/rút.

### 🔧 Fixed

- `readFirstArrival()` / `readLastRemoval()` — trước đây là stub trả về `""`;
  nay đọc thật FILETIME từ `Enum\<pnp>\Properties\{83da6326-...}\0064/0066`
  (internal/hardware/usb.go).
- `scanMountedDevices` — decode UTF-16 sai offset (4 thay vì 8) làm gán ký tự
  ổ đĩa thất bại; nay decode bền với nhiều offset + pattern scan `USBSTOR`.
- `parseLastConnect` (Wi-Fi) — stub `""`; nay trả "Đang kết nối (mới nhất)"
  hoặc max(DateLastConnected) từ NetworkList registry.
- `readWirelessProfilesFromRegistry` — stub `""`; nay đọc ProfileName từ
  NetworkList\Profiles làm fallback khi netsh lỗi.
- `scanGhostedDevices` — PNPDeviceID thiếu tiền tố `USB\` khiến tra Properties
  luôn thất bại.
- `scanAutoruns` (module Mã độc) — stub rỗng; nay đọc 3 nhánh Run
  (HKLM/HKLM-WOW64/HKCU), gắn cờ path lạ, exe không tồn tại, PowerShell -enc,
  tên mạo danh hệ thống (Type mới: "Persistence", field `reason`).
- `isLikelyKeylogFile` — luôn trả `true` (báo động giả); nay heuristic thật:
  token `[ENTER]/[CTRL]...` + API token `vk_/keydown...`, loại file binary.
- `extractVIDPID` — không tách được VID/PID từ PNP ID đầy đủ `USB\VID_...`;
  nay dùng tìm chuỗi con + token hex.
- Matching thiết bị khi merge current/history: từ khớp chính xác HardwareID
  mở rộng thành serial-tail + VID/PID (`sameDevice`).
- `PNPDeviceID` ghosted device chuẩn hoá (USB\VID_xxx&PID_xxx\serial).

### 🧪 Tests

- Thêm unit test cho: parse XML Kernel-PnP, dựng session, vnDuration,
  normalizePnpID, parse SetupAPI, sameDevice, merge, extractVIDPID,
  heuristic keylog, classifyAutorun (`go test ./internal/...`).

---

## [1.0.0] - 2026-09-12

### 🎉 Initial Release

Phiên bản đầu tiên của BCY-VKS - Offline Security Audit & Forensics Suite.

### ✨ Added - 5 Nhóm Quét Đầu Vào

#### Module 1: License Audit (`internal/license/`)
- ✅ Quét công cụ crack phổ biến: KMSpico, KMSAuto, MAS, HWIDGen, Adobe Zii
- ✅ Đọc Registry KeyManagementServiceName để phát hiện KMS server lậu
- ✅ Quét Task Scheduler cho task AutoActivation ngầm
- ✅ Trích xuất Product ID, Licensing Channel, BIOS OEM Key
- ✅ 60+ chữ ký crack trong `assets/crack_signatures.txt`

#### Module 2: Network & Pentest (`internal/network/`)
- ✅ Kiểm tra trạng thái Internet (TCP tới 8.8.8.8:53)
- ✅ Trích xuất IP/MAC/ISP hiện tại
- ✅ Đọc NetworkList profiles (lịch sử mạng)
- ✅ Quét 24 cổng TCP nguy hiểm (21, 445, 3389...)
- ✅ Đối chiếu phần mềm đã cài với 60+ CVE offline
- ✅ Giả lập pentest - đánh giá rủi ro không khai thác thực sự
- ✅ Kiểm tra SMBv1/RDP enabled từ Registry

#### Module 3: Hardware Audit (`internal/hardware/`)
- ✅ Quét Win32_NetworkAdapter qua PowerShell Get-NetAdapter
- ✅ Phát hiện card mạng gắn ngoài USB (TP-Link, D-Com 4G/5G)
- ✅ Đọc ghosted devices từ `Enum\USB`
- ✅ Trích xuất wireless profiles qua `netsh wlan show profiles`
- ✅ Đọc thời gian kết nối Wi-Fi gần nhất

#### Module 4: Peripheral Forensics (`internal/hardware/`)
- ✅ Đọc `HKLM\SYSTEM\CurrentControlSet\Enum\USBSTOR` cho storage
- ✅ Đọc `HKLM\SYSTEM\CurrentControlSet\Enum\USB` cho VID/PID
- ✅ Đọc `MountedDevices` để lấy ký tự ổ đĩa từng gán
- ️ Đối chiếu 15 BadUSB VID/PID (Atmel, Teensy, Arduino, Hak5)
- ✅ Merge current + history (FirstPlug, PlugCount)

#### Module 5: Malware & Memory Forensics (`internal/malware/`)
- ✅ Quét tiến trình nghi vấn qua gopsutil
- ✅ Phát hiện keylogger (SetWindowsHookEx pattern)
- ✅ Phát hiện tiến trình mạo danh (svch0st, lsass1...)
- ✅ Kiểm tra đường dẫn bất thường (Downloads/Public/Temp)
- ✅ Đọc Event Log Security cho Event ID 1102 (Log cleared)
- ✅ ReadProcessMemory để dò fileless malware (MZ header, shellcode)
- ✅ Phát hiện C2 server từ netstat connections

### ✨ Added - 5 Bảng Đầu Ra + Dashboard

- ✅ 8 thẻ thống kê real-time (crack, ports, CVE, adapters, USB, BadUSB, keylogger, duration)
- ✅ 5 bảng dữ liệu trong tabs chuyển nhanh
- ✅ Dark mode security dashboard theme
- ✅ Responsive layout (min 1100x700)

### ✨ Added - 2 Nút Lệnh Cao Cấp

#### Nút 1: Đề Xuất Xử Lý Khắc Phục (`internal/output/`)
- ✅ Chế độ Pop-up HTML nhanh (inline style trong modal)
- ✅ Xuất file `.html` hoàn chỉnh (in được, có @page A4)
- ✅ Xuất file `.docx` (HTML-as-DOCX format, Word 2010+ mở được)
- ✅ Lệnh Command sẵn sàng copy-paste (slmgr, netsh advfirewall, reg add, taskkill)
- ✅ Tự động lưu vào `Documents\BCY-VKS-Reports\` với timestamp

#### Nút 2: Phòng Thủ / Anti-Forensics (`internal/anti/`)
- ✅ Yêu cầu mã `WIPE-CONFIRM-2026` chống click nhầm
- ✅ Wipe license: `slmgr /upk /ckms /cpky /rearm`, xóa task scheduler KMS
- ✅ Wipe network: `ipconfig /flushdns`, xóa NetworkList, browser history
- ✅ Wipe USB: xóa registry `USBSTOR`, `USB`, `MountedDevices`
- ✅ Wipe logs: `wevtutil cl Security/System/Application`, `auditpol /clear`
- ✅ Xóa Prefetch + Recent Documents
- ✅ Reset WinHTTP proxy + Winsock
- ✅ Memory purge (ghi đè RAM trống)

### 🏗️ Added - Architecture & Build

- ✅ Single-EXE binary (~25-35 MB)
- ✅ Wails v2.9.1 framework
- ✅ `go:embed` cho assets (offline, không cần internet)
- ✅ Cross-platform codebase (build tags `//go:build windows`)
- ✅ `build.bat` script tự động cài Wails + build
- ✅ NSIS installer option (`-nsis` flag)
- ✅ `wails.json` config cho single-file setup

### 📚 Added - Documentation

- ✅ README.md hoàn chỉnh với badges, hướng dẫn build & cài đặt
- ✅ CONTRIBUTING.md cho contributor
- ✅ LICENSE với điều khoản pháp lý rõ ràng
- ✅ CHANGELOG.md (file này)

### 🔒 Security

- ✅ Mã xác nhận `WIPE-CONFIRM-2026` cho nút Anti-Forensics
- ✅ Audit log ghi lại mọi thao tác scan & wipe
- ✅ Không có telemetry, không gửi dữ liệu ra ngoài
- ✅ Không lưu dữ liệu quét ra ổ đĩa (chỉ trong RAM)

---

[Unreleased]: https://github.com/USER/BCY-VKS/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/USER/BCY-VKS/releases/tag/v1.0.0
