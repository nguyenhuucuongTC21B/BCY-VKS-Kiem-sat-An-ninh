@echo off
:: BCY-VKS Build Script for Windows
:: Run this on a Windows 10/11 x64 machine with Go 1.22+ and Wails CLI installed

setlocal

echo ==========================================
echo  BCY-VKS BUILD SCRIPT
echo ==========================================
echo.

:: 1. Check Go
where go >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go is not installed. Please install Go 1.22+ from https://go.dev/dl/
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
echo [1/5] Running 'go mod tidy'...
go mod tidy
if errorlevel 1 (
    echo [ERROR] 'go mod tidy' failed
    exit /b 1
)

:: 4. Sanity check
echo [2/5] Checking build with 'go vet'...
go vet ./...
if errorlevel 1 (
    echo [WARN] 'go vet' reported issues - proceeding anyway
)

:: 5. Build PRODUCTION (stripped, small ~8-12 MB)
echo [3/5] Building BCY-VKS.exe (PRODUCTION, stripped, ~8-12 MB)...
wails build -platform windows/amd64 -clean -trimpath -ldflags="-s -w"
if errorlevel 1 (
    echo [ERROR] Production build failed
    exit /b 1
)
echo ✅ Production build successful!

:: 6. Build DEBUG (full symbols, ~20-30 MB)
echo [4/5] Building BCY-VKS-debug.exe (DEBUG, full symbols, ~20-30 MB)...
wails build -platform windows/amd64 -clean -trimpath -o BCY-VKS-debug.exe
if errorlevel 1 (
    echo [WARN] Debug build failed - skipping
) else (
    echo ✅ Debug build successful!
)

:: 7. Verify embedded assets
echo [5/5] Verifying embedded assets in production exe...
if exist "build\bin\BCY-VKS.exe" (
    for %%A in ("build\bin\BCY-VKS.exe") do echo   File size: %%~zA bytes
    echo.
    echo To verify embedded assets, run:
    echo   strings build\bin\BCY-VKS.exe ^| findstr /i "index.html style.css main.js logo-bcy crack_signatures cve_db badusb"
)

echo.
echo ==========================================
echo  BUILD COMPLETE
echo ==========================================
echo.
echo Output:
echo   .\build\bin\BCY-VKS.exe         (production, stripped, ~8-12 MB)
echo   .\build\bin\BCY-VKS-debug.exe   (debug, full, ~20-30 MB)
echo.
echo To run: Right-click BCY-VKS.exe -^> Run as Administrator
echo.

endlocal
