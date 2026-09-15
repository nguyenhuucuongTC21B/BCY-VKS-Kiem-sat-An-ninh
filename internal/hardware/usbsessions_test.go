package hardware

import (
        "strings"
        "testing"
        "time"
)

// TestParseKernelPnPXMLEvents kiểm tra parse output XML của wevtutil
func TestParseKernelPnPXMLEvents(t *testing.T) {
        raw := `<Event xmlns='http://schemas.microsoft.com/win/2004/08/events/event'><System><Provider Name='Microsoft-Windows-Kernel-PnP'/><EventID>410</EventID><TimeCreated SystemTime='2026-08-12T03:15:22.1234567Z'/></System><EventData><Data Name='DeviceInstanceId'>USBSTOR\DISK&amp;VEN_SANDISK&amp;PROD_ULTRA&amp;REV_1.00\07815FB8D3A3&amp;0</Data></EventData></Event><Event xmlns='http://schemas.microsoft.com/win/2004/08/events/event'><System><Provider Name='Microsoft-Windows-Kernel-PnP'/><EventID>420</EventID><TimeCreated SystemTime='2026-08-12T04:02:41.0000000Z'/></System><EventData><Data Name='DeviceInstanceId'>USBSTOR\DISK&amp;VEN_SANDISK&amp;PROD_ULTRA&amp;REV_1.00\07815FB8D3A3&amp;0</Data></EventData></Event>`

        events := parseKernelPnPXMLEvents(raw)
        if len(events) != 2 {
                t.Fatalf("mong đợi 2 event, nhận được %d", len(events))
        }
        if events[0].EventID != 410 || events[1].EventID != 420 {
                t.Errorf("EventID sai: %d, %d", events[0].EventID, events[1].EventID)
        }
        want := "usbstor\\disk&ven_sandisk&prod_ultra&rev_1.00\\07815fb8d3a3&0"
        if events[0].DeviceID != want {
                t.Errorf("DeviceID = %q, muốn %q", events[0].DeviceID, want)
        }
        if events[0].Time.Year() != 2026 || events[0].Time.Month() != time.August {
                t.Errorf("Time parse sai: %v", events[0].Time)
        }
}

// TestParseKernelPnPXMLEventsEmpty kiểm tra input rỗng không gây lỗi
func TestParseKernelPnPXMLEventsEmpty(t *testing.T) {
        if evs := parseKernelPnPXMLEvents(""); len(evs) != 0 {
                t.Errorf("mong đợi 0 event, nhận %d", len(evs))
        }
        if evs := parseKernelPnPXMLEvents("rác không phải XML"); len(evs) != 0 {
                t.Errorf("input rác -> mong 0 event, nhận %d", len(evs))
        }
}

// TestBuildDeviceSessions kiểm tra dựng session: cắm -> rút -> cắm (dedupe)
func TestBuildDeviceSessions(t *testing.T) {
        base := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
        mk := func(off time.Duration, id int, dev string) PnPEvent {
                return PnPEvent{Time: base.Add(off), EventID: id, DeviceID: dev}
        }
        events := []PnPEvent{
                // Lần 1: cắm 08:00, rút 08:41
                mk(0, 410, "usb\\vid_0781&pid_5590\\abc123"),
                mk(41*time.Minute, 420, "usb\\vid_0781&pid_5590\\abc123"),
                // Lần 2: cắm 09:00 (400 + 410 trùng nhau -> chỉ tính 1 lần)
                mk(time.Hour, 400, "usb\\vid_0781&pid_5590\\abc123"),
                mk(time.Hour+2*time.Second, 410, "usb\\vid_0781&pid_5590\\abc123"),
        }
        m := buildDeviceSessions(events)
        ds := m["usb\\vid_0781&pid_5590\\abc123"]
        if ds == nil {
                t.Fatal("không tìm thấy device trong map")
        }
        if len(ds.Sessions) != 2 {
                t.Fatalf("mong đợi 2 session, nhận %d: %+v", len(ds.Sessions), ds.Sessions)
        }
        if ds.Arrivals != 2 {
                t.Errorf("Arrivals = %d, muốn 2", ds.Arrivals)
        }
        s1 := ds.Sessions[0]
        if s1.Arrival != "2026-09-01 08:00:00" || s1.Removal != "2026-09-01 08:41:00" {
                t.Errorf("session 1 sai: %+v", s1)
        }
        if s1.Duration != "41 phút" {
                t.Errorf("Duration = %q, muốn '41 phút'", s1.Duration)
        }
        // Session 2 vẫn đang mở (chưa rút) -> Removal rỗng
        if ds.Sessions[1].Removal != "" {
                t.Errorf("session 2 phải đang mở, Removal = %q", ds.Sessions[1].Removal)
        }
}

