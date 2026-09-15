@echo off
:: BCY-VKS Build Script for Windows
:: Run this on a Windows 10/11 x64 machine with Go 1.25+ and Wails CLI installed

setlocal

echo ==========================================
echo  BCY-VKS BUILD SCRIPT
echo ==========================================
echo.

:: 1. Check Go
where go >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go is not installed. Please install Go 1.25+ from https://go.dev/dl/
    exit /b 1
)

:: 2. Check Wails
where wails >nul 2>&1
if errorlevel 1 (
    echo [INFO] Wails CLI not found. Installing...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if errorlevel 1 (
        echo [ERROR] Failed to install Wails CLI
        exit /b 1
    )
)

:: 3. Tidy dependencies
echo [1/6] Running 'go mod tidy'...
go mod tidy
if errorlevel 1 (
    echo [ERROR] 'go mod tidy' failed
    exit /b 1
)

:: 4. Sanity check
echo [2/6] Checking build with 'go vet'...
go vet ./...
if errorlevel 1 (
    echo [WARN] 'go vet' reported issues - proceeding anyway
)

:: 5. Build single-exe
echo [3/6] Building BCY-VKS.exe...
wails build -platform windows/amd64 -clean -trimpath
if errorlevel 1 (
    echo [ERROR] Build failed
    exit /b 1
)
echo ✅ Build successful!

:: 6. Code signing (optional - asks user)
echo [4/6] Code signing check...
if exist "build\bin\BCY-VKS.exe" (
    :: Check if file has signature
    powershell -Command "$sig = Get-AuthenticodeSignature -FilePath 'build\bin\BCY-VKS.exe'; if ($sig.Status -eq 'NotSigned') { exit 1 } else { exit 0 }" 2>nul
    if errorlevel 1 (
        echo.
        echo File BCY-VKS.exe chua duoc ky so (unsigned).
        echo Windows SmartScreen se hien canh bao khi chay.
        echo.
        set /p SIGN_CHOICE="Ban co muon ky so ngay khong? (y/N): "
        if /i "!SIGN_CHOICE!"=="y" (
            echo.
            echo Chon che do ky:
            echo   1. Self-signed (free, chi tren may nay)
            echo   2. Production cert .pfx (tu DigiCert/Sectigo)
            set /p SIGN_MODE="Nhap 1 hoac 2: "
            if "!SIGN_MODE!"=="1" (
                powershell -ExecutionPolicy Bypass -File sign-code.ps1 selfsigned
            ) else if "!SIGN_MODE!"=="2" (
                set /p PFX_PATH="Nhap duong dan .pfx: "
                set /p PFX_PWD="Nhap mat khau .pfx: "
                powershell -ExecutionPolicy Bypass -File sign-code.ps1 production -PfxPath "!PFX_PATH!" -CertPassword "!PFX_PWD!"
            ) else (
                echo Bo qua ky so.
            )
        ) else (
            echo Bo qua ky so. SmartScreen se canh bao khi chay .exe.
            echo De ky so sau: powershell -ExecutionPolicy Bypass -File sign-code.ps1 selfsigned
        )
    ) else (
        echo ✅ File da duoc ky so.
    )
)

:: 7. Verify embedded assets
echo [5/6] Verifying embedded assets...
powershell -ExecutionPolicy Bypass -File verify-binary.ps1

:: 8. Final info
echo [6/6] Build complete!
echo.
echo ==========================================
echo  BUILD OUTPUT
echo ==========================================
echo.
echo File output:
if exist "build\bin\BCY-VKS.exe" (
    for %%A in ("build\bin\BCY-VKS.exe") do echo   BCY-VKS.exe  - %%~zA bytes
)
echo.
echo To run:
echo   1. Right-click BCY-VKS.exe -^> Run as Administrator
echo   2. If SmartScreen warns:
echo      a. Click "More info"
echo      b. Click "Run anyway"
echo      OR run: powershell -ExecutionPolicy Bypass -File unblock-file.ps1
echo      OR sign with: powershell -ExecutionPolicy Bypass -File sign-code.ps1 selfsigned
echo.

endlocal
