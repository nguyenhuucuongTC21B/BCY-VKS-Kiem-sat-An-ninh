# Đóng Góp Cho BCY-VKS

Cảm ơn bạn đã quan tâm đóng góp cho dự án **BCY-VKS Kiểm Sát An Ninh**! 🎉

Đây là hướng dẫn giúp bạn đóng góp hiệu quả.

---

## 🚀 Bắt Đầu Nhanh

### 1. Fork Repository

Vào [repository chính](https://github.com/USER/BCY-VKS), click nút **Fork** ở góc trên phải.

### 2. Clone Fork Về Máy

```bash
git clone https://github.com/YOUR_USERNAME/BCY-VKS.git
cd BCY-VKS
git remote add upstream https://github.com/USER/BCY-VKS.git
```

### 3. Tạo Branch Mới

```bash
# Tạo branch mới từ main
git checkout -b feature/them-chu-ky-crack-moi

# Hoặc fix bug
git checkout -b fix/loi-quet-usb
```

### 4. Cài Đặt Môi Trường Dev

```bash
# Cài Go 1.22+ từ https://go.dev/dl/
# Cài Node.js 18+ từ https://nodejs.org/
# Cài Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Tải dependencies
go mod download

# Chạy dev mode với hot-reload
wails dev
```

### 5. Code → Commit → Push

```bash
git add .
git commit -m "feat: thêm 5 chữ ký crack mới (KMSAuto++, AACT...)"

# Push lên fork
git push origin feature/them-chu-ky-crack-moi
```

### 6. Tạo Pull Request

Vào GitHub fork của bạn → tab **Pull requests** → **New pull request** → Chọn branch của bạn → **Create pull request**.

---

## 📋 Quy Tắc Đóng Góp

### Phong Cách Code

- **Ngôn ngữ:** Code + comment bằng tiếng Việt có dấu (Unicode)
- **Naming:** Function/variable dùng tiếng Anh snake_case hoặc camelCase
- **Format:** Chạy `gofmt -w .` trước khi commit
- **Build tags:** File Windows-specific có `//go:build windows`, file non-Windows có `//go:build !windows`

### Convention Commit

Sử dụng format [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

| Type | Mô Tả |
|------|-------|
| `feat` | Tính năng mới |
| `fix` | Sửa lỗi |
| `docs` | Thay đổi documentation |
| `style` | Format code, không thay đổi logic |
| `refactor` | Refactor code, không thay đổi hành vi |
| `perf` | Cải thiện performance |
| `test` | Thêm/sửa tests |
| `chore` | Maintenance, dependencies... |

**Ví dụ:**
```
feat(license): thêm quét KMSAuto Lite v1.x

fix(network): sửa lỗi parse output Get-NetAdapter CSV
docs: cập nhật README với hướng dẫn build
```

### Phạm Vi Code

#### ✅ Nên Đóng Góp:
- Thêm chữ ký crack mới vào `assets/crack_signatures.txt`
- Thêm CVE mới vào `assets/cve_db.csv`
- Thêm VID/PID BadUSB vào `assets/badusb_vid_pid.txt`
- Sửa bug, tối ưu performance
- Cải thiện UI/UX (CSS, layout)
- Thêm tests
- Cải thiện documentation (README, code comments)
- Hỗ trợ i18n (tiếng Anh)

#### ❌ Không Đóng Góp:
- Thêm tính năng bypass antivirus (AV/EDR evasion)
- Thêm tính năng khai thác (exploit) thực sự — chỉ được phép làm mô phỏng
- Thêm tính năng phá hoại (destructive) ngoài wipe đã có sẵn
- Thêm telemetry/phone-home (phần mềm phải luôn offline)
- Thêm dependency có license Copyleft mạnh (GPL, AGPL)

---

## 🧪 Testing

### Test Các Module Backend

```bash
# Test license module
go test ./internal/license/... -v

# Test network module
go test ./internal/network/... -v

# Test toàn bộ
go test ./... -v -cover
```

### Test Build Trên Nhiều Môi Trường

- Windows 10 22H2 x64 ✅
- Windows 11 23H2 x64 ✅
- Windows Server 2022 ✅
- Cross-compile từ Ubuntu 22.04 (chỉ syntax check) ✅

---

## 📝 Báo Cáo Bug

Nếu bạn phát hiện bug, vui lòng [mở issue](https://github.com/USER/BCY-VKS/issues/new) với format:

```markdown
## 🐛 Bug Report

**Mô tả ngắn gọn:**
[1-2 câu mô tả bug]

**Các bước tái hiện:**
1. ...
2. ...
3. ...

**Kết quả mong đợi:**
[... ]

**Kết quả thực tế:**
[... ]

**Môi trường:**
- OS: Windows 11 23H2 x64
- BCY-VKS version: 1.0.0
- Build: stable release
- Quyền: Administrator

**Screenshot/Log:**
[Đính kèm nếu có]
```

---

## 💡 Đề Xuất Tính Năng

Mở issue với label `enhancement` để đề xuất tính năng mới. Mô tả rõ:

- **Vấn đề/Tình huống sử dụng:** Tại sao cần tính năng này?
- **Giải pháp đề xuất:** Bạn hình dung tính năng hoạt động thế nào?
- **Alternatives considered:** Có giải pháp khác không?

---

## 📜 Code Of Conduct

- 🤝 Tôn trọng mọi người, bất kể trình độ
- 📚 Giúp đỡ người mới học
- 🔒 Tôn trọng tính nhạy cảm của phần mềm (đặc biệt là module Anti-Forensics)
- 🚫 Không chấp nhận spam, quảng cáo, harassment

---

<div align="center">

Cảm ơn bạn đã đóng góp cho BCY-VKS! 🙏

Made with ❤️ by BCY-VKS Team

</div>
