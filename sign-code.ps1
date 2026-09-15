# ============================================================
# BCY-VKS Code Signing Script (Enhanced - certerator principles)
# ============================================================
# Source tham khảo:
#   - DigiCert Software Trust Manager: cloud-based signing với MFA + audit
#   - certerator: cách tạo Authenticode cert chuẩn + cài CA cho máy đích
#   - gsudo: chạy lệnh với quyền Admin trong cùng console
#
# 2 chế độ:
#   1. selfsigned (free, EV-like attributes, dùng cho test nội bộ)
#   2. production (.pfx từ CA như DigiCert/Sectigo)
#   3. digicert (DigiCert Software Trust Manager - cloud, MFA, audit logs)
# ============================================================

param(
    [Parameter(Position=0)]
    [string]$Mode = "selfsigned",
    
    [string]$ExePath = "build\bin\BCY-VKS.exe",
    [string]$PfxPath = "",
    [string]$CertPassword = "",
    
    # DigiCert Software Trust Manager params
    [string]$SmHost = "",
    [string]$SmApiKey = "",
    [string]$SmClientCertFile = "",
    [string]$SmClientCertPassword = "",
    [string]$SmCertKey = "",
    
    # Cert subject info (theo nguyên tắc certerator - Subject phải đầy đủ)
    [string]$SubjectCN = "BCY-VKS - Ban Co Yeu Trung Uong",
    [string]$SubjectO = "Ban Co Yeu Trung Uong",
    [string]$SubjectOU = "Cyber Security Department",
    [string]$SubjectC = "VN",
    [string]$SubjectST = "Ha Noi"
)

$ErrorActionPreference = "Stop"

function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Ok($msg) { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "[ERR]  $msg" -ForegroundColor Red }

Write-Host "============================================================" -ForegroundColor Yellow
Write-Host " BCY-VKS Code Signing Script (Enhanced)" -ForegroundColor Yellow
Write-Host " Mode: $Mode" -ForegroundColor Yellow
Write-Host "============================================================" -ForegroundColor Yellow
Write-Host ""

# Verify .exe
if (-not (Test-Path $ExePath)) {
    Write-Err "Khong tim thay file .exe: $ExePath"
    Write-Host "Hay chay build.bat truoc."
    exit 1
}

# Tìm signtool.exe
function Find-SignTool {
    $signtool = Get-Command signtool.exe -ErrorAction SilentlyContinue
    if ($signtool) { return $signtool.Source }
    
    $sdkPaths = @(
        "C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\signtool.exe",
        "C:\Program Files (x86)\Windows Kits\10\bin\10.0.22000.0\x64\signtool.exe",
        "C:\Program Files (x86)\Windows Kits\10\bin\10.0.19041.0\x64\signtool.exe"
    )
    foreach ($p in $sdkPaths) {
        if (Test-Path $p) { return $p }
    }
    
    $kitsRoot = "C:\Program Files (x86)\Windows Kits\10\bin"
    if (Test-Path $kitsRoot) {
        $latest = Get-ChildItem -Path $kitsRoot -Directory | Sort-Object Name -Descending | Select-Object -First 1
        if ($latest) {
            $candidate = Join-Path $latest.FullName "x64\signtool.exe"
            if (Test-Path $candidate) { return $candidate }
        }
    }
    return $null
}

$signtool = Find-SignTool

# Subject chuẩn (theo certerator - cần CN, O, OU, C, ST)
$subjectDN = "CN=$SubjectCN, O=$SubjectO, OU=$SubjectOU, C=$SubjectC, ST=$SubjectST"
Write-Info "Subject: $subjectDN"

