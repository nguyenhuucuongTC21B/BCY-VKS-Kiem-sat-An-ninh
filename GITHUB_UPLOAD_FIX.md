# 🚨 Xử lý lỗi "Looks like something went wrong" trên GitHub

## ❌ Vấn đề bạn gặp

Khi upload file zip qua **Web UI** (kéo thả file vào github.com), đôi khi GitHub báo:

> **"Looks like something went wrong! We track these errors automatically..."**

**Nguyên nhân:**
- Web UI giới hạn số file và kích thước upload (khuyến nghị <100 file/lần)
- Web UI hay timeout với nhiều file nhị phân (.ico, .png)
- GitHub server tạm thời quá tải
- Network giữa máy bạn và GitHub bị gián đoạn

**Trạng thái GitHub hiện tại:** ✅ "All Systems Operational" (đã verify lúc 14:44 UTC)

Vậy vấn đề là do **Web UI**, không phải file zip của chúng ta.

---

## ✅ Giải pháp 1: Upload qua COMMAND LINE (khuyến nghị)

**Lý do nên dùng command line:**
- Git tự động retry khi network fail
- Không giới hạn số file
- Hỗ trợ authentication tốt hơn (PAT, gh CLI)
- Có progress bar

### Bước 1: Cài Git (nếu chưa có)

Tải từ: https://git-scm.com/download/win
Hoặc: `winget install Git.Git`

### Bước 2: Tạo Personal Access Token (PAT)

PAT thay thế password khi push code:

1. Vào https://github.com/settings/tokens
2. Click **"Generate new token (classic)"**
3. Note: `BCY-VKS push token`
4. Expiration: 90 days
5. Scope: tick **`repo`** (Full control of private repositories)
6. Click **"Generate token"**
7. **Copy token** (chỉ hiện 1 lần!) - dạng `ghp_xxxxxxxxxxxxxxxxxxxxxxxx`

### Bước 3: Tạo repo mới trên GitHub

1. Vào https://github.com/new
2. Repository name: `BCY-VKS-Kiem-sat-An-ninh`
3. Visibility: **Private** (recommended)
4. **KHÔNG tick** "Add a README file"
5. **KHÔNG tick** ".gitignore"  
6. **KHÔNG tick** "Choose a license"
7. Click **"Create repository"**

### Bước 4: Upload qua command line

```bash
# 1. Giải nén file zip
unzip BCY-VKS-src.zip -d BCY-VKS-Kiem-sat-An-ninh
cd BCY-VKS-Kiem-sat-An-ninh

# 2. Khởi tạo git
git init
git branch -M main

# 3. Cấu hình user (chỉ 1 lần)
git config --global user.name "TEN_BAN"
git config --global user.email "email@example.com"

# 4. Add + commit tất cả file
git add .
git commit -m "Initial commit: BCY-VKS v1.0.0 - Offline Security Audit Suite"

# 5. Add remote (thay USERNAME bằng GitHub username của bạn)
git remote add origin https://github.com/USERNAME/BCY-VKS-Kiem-sat-An-ninh.git

# 6. Push
git push -u origin main
# Khi hỏi Username: nhập GitHub username
# Khi hỏi Password: PASTE PAT (ghp_xxxxx...) - KHÔNG paste password thật
```

### Bước 5: Nếu push lỗi, thử lại với --force

```bash
# Nếu lỗi do repo đã có history (dù không add README):
git push -u -f origin main
```

---

## ✅ Giải pháp 2: Dùng GitHub CLI (đơn giản nhất)

### Bước 1: Cài GitHub CLI

```powershell
winget install GitHub.cli
```

### Bước 2: Login

```powershell
gh auth login
# Chọn: GitHub.com
# Chọn: HTTPS
# Chọn: Y - authenticate Git with GitHub credentials
# Chọn: Login with web browser
# Mở browser, copy 1-time code, login
```

### Bước 3: Tạo repo + push 1 lệnh

```powershell
# Vào thư mục code đã giải nén
cd BCY-VKS-Kiem-sat-An-ninh

# Khởi tạo + tạo repo + push 1 lệnh
git init
git add .
git commit -m "Initial commit: BCY-VKS v1.0.0"
gh repo create BCY-VKS-Kiem-sat-An-ninh --private --source=. --push
```

---

## ✅ Giải pháp 3: Dùng script tự động

Tôi đã tạo sẵn script `upload-to-github.ps1`:

```powershell
# Chạy với quyền thường (không cần Admin)
powershell -ExecutionPolicy Bypass -File upload-to-github.ps1 -CreateNewRepo

# Hoặc nếu đã tạo repo trước:
powershell -ExecutionPolicy Bypass -File upload-to-github.ps1
```

