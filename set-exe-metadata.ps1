# ============================================================
# BCY-VKS Set EXE Metadata Script
# ============================================================
# Set metadata cho file .exe (FileDescription, CompanyName, ProductName, etc.)
# Đây là yếu tố QUAN TRỌNG ảnh hưởng SmartScreen reputation
# (tham khảo certerator - https://github.com/stufus/certerator)
# ============================================================

param(
    [string]$ExePath = "build\bin\BCY-VKS.exe",
    [string]$FileDescription = "BCY-VKS - Kiem Sat An Ninh",
    [string]$CompanyName = "Ban Co Yeu Trung Uong",
    [string]$ProductName = "BCY-VKS Kiem Sat An Ninh",
    [string]$OriginalFilename = "BCY-VKS.exe",
    [string]$LegalCopyright = "Copyright (c) 2026 Ban Co Yeu. All rights reserved.",
    [string]$ProductVersion = "1.0.0.0",
    [string]$FileVersion = "1.0.0.0"
)

$ErrorActionPreference = "Stop"

function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Ok($msg) { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Err($msg) { Write-Host "[ERR]  $msg" -ForegroundColor Red }

if (-not (Test-Path $ExePath)) {
    Write-Err "Khong tim thay file: $ExePath"
    exit 1
}

Write-Info "Setting EXE metadata: $ExePath"

# Method 1: Dùng Resource Hacker (nếu có)
$reshacker = Get-Command "ResourceHacker.exe" -ErrorAction SilentlyContinue
if ($reshacker) {
    Write-Info "Found Resource Hacker at: $($reshacker.Source)"
    # Tạo file resource script
    $rcFile = "build\BCY-VKS-version.rc"
    $rcContent = @"
1 VERSIONINFO
FILEVERSION $($FileVersion -replace '\.', ',')
PRODUCTVERSION $($ProductVersion -replace '\.', ',')
FILEOS 0x40004
FILETYPE 0x1
{
  BLOCK "StringFileInfo"
  {
    BLOCK "040904b0"
    {
      VALUE "FileDescription", "$FileDescription"
      VALUE "FileVersion", "$FileVersion"
      VALUE "CompanyName", "$CompanyName"
      VALUE "ProductName", "$ProductName"
      VALUE "OriginalFilename", "$OriginalFilename"
      VALUE "LegalCopyright", "$LegalCopyright"
      VALUE "ProductVersion", "$ProductVersion"
    }
  }
  BLOCK "VarFileInfo"
  {
    VALUE "Translation", 0x0409, 0x04B0
  }
}
"@
    $rcContent | Out-File -FilePath $rcFile -Encoding ASCII
    & $reshacker.Source -open $ExePath -save $ExePath -resource $rcFile -action addoverwrite
    Write-Ok "Metadata set via Resource Hacker"
} else {
    # Method 2: Dùng Windows SDK Resource Compiler (rc.exe) nếu có
    $rc = Get-Command "rc.exe" -ErrorAction SilentlyContinue
    if ($rc) {
        Write-Info "Using Resource Compiler: $($rc.Source)"
        $rcFile = "build\BCY-VKS-version.rc"
        $resFile = "build\BCY-VKS-version.res"
        $rcContent = @"
1 VERSIONINFO
FILEVERSION $($FileVersion -replace '\.', ',')
PRODUCTVERSION $($ProductVersion -replace '\.', ',')
BEGIN
  BLOCK "StringFileInfo"
  BEGIN
    BLOCK "040904b0"
    BEGIN
      VALUE "FileDescription", "$FileDescription"
      VALUE "FileVersion", "$FileVersion"
      VALUE "CompanyName", "$CompanyName"
      VALUE "ProductName", "$ProductName"
      VALUE "OriginalFilename", "$OriginalFilename"
      VALUE "LegalCopyright", "$LegalCopyright"
      VALUE "ProductVersion", "$ProductVersion"
    END
  END
  BLOCK "VarFileInfo"
  BEGIN
    VALUE "Translation", 0x0409, 0x04B0
  END
END
"@
        $rcContent | Out-File -FilePath $rcFile -Encoding ASCII
        & $rc.Source /r $rcFile /fo $resFile
        # Cần link res vào exe - phức tạp hơn, dùng Method 3 nếu có
        Write-Warn "rc.exe available but linking res to exe requires linker"
    }
    
    # Method 3: Dùng PowerShell + UpdateExplorer or SetContent (không thể modify binary trực tiếp)
    Write-Warn "Neither Resource Hacker nor rc.exe found."
    Write-Host "EXE metadata NOT set. SmartScreen reputation may be reduced." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Cai dat Resource Hacker (free): https://www.angusj.com/resourcehacker/"
    Write-Host "Hoac Windows SDK cho rc.exe: https://developer.microsoft.com/windows/downloads/windows-sdk/"
    Write-Host ""
    Write-Host "WORKAROUND: Khi build Wails, metadata co the set qua wails.json 'info' field:"
    Write-Host '  wails.json -> "info": { "productName": "BCY-VKS Kiem Sat An Ninh", ... }'
    Write-Host "Da co san trong wails.json cua BCY-VKS, khong can chay script nay."
}

# Verify metadata đã có
Write-Info "Verifying metadata..."
$fileInfo = Get-Item $ExePath
if ($fileInfo.VersionInfo) {
    Write-Ok "Current EXE metadata:"
    Write-Host "  FileDescription: $($fileInfo.VersionInfo.FileDescription)"
    Write-Host "  CompanyName:     $($fileInfo.VersionInfo.CompanyName)"
    Write-Host "  ProductName:     $($fileInfo.VersionInfo.ProductName)"
    Write-Host "  FileVersion:     $($fileInfo.VersionInfo.FileVersion)"
    Write-Host "  ProductVersion:  $($fileInfo.VersionInfo.ProductVersion)"
    Write-Host "  LegalCopyright:  $($fileInfo.VersionInfo.LegalCopyright)"
}
