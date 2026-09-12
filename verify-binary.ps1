# BCY-VKS Binary Verification Script
# Chạy sau khi build để kiểm tra .exe có đầy đủ assets không

param(
    [string]$ExePath = "build\bin\BCY-VKS.exe"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " BCY-VKS Binary Verification" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $ExePath)) {
    Write-Host "❌ File không tồn tại: $ExePath" -ForegroundColor Red
    Write-Host "Hãy chạy build.bat trước rồi mới verify."
    exit 1
}

$fileInfo = Get-Item $ExePath
$sizeBytes = $fileInfo.Length
$sizeMB = [math]::Round($sizeBytes / 1MB, 2)
$sizeKB = [math]::Round($sizeBytes / 1KB, 2)

Write-Host "📄 File: $ExePath" -ForegroundColor Yellow
Write-Host "📦 Size: $sizeMB MB ($sizeKB KB, $sizeBytes bytes)" -ForegroundColor Yellow
Write-Host ""

# Đọc toàn bộ file thành chuỗi ASCII
Write-Host "🔍 Đang phân tích nội dung binary..." -ForegroundColor Cyan
$bytes = [System.IO.File]::ReadAllBytes($ExePath)
$text = [System.Text.Encoding]::ASCII.GetString($bytes)

Write-Host ""
Write-Host "=== KIỂM TRA ASSETS EMBED ===" -ForegroundColor Cyan
Write-Host ""

$assets = @(
    @{ Name = "index.html (frontend)"; Pattern = "index.html" },
    @{ Name = "style.css (frontend)"; Pattern = "style.css" },
    @{ Name = "main.js (frontend)"; Pattern = "main.js" },
    @{ Name = "logo-bcy.png (logo)"; Pattern = "logo-bcy" },
    @{ Name = "crack_signatures.txt"; Pattern = "crack_signatures" },
    @{ Name = "cve_db.csv"; Pattern = "cve_db" },
    @{ Name = "badusb_vid_pid.txt"; Pattern = "badusb" }
)

$allAssetsFound = $true
foreach ($asset in $assets) {
    if ($text -match $asset.Pattern) {
        Write-Host "  ✅ $($asset.Name)" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $($asset.Name) - KHÔNG TÌM THẤY!" -ForegroundColor Red
        $allAssetsFound = $false
    }
}

Write-Host ""
Write-Host "=== KIỂM TRA FRAMEWORK ===" -ForegroundColor Cyan
Write-Host ""

$frameworks = @(
    @{ Name = "Wails framework"; Pattern = "wails" },
    @{ Name = "gopsutil (process scan)"; Pattern = "gopsutil" },
    @{ Name = "Go runtime"; Pattern = "runtime.main" },
    @{ Name = "Windows syscalls"; Pattern = "syscall" }
)

foreach ($fw in $frameworks) {
    if ($text -match $fw.Pattern) {
        Write-Host "  ✅ $($fw.Name)" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  $($fw.Name) - không tìm thấy (có thể được strip)" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "=== KIỂM TRA CHỨC NĂNG ===" -ForegroundColor Cyan
Write-Host ""

$features = @(
    @{ Name = "License Audit"; Pattern = "ScanCrackTools" },
    @{ Name = "Network Scan"; Pattern = "ScanPort" },
    @{ Name = "Hardware Scan"; Pattern = "ScanNIC" },
    @{ Name = "Peripheral Forensics"; Pattern = "ScanPeripherals" },
    @{ Name = "Malware Forensics"; Pattern = "scanProcessMemory" },
    @{ Name = "Anti-Forensics"; Pattern = "RunAllWipe" },
    @{ Name = "Remediation Report"; Pattern = "GenerateRemediation" }
)

foreach ($f in $features) {
    if ($text -match $f.Pattern) {
        Write-Host "  ✅ $($f.Name)" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $($f.Name) - KHÔNG TÌM THẤY!" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host " TỔNG KẾT" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if ($sizeMB -lt 5) {
    Write-Host "⚠️  Binary quá nhỏ ($sizeMB MB) - có thể thiếu thư viện" -ForegroundColor Yellow
} elseif ($sizeMB -lt 12) {
    Write-Host "✅ Binary production-stripped ($sizeMB MB) - kích thước bình thường" -ForegroundColor Green
} elseif ($sizeMB -lt 25) {
    Write-Host "✅ Binary stripped ($sizeMB MB) - kích thước bình thường" -ForegroundColor Green
} elseif ($sizeMB -lt 40) {
    Write-Host "✅ Binary full debug ($sizeMB MB) - kích thước bình thường" -ForegroundColor Green
} else {
    Write-Host "⚠️  Binary quá lớn ($sizeMB MB) - có thể có vấn đề" -ForegroundColor Yellow
}

if ($allAssetsFound) {
    Write-Host "✅ Tất cả assets đã được embed vào binary" -ForegroundColor Green
} else {
    Write-Host "❌ MỘT SỐ ASSETS THIẾU - build có vấn đề!" -ForegroundColor Red
    Write-Host "Kiểm tra lại:" -ForegroundColor Yellow
    Write-Host "  1. File //go:embed trong main.go"
    Write-Host "  2. Thư mục frontend/ và assets/ có đầy đủ file"
    Write-Host "  3. Chạy 'go mod tidy' trước khi build"
}

Write-Host ""
