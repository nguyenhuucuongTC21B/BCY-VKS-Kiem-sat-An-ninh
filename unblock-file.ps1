# ============================================================
# BCY-VKS Unblock File Script
# ============================================================
# Dùng khi tải file .exe về máy và Windows "mark of the web"
# đánh dấu là tải từ Internet → SmartScreen cảnh báo.
# Script này sẽ unblock file để không cần bấm "More info" → "Run anyway"
# ============================================================

param(
    [Parameter(Position=0)]
    [string]$ExePath = "BCY-VKS.exe"
)

$ErrorActionPreference = "Stop"

function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Ok($msg) { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "[ERR]  $msg" -ForegroundColor Red }

Write-Host "============================================================" -ForegroundColor Yellow
Write-Host " BCY-VKS Unblock File Script" -ForegroundColor Yellow
Write-Host "============================================================" -ForegroundColor Yellow
Write-Host ""

if (-not (Test-Path $ExePath)) {
    Write-Err "Khong tim thay file: $ExePath"
    Write-Host "Cach dung:"
    Write-Host "  .\unblock-file.ps1 -ExePath 'C:\path\to\BCY-VKS.exe'"
    Write-Host "  Hoac dat file BCY-VKS.exe cung thu muc voi script nay"
    exit 1
}

# Lấy thông tin file trước khi unblock
Write-Info "Kiem tra trang thai file truoc khi unblock..."
$zoneInfo = Get-Item -Path $ExePath -Stream Zone.Identifier -ErrorAction SilentlyContinue
if ($zoneInfo) {
    Write-Host "File co 'Mark of the Web' (download tu Internet)" -ForegroundColor Yellow
    Write-Host "  → SmartScreen se canh bao khi chay" -ForegroundColor Yellow
} else {
    Write-Ok "File KHONG co Mark of the Web - khong can unblock"
    Write-Host "  Neu van gap SmartScreen, do la loi 'unknown publisher' (chua ky so)"
    exit 0
}

# Unblock file
Write-Info "Dang unblock file..."
Unblock-File -Path $ExePath

# Verify
$zoneInfoAfter = Get-Item -Path $ExePath -Stream Zone.Identifier -ErrorAction SilentlyContinue
if ($zoneInfoAfter) {
    Write-Err "Unblock KHONG thanh cong. File van co Mark of the Web"
    Write-Host "Thu cach khac: Right-click .exe → Properties → tick 'Unblock' checkbox → Apply" -ForegroundColor Yellow
    exit 1
}

Write-Ok "✅ File da duoc unblock!"
Write-Host ""
Write-Host "Lan sau khi chay .exe, Windows se KHONG hien SmartScreen popup." -ForegroundColor Green
Write-Host ""
Write-Host "Luu y:" -ForegroundColor Yellow
Write-Host "  - Unblock chi loai bo 'Mark of the Web', KHONG thay doi signature"
Write-Host "  - Neu .exe chua ky so, mot so may van hien canh bao 'unknown publisher'"
Write-Host "  - De KHAC PHUC HOAN TOAN, dung sign-code.ps1 de ky so .exe"
Write-Host ""

# Hiển thị thông tin signature (nếu có)
Write-Info "Kiem tra signature cua file..."
$sig = Get-AuthenticodeSignature -FilePath $ExePath
if ($sig.Status -eq "Valid") {
    Write-Ok "File da duoc ky so hop le"
    Write-Host "  Subject: $($sig.SignerCertificate.Subject)"
    Write-Host "  Status:  $($sig.Status)"
} elseif ($sig.Status -eq "NotSigned") {
    Write-Warn "File CHUA duoc ky so - SmartScreen se canh bao 'unknown publisher'"
    Write-Host ""
    Write-Host "  De ky so: .\sign-code.ps1 selfsigned  (cho test)"
    Write-Host "           .\sign-code.ps1 production -PfxPath cert.pfx -CertPassword 'pw'"
} else {
    Write-Warn "Signature status: $($sig.Status)"
    Write-Host "  Message: $($sig.StatusMessage)"
}
