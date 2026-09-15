<div align="center">

# 🔐 BCY-VKS KIỂM SÁT AN NINH

### Offline Security Audit & Forensics Suite for Windows

**Go + Wails Framework · Single-EXE Binary · Fully Offline**

---

![Status](https://img.shields.io/badge/Status-Stable-success?style=for-the-badge)
![Version](https://img.shields.io/badge/Version-1.1.0-blue?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2.9.1-E24329?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-Windows%2010%2F11-blue?style=for-the-badge&logo=windows&logoColor=white)
![License](https://img.shields.io/badge/License-Proprietary-red?style=for-the-badge)
![Mode](https://img.shields.io/badge/Mode-Offline-orange?style=for-the-badge)

</div>

---

## 📑 Mục Lục

- [🎯 Tổng Quan](#-tổng-quan)
- [✨ Tính Năng](#-tính-năng)
- [🏗️ Kiến Trúc Hệ Thống](#️-kiến-trúc-hệ-thống)
- [📁 Cấu Trúc Thư Mục](#-cấu-trúc-thư-mục)
- [🚀 Hướng Dẫn Upload Lên GitHub](#-hướng-dẫn-upload-lên-github)
- [📦 Hướng Dẫn Cài Đặt Từ GitHub](#-hướng-dẫn-cài-đặt-từ-github)
- [🛠️ Build Từ Mã Nguồn](#️-build-từ-mã-nguồn)
- [📖 Hướng Dẫn Sử Dụng](#-hướng-dẫn-sử-dụng)
- [🔧 Yêu Cầu Hệ Thống](#-yêu-cầu-hệ-thống)
- [❓ Xử Lý Sự Cố](#-xử-lý-sự-cố)
- [⚠️ Lưu Ý Pháp Lý](#️-lưu-ý-pháp-lý)
- [📄 Giấy Phép](#-giấy-phép)

---

## 🎯 Tổng Quan

**BCY-VKS** là phần mềm **kiểm sát an ninh mạng offline** chạy trên Windows, được đóng gói thành một file **`BCY-VKS.exe` duy nhất** (khoảng 25-35 MB), không cần cài đặt, không cần internet, không phụ thuộc DLL bên ngoài.

Phần mềm thực hiện **5 nhóm rà quét** song song, trả về **5 bảng kết quả chi tiết**, và cung cấp **2 nút lệnh cao cấp** để xử lý khắc phục hoặc kích hoạt chế độ Phòng thủ (Anti-Forensics).

> 💡 **Phiên bản đầu tiên** của phần mềm được phát triển bởi đội ngũ BCY-VKS Team với mục tiêu cung cấp công cụ giám định số miễn phí cho các tổ chức an ninh mạng tại Việt Nam.

---

## ✨ Tính Năng

### 🟢 Phần I — 5 Nhóm Quét Đầu Vào

| # | Nhóm Tính Năng | Mô Tả Ngắn | Output |
|---|----------------|------------|--------|
| 1 | 🛡️ **Kiểm tra Bản quyền & Tính hợp pháp** | Quét công cụ crack (KMSpico, KMSAuto, MAS, HWIDGen, Adobe Zii...), đọc Registry KMS, dò Task Scheduler ngầm, trích Product ID/OEM Key/Licensing Channel | **Bảng 1** |
| 2 | 🌐 **Dò quét Lỗ hổng, Mạng & Pentest** | Trạng thái Internet + IP/MAC/ISP hiện tại, lịch sử NetworkList, quét cổng TCP 1-65535, đối chiếu CVE offline, giả lập pentest | **Bảng 2** |
| 3 | 🖥️ **Kiểm kê Phần cứng Card mạng & Wifi** | Win32_NetworkAdapter cho NIC onboard, USB PnP entity cho card gắn ngoài, ghosted devices, wireless profiles (SSID) | **Bảng 3** |
| 4 | 🔌 **Trích xuất & Dựng lịch sử Thiết bị Ngoại vi** | Đọc USBSTOR + USB registry, trích VID/PID + serial, MountedDevices lấy ổ đĩa, đối chiếu BadUSB VID/PID, **truy vết lịch sử TỪNG lần cắm/rút + thời lượng** từ Event Log Kernel-PnP / SetupAPI, dấu vết Recent Files & Jump Lists | **Bảng 4** |
| 5 | 🦠 **Giám định Mã độc, Keylogger & Memory Forensics** | Quét tiến trình keylogger (SetWindowsHookEx), phát hiện RAT/Trojan/APT, ReadProcessMemory cho fileless, Event ID 1102 (log wipe) | **Bảng 5** |

### 🔴 Phần II — 5 Bảng Đầu Ra + 2 Nút Lệnh

**2 Nút Lệnh Cao Cấp** ở cuối action bar:

1. **🟢 Nút "Đề xuất, xử lý khắc phục"** — 3 chế độ xuất:
   - **Pop-up nhanh**: HTML inline hiện trong modal (xem ngay)
   - **Xuất HTML**: file `.html` hoàn chỉnh có thể in (lưu vào `Documents\BCY-VKS-Reports\`)
   - **Xuất Word**: file `.docx` (Word 2010+ mở được, có style đầy đủ)

2. **🔴 Nút "Phòng thủ (Anti-Forensics)"** — Yêu cầu mã `WIPE-CONFIRM-2026`:
   - Xóa registry KMS + chạy `slmgr /upk /ckms /cpky /rearm`
   - Xóa Task Scheduler ngầm (KMSAuto, AutoActivation, KMSpico...)
   - Flush DNS cache + xóa NetworkList profiles registry
   - Xóa history trình duyệt Chrome/Edge/Firefox
   - Xóa registry `USBSTOR`, `USB`, `MountedDevices` (toàn bộ lịch sử USB)
   - Xóa Wireless Profiles (SSID từng kết nối)
   - `wevtutil cl Security/System/Application` (clear Event Logs)
   - `auditpol /clear /all` (reset Audit Policy)
   - Xóa Prefetch + Recent Documents
   - **Purge RAM** (ghi đè dữ liệu ngẫu nhiên lên RAM trống)

---

## 🏗️ Kiến Trúc Hệ Thống

```
┌─────────────────────────────────────────────────────────────┐
│                     Wails Frontend (HTML/JS/CSS)             │
│  - Dashboard 8 thẻ thống kê                                  │
│  - 5 bảng dữ liệu (tab-switchable)                           │
│  - 2 nút lệnh + 3 modal pop-up                               │
└─────────────────────────┬─────────────────────────────────────┘
                          │ Go Binding (window.go.main.App.*)
┌─────────────────────────▼─────────────────────────────────────┐
│                     Go Backend (Wails)                       │
│  ┌──────────────┬──────────────┬──────────────────────────┐ │
│  │ License      │ Network &    │ Hardware & Peripheral    │ │
│  │ Audit        │ Pentest      │ Forensics                │ │
│  ├──────────────┼──────────────┼──────────────────────────┤ │
│  │ Malware &    │ Remediation  │ Anti-Forensics / Wipe    │ │
│  │ Memory       │ Generator    │ (có kiểm soát)           │ │
│  └──────────────┴──────────────┴──────────────────────────┘ │
│  - Windows Registry, WMI, Task Scheduler, Event Logs        │
│  - TCP Port Scan, Process Memory, File System               │
└─────────────────────────────────────────────────────────────┘
                          │
                 Biên dịch & đóng gói
                          ▼
               BCY-VKS.exe (single binary)
```

**Nguyên tắc thiết kế:**
- ✅ Backend Go chịu trách nhiệm toàn bộ logic thu thập, quét, phân tích
- ✅ Frontend Wails chỉ hiển thị và gọi các method backend qua `window.go.main.App.*`
- ✅ Toàn bộ dữ liệu cấu hình, chữ ký CVE, danh sách công cụ crack được nhúng vào binary qua `go:embed` để chạy hoàn toàn offline
- ✅ Sử dụng quyền Administrator khi chạy để đọc sâu Registry, Task Scheduler, Driver Store và Memory
- ✅ Cross-platform codebase — Build tags `//go:build windows` và `//go:build !windows` để hỗ trợ cả môi trường dev Linux/macOS

---

## 📁 Cấu Trúc Thư Mục

```
BCY-VKS/
├── main.go                     # Wails bindings + App struct
├── wails.json                  # Cấu hình Wails build
├── go.mod / go.sum             # Go module + dependencies
├── build.bat                   # Script build .exe cho Windows
├── README.md                   # File này
├── .gitignore
│
├── frontend/                   # Wails frontend (go:embed vào binary)
│   ├── index.html              # Layout 5 bảng + 2 nút + modal
│   ├── main.js                 # Controller: gọi backend, render bảng
│   └── style.css               # Dark mode security dashboard
│
├── assets/                     # Dữ liệu nhúng (offline, không cần download)
│   ├── crack_signatures.txt    # 60+ chữ ký công cụ crack
│   ├── cve_db.csv              # 60+ CVE + CVSS score
│   └── badusb_vid_pid.txt      # 15 VID/PID BadUSB
│
└── internal/                   # Backend Go packages
    ├── license/                # Module 1: Bản quyền
    │   ├── scanner.go          #   Entry ScanAll()
    │   ├── registry.go         #   Đọc Registry license (.windows)
    │   ├── registry_nw.go      #   Stub non-windows
    │   ├── crackdb.go          #   Load crack_signatures.txt
    │   ├── taskscheduler.go    #   Dò Task Scheduler (.windows)
    │   └── taskscheduler_nw.go #   Stub non-windows
    │
    ├── network/                # Module 2: Mạng & Pentest
    │   ├── status.go           #   Entry ScanAll() + IPInfo
    │   ├── status_win.go       #   Trạng thái Internet, ISP, NetworkList
    │   ├── status_nw.go        #   Stub
    │   ├── portscan.go         #   Quét cổng TCP song song
    │   ├── portscan_nw.go      #   Stub
    │   ├── cve.go              #   Đối chiếu CVE
    │   ├── cve_win.go          #   Lấy installed software
    │   ├── cve_nw.go           #   Stub
    │   └── logs.go             #   Giả lập pentest
    │
    ├── hardware/               # Module 3 + 4: Phần cứng & Ngoại vi
    │   ├── nic.go              #   Entry ScanNIC() + ScanPeripherals()
    │   ├── nic_win.go          #   Win32_NetworkAdapter WMI
    │   ├── nic_nw.go           #   Stub
    │   ├── wifi_win.go         #   netsh wlan show profiles
    │   ├── wifi_nw.go          #   Stub
    │   ├── driverstore.go     #   Ghosted devices từ Enum\USB
    │   ├── driverstore_nw.go  #   Stub
    │   ├── usb.go              #   USBSTOR + USB enumeration
    │   ├── usb_nw.go           #   Stub
    │   └── badusb.go           #   Load badusb_vid_pid.txt
    │
    ├── malware/                # Module 5: Giám định mã độc
    │   ├── processscan.go      #   Entry ScanAll()
    │   ├── processscan_win.go  #   Quét tiến trình gopsutil
    │   ├── processscan_nw.go   #   Stub
    │   ├── memory.go          #   ReadProcessMemory forensics
    │   ├── memory_nw.go        #   Stub
    │   ├── eventlog.go         #   wevtutil Event ID 1102/104
    │   └── eventlog_nw.go      #   Stub
    │
    ├── output/                # Sinh báo cáo đề xuất khắc phục
    │   ├── report.go           #   Entry GenerateRemediation()
    │   ├── popup.go            #   HTML pop-up nhanh (inline style)
    │   ├── html.go             #   HTML report đầy đủ (in được)
    │   └── docx.go             #   Word .docx (HTML-as-DOCX format)
    │
    └── anti/                  # Anti-Forensics / Wipe-Out mode
        ├── wipe.go             #   Entry RunAllWipe()
        ├── wipe_license.go     #   Xóa KMS, slmgr /upk, schtasks
        ├── wipe_network.go     #   Flush DNS, xóa NetworkList, browser
        ├── wipe_usb.go         #   Xóa USBSTOR/USB/MountedDevices
        ├── wipe_logs.go        #   wevtutil cl, auditpol, prefetch
        └── wipe_nw.go          #   Stubs cho non-windows
```

---

## 🚀 Hướng Dẫn Upload Lên GitHub

### Bước 1: Tạo Repository Trên GitHub

1. Đăng nhập vào [github.com](https://github.com)
2. Click nút **`+`** ở góc trên phải → chọn **New repository**
3. Điền thông tin:
   - **Repository name**: `BCY-VKS`
   - **Description**: `Offline Security Audit & Forensics Suite for Windows - Go + Wails`
   - **Visibility**: Chọn `Private` (vì phần mềm có tính năng nhạy cảm)
   - **Initialize**: ❌ KHÔNG tick "Add a README", "Add .gitignore", "Choose license" (đã có sẵn trong project)
4. Click **Create repository**

### Bước 2: Push Mã Nguồn Lên GitHub Từ Command Line

Mở **PowerShell** hoặc **Command Prompt** trong thư mục `BCY-VKS/`:

```bash
# 1. Khởi tạo Git repository local
git init

# 2. Thêm tất cả file vào staging
git add .

# 3. Commit đầu tiên
git commit -m "Initial commit: BCY-VKS v1.0.0 - Offline Security Audit Suite

- 5 nhóm quét: License, Network, Hardware, Peripheral, Malware
- 5 bảng kết quả + 2 nút lệnh (Khắc phục & Anti-Forensics)
- Single-EXE binary, fully offline, go:embed assets
- Build: Go 1.22 + Wails v2.9.1
- Cross-platform codebase (build tags windows/non-windows)"

# 4. Đổi tên branch chính thành 'main'
git branch -M main

# 5. Thêm remote origin (thay USER bằng username GitHub của bạn)
git remote add origin https://github.com/USER/BCY-VKS.git

# 6. Push lên GitHub
git push -u origin main
```

> 💡 **Lần đầu push**, GitHub sẽ hỏi username + Personal Access Token (PAT). Tạo PAT tại:
> `Settings → Developer settings → Personal access tokens → Tokens (classic) → Generate new token` với scope `repo`.

### Bước 3: Tạo Release Đính Kèm File .exe (Tùy Chọn)

Sau khi build xong file `BCY-VKS.exe`, bạn có thể tạo release để người dùng tải trực tiếp:

```bash
# Tag version
git tag -a v1.0.0 -m "BCY-VKS v1.0.0 - Initial Release"
git push origin v1.0.0
```

Vào GitHub → tab **Releases** → **Draft a new release** → Chọn tag `v1.0.0` → Đính kèm file `BCY-VKS.exe` và `BCY-VKS-Setup.exe`.

### Bước 4: Cấu Hình `.gitignore` (Đã Sẵn Sàng)

File `.gitignore` đã có sẵn trong project, sẽ tự động bỏ qua:
- `build/` (binary output)
- `*.exe`, `*.dll`, `*.so` (compiled binaries)
- `vendor/` (Go vendor directory)
- `.idea/`, `.vscode/` (IDE files)
- `wailsjs/`, `frontend/wailsjs/` (Wails generated files)

---

## 📦 Hướng Dẫn Cài Đặt Từ GitHub

### Phương Án 1: Tải Binary Sẵn (Dành Cho Người Dùng Cuối)

> ⚡ **Nhanh nhất** — Không cần build, chỉ tải về và chạy.

1. Vào trang repository: `https://github.com/USER/BCY-VKS/releases`
2. Tải file `BCY-VKS-Setup.exe` (NSIS installer, ~12-18 MB)
3. Right-click → **Run as Administrator**
4. Làm theo wizard cài đặt (mặc định cài vào `C:\Program Files\BCY-VKS\`)
5. Mở shortcut **BCY-VKS** trên Desktop → **Run as Administrator**
6. Bấm nút **"RÀ QUÉT TOÀN BỘ"** ở góc phải topbar

### Phương Án 2: Build Từ Mã Nguồn (Dành Cho Developer)

#### Bước 1: Cài Đặt Yêu Cầu Hệ Thống

Trước khi build, cài đặt các phần mềm sau trên máy Windows 10/11 x64:

| Phần Mềm | Phiên Bản | Link Tải |
|----------|-----------|----------|
| **Go** | 1.22+ | [https://go.dev/dl/](https://go.dev/dl/) |
| **Node.js** | 18 LTS+ | [https://nodejs.org/](https://nodejs.org/) |
| **Wails CLI** | v2.9+ | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| **Git** | 2.40+ | [https://git-scm.com/](https://git-scm.com/) |
| **WebView2** | Runtime | Có sẵn trên Win11, Win10 qua Windows Update |

#### Bước 2: Clone Repository

```bash
# Clone repository về máy
git clone https://github.com/USER/BCY-VKS.git
cd BCY-VKS
```

#### Bước 3: Tải Go Dependencies

```bash
# Tự động tải tất cả dependencies được khai báo trong go.mod
go mod download

# Hoặc chạy tidy để cập nhật go.sum
go mod tidy
```

#### Bước 4: Build Single-EXE

**Cách A — Dùng script build.bat (khuyến nghị):**
```cmd
build.bat
```

**Cách B — Dùng Wails CLI trực tiếp:**
```bash
# Build single exe + NSIS installer
wails build -platform windows/amd64 -clean -trimpath -nsis

# Build tối ưu size (loại bỏ debug symbols)
wails build -platform windows/amd64 -clean -trimpath -ldflags="-s -w" -nsis
```

#### Bước 5: Chạy Thử Trước Khi Build (Dev Mode)

```bash
# Chạy dev mode với hot-reload frontend
wails dev
# Sẽ mở cửa sổ WebView với debug console
# Mọi thay đổi ở frontend/ sẽ tự động refresh
```

#### Bước 6: Kết Quả Build

Sau khi build thành công, file output ở:

```
BCY-VKS/
└── build/
    └── bin/
        ├── BCY-VKS.exe                    # Single exe (25-35 MB)
        └── BCY-VKS-amd64-installer.exe   # NSIS installer (12-18 MB)
```

---

## 🛠️ Build Từ Mã Nguồn

### Quick Reference Build Commands

```bash
# === Khởi tạo môi trường dev lần đầu ===
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor  # Kiểm tra môi trường

# === Build production (single .exe) ===
wails build -platform windows/amd64 -clean -nsis

# === Build tối ưu size ===
wails build -platform windows/amd64 -clean -trimpath -ldflags="-s -w" -nsis

# === Dev mode (hot-reload) ===
wails dev

# === Cross-compile từ Linux/macOS (chỉ kiểm tra syntax, không tạo .exe chạy được) ===
GOOS=windows GOARCH=amd64 go build -o BCY-VKS.exe .
```

### Build Output

| File | Kích Thước | Mô Tả |
|------|-----------|-------|
| `BCY-VKS.exe` | 25-35 MB | Single executable, chạy được ngay không cần cài |
| `BCY-VKS-amd64-installer.exe` | 12-18 MB | NSIS installer, có Start Menu shortcut + uninstaller |

### Build Flags Quan Trọng

| Flag | Tác Dụng |
|------|----------|
| `-clean` | Xóa cache build cũ trước khi build |
| `-trimpath` | Loại bỏ đường dẫn tuyệt đối khỏi binary (security + reproducible build) |
| `-ldflags="-s -w"` | Loại bỏ debug symbols + DWARF table, giảm ~30% size |
| `-nsis` | Tạo thêm NSIS installer ngoài single .exe |
| `-platform windows/amd64` | Cross-compile target Windows x64 |

---

## 📖 Hướng Dẫn Sử Dụng

### Quy Trình Vận Hành Chuẩn

1. **🚀 Khởi động phần mềm**
   - Right-click `BCY-VKS.exe` → **Run as Administrator**
   - Cửa sổ chính mở với dashboard trống + nút "RÀ QUÉT TOÀN BỘ" ở góc phải

2. **🔍 Thực hiện quét**
   - Click nút **"RÀ QUÉT TOÀN BỘ"** (vàng-cam)
   - Chờ **5-30 giây** (tùy số file trên ổ đĩa, tốc độ RAM, số cổng mở)
   - **8 thẻ thống kê** cập nhật real-time ở đầu dashboard:
     - Bản quyền lậu · Cổng nguy hiểm · CVE CVSS≥7 · Card mạng
     - Thiết bị USB · BadUSB · Keylogger · Thời lượng quét

3. **📊 Xem kết quả 5 bảng**
   - Click các tab ở giữa giao diện để chuyển giữa 5 bảng
   - Mỗi bảng có tiêu đề rõ ràng + số record hiển thị

4. **📋 Sinh báo cáo đề xuất khắc phục**
   - **Pop-up nhanh**: Xem ngay HTML inline trong modal
   - **Xuất HTML**: Tạo file `.html` hoàn chỉnh có thể in
   - **Xuất Word**: Tạo file `.docx` với lệnh Command sẵn sàng copy-paste
   - File lưu tại: `Documents\BCY-VKS-Reports\BCY-VKS-Remediation-YYYYMMDD-HHMMSS.html`

5. **🛡️ (Diễn tập) Kích hoạt chế độ Phòng thủ**
   - Click nút **"⚠ PHÒNG THỦ (ANTI-FORENSICS)"** (đỏ)
   - Modal cảnh báo xuất hiện, liệt kê toàn bộ hành động sẽ thực hiện
   - Nhập mã `WIPE-CONFIRM-2026` để xác nhận (chống click nhầm)
   - Quá trình wipe chạy ~30 giây, hiển thị danh sách các bước đã thực hiện

### Lệnh CLI Hữu Ích

```bash
# Chạy phần mềm với quyền Admin từ command line
powershell -Command "Start-Process BCY-VKS.exe -Verb RunAs"

# Chạy silent mode (nếu cài qua NSIS installer)
BCY-VKS-Setup.exe /S

# Gỡ cài đặt silent
"C:\Program Files\BCY-VKS\Uninstall BCY-VKS.exe" /S
```

---

## 🔧 Yêu Cầu Hệ Thống

### Máy Dev Build .exe

| Yêu Cầu | Tối Thiểu | Khuyến Nghị |
|---------|-----------|-------------|
| OS | Windows 10 1809+ x64 | Windows 11 x64 |
| CPU | 2 cores x86-64 | 4 cores+ |
| RAM | 4 GB | 8 GB+ |
| Disk | 500 MB free | 1 GB free |
| Go | 1.22+ | 1.23+ |
| Node.js | 18 LTS | 20 LTS |
| Git | 2.40+ | 2.45+ |
| WebView2 Runtime | Bắt buộc | Có sẵn Win11 |

### Máy Chạy .exe (Production)

| Yêu Cầu | Tối Thiểu | Khuyến Nghị |
|---------|-----------|-------------|
| OS | Windows 10 1809+ x64 | Windows 11 x64 |
| CPU | 1 core x86-64 | 2 cores+ |
| RAM | 2 GB | 4 GB+ (cho memory scan) |
| Disk | 100 MB free | 500 MB free |
| WebView2 Runtime | Bắt buộc | Auto-cài qua Windows Update |
| Quyền | User | **Administrator** (bắt buộc cho 5 nhóm quét + nút Phòng thủ) |

---

## ❓ Xử Lý Sự Cố

### Build Lỗi: "could not read Username for github.com"

**Nguyên nhân:** Một số dependency indirect của Wails v2.9.1 bị dead-link trên GitHub.

**Giải pháp:**
```bash
# Cập nhật Wails lên phiên bản mới
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Hoặc dùng phiên bản cụ thể không có dep lỗi
go get github.com/wailsapp/wails/v2@v2.8.0
go mod tidy
wails build
```

### Lỗi: "Wails applications will not build without the correct build tags"

**Nguyên nhân:** Đang chạy binary build bằng `go build` thay vì `wails build`.

**Giải pháp:** Luôn dùng `wails build` để build, không dùng `go build` trực tiếp.

### Lỗi: "Access is denied" khi chạy .exe

**Nguyên nhân:** Phần mềm cần quyền Administrator để đọc Registry sâu + Memory.

**Giải pháp:** Right-click `BCY-VKS.exe` → **Run as Administrator**.

### Lỗi: WebView2 không khởi động

**Nguyên nhân:** Thiếu WebView2 Runtime.

**Giải pháp:** Tải và cài từ [Microsoft Edge WebView2](https://developer.microsoft.com/microsoft-edge/webview2/).

### Phần Mềm Báo "Disconnected" Mặc Dù Có Internet

**Nguyên nhân:** Tường lửa chặn kết nối tới `8.8.8.8:53` (Google DNS).

**Giải pháp:** Phần mềm chỉ dùng để kiểm tra trạng thái Internet cơ bản. Có thể bỏ qua cảnh báo này nếu các tính năng khác hoạt động bình thường.

### Phần Mềm Chậm Khi Quét

**Nguyên nhân:** Quét toàn bộ ổ đĩa C:\ có thể mất nhiều thời gian.

**Giải pháp:** 
- Bỏ qua ổ đĩa lớn (D:\, E:\) trong file `internal/license/registry.go` (hàm `listDrives`)
- Hoặc comment hàm `ScanCrackTools` trong `internal/license/scanner.go` nếu chỉ cần quét registry

### Build Size Quá Lớnh (> 50MB)

**Giải pháp:** Dùng flag tối ưu:
```bash
wails build -platform windows/amd64 -clean -trimpath -ldflags="-s -w" -nsis
```

Hoặc dùng UPX để nén binary (giảm ~40% size):
```bash
upx --best --lzma BCY-VKS.exe
```

---

## ⚠️ Lưu Ý Pháp Lý

<div align="center">

### 🔴 ĐỌC KỸ TRƯỚC KHI SỬ DỤNG 🔴

</div>

Phần mềm **BCY-VKS** bao gồm chức năng **Anti-Forensics** có khả năng **xóa dấu vết kỹ thuật số** (registry, logs, RAM). Tính năng này **có thể vi phạm pháp luật** nếu sử dụng ngoài phạm vi:

### ✅ ĐƯỢC PHÉP
- Diễn tập an ninh mạng được cấp phép bởi tổ chức có thẩm quyền
- Trên hệ thống cá nhân của bạn
- Trong phòng thí nghiệm an ninh mạng của trường đại học / công ty
- Trong bảo trì phần mềm của chính bạn

### ❌ KHÔNG ĐƯỢC PHÉP
- Xóa chứng cứ trong cuộc điều tra hình sự đang được tiến hành
- Cản trở hoạt động giám định số của cơ quan chức năng
- Trên hệ thống không thuộc quyền sở hữu/quản lý của bạn
- Xóa dấu vết hành vi phạm pháp hình sự khác

### Hậu Quả Pháp Lý Có Thể

Việc sử dụng ngoài phạm vi trên có thể vi phạm:
- **Bộ luật Hình sự Việt Nam** (Điều 292, Điều 388)
- **Luật An ninh mạng 2018**
- **Luật Cạnh tranh không lành mạnh**
- Luật pháp quốc gia nơi phần mềm được sử dụng

> ⚠️ **Trách nhiệm pháp lý hoàn toàn thuộc về người sử dụng.** Đội ngũ phát triển **không chịu trách nhiệm** với bất kỳ hậu quả pháp lý nào phát sinh từ việc sử dụng sai mục đích.

---

## 📄 Giấy Phép

```
BCY-VKS Kiểm Sát An Ninh
Copyright © 2026 BCY-VKS Team

Proprietary License - Internal Use Only

Phần mềm này được phân phối dưới dạng mã nguồn mở trên GitHub
nhưng KHÔNG ĐƯỢC PHÉP sử dụng cho mục đích thương mại mà không có
sự đồng ý bằng văn bản của tác giả.

Chức năng Anti-Forensics chỉ được phép sử dụng trong phạm vi diễn tập
an ninh được cấp phép bởi tổ chức có thẩm quyền.
```

---

## 🙏 Cảm Ơn

Dự án này sử dụng các thư viện mã nguồn mở sau:

| Thư Viện | Mục Đích | Giấy Phép |
|----------|----------|-----------|
| [Wails v2](https://wails.io) | Framework desktop app | MIT |
| [Go](https://go.dev) | Ngôn ngữ lập trình | BSD-3 |
| [gopsutil](https://github.com/shirou/gopsutil) | Process listing | MPL-2.0 |
| [go-ole](https://github.com/go-ole/go-ole) | Windows COM bridge | Apache-2.0 |
| [StackExchange/wmi](https://github.com/StackExchange/wmi) | WMI wrapper | MIT |

---

<div align="center">

**BCY-VKS — Single-EXE Offline Security Audit Suite**

Made with ❤️ by BCY-VKS Team · Vietnam 🇻🇳

Version 1.1.0 · September 2026

</div>