// TestBuildDeviceSessionsEventsRỗng
func TestBuildDeviceSessionsEmpty(t *testing.T) {
        m := buildDeviceSessions(nil)
        if len(m) != 0 {
                t.Errorf("mong map rỗng, nhận %d phần tử", len(m))
        }
}

// TestVnDuration kiểm tra định dạng thời lượng tiếng Việt
func TestVnDuration(t *testing.T) {
        cases := []struct {
                d    time.Duration
                want string
        }{
                {30 * time.Second, "30 giây"},
                {41 * time.Minute, "41 phút"},
                {2*time.Hour + 21*time.Minute, "2 giờ 21 phút"},
                {1*time.Hour + 47*time.Minute, "1 giờ 47 phút"},
                {25 * time.Hour, "1 ngày 1 giờ"},
                {0, "—"},
        }
        for _, c := range cases {
                if got := vnDuration(c.d); got != c.want {
                        t.Errorf("vnDuration(%v) = %q, muốn %q", c.d, got, c.want)
                }
        }
}

// TestNormalizePnpID
func TestNormalizePnpID(t *testing.T) {
        cases := []struct {
                in, want string
        }{
                // "#" được đổi thành "\" để đồng nhất với DeviceInstanceId của Event Log
                {`\\?\USBSTOR#Disk&Ven_X#serial#{guid}`, `usbstor\disk&ven_x\serial\{guid}`},
                {`\??\USB\VID_045E&PID_07C0\6&2f3a&0&3`, `usb\vid_045e&pid_07c0\6&2f3a&0&3`},
                {`  USB\VID_1&PID_2\SER  `, `usb\vid_1&pid_2\ser`},
        }
        for _, c := range cases {
                if got := normalizePnpID(c.in); got != c.want {
                        t.Errorf("normalizePnpID(%q) = %q, muốn %q", c.in, got, c.want)
                }
        }
}

// TestParseSetupAPILogs kiểm tra parse setupapi.dev.log
func TestParseSetupAPILogs(t *testing.T) {
        content := ">>>  [Device Install (Hardware initiated) - USBSTOR\\DISK&VEN_A&PROD_B\\SERIAL&0]\n" +
                ">>>  Section start 2026/08/12 10:15:22.123\n" +
                "<<<  Section end 2026/08/12 10:15:30.000\n" +
                ">>>  [Device Install (DiInstallDriver) - USB\\VID_046D&PID_C077\\5&2f3a&0&3]\n" +
                ">>>  Section start 2026/09/01 09:00:00.000\n"
        events := parseSetupAPILogs(content)
        if len(events) != 2 {
                t.Fatalf("mong 2 event, nhận %d", len(events))
        }
        if !strings.HasPrefix(events[0].DeviceID, "usbstor\\") {
                t.Errorf("event 0 DeviceID = %q", events[0].DeviceID)
        }
        if events[0].Time.Format(timeLayout) != "2026-08-12 10:15:22" {
                t.Errorf("event 0 Time = %v", events[0].Time)
        }
}