Script tự động:
1. Cài đặt git config (user name + email)
2. Tạo repo mới qua `gh` CLI (nếu có)
3. `git init`, `git add`, `git commit`
4. `git push` với 3 lần retry nếu fail
5. Hiển thị URL repo + hướng dẫn download artifact

---

## ✅ Giải pháp 4: GitHub Desktop (GUI đơn giản)

Tải: https://desktop.github.com/

1. Mở GitHub Desktop → Sign in to GitHub
2. File → New Repository
   - Name: `BCY-VKS-Kiem-sat-An-ninh`
   - Local path: chọn thư mục đã giải nén
   - Initialize with README: **NO** (đã có sẵn)
   - Git ignore: **None**
   - License: **None**
3. Click **"Create Repository"**
4. Click **"Publish repository"** (private recommended)
5. Done

---

## ⚠️ Lỗi thường gặp khi push

### Lỗi 1: `fatal: Authentication failed`

**Nguyên nhân:** Dùng password thật thay vì PAT.

**Fix:** GitHub không chấp nhận password từ 2021-08. Phải dùng PAT:
1. Tạo PAT: https://github.com/settings/tokens (scope: `repo`)
2. Khi push: paste PAT vào ô Password (KHÔNG phải password GitHub)

### Lỗi 2: `fatal: remote origin already exists`

**Fix:**
```bash
git remote remove origin
git remote add origin https://github.com/USERNAME/BCY-VKS-Kiem-sat-An-ninh.git
```

### Lỗi 3: `error: failed to push some refs` (history không khớp)

**Fix:**
```bash
git push -u -f origin main  # --force để ghi đè
```

### Lỗi 4: `remote contains work that you do not have`

**Nguyên nhân:** GitHub repo đã có commit (vd: tự tạo README).

**Fix:**
```bash
# Pull trước rồi push
git pull origin main --allow-unrelated-histories
git push -u origin main

# Hoặc nếu chỉ muốn ghi đè:
git push -u -f origin main
```

### Lỗi 5: Push báo `403 Forbidden`

**Nguyên nhân:** PAT không có scope `repo`, hoặc repo private mà PAT chỉ có `public_repo`.

**Fix:** Tạo lại PAT với scope **`repo`** (full control of private repositories).

---

## 📊 So sánh 4 cách upload

| Cách | Độ khó | Độ ổn định | Tốc độ | Khi nào dùng |
|------|--------|------------|--------|---------------|
| **Web UI** | Dễ nhất | ⚠️ Hay lỗi | Chậm | Repo nhỏ (<10 file) |
| **Command line (git push)** | Trung bình | ✅ Rất ổn | Nhanh | Khuyến nghị |
| **GitHub CLI (gh)** | Dễ | ✅ Rất ổn | Nhanh | Có gh CLI |
| **GitHub Desktop** | Dễ | ✅ Ổn | Trung bình | Thích GUI |
| **Script tự động** | Dễ | ✅ Có retry | Nhanh | Có script sẵn |

---

## 🆘 Vẫn lỗi? Thu thập thông tin debug

Nếu vẫn không push được, chạy lệnh sau để biết lỗi chính xác:

```bash
# Debug push với verbose
GIT_CURL_VERBOSE=1 git push -u origin main 2>&1 | head -50

# Kiểm tra remote URL
git remote -v

# Kiểm tra git config
git config --list --global

# Test kết nối tới GitHub
ssh -T git@github.com  # nếu dùng SSH
curl -I https://github.com  # nếu dùng HTTPS
```

Sau đó gửi output cho tôi để debug tiếp.

---

## 📞 Hỗ trợ

- GitHub Status: https://www.githubstatus.com/
- GitHub Support: https://support.github.com/
- Tạo Personal Access Token: https://github.com/settings/tokens
- GitHub CLI docs: https://cli.github.com/manual/

---

## 💡 Khuyến nghị

Cho trường hợp của bạn (99 files, 1.7MB), **dùng GitHub CLI** là đơn giản và ổn định nhất:

```powershell
# 1. Cài GitHub CLI
winget install GitHub.cli

# 2. Login (1 lần)
gh auth login

# 3. Vào thư mục code đã giải nén
cd BCY-VKS-Kiem-sat-An-ninh

# 4. Tạo repo + push 1 lệnh
git init
git add .
git commit -m "Initial commit: BCY-VKS v1.0.0"
gh repo create BCY-VKS-Kiem-sat-An-ninh --private --source=. --push
```

→ 100% sẽ thành công, không cần bận tâm về Web UI.
