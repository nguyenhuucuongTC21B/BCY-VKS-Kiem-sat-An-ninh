package main

import (
        "context"
        "embed"
        "log"
        "os"
        "path/filepath"
        "runtime"
        "sync"
        "time"

        "bcy-vks/internal/anti"
        "bcy-vks/internal/hardware"
        "bcy-vks/internal/license"
        "bcy-vks/internal/malware"
        "bcy-vks/internal/network"
        "bcy-vks/internal/output"

        "github.com/wailsapp/wails/v2"
        "github.com/wailsapp/wails/v2/pkg/options"
        "github.com/wailsapp/wails/v2/pkg/options/assetserver"
        "github.com/wailsapp/wails/v2/pkg/options/windows"
        wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend
var assets embed.FS

//go:embed all:assets
var dataAssets embed.FS

// App là cấu trúc chính gắn với frontend Wails.
// Tất cả method public (viết hoa chữ đầu) sẽ tự động được expose sang
// frontend qua window.go.main.App.<Method>
type App struct {
        ctx          context.Context
        mu           sync.Mutex
        lastResult   *ScanResult
        auditLogPath string
}

// ScanEstimate mô tả thời gian dự kiến cho một nhóm quét
type ScanEstimate struct {
        ID         string `json:"id"`
        Group      string `json:"group"`
        MinSec     int    `json:"min_sec"`
        MaxSec     int    `json:"max_sec"`
        Status     string `json:"status"` // "pending" | "running" | "done" | "timeout"
}

// GetScanEstimates trả về bảng dự báo thời gian quét cho frontend
// (gọi ngay khi mở app để UI hiển thị bảng dự báo)
func (a *App) GetScanEstimates() []ScanEstimate {
        return []ScanEstimate{
                {"bang1", "Bản quyền & Crack Tools", 2, 8, "pending"},
                {"bang2", "Mạng & Pentest", 5, 20, "pending"},
                {"bang3", "Card mạng & Wi-Fi", 1, 4, "pending"},
                {"bang4", "USB & Ngoại vi", 1, 5, "pending"},
                {"bang5", "Mã độc & Memory", 3, 15, "pending"},
        }
}

// ScanResult gói kết quả 5 bảng trả về cho frontend
type ScanResult struct {
        Bang1 []license.LicenseRecord   `json:"bang1"`
        Bang2 []network.NetworkRecord   `json:"bang2"`
        Bang3 []hardware.HardwareRecord `json:"bang3"`
        Bang4 []hardware.PeripheralRec  `json:"bang4"`
        Bang5 []malware.MalwareRecord   `json:"bang5"`
        Stats ScanStats                 `json:"stats"`
}

// ScanStats là số liệu tóm tắt để hiển thị trên dashboard
type ScanStats struct {
        TotalCrack     int     `json:"total_crack"`
        OpenPorts      int     `json:"open_ports"`
        CriticalCVE    int     `json:"critical_cve"`
        Adapters       int     `json:"adapters"`
        Peripherals    int     `json:"peripherals"`
        BadUSB         int     `json:"badusb"`
        Keyloggers     int     `json:"keyloggers"`
        SuspiciousProc int     `json:"suspicious_proc"`
        StartedAt      string  `json:"started_at"`
        DurationSec    float64 `json:"duration_sec"`
}

// NewApp tạo instance App mới
func NewApp() *App {
        home, _ := os.UserHomeDir()
        auditDir := filepath.Join(home, ".bcy-vks", "logs")
        _ = os.MkdirAll(auditDir, 0o755)
        return &App{
                auditLogPath: filepath.Join(auditDir, "audit.bcy"),
        }
}

// Startup được Wails gọi khi ứng dụng vừa khởi động
func (a *App) startup(ctx context.Context) {
        a.ctx = ctx
        a.writeAuditLog("APP_START", "BCY-VKS khởi động")
}

// Shutdown được Wails gọi khi ứng dụng chuẩn bị đóng
func (a *App) shutdown(ctx context.Context) {
        a.writeAuditLog("APP_STOP", "BCY-VKS đóng")
}

// ScanAll chạy toàn bộ 5 nhóm quét song song và trả về ScanResult
// Mỗi nhóm có timeout 60 giây để chống treo vô hạn
// Phát progress events qua wailsRuntime.EventsEmit để UI hiển thị tiến độ real-time
func (a *App) ScanAll() (*ScanResult, error) {
        a.mu.Lock()
        defer a.mu.Unlock()

        a.writeAuditLog("SCAN_START", "Bắt đầu quét toàn bộ hệ thống")
        start := time.Now()

        // Phát event "scan_progress" để frontend cập nhật UI
        a.emitProgress("init", "Đang khởi tạo trình quét...", 5)

        if runtime.GOOS != "windows" {
                // Trong môi trường dev/demo không phải Windows, trả dữ liệu mẫu
                a.emitProgress("demo", "Demo mode - đang sinh dữ liệu mẫu...", 50)
                time.Sleep(500 * time.Millisecond) // giả lập delay
                res := a.demoResult()
                res.Stats.DurationSec = time.Since(start).Seconds()
                res.Stats.StartedAt = start.Format(time.RFC3339)
                a.lastResult = res
                a.emitProgress("done", "Quét hoàn tất (demo)", 100)
                a.writeAuditLog("SCAN_DONE", "Demo mode - quét hoàn tất")
                return res, nil
        }

        a.emitProgress("license", "Đang quét bản quyền & công cụ crack (dự kiến 2-8s)...", 15)
        a.emitProgress("network", "Đang quét cổng mạng & đối chiếu CVE (dự kiến 5-20s)...", 25)
        a.emitProgress("hardware", "Đang kiểm kê card mạng & Wi-Fi (dự kiến 1-4s)...", 40)
        a.emitProgress("peripheral", "Đang trích xuất lịch sử USB (dự kiến 1-5s)...", 55)
        a.emitProgress("malware", "Đang giám định mã độc & keylogger (dự kiến 3-15s)...", 75)

        out := &ScanResult{}
        var wg sync.WaitGroup

        // scanGroup chạy một nhóm quét với timeout 60 giây
        // Nếu quá 60s, sẽ emit progress "timeout" và trả kết quả rỗng cho nhóm đó
        scanGroup := func(wg *sync.WaitGroup, groupID, doneMsg string, percent int, fn func() error) {
                defer wg.Done()
                done := make(chan struct{})
                go func() {
                        defer close(done)
                        _ = fn()
                }()
                select {
                case <-done:
                        a.emitProgress(groupID+"_done", doneMsg, percent)
                case <-time.After(60 * time.Second):
                        a.emitProgress(groupID+"_timeout", "⚠ Timeout nhóm "+groupID+" sau 60s", percent)
                        a.writeAuditLog("SCAN_TIMEOUT", "Nhóm "+groupID+" timeout sau 60s")
                }
        }

        wg.Add(5)
        go scanGroup(&wg, "license", "✓ Hoàn tất quét bản quyền", 30, func() error {
                out.Bang1, _ = license.ScanAll(dataAssets)
                return nil
        })
        go scanGroup(&wg, "network", "✓ Hoàn tất quét mạng", 45, func() error {
                out.Bang2, _ = network.ScanAll(dataAssets)
                return nil
        })
        go scanGroup(&wg, "hardware", "✓ Hoàn tất kiểm kê card mạng", 60, func() error {
                out.Bang3, _ = hardware.ScanNIC()
                return nil
        })
        go scanGroup(&wg, "peripheral", "✓ Hoàn tất quét USB", 80, func() error {
                out.Bang4, _ = hardware.ScanPeripherals(dataAssets)
                return nil
        })
        go scanGroup(&wg, "malware", "✓ Hoàn tất giám định mã độc", 95, func() error {
                out.Bang5, _ = malware.ScanAll(dataAssets)
                return nil
        })
        wg.Wait()

        out.Stats = ScanStats{
                TotalCrack:     len(out.Bang1),
                OpenPorts:      countOpenPorts(out.Bang2),
                CriticalCVE:    countCriticalCVE(out.Bang2),
                Adapters:       len(out.Bang3),
                Peripherals:    len(out.Bang4),
                BadUSB:         countBadUSB(out.Bang4),
                Keyloggers:     countKeyloggers(out.Bang5),
                SuspiciousProc: len(out.Bang5),
                StartedAt:      start.Format(time.RFC3339),
                DurationSec:    time.Since(start).Seconds(),
        }

        a.lastResult = out
        a.emitProgress("done", "✓ Quét hoàn tất", 100)
        a.writeAuditLog("SCAN_DONE", "Quét hoàn tất")
        return out, nil
}

// emitProgress phát event "scan_progress" tới frontend
// Wails frontend sẽ nhận qua wails.Events.On('scan_progress', ...)
func (a *App) emitProgress(step, message string, percent int) {
        if a.ctx == nil {
                return
        }
        wailsRuntime.EventsEmit(a.ctx, "scan_progress", map[string]interface{}{
                "step":    step,
                "message": message,
                "percent": percent,
        })
}

// GetLastResult trả về kết quả quét gần nhất
func (a *App) GetLastResult() (*ScanResult, error) {
        a.mu.Lock()
        defer a.mu.Unlock()
        if a.lastResult == nil {
                return a.ScanAll()
        }
        return a.lastResult, nil
}

// ExportRemediationReport sinh báo cáo đề xuất khắc phục
// format = "docx" | "html" | "popup"
func (a *App) ExportRemediationReport(format string) (string, error) {
        a.writeAuditLog("EXPORT_REMEDIATION", "format="+format)
        if a.lastResult == nil {
                r, err := a.ScanAll()
                if err != nil {
                        return "", err
                }
                a.lastResult = r
        }
        return output.GenerateRemediation(a.lastResult, format, dataAssets)
}

// RunAntiForensics kích hoạt chế độ Phòng thủ (Wipe-Out)
// Yêu cầu mã xác nhận WIPE-CONFIRM-2026 để chống click nhầm
func (a *App) RunAntiForensics(confirmCode string) ([]anti.WipeStep, error) {
        a.writeAuditLog("ANTI_FORENSICS_REQ", "confirm="+confirmCode)
        if confirmCode != "WIPE-CONFIRM-2026" {
                return nil, errInvalidConfirm
        }
        steps := anti.RunAllWipe()
        a.writeAuditLog("ANTI_FORENSICS_DONE", "đã thực hiện các bước wipe")
        return steps, nil
}

// ReadAuditLog trả về nội dung file audit log nội bộ
func (a *App) ReadAuditLog() (string, error) {
        b, err := os.ReadFile(a.auditLogPath)
        if err != nil {
                return "", err
        }
        return string(b), nil
}

// ===================== helpers =====================

func (a *App) writeAuditLog(action, detail string) {
        line := time.Now().Format(time.RFC3339) + " | " + action + " | " + detail + "\n"
        f, err := os.OpenFile(a.auditLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
        if err != nil {
                return
        }
        defer f.Close()
        _, _ = f.WriteString(line)
}

func countOpenPorts(rs []network.NetworkRecord) int {
        n := 0
        for _, r := range rs {
                if r.OpenPorts != "" {
                        n += countCSV(r.OpenPorts)
                }
        }
        return n
}

func countCriticalCVE(rs []network.NetworkRecord) int {
        n := 0
        for _, r := range rs {
                if r.CVEID != "" && r.CVSSScore >= 7.0 {
                        n++
                }
        }
        return n
}

func countBadUSB(rs []hardware.PeripheralRec) int {
        n := 0
        for _, r := range rs {
                if r.BadUSBWarning {
                        n++
                }
        }
        return n
}

func countKeyloggers(rs []malware.MalwareRecord) int {
        n := 0
        for _, r := range rs {
                if r.Type == "Keylogger" {
                        n++
                }
        }
        return n
}

func countCSV(s string) int {
        if s == "" {
                return 0
        }
        c := 1
        for i := 0; i < len(s); i++ {
                if s[i] == ',' {
                        c++
                }
        }
        return c
}

// errInvalidConfirm là lỗi khi mã xác nhận Anti-Forensics không khớp
var errInvalidConfirm = &appErr{Code: 401, Msg: "Mã xác nhận không hợp lệ. Yêu cầu chuỗi WIPE-CONFIRM-2026"}

type appErr struct {
        Code int
        Msg  string
}

func (e *appErr) Error() string { return e.Msg }

// ===================== Demo data (cho dev/Linux) =====================

func (a *App) demoResult() *ScanResult {
        return &ScanResult{
                Bang1: []license.LicenseRecord{
                        {
                                SoftwareName:     "Microsoft Windows 11 Pro",
                                Version:          "23H2 (Build 22631.3737)",
                                ProductID:        "00330-50000-00000-AAOEM",
                                LicensingChannel: "OEM:SLP",
                                BIOSOEMKey:       "VK7JG-NPHTM-C97JM-9MPGT-3V66T",
                                CrackTool:        "KMSpico",
                                CrackPath:        "C:\\Users\\Public\\Downloads\\KMSpico_setup.exe",
                                Legal:            false,
                        },
                },
                Bang2: []network.NetworkRecord{
                        {
                                InternetStatus:    "Connected",
                                CurrentIP:         "192.168.1.105",
                                CurrentMAC:        "AA:BB:CC:11:22:33",
                                ISP:               "Viettel",
                                ConnectionHistory: "2026-09-10 08:23 | 192.168.1.105 | DHCP",
                                OpenPorts:         "445,3389,135,139",
                                CVEID:             "CVE-2017-0144",
                                CVSSScore:         8.1,
                                ExploitResult:     "SUCCESS",
                                Notes:             "SMBv1 đang bật - EternalBlue rủi ro cao",
                        },
                },
                Bang3: []hardware.HardwareRecord{
                        {
                                AdapterType:   "LAN",
                                ConnectionPos: "Internal",
                                DeviceName:    "Intel I219-V Gigabit",
                                Serial:        "PCI\\VEN_8086&DEV_15BB",
                                MAC:           "AA:BB:CC:11:22:33",
                                DriverStatus:  "OK",
                                SSIDList:      "",
                                LastConnect:   "",
                        },
                        {
                                AdapterType:   "Wi-Fi",
                                ConnectionPos: "Internal (M.2)",
                                DeviceName:    "Intel Wi-Fi 6E AX211",
                                Serial:        "PCI\\VEN_8086&DEV_51F0",
                                MAC:           "BB:CC:DD:22:33:44",
                                DriverStatus:  "OK",
                                SSIDList:      "BCY-Office,Wifi-Cafe",
                                LastConnect:   "2026-09-11T19:42:00Z",
                        },
                },
                Bang4: []hardware.PeripheralRec{
                        {
                                DeviceType:  "USB Flash Drive",
                                VendorModel: "SanDisk Ultra Flair 64GB",
                                HardwareID:  "disk&ven_sandisk&prod_ultra_flair&rev_1.00\\4c530001234567890123&0",
                                VIDPID:      "VID_0781&PID_5590",
                                DriveLetter: "E:",
                                FirstPlug:   "2026-08-12 10:15:22",
                                LastPlug:    "2026-09-11 15:08:01",
                                PlugCount:   6,
                                BadUSBWarning: false,
                                RecentFilesSummary: "8 shortcut Recent: bao-cao-q3.docx, danh-sach-can-bo.xlsx, anh-hoi-nghi.zip | " +
                                        "Gốc ổ E:: 18 mục, mới nhất: BAO-CAO-Q3.docx (11/09/2026 14:55)",
                                Sessions: []hardware.ConnectSession{
                                        {Arrival: "2026-09-11 15:08:01", Removal: "", Duration: "—"},
                                        {Arrival: "2026-09-10 08:23:44", Removal: "2026-09-10 09:05:12", Duration: "41 phút"},
                                        {Arrival: "2026-09-05 14:02:10", Removal: "2026-09-05 14:47:39", Duration: "45 phút"},
                                        {Arrival: "2026-08-29 09:11:00", Removal: "2026-08-29 11:32:15", Duration: "2 giờ 21 phút"},
                                        {Arrival: "2026-08-20 16:40:33", Removal: "2026-08-20 16:44:58", Duration: "4 phút"},
                                        {Arrival: "2026-08-12 10:15:22", Removal: "2026-08-12 12:02:41", Duration: "1 giờ 47 phút"},
                                },
                        },
                        {
                                DeviceType:  "USB Mouse",
                                VendorModel: "Logitech USB Optical Mouse",
                                HardwareID:  "5&1f2e3d4c&0&3",
                                VIDPID:      "VID_046D&PID_C077",
                                FirstPlug:   "2026-08-12 09:50:01",
                                LastPlug:    "2026-09-11 15:08:00",
                                PlugCount:   3,
                                BadUSBWarning: false,
                                Sessions: []hardware.ConnectSession{
                                        {Arrival: "2026-09-11 15:08:00", Removal: "", Duration: "—"},
                                        {Arrival: "2026-08-12 10:14:55", Removal: "2026-08-12 18:22:10", Duration: "8 giờ 7 phút"},
                                        {Arrival: "2026-08-12 09:50:01", Removal: "2026-08-12 10:01:23", Duration: "11 phút"},
                                },
                        },
                },
                Bang5: []malware.MalwareRecord{
                        {
                                ProcessName:      "svch0st.exe",
                                PID:              4812,
                                Type:             "Keylogger",
                                DangerLevel:      "CRITICAL",
                                RunningInRAM:     true,
                                FilePath:         "C:\\Users\\Public\\svch0st.exe",
                                C2Server:         "45.137.21.88:8443",
                                LogWipeEvidence:  "Đã phát hiện Event ID 1102 lúc 03:14",
                                Reason:           "Trojan (mạo danh svchost)",
                        },
                },
                Stats: ScanStats{
                        TotalCrack:     1,
                        OpenPorts:      4,
                        CriticalCVE:    1,
                        Adapters:       2,
                        Peripherals:    2,
                        BadUSB:         0,
                        Keyloggers:     1,
                        SuspiciousProc: 1,
                },
        }
}

// ===================== main =====================

func main() {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
        app := NewApp()

        err := wails.Run(&options.App{
                Title:     "BCY-VKS Kiểm Sát An Ninh",
                Width:     1440,
                Height:    900,
                MinWidth:  1100,
                MinHeight: 700,
                AssetServer: &assetserver.Options{
                        Assets: assets,
                },
                BackgroundColour: &options.RGBA{R: 18, G: 22, B: 30, A: 1},
                OnStartup:        app.startup,
                OnShutdown:       app.shutdown,
                Bind: []interface{}{
                        app,
                },
                Windows: &windows.Options{
                        WebviewIsTransparent: false,
                        WindowIsTranslucent:  false,
                },
        })
        if err != nil {
                log.Fatalf("BCY-VKS lỗi khởi động: %v", err)
        }
}