// TestSameDevice kiểm tra matching thiết bị giữa current (registry) và history (event log)
func TestSameDevice(t *testing.T) {
        cur := PeripheralRec{
                HardwareID: `Disk&Ven_SanDisk&Prod_Ultra&Rev_1.00\07815FB8D3A3&0`,
                VIDPID:     "VID_0781&PID_5590",
        }
        hist := PeripheralRec{
                HardwareID: `usbstor\disk&ven_sandisk&prod_ultra&rev_1.00\07815fb8d3a3&0`,
                VIDPID:     "VID_0781&PID_5590",
        }
        if !sameDevice(cur, hist) {
                t.Error("hai record cùng serial tail phải được coi là cùng thiết bị")
        }

        // Khác serial, cùng VID/PID + một bên không có serial tail riêng
        a := PeripheralRec{HardwareID: "VID_046D&PID_C077", VIDPID: "VID_046D&PID_C077"}
        b := PeripheralRec{HardwareID: "", VIDPID: "VID_046D&PID_C077"}
        if !sameDevice(a, b) {
                t.Error("cùng VID/PID khi một bên không có serial phải match")
        }

        // Hai thiết bị khác nhau
        c := PeripheralRec{HardwareID: `usb\vid_046d&pid_c077\aaa`, VIDPID: "VID_046D&PID_C077"}
        d := PeripheralRec{HardwareID: `usb\vid_046d&pid_c077\bbb`, VIDPID: "VID_046D&PID_C077"}
        if sameDevice(c, d) {
            t.Error("serial khác nhau -> không được match dù cùng VID/PID")
        }
}

// TestMergePeripheralRecords kiểm tra merge không nhân bản record
func TestMergePeripheralRecords(t *testing.T) {
        current := []PeripheralRec{
                {HardwareID: "ser123", VendorModel: "SanDisk", DriveLetter: "E:"},
        }
        history := []PeripheralRec{
                {
                        HardwareID: "usb\\vid_0781&pid_5590\\ser123",
                        VIDPID:     "VID_0781&PID_5590",
                        FirstPlug:  "2026-08-01 10:00:00",
                        LastPlug:   "2026-09-01 10:00:00",
                        PlugCount:  3,
                        Sessions: []ConnectSession{
                                {Arrival: "2026-09-01 10:00:00"},
                                {Arrival: "2026-08-20 09:00:00"},
                                {Arrival: "2026-08-01 10:00:00"},
                        },
                },
        }
        merged := mergePeripheralRecords(current, history)
        if len(merged) != 1 {
                t.Fatalf("merge phải gộp thành 1 record, nhận %d", len(merged))
        }
        m := merged[0]
        if m.PlugCount != 3 {
                t.Errorf("PlugCount = %d, muốn 3", m.PlugCount)
        }
        if len(m.Sessions) != 3 {
                t.Errorf("Sessions = %d, muốn 3", len(m.Sessions))
        }
        if m.VIDPID == "" {
                t.Error("VIDPID phải được kế thừa từ history")
        }
        if m.DriveLetter != "E:" {
                t.Error("DriveLetter phải được giữ từ current")
        }
}

// TestPnpSerialTail
func TestPnpSerialTail(t *testing.T) {
        cases := []struct {
                in, want string
        }{
                {`USBSTOR\Disk&Ven_X\SERIAL&0`, `SERIAL&0`},
                {"vidpidonly", "vidpidonly"},
                {"", ""},
        }
        for _, c := range cases {
                if got := pnpSerialTail(c.in); got != c.want {
                        t.Errorf("pnpSerialTail(%q) = %q, muốn %q", c.in, got, c.want)
                }
        }
}

// TestExtractVIDPID
func TestExtractVIDPID(t *testing.T) {
        vid, pid := extractVIDPID("VID_0781&PID_5590")
        if vid != "0781" || pid != "5590" {
                t.Errorf("extractVIDPID = %q, %q; muốn 0781, 5590", vid, pid)
        }
        if s := extractVIDPIDStr(`USB\VID_045E&PID_07C0\6&2f3a`); s != "VID_045E&PID_07C0" {
                t.Errorf("extractVIDPIDStr = %q", s)
        }
        if s := extractVIDPIDStr("usbstor\\disk\\ser"); s != "" {
                t.Errorf("không có VID -> phải rỗng, nhận %q", s)
        }
}
