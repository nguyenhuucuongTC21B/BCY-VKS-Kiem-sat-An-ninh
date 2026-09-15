# ============================================================
# BCY-VKS Launcher Script - Auto-Elevation via gsudo
# ============================================================
# Phát hiện nếu không chạy với quyền Admin:
#   1. Thử dùng gsudo (https://github.com/gerardog/gsudo) nếu đã cài
#   2. Nếu chưa có gsudo, đề nghị cài đặt (1 lệnh winget)
#   3. Fallback: dùng PowerShell Start-Process -Verb RunAs (UAC popup)
# ============================================================

param(
    [string]$ExePath = "BCY-VKS.exe"
)

$ErrorActionPreference = "Stop"

function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Ok($msg) { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "[ERR]  $msg" -ForegroundColor Red }

# Kiểm tra quyền Admin
function Test-IsAdmin {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($identity)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Kiểm tra gsudo có trong PATH không
function Find-gsudo {
    $gsudo = Get-Command gsudo -ErrorAction SilentlyContinue
    if ($gsudo) { return $gsudo.Source }
    # Kiểm tra đường dẫn phổ biến
    $commonPaths = @(
        "$env:LOCALAPPDATA\Microsoft\WinGet\Packages\gerardog.gsudo_Microsoft.Winget.Source_8wekyb3d8bbwe\gsudo\x64\gsudo.exe",
        "C:\ProgramData\chocolatey\bin\gsudo.exe",
        "$env:USERPROFILE\scoop\apps\gsudo\current\bin\gsudo.exe"
    )
    foreach ($p in $commonPaths) {
        if (Test-Path $p) { return $p }
    }
    return $null
}

Write-Host "============================================================" -ForegroundColor Yellow
Write-Host " BCY-VKS Auto-Elevation Launcher" -ForegroundColor Yellow
Write-Host "============================================================" -ForegroundColor Yellow
Write-Host ""

if (-not (Test-Path $ExePath)) {
    # Tìm file .exe cùng thư mục
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    $ExePath = Join-Path $scriptDir "BCY-VKS.exe"
    if (-not (Test-Path $ExePath)) {
        # Thử tìm trong build/bin
        $ExePath = Join-Path $scriptDir "build\bin\BCY-VKS.exe"
        if (-not (Test-Path $ExePath)) {
            Write-Err "Khong tim thay BCY-VKS.exe"
            Write-Host "Cach dung:"
            Write-Host "  .\launch-as-admin.ps1 -ExePath 'C:\path\to\BCY-VKS.exe'"
            Write-Host "  Hoac dat BCY-VKS.exe cung thu muc voi script nay"
            exit 1
        }
    }
}

if (Test-IsAdmin) {
    Write-Ok "Da co quyen Admin - chay .exe truc tiep"
    Start-Process -FilePath $ExePath -WorkingDirectory (Split-Path $ExePath) -NoNewWindow
    exit 0
}

Write-Warn "Hien dang chay voi quyen User thuong."
Write-Host "BCY-VKS can quyen Admin de:"
Write-Host "  - Doc sâu Registry, Task Scheduler, Driver Store"
Write-Host "  - Quet Memory Forensics (ReadProcessMemory)"
Write-Host "  - Quet Windows Event Logs"
Write-Host "  - Ky so .exe (neu can)"
Write-Host ""

# Phương án 1: gsudo (ưu tiên - tốt UX nhất)
$gsudo = Find-gsudo
if ($gsudo) {
    Write-Ok "Phat hien gsudo: $gsudo"
    Write-Info "Dang nang quyen qua gsudo (1 UAC popup, nhieu lenh)..."
    & $gsudo $ExePath
    exit 0
}

# Phương án 2: Đề nghị cài gsudo
Write-Warn "Khong tim thay gsudo (https://github.com/gerardog/gsudo)"
Write-Host ""
Write-Host "Ban co the cai gsudo de UX tot hon:" -ForegroundColor Cyan
Write-Host "  winget install gerardog.gsudo"
Write-Host "  Hoac: scoop install gsudo"
Write-Host "  Hoac: choco install gsudo"
Write-Host ""
Write-Host "Hoac dung phuong an 3 (PowerShell UAC):" -ForegroundColor Yellow

# Phương án 3: PowerShell Start-Process -Verb RunAs (UAC popup mỗi lần)
$choice = Read-Host "Tiep tuc voi PowerShell UAC? (Y/n)"
if ($choice -eq "n" -or $choice -eq "N") {
    Write-Warn "Huy bo. Vui long right-click BCY-VKS.exe -> Run as Administrator"
    exit 1
}

Write-Info "Khoi dong BCY-VKS.exe voi quyen Admin (UAC popup)..."
try {
    $proc = Start-Process -FilePath $ExePath -Verb RunAs -WorkingDirectory (Split-Path $ExePath) -PassThru
    Write-Ok "BCY-VKS da khoi dong voi PID $($proc.Id)"
} catch {
    Write-Err "Khong the khoi dong voi Admin: $_"
    Write-Host "Vui long right-click BCY-VKS.exe -> Run as Administrator"
    exit 1
}
