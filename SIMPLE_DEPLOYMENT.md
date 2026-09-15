# 🚀 BCY-VKS - Hướng dẫn deploy đơn giản nhất

## 🎯 Mục tiêu

- Build .exe
- Copy vào bất kỳ máy nào
- Chạy được (bỏ qua cảnh báo SmartScreen mỗi lần)

**KHÔNG cần ký số, KHÔNG cần cài cert.**

---

## 📋 Quy trình 4 bước

### Bước 1: Download + Upload lên GitHub

```bash
# 1. Download file BCY-VKS-src.zip
# 2. Giải nén
unzip BCY-VKS-src.zip -d BCY-VKS-Kiem-sat-An-ninh
cd BCY-VKS-Kiem-sat-An-ninh

# 3. Tạo repo mới trên github.com (KHÔNG tick "Add README")
# 4. Push code lên:
git init
git branch -M main
git add .
git commit -m "Initial commit: BCY-VKS v1.0.0"
git remote add origin https://github.com/USERNAME/BCY-VKS-Kiem-sat-An-ninh.git
git push -u origin main
```

### Bước 2: GitHub Actions tự build .exe

- Vào tab **Actions** trên GitHub repo
- Đợi workflow "Build BCY-VKS" chạy xong (~3-5 phút)
- Click vào build run mới nhất
- Kéo xuống cuối → download artifact **`BCY-VKS-exe`**
- Giải nén → được file `BCY-VKS.exe` (~8 MB)

### Bước 3: Copy vào USB hoặc share network

```
USB Stick /
└── BCY-VKS.exe       (8 MB - file duy nhất)
```

### Bước 4: Chạy trên bất kỳ máy nào

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Double-click BCY-VKS.exe                                │
│                                                             │
│ 2. ⚠️ Popup xuất hiện:                                     │
│    "Windows protected your PC"                             │
│    Microsoft Defender SmartScreen prevented...             │
│    [Don't run]                                              │
│                                                             │
│ 3. Click "More info" (link dưới)                           │
│    → Hiện thêm nút [Run anyway]                            │
│                                                             │
│ 4. Click "Run anyway"                                       │
│    → UAC popup (nếu cần Admin)                             │
│    → Phần mềm mở cửa sổ → chạy bình thường                │
└─────────────────────────────────────────────────────────────┘
```

**Quan trọng:** Phải làm "More info → Run anyway" mỗi lần chạy (mỗi máy). Đây là chi phí của việc không ký số - chấp nhận được.

---

## 💻 Yêu cầu trên máy đích

### Bắt buộc
- Windows 10 1809+ / Windows 11 x64
- **WebView2 Runtime** (đã có sẵn trên Win11; Win10 tự cài qua Windows Update)
- 100 MB disk free
- 4 GB RAM

### Khuyến nghị (để scan đầy đủ)
- **Quyền Administrator** (Right-click → Run as Administrator)
  - Cần để đọc sâu Registry, Driver Store, Memory, Event Logs
  - Nếu chạy quyền User, một số bảng có thể thiếu dữ liệu

---

## 🎯 Tóm tắt

| Bước | Thời gian | Ai làm |
|------|-----------|---------|
| Download zip + upload GitHub | 5 phút | Bạn (1 lần) |
| GitHub Actions build | 3-5 phút | Tự động |
| Download .exe artifact | 30 giây | Bạn |
| Copy sang máy khác | Tùy | Bạn |
| Chạy + "More info → Run anyway" | 10 giây/máy | User cuối |

**Sau khi setup 1 lần:**
- Mỗi lần cần build lại → chỉ cần push code mới lên GitHub
- Build .exe mới → download → copy đi máy khác
- Lặp lại "More info → Run anyway" trên mỗi máy mới

---

## ⚡ Lợi ích của workflow này

✅ **Đơn giản**: không cần cài Go, Wails, signtool trên máy dev
✅ **Free**: GitHub Actions free cho repo public, không cần cert
✅ **Cross-machine**: copy .exe đi đâu cũng chạy được
✅ **Reproducible**: mỗi lần build ra cùng kết quả
✅ **No setup on target**: chỉ cần Windows 10/11 có WebView2

## ⚠️ Hạn chế

- **SmartScreen popup mỗi lần** (chấp nhận được)
- **Cần quyền Admin** để scan đầy đủ (cảnh báo nếu không)
- **Cần WebView2 Runtime** (đã có sẵn Win11, Win10 qua Windows Update)

---

## 🔧 Build local (nếu muốn build tay, không qua GitHub)

```cmd
:: 1. Cài Go 1.25+: https://go.dev/dl/
:: 2. Cài Wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest

:: 3. Vào thư mục BCY-VKS
cd BCY-VKS-Kiem-sat-An-ninh

:: 4. Build
.\build.bat

:: 5. File .exe ở: build\bin\BCY-VKS.exe
:: Copy đi máy khác, chạy "More info → Run anyway"
```

---

## ❓ Câu hỏi thường gặp

**Q: Có cần cài gì trên máy đích không?**
A: KHÔNG. Chỉ cần Windows 10/11 có WebView2 Runtime (đã có sẵn Win11).

**Q: File .exe có virus không?**
A: Không. Build từ source code trên GitHub, không có mã độc. Source code audit được.

**Q: Tại sao có popup SmartScreen?**
A: Vì file .exe chưa được ký số với cert từ CA. Windows không biết ai phát hành → cảnh báo.

**Q: Có cách nào bỏ popup SmartScreen vĩnh viễn không?**
A: CÓ - cần ký số. 2 cách:
   - Miễn phí: tự ký self-signed + cài cert vào từng máy (1 lần)
   - Trả phí: mua OV/EV cert từ DigiCert/Sectigo ($289+/năm)
   Đọc file `CERT_DEPLOYMENT_GUIDE.md` để biết chi tiết.

**Q: Mỗi lần chạy phải bấm "More info → Run anyway" đúng không?**
A: ĐÚNG. Đây là chi phí của việc không ký số. Mất khoảng 5 giây mỗi lần.

**Q: Nếu chạy nhiều lần trên cùng 1 máy, có phải bấm nhiều lần?**
A: CÓ, mỗi lần chạy .exe đều có popup. Đây là cơ chế bảo mật của Windows.