if ($Mode -eq "selfsigned") {
    # ============================================================
    # MODE 1: Self-signed cert với EV-like attributes (theo certerator)
    # ============================================================
    if (-not $signtool) {
        Write-Err "Khong tim thay signtool.exe"
        Write-Host "Cai dat Windows SDK: https://developer.microsoft.com/windows/downloads/windows-sdk/"
        exit 1
    }
    Write-Ok "signtool.exe: $signtool"
    
    Write-Info "Tao self-signed Authenticode cert voi EV-like attributes..."
    
    # Tạo cert với đầy đủ KeyUsage + EnhancedKeyUsage (theo certerator)
    # - Type: CodeSigningCert
    # - KeyUsage: DigitalSignature (chỉ ký số, không encrypt)
    # - EnhancedKeyUsage: Code Signing (1.3.6.1.5.5.7.3.3) + Time Stamping
    # - KeyAlgorithm: RSA 2048-bit
    # - NotAfter: 3 năm (tuổi thọ cert)
    $cert = New-SelfSignedCertificate `
        -Type CodeSigningCert `
        -Subject $subjectDN `
        -KeyUsage DigitalSignature `
        -KeyAlgorithm RSA -KeyLength 2048 -HashAlgorithm SHA256 `
        -NotAfter (Get-Date).AddYears(3) `
        -CertStoreLocation "Cert:\CurrentUser\My"
    
    if (-not $cert) {
        Write-Err "Khong tao duoc self-signed cert"
        exit 1
    }
    
    Write-Ok "Self-signed cert da tao (EV-like)"
    Write-Host "  Subject:    $($cert.Subject)"
    Write-Host "  Issuer:     $($cert.Issuer)"
    Write-Host "  Thumbprint: $($cert.Thumbprint)"
    Write-Host "  Valid:      $($cert.NotBefore) -> $($cert.NotAfter)"
    Write-Host "  Algorithm:  RSA 2048-bit + SHA256"
    
    # Export backup .pfx
    $pfxFile = "build\BCY-VKS-code-signing.pfx"
    $pfxPassword = "BCY-VKS-2026" | ConvertTo-SecureString -AsPlainText -Force
    Export-PfxCertificate -Cert $cert -FilePath $pfxFile -Password $pfxPassword | Out-Null
    Write-Ok "Backup .pfx: $pfxFile (password: BCY-VKS-2026)"
    
    # Export .cer (public cert) để phân phối cho máy đích
    $cerFile = "build\BCY-VKS-code-signing.cer"
    Export-Certificate -Cert $cert -FilePath $cerFile | Out-Null
    Write-Ok "Public cert .cer: $cerFile (phan phoi cho may khac)"
    
    # Cài cert vào CẢ 2 store (theo certerator - bắt buộc để SmartScreen không cảnh báo)
    # 1. LocalMachine\Root (Trusted Root CA) - để Windows trust cert chain
    # 2. LocalMachine\TrustedPublisher - để SmartScreen không cảnh báo "unknown publisher"
    
    # Import vào Trusted Root (cần Admin)
    Write-Info "Cai cert vao Trusted Root CA (can Admin)..."
    try {
        $rootStore = New-Object System.Security.Cryptography.X509Certificates.X509Store("Root", "LocalMachine")
        $rootStore.Open("ReadWrite")
        $rootStore.Add($cert)
        $rootStore.Close()
        Write-Ok "✓ Cert installed vao LocalMachine\Root (Trusted Root CA)"
    } catch {
        Write-Err "Khong the cai vao Root store - co the do khong co Admin"
        Write-Host "  Vui long chay script voi Admin: gsudo powershell -File sign-code.ps1 selfsigned"
        exit 1
    }
    
    # Import vào Trusted Publishers
    Write-Info "Cai cert vao Trusted Publishers..."
    try {
        $pubStore = New-Object System.Security.Cryptography.X509Certificates.X509Store("TrustedPublisher", "LocalMachine")
        $pubStore.Open("ReadWrite")
        $pubStore.Add($cert)
        $pubStore.Close()
        Write-Ok "✓ Cert installed vao LocalMachine\TrustedPublisher"
    } catch {
        Write-Warn "Khong the cai vao TrustedPublisher - cert Root da du de trust"
    }
    
    # Cài vào LocalMachine\My để signtool có thể dùng SHA1 để ký
    Write-Info "Cai cert vao LocalMachine\My..."
    try {
        $myStore = New-Object System.Security.Cryptography.X509Certificates.X509Store("My", "LocalMachine")
        $myStore.Open("ReadWrite")
        $myStore.Add($cert)
        $myStore.Close()
        Write-Ok "✓ Cert installed vao LocalMachine\My"
    } catch {
        Write-Warn "Khong the cai vao My store LocalMachine (khong critical)"
    }
    
    # Ký .exe với timestamp (theo certerator - cần timestamp để cert hết hạn vẫn verify được)
    Write-Info "Ky so file: $ExePath"
    Write-Host "  - Dung SHA256 (modern)" 
    Write-Host "  - Timestamp: http://timestamp.digicert.com (RFC3161)"
    
    $signArgs = @(
        "sign",
        "/sha1", $cert.Thumbprint,
        "/fd", "sha256",           # File digest algorithm
        "/td", "sha256",           # Timestamp digest algorithm
        "/tr", "http://timestamp.digicert.com",  # RFC 3161 timestamp server
        $ExePath
    )
    & $signtool $signArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Warn "Signtool voi timestamp fail (co the do khong co Internet)"
        Write-Info "Thu lai khong timestamp..."
        & $signtool sign /sha1 $cert.Thumbprint /fd sha256 $ExePath
        if ($LASTEXITCODE -ne 0) {
            Write-Err "Signtool failed (exit $LASTEXITCODE)"
            Write-Host "Fallback: Dung Set-AuthenticodeSignature..."
            $signed = Set-AuthenticodeSignature -FilePath $ExePath -Certificate $cert -HashAlgorithm SHA256
            if ($signed.Status -ne 0) {
                Write-Err "Set-AuthenticodeSignature failed: $($signed.Status)"
                exit 1
            }
            Write-Ok "Da ky bang Set-AuthenticodeSignature"
        }
    }
    
    Write-Ok "✅ Code signing successful!"
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Green
    Write-Host "  File $ExePath da duoc ky" -ForegroundColor Green
    Write-Host "  Cert da cai vao Trusted Root + Trusted Publishers" -ForegroundColor Green
    Write-Host "  SmartScreen SE KHONG con hien popup tren may nay" -ForegroundColor Green
    Write-Host "============================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "DE PHAN PHOI CHO MAY KHAC:" -ForegroundColor Yellow
    Write-Host "  1. Copy cac file sau sang may khac:"
    Write-Host "     - $ExePath"
    Write-Host "     - $cerFile"
    Write-Host "  2. Tren may khac (Admin), chay:"
    Write-Host "     certutil -addstore -f Root `"$cerFile`""
    Write-Host "     certutil -addstore -f TrustedPublisher `"$cerFile`""
    Write-Host "  3. Sau do chay .exe - SmartScreen khong hien popup"
    Write-Host ""
    
} elseif ($Mode -eq "production") {
    # ============================================================
    # MODE 2: Production cert .pfx từ CA (DigiCert, Sectigo, etc.)
    # ============================================================
    if (-not $signtool) {
        Write-Err "Khong tim thay signtool.exe"
        exit 1
    }
    if (-not $PfxPath -or -not (Test-Path $PfxPath)) {
        Write-Err "Production mode can file .pfx"
        Write-Host "Cach dung:"
        Write-Host "  .\sign-code.ps1 production -PfxPath C:\certs\cert.pfx -CertPassword 'your_pw'"
        exit 1
    }
    if (-not $CertPassword) {
        $securePwd = Read-Host "Nhap mat khau .pfx" -AsSecureString
        $CertPassword = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
            [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePwd))
    }
    
    Write-Info "Ky so voi production cert: $PfxPath"
    
    $signArgs = @(
        "sign",
        "/f", $PfxPath,
        "/p", $CertPassword,
        "/fd", "sha256",
        "/td", "sha256",
        "/tr", "http://timestamp.digicert.com",
        $ExePath
    )
    & $signtool $signArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Err "signtool failed (exit $LASTEXITCODE)"
        exit 1
    }
    Write-Ok "✅ Production code signing successful!"
    
} elseif ($Mode -eq "digicert") {
    # ============================================================
    # MODE 3: DigiCert Software Trust Manager (cloud, MFA, audit logs)
    # Tham khao: https://github.com/digicert/code-signing-software-trust-action
    # ============================================================
    Write-Info "DigiCert Software Trust Manager mode (cloud signing)"
    Write-Host "  Yeu cau:"
    Write-Host "    - DigiCert ONE account"
    Write-Host "    - SM_HOST, SM_API_KEY, SM_CLIENT_CERT_FILE, SM_CLIENT_CERT_PASSWORD"
    Write-Host "    - SM_CERT_KEY (keypair fingerprint)"
    Write-Host ""
    
    if (-not $SmHost -or -not $SmApiKey -or -not $SmClientCertFile) {
        Write-Err "Thieu params cho DigiCert Software Trust Manager"
        Write-Host "Cach dung:"
        Write-Host "  .\sign-code.ps1 digicert -SmHost 'https://client.digicert.com/mtls' -SmApiKey 'xxx' -SmClientCertFile 'C:\certs\client.p12' -SmClientCertPassword 'pw' -SmCertKey 'abc123'"
        Write-Host ""
        Write-Host "De automate trong CI/CD: dung GitHub Action"
        Write-Host "  uses: digicert/code-signing-software-trust-action@v1"
        Write-Host "  env: SM_HOST, SM_API_KEY, SM_CLIENT_CERT_FILE, SM_CLIENT_CERT_PASSWORD"
        Write-Host "  with: signing-cert: <keypair-id>"
        Write-Host ""
        Write-Host "Loi ich: MFA, audit logs, no .pfx on disk, cross-platform"
        exit 1
    }
    
    # Tải SM Toolkit (DigiCert Software Trust client tools)
    Write-Info "Downloading DigiCert Software Trust Manager tools..."
    $smtoolsUrl = "https://github.com/digicert/ssm-go-tools/releases/latest/download/smtools-windows-x64.zip"
    $smtoolsZip = "build\smtools.zip"
    $smtoolsDir = "build\smtools"
    
    if (-not (Test-Path $smtoolsDir)) {
        New-Item -ItemType Directory -Path $smtoolsDir -Force | Out-Null
        Invoke-WebRequest -Uri $smtoolsUrl -OutFile $smtoolsZip
        Expand-Archive -Path $smtoolsZip -DestinationPath $smtoolsDir -Force
    }
    
    $env:SM_HOST = $SmHost
    $env:SM_API_KEY = $SmApiKey
    $env:SM_CLIENT_CERT_FILE = $SmClientCertFile
    $env:SM_CLIENT_CERT_PASSWORD = $SmClientCertPassword
    
    # Sign with smctl.exe (DigiCert cloud signing)
    $smctl = Join-Path $smtoolsDir "smctl.exe"
    if (-not (Test-Path $smctl)) {
        Write-Err "smctl.exe not found in $smtoolsDir"
        exit 1
    }
    
    Write-Info "Signing with DigiCert cloud (MFA + audit logs)..."
    & $smctl sign --keypair-id $SmCertKey --certificate "code_signing" $ExePath
    if ($LASTEXITCODE -ne 0) {
        Write-Err "smctl failed"
        exit 1
    }
    Write-Ok "✅ DigiCert cloud signing successful!"
    Write-Host "  Audit log: https://www.digicert.com/software-trust-manager/audit"
    Write-Host "  Cert subject: $subjectDN"
    
} else {
    Write-Err "Mode khong hop le. Dung: selfsigned | production | digicert"
    exit 1
}

# Verify signature
Write-Info "Verifying signature..."
if ($signtool) {
    $verifyArgs = @("verify", "/pa", "/v", $ExePath)
    & $signtool $verifyArgs 2>&1 | ForEach-Object { Write-Host $_ }
}

# Show cert info
Write-Host ""
Write-Host "=== Certificate info ===" -ForegroundColor Cyan
$fileCert = Get-AuthenticodeSignature -FilePath $ExePath
if ($fileCert.SignerCertificate) {
    Write-Host "Subject:  $($fileCert.SignerCertificate.Subject)" -ForegroundColor Green
    Write-Host "Issuer:   $($fileCert.SignerCertificate.Issuer)"
    Write-Host "Valid:    $($fileCert.SignerCertificate.NotBefore) -> $($fileCert.SignerCertificate.NotAfter)"
    Write-Host "Status:   $($fileCert.Status)"
    Write-Host "Message:  $($fileCert.StatusMessage)"
}
