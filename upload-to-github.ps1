# ============================================================
# BCY-VKS GitHub Upload Script - Upload qua command line
# ============================================================
# Dùng khi Web UI báo lỗi "Looks like something went wrong"
# Cách này ổn định hơn vì git push xử lý retries, authentication
# ============================================================

param(
    [Parameter(Mandatory=$true)]
    [string]$RepoName = "BCY-VKS-Kiem-sat-An-ninh",
    
    [string]$Description = "BCY-VKS v1.0.0 - Offline Security Audit Suite",
    
    [string]$SourceDir = ".",  # Thư mục chứa code đã giải nén
    
    [string]$GitUserName = "",  # GitHub username (cần nhập nếu chưa set)
    [string]$GitEmail = "",     # GitHub email (cần nhập nếu chưa set)
    
    [switch]$CreateNewRepo,     # Tự tạo repo mới qua gh CLI hoặc curl
    [switch]$Force              # Bỏ qua git history nếu đã có repo
)

$ErrorActionPreference = "Stop"

function Write-Info($msg)  { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Ok($msg)   { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err($msg)  { Write-Host "[ERR]  $msg" -ForegroundColor Red }

Write-Host "============================================================" -ForegroundColor Yellow
Write-Host " BCY-VKS GitHub Upload Script (command line)" -ForegroundColor Yellow
Write-Host " Repo: $RepoName" -ForegroundColor Yellow
Write-Host "============================================================" -ForegroundColor Yellow
Write-Host ""

# 1. Kiểm tra git đã cài
$git = Get-Command git -ErrorAction SilentlyContinue
if (-not $git) {
    Write-Err "Git chưa cài. Tải từ: https://git-scm.com/download/win"
    exit 1
}
Write-Ok "Git: $($git.Source)"

# 2. Set user name + email nếu chưa có
if ($GitUserName) {
    git config --global user.name $GitUserName
    Write-Ok "Git user.name: $GitUserName"
} else {
    $currentName = git config --global user.name 2>$null
    if (-not $currentName) {
        $GitUserName = Read-Host "Nhap GitHub username (vd: nguyenhuucuong)"
        git config --global user.name $GitUserName
        Write-Ok "Git user.name set: $GitUserName"
    } else {
        Write-Ok "Git user.name existing: $currentName"
    }
}

if ($GitEmail) {
    git config --global user.email $GitEmail
    Write-Ok "Git user.email: $GitEmail"
} else {
    $currentEmail = git config --global user.email 2>$null
    if (-not $currentEmail) {
        $GitEmail = Read-Host "Nhap GitHub email"
        git config --global user.email $GitEmail
        Write-Ok "Git user.email set: $GitEmail"
    } else {
        Write-Ok "Git user.email existing: $currentEmail"
    }
}

# 3. Vào thư mục source
if ($SourceDir -ne ".") {
    if (-not (Test-Path $SourceDir)) {
        Write-Err "Thư mục source không tồn tại: $SourceDir"
        exit 1
    }
    Set-Location $SourceDir
}
Write-Info "Working dir: $(Get-Location)"

# 4. Verify cấu trúc
if (-not (Test-Path "main.go")) {
    Write-Err "Không tìm thấy main.go. Đảm bảo đang ở thư mục gốc của BCY-VKS"
    Write-Host "Cấu trúc phải có:"
    Write-Host "  main.go, wails.json, go.mod"
    Write-Host "  internal/, frontend/, build/, assets/"
    Write-Host "  .github/workflows/build.yml"
    exit 1
}
Write-Ok "Cấu trúc thư mục OK"

# 5. Tạo repo mới trên GitHub nếu cần
$remoteUrl = "https://github.com/$GitUserName/$RepoName.git"

if ($CreateNewRepo) {
    Write-Info "Tạo repo mới: $RepoName"
    
    # Phương án A: Dùng gh CLI (khuyến nghị)
    $gh = Get-Command gh -ErrorAction SilentlyContinue
    if ($gh) {
        Write-Ok "GitHub CLI (gh) đã cài"
        Write-Info "Tạo repo qua gh CLI..."
        # Kiểm tra đã login
        $authStatus = gh auth status 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Warn "Chưa login GitHub CLI. Đang mở browser để login..."
            gh auth login --web --git-protocol https
        }
        # Tạo repo private (khuyến nghị cho phần mềm nội bộ)
        gh repo create $RepoName --private --description $Description
        if ($LASTEXITCODE -eq 0) {
            Write-Ok "✅ Repo đã tạo: $remoteUrl"
        } else {
            Write-Err "Không tạo được repo qua gh CLI"
            exit 1
        }
    } else {
        # Phương án B: Hướng dẫn tạo manual
        Write-Warn "GitHub CLI (gh) chưa cài"
        Write-Host ""
        Write-Host "Vui lòng tạo repo manual:" -ForegroundColor Yellow
        Write-Host "  1. Vào https://github.com/new"
        Write-Host "  2. Repository name: $RepoName"
        Write-Host "  3. Visibility: Private (recommended)"
        Write-Host "  4. KHÔNG tick 'Add a README file'"
        Write-Host "  5. KHÔNG tick 'Add .gitignore'"
        Write-Host "  6. KHÔNG tick 'Choose a license'"
        Write-Host "  7. Click 'Create repository'"
        Write-Host ""
        $continue = Read-Host "Đã tạo repo xong? Nhấn Enter để tiếp tục..."
    }
} else {
    Write-Info "Giả định repo đã tồn tại: $remoteUrl"
}

# 6. Khởi tạo git repo
if (Test-Path ".git") {
    if ($Force) {
        Write-Warn "Xóa .git cũ (--Force)"
        Remove-Item -Recurse -Force ".git"
    } else {
        Write-Info ".git đã tồn tại - sử dụng repo hiện có"
    }
}

if (-not (Test-Path ".git")) {
    Write-Info "git init..."
    git init
    git branch -M main
}

# 7. Add tất cả file
Write-Info "git add . (thêm toàn bộ file)"
git add .

# 8. Show status để verify
Write-Info "git status - số file staged:"
$statusOutput = git status --short
$stagedCount = ($statusOutput | Where-Object { $_ -match "^[AM]" }).Count
Write-Host "  $stagedCount file đã staged"

# 9. Commit
Write-Info "git commit..."
$commitMsg = @"
Initial commit: BCY-VKS v1.0.0 - Offline Security Audit Suite

- 5 nhóm quét: License, Network, Hardware, Peripheral, Malware
- 5 bảng kết quả + 2 nút lệnh (Khắc phục & Anti-Forensics)
- 6 tính năng tinh hoa (DNS, Browser, LAN, Document, Component, Recent)
- Custom Vietnam government frontend (banner vàng + 3 tầng header)
- Popup chi tiết cho 8 ô thống kê
- Căn cứ pháp lý: Luật An ninh mạng, BLHS, Luật Bảo vệ bí mật nhà nước
- File .docx THẬT (OOXML) + file HTML đầy đủ căn cứ pháp lý
- Race condition fix: timeout mềm chờ kết quả đầy đủ
- Build: Go 1.25 + Wails v2.9.1
- Cross-platform codebase (build tags windows/non-windows)
"@
git commit -m $commitMsg

if ($LASTEXITCODE -ne 0) {
    Write-Err "git commit thất bại"
    exit 1
}
Write-Ok "Commit thành công"

# 10. Add remote
Write-Info "git remote add origin..."
$existingRemote = git remote get-url origin 2>$null
if ($existingRemote) {
    if ($existingRemote -ne $remoteUrl) {
        Write-Warn "Remote cũ khác - đổi sang URL mới"
        git remote set-url origin $remoteUrl
    } else {
        Write-Ok "Remote đã đúng URL"
    }
} else {
    git remote add origin $remoteUrl
}
Write-Ok "Remote: $remoteUrl"

# 11. Push
Write-Info "git push -u origin main..."
Write-Host "  (Sẽ hỏi username + Personal Access Token nếu lần đầu)"
Write-Host ""

# Try push with retry
$maxRetries = 3
$retry = 0
$pushed = $false

while (-not $pushed -and $retry -lt $maxRetries) {
    $retry++
    Write-Host "  Attempt $retry/$maxRetries..." -ForegroundColor Cyan
    
    if ($Force) {
        git push -u -f origin main
    } else {
        git push -u origin main
    }
    
    if ($LASTEXITCODE -eq 0) {
        $pushed = $true
        Write-Ok "✅ Push thành công!"
    } else {
        Write-Warn "Push attempt $retry thất bại"
        if ($retry -lt $maxRetries) {
            Write-Host "  Đợi 5 giây rồi thử lại..." -ForegroundColor Yellow
            Start-Sleep -Seconds 5
        }
    }
}

if (-not $pushed) {
    Write-Err "Push thất bại sau $maxRetries lần"
    Write-Host ""
    Write-Host "CÁCH KHẮC PHỤC:" -ForegroundColor Yellow
    Write-Host "  1. Kiểm tra Personal Access Token (PAT) còn hiệu lực"
    Write-Host "     https://github.com/settings/tokens"
    Write-Host "     Scope: 'repo' full control"
    Write-Host "  2. Đảm bảo username nhập ĐÚNG (không có @)"
    Write-Host "  3. Repo đã tạo trên GitHub và URL đúng"
    Write-Host "  4. Thử: git push -u origin main --verbose"
    Write-Host ""
    Write-Host "Hoặc dùng gh CLI:"
    Write-Host "  winget install GitHub.cli"
    Write-Host "  gh auth login"
    Write-Host "  gh repo create $RepoName --private --source=. --push"
    exit 1
}

Write-Host ""
Write-Host "============================================================" -ForegroundColor Green
Write-Host "  ✅ UPLOAD THÀNH CÔNG!" -ForegroundColor Green
Write-Host "============================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Repo URL: https://github.com/$GitUserName/$RepoName" -ForegroundColor Cyan
Write-Host ""
Write-Host "Bước tiếp theo:" -ForegroundColor Yellow
Write-Host "  1. Vào tab Actions"
Write-Host "  2. Đợi workflow 'Build BCY-VKS' chạy (3-5 phút)"
Write-Host "  3. Click vào build run mới nhất"
Write-Host "  4. Kéo xuống cuối → download artifact 'BCY-VKS-exe'"
Write-Host "  5. Giải nén → được file BCY-VKS.exe"
Write-Host "  6. Copy sang máy khác, chạy 'More info' → 'Run anyway'"
Write-Host ""
