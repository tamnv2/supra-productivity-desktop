//go:build windows

package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/tamnv2/supra-productivity-desktop/internal/core"
	liveio "github.com/tamnv2/supra-productivity-desktop/internal/live"
)

var appVersion = "dev"

const appName = "SUPRA PRODUCTIVITY"
const updateRepo = "tamnv2/supra-productivity-desktop"

const (
	WS_OVERLAPPEDWINDOW          = 0x00CF0000
	WS_VISIBLE                   = 0x10000000
	WS_CHILD                     = 0x40000000
	WS_TABSTOP                   = 0x00010000
	WS_BORDER                    = 0x00800000
	WS_VSCROLL                   = 0x00200000
	WS_HSCROLL                   = 0x00100000
	WS_CAPTION                   = 0x00C00000
	WS_SYSMENU                   = 0x00080000
	WS_MINIMIZEBOX               = 0x00020000
	WS_CLIPCHILDREN              = 0x02000000
	WM_CREATE                    = 0x0001
	WM_DESTROY                   = 0x0002
	WM_CLOSE                     = 0x0010
	WM_SETREDRAW                 = 0x000B
	WM_SYSCOMMAND                = 0x0112
	WM_SIZE                      = 0x0005
	WM_SETTINGCHANGE             = 0x001A
	WM_DISPLAYCHANGE             = 0x007E
	WM_COMMAND                   = 0x0111
	WM_NOTIFY                    = 0x004E
	WM_TIMER                     = 0x0113
	WM_SETFONT                   = 0x0030
	WM_APP                       = 0x8000
	WM_APP_STATUS                = WM_APP + 1
	WM_APP_REFRESH               = WM_APP + 2
	WM_APP_FIT_WORKAREA          = WM_APP + 3
	WM_APP_PING                  = WM_APP + 4
	WM_APP_METRICS               = WM_APP + 5
	WM_APP_LOGTEXT               = WM_APP + 6
	SW_SHOW                      = 5
	SW_MAXIMIZE                  = 3
	CW_USEDEFAULT                = 0x80000000
	SS_LEFT                      = 0x00000000
	ES_MULTILINE                 = 0x0004
	ES_AUTOVSCROLL               = 0x0040
	ES_READONLY                  = 0x0800
	ES_AUTOHSCROLL               = 0x0080
	BS_PUSHBUTTON                = 0x00000000
	BS_GROUPBOX                  = 0x00000007
	BS_AUTOCHECKBOX              = 0x00000003
	CBS_DROPDOWNLIST             = 0x0003
	CB_ADDSTRING                 = 0x0143
	CB_SETCURSEL                 = 0x014E
	CB_GETCURSEL                 = 0x0147
	CB_GETLBTEXTLEN              = 0x0149
	CB_GETLBTEXT                 = 0x0148
	BM_GETCHECK                  = 0x00F0
	BM_SETCHECK                  = 0x00F1
	BST_CHECKED                  = 1
	LVS_REPORT                   = 0x0001
	LVS_SHOWSELALWAYS            = 0x0008
	LVS_EX_FULLROWSELECT         = 0x20
	LVS_EX_GRIDLINES             = 0x1
	LVS_EX_DOUBLEBUFFER          = 0x00010000
	LVM_FIRST                    = 0x1000
	LVM_SETEXTENDEDLISTVIEWSTYLE = LVM_FIRST + 54
	LVM_INSERTCOLUMNW            = LVM_FIRST + 97
	LVM_INSERTITEMW              = LVM_FIRST + 77
	LVM_SETITEMTEXTW             = LVM_FIRST + 116
	LVM_DELETEALLITEMS           = LVM_FIRST + 9
	LVM_GETNEXTITEM              = LVM_FIRST + 12
	LVIF_TEXT                    = 0x0001
	LVCF_FMT                     = 0x0001
	LVCF_WIDTH                   = 0x0002
	LVCF_TEXT                    = 0x0004
	LVCFMT_LEFT                  = 0
	LVNI_SELECTED                = 0x2
	LVN_FIRST                    = -100
	LVN_ITEMCHANGED              = LVN_FIRST - 1
	LVN_COLUMNCLICK              = LVN_FIRST - 8
	NM_DBLCLK                    = -3
	MB_OK                        = 0
	MB_ICONINFORMATION           = 0x40
	MB_ICONWARNING               = 0x30
	CRYPTPROTECT_UI_FORBIDDEN    = 0x1
	SC_SIZE                      = 0xF000
	SC_MOVE                      = 0xF010
	SC_RESTORE                   = 0xF120
	SC_MAXIMIZE                  = 0xF030
	SIZE_RESTORED                = 0
	SIZE_MINIMIZED               = 1
	MONITOR_DEFAULTTONEAREST     = 2
	COLOR_WINDOW                 = 5
	BS_AUTORADIOBUTTON           = 0x00000009
	BS_PUSHLIKE                  = 0x00001000
	WS_GROUP                     = 0x00020000
	OFN_OVERWRITEPROMPT          = 0x00000002
	OFN_PATHMUSTEXIST            = 0x00000800
)

const (
	ID_NAV_OVERVIEW = 101 + iota
	ID_NAV_ACTIVE
	ID_NAV_PICK
	ID_NAV_PACK
	ID_NAV_SHIFT
	ID_NAV_USERPDA
	ID_NAV_LOG
	ID_NAV_SETTINGS
)
const (
	ID_SYNC            = 200
	ID_UPDATE          = 201
	ID_CURL_IMPORT     = 300
	ID_SECRET_TOGGLE   = 301
	ID_NET_TEST        = 302
	ID_SOURCE_KIND     = 303
	ID_SOURCE_CLEAR    = 304
	ID_LOG_OPEN        = 400
	ID_LOG_EXPORT      = 401
	ID_ACTIVE_STATUS   = 500
	ID_PICK_SHIFT      = 510
	ID_PICK_DEDUCT     = 511
	ID_PICK_REQUIRE    = 512
	ID_PICK_ALLSITE    = 513
	ID_PICK_1C1L       = 514
	ID_PICK_INCOMPLETE = 515
	ID_PICK_1C1LERR    = 516
	ID_PACK_SHIFT      = 520
	ID_SHIFT_MANUAL    = 530
)

const TIMER_METRICS = 9001

type WNDCLASSEX struct {
	CbSize                                   uint32
	Style                                    uint32
	LpfnWndProc                              uintptr
	CbClsExtra, CbWndExtra                   int32
	HInstance, HIcon, HCursor, HbrBackground uintptr
	LpszMenuName, LpszClassName              *uint16
	HIconSm                                  uintptr
}
type MSG struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             POINT
}
type POINT struct{ X, Y int32 }
type RECT struct{ Left, Top, Right, Bottom int32 }
type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}
type NMHDR struct {
	HwndFrom, IdFrom uintptr
	Code             uint32
}
type NMLISTVIEW struct {
	Hdr                            NMHDR
	IItem, ISubItem                int32
	UNewState, UOldState, UChanged uint32
	PtAction                       POINT
	LParam                         uintptr
}
type LVCOLUMN struct {
	Mask                                 uint32
	Fmt, Cx                              int32
	PszText                              *uint16
	CchTextMax, ISubItem, IImage, IOrder int32
	CxMin, CxDefault, CxIdeal            int32
}
type LVITEM struct {
	Mask               uint32
	IItem, ISubItem    int32
	State, StateMask   uint32
	PszText            *uint16
	CchTextMax, IImage int32
	LParam             uintptr
	IIndent            int32
	IGroupId           uint32
	CColumns           uint32
	PuColumns          *uint32
	PiColFmt           *int32
	IGroup             int32
}
type DATA_BLOB struct {
	CbData uint32
	PbData *byte
}
type MEMORYSTATUSEX struct {
	Length, MemoryLoad                                                                                   uint32
	TotalPhys, AvailPhys, TotalPageFile, AvailPageFile, TotalVirtual, AvailVirtual, AvailExtendedVirtual uint64
}

type OPENFILENAME struct {
	LStructSize       uint32
	HwndOwner         uintptr
	HInstance         uintptr
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
	PvReserved        uintptr
	DwReserved        uint32
	FlagsEx           uint32
}

type credentials struct {
	URL, Method, Authorization, Token, APISID, USID, Signature, Nonce, Body, UserAgent, ImportedAt string
	OtherHeaders                                                                                   map[string]string
}
type appSettings struct {
	DataFolder string                `json:"data_folder"`
	Business   core.BusinessSettings `json:"business"`
}
type runtimeBinding = liveio.Binding
type runtimeProfile struct {
	SchemaVersion int                       `json:"schema_version"`
	ProfileID     string                    `json:"profile_id"`
	Bindings      map[string]runtimeBinding `json:"bindings"`
}
type liveState struct {
	Pick, Pack, Shift, Active, UserPDA core.Table
	Payroll                            []core.PayrollRow
	LastSync                           time.Time
	LastError                          string
	Route                              string
}

type tableModel struct {
	hwnd    uintptr
	data    core.Table
	kinds   []colKind
	sortCol int
	asc     bool
}
type colKind int

const (
	kindText colKind = iota
	kindNumber
	kindPercent
	kindDate
	kindTenure
	kindBool
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	gdi32                = syscall.NewLazyDLL("gdi32.dll")
	comctl32             = syscall.NewLazyDLL("comctl32.dll")
	crypt32              = syscall.NewLazyDLL("crypt32.dll")
	shell32              = syscall.NewLazyDLL("shell32.dll")
	psapi                = syscall.NewLazyDLL("psapi.dll")
	comdlg32              = syscall.NewLazyDLL("comdlg32.dll")
	pRegisterClass       = user32.NewProc("RegisterClassExW")
	pCreateWindow        = user32.NewProc("CreateWindowExW")
	pDefWindowProc       = user32.NewProc("DefWindowProcW")
	pShowWindow          = user32.NewProc("ShowWindow")
	pUpdateWindow        = user32.NewProc("UpdateWindow")
	pGetMessage          = user32.NewProc("GetMessageW")
	pTranslateMessage    = user32.NewProc("TranslateMessage")
	pDispatchMessage     = user32.NewProc("DispatchMessageW")
	pPostQuit            = user32.NewProc("PostQuitMessage")
	pPostMessage         = user32.NewProc("PostMessageW")
	pSendMessage         = user32.NewProc("SendMessageW")
	pDestroyWindow       = user32.NewProc("DestroyWindow")
	pGetClientRect       = user32.NewProc("GetClientRect")
	pMoveWindow          = user32.NewProc("MoveWindow")
	pSetWindowText       = user32.NewProc("SetWindowTextW")
	pGetWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	pGetWindowText       = user32.NewProc("GetWindowTextW")
	pLoadCursor          = user32.NewProc("LoadCursorW")
	pSetTimer            = user32.NewProc("SetTimer")
	pKillTimer           = user32.NewProc("KillTimer")
	pMessageBox          = user32.NewProc("MessageBoxW")
	pInvalidateRect      = user32.NewProc("InvalidateRect")
	pMonitorFromWindow   = user32.NewProc("MonitorFromWindow")
	pGetMonitorInfo      = user32.NewProc("GetMonitorInfoW")
	pCreateFont          = gdi32.NewProc("CreateFontW")
	pInitCommon          = comctl32.NewProc("InitCommonControls")
	pCryptProtect        = crypt32.NewProc("CryptProtectData")
	pCryptUnprotect      = crypt32.NewProc("CryptUnprotectData")
	pLocalFree           = kernel32.NewProc("LocalFree")
	pGlobalMemory        = kernel32.NewProc("GlobalMemoryStatusEx")
	pShellExecute        = shell32.NewProc("ShellExecuteW")
	pGetProcMem          = psapi.NewProc("GetProcessMemoryInfo")
	pGetCurrentProcess   = kernel32.NewProc("GetCurrentProcess")
	pGetSaveFileName     = comdlg32.NewProc("GetSaveFileNameW")
)

var (
	mainWnd, statusText, headerText, tableWnd, settingsCurl, settingsSummary, logEdit, overviewMetric uintptr
	nav                                                                                               = map[int]uintptr{}
	pageControls                                                                                      []uintptr
	fontNormal, fontSmall, fontBold                                                                   uintptr
	currentPage                                                                                       = ID_NAV_OVERVIEW
	currentTable                                                                                      *tableModel
	settings                                                                                          appSettings
	creds                                                                                             credentials
	profile                                                                                           runtimeProfile
	revealSecrets                                                                                     bool
	live                                                                                              liveState
	liveMu                                                                                            sync.RWMutex
	busy                                                                                              atomic.Bool
	pendingStatus                                                                                     string
	statusMu                                                                                          sync.Mutex
	activeCombo, pickShiftCombo, packShiftCombo, shiftManualCombo                                     uintptr
	settingsSourceCombo                                                                               uintptr
	syncButton, updateButton                                                                            uintptr
	navOrder                                                                                             = []int{ID_NAV_OVERVIEW, ID_NAV_ACTIVE, ID_NAV_PICK, ID_NAV_PACK, ID_NAV_SHIFT, ID_NAV_USERPDA, ID_NAV_LOG, ID_NAV_SETTINGS}
	navLabels                                                                                            = map[int]string{ID_NAV_OVERVIEW: "TỔNG QUAN", ID_NAV_ACTIVE: "ĐANG LẤY HÀNG", ID_NAV_PICK: "PICK", ID_NAV_PACK: "PACK", ID_NAV_SHIFT: "PHÂN CA", ID_NAV_USERPDA: "USER / PDA", ID_NAV_LOG: "LOG", ID_NAV_SETTINGS: "THIẾT LẬP"}
	logQueue                                                                                             = make(chan string, 4096)
	logFlush                                                                                             = make(chan chan struct{})
	logDropped                                                                                           atomic.Uint64
	rendering                                                                                            atomic.Bool
	lastTelemetry                                                                                        time.Time
	fittingWindow                                                                                        atomic.Bool
	uiPingSeq                                                                                            atomic.Uint64
	uiPongSeq                                                                                            atomic.Uint64
	metricsBusy                                                                                          atomic.Bool
	metricText                                                                                           string
	metricMu                                                                                             sync.Mutex
	pendingLogText                                                                                       string
	pendingLogMu                                                                                         sync.Mutex
	logExportBusy                                                                                        atomic.Bool
	windowWidth                                                                                          atomic.Int64
	windowHeight                                                                                         atomic.Int64
	currentTableTop                                                                                      = 170
	currentLogTop                                                                                        = 170
)

func maxInt(a, b int) int { if a > b { return a }; return b }

func clientSize() (int, int) {
	var rc RECT
	pGetClientRect.Call(mainWnd, uintptr(unsafe.Pointer(&rc)))
	return int(rc.Right), int(rc.Bottom)
}

func groupBox(title string, x, y, w, h int) uintptr {
	g := create("BUTTON", title, WS_CHILD|WS_VISIBLE|BS_GROUPBOX, x, y, w, h, mainWnd, 0)
	setFont(g, fontSmall)
	addPage(g)
	return g
}

func sectionLabel(text string, x, y, w int) uintptr {
	c := static(text, x, y, w, 24, true)
	return c
}

func requestOverviewMetrics() {
	if currentPage != ID_NAV_OVERVIEW || !metricsBusy.CompareAndSwap(false, true) { return }
	go func() {
		defer metricsBusy.Store(false)
		var ms MEMORYSTATUSEX
		ms.Length = uint32(unsafe.Sizeof(ms))
		pGlobalMemory.Call(uintptr(unsafe.Pointer(&ms)))
		var rm runtime.MemStats
		runtime.ReadMemStats(&rm)
		used := ms.TotalPhys - ms.AvailPhys
		txt := fmt.Sprintf("RAM máy: %.1f / %.1f GB (%d%%)   •   RAM ứng dụng: %.1f MB   •   Goroutine: %d",
			float64(used)/1e9, float64(ms.TotalPhys)/1e9, int(ms.MemoryLoad),
			float64(rm.Alloc)/1024/1024, runtime.NumGoroutine())
		metricMu.Lock()
		metricText = txt
		metricMu.Unlock()
		if mainWnd != 0 { pPostMessage.Call(mainWnd, WM_APP_METRICS, 0, 0) }
	}()
}

func loadLogTailAsync() {
	go func() {
		text := readLogTail(220)
		pendingLogMu.Lock()
		pendingLogText = text
		pendingLogMu.Unlock()
		if mainWnd != 0 { pPostMessage.Call(mainWnd, WM_APP_LOGTEXT, 0, 0) }
	}()
}

func fitWindowToWorkArea() {
	if mainWnd == 0 || !fittingWindow.CompareAndSwap(false, true) { return }
	defer fittingWindow.Store(false)
	monitor, _, _ := pMonitorFromWindow.Call(mainWnd, MONITOR_DEFAULTTONEAREST)
	if monitor == 0 { return }
	mi := MONITORINFO{CbSize: uint32(unsafe.Sizeof(MONITORINFO{}))}
	ok, _, _ := pGetMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&mi)))
	if ok == 0 { return }
	w := int(mi.RcWork.Right - mi.RcWork.Left)
	h := int(mi.RcWork.Bottom - mi.RcWork.Top)
	if w <= 0 || h <= 0 { return }
	pMoveWindow.Call(mainWnd,
		uintptr(mi.RcWork.Left), uintptr(mi.RcWork.Top),
		uintptr(w), uintptr(h), 1)
	logEvent("INFO", "WINDOW_FIT_WORKAREA",
		"x", strconv.Itoa(int(mi.RcWork.Left)),
		"y", strconv.Itoa(int(mi.RcWork.Top)),
		"w", strconv.Itoa(w),
		"h", strconv.Itoa(h))
}

func uiWatchdog() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	var lastReported uint64
	for range ticker.C {
		if mainWnd == 0 { continue }
		seq := uiPingSeq.Add(1)
		pPostMessage.Call(mainWnd, WM_APP_PING, uintptr(seq), 0)
		time.Sleep(5 * time.Second)
		if uiPongSeq.Load() >= seq { continue }
		if lastReported == seq { continue }
		lastReported = seq
		buf := make([]byte, 128*1024)
		n := runtime.Stack(buf, true)
		logEvent("ERROR", "UI_WATCHDOG_STALL",
			"ping_seq", strconv.FormatUint(seq, 10),
			"page", strconv.Itoa(currentPage),
			"stack", string(buf[:n]))
	}
}

func ptr(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func create(class, text string, style uint32, x, y, w, h int, parent, menu uintptr) uintptr {
	r, _, _ := pCreateWindow.Call(0, uintptr(unsafe.Pointer(ptr(class))), uintptr(unsafe.Pointer(ptr(text))), uintptr(style), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, menu, 0, 0)
	return r
}
func setFont(h, f uintptr) { pSendMessage.Call(h, WM_SETFONT, f, 1) }
func setText(h uintptr, s string) {
	if h != 0 {
		pSetWindowText.Call(h, uintptr(unsafe.Pointer(ptr(s))))
	}
}
func getText(h uintptr) string {
	n, _, _ := pGetWindowTextLength.Call(h)
	if n == 0 { return "" }
	b := make([]uint16, n+1)
	pGetWindowText.Call(h, uintptr(unsafe.Pointer(&b[0])), n+1)
	return syscall.UTF16ToString(b)
}
func move(h uintptr, x, y, w, hh int) {
	if h != 0 { pMoveWindow.Call(h, uintptr(x), uintptr(y), uintptr(w), uintptr(hh), 1) }
}
func addPage(h uintptr) { pageControls = append(pageControls, h) }
func destroyPage() {
	for _, h := range pageControls { pDestroyWindow.Call(h) }
	pageControls = nil
	tableWnd = 0
	currentTable = nil
	settingsCurl = 0
	settingsSummary = 0
	logEdit = 0
	activeCombo = 0
	pickShiftCombo = 0
	packShiftCombo = 0
	shiftManualCombo = 0
	settingsSourceCombo = 0
	overviewMetric = 0
	currentTableTop = 170
	currentLogTop = 170
}
func static(text string, x, y, w, h int, bold bool) uintptr {
	c := create("STATIC", text, WS_CHILD|WS_VISIBLE|SS_LEFT, x, y, w, h, mainWnd, 0)
	if bold { setFont(c, fontBold) } else { setFont(c, fontNormal) }
	addPage(c)
	return c
}
func button(id int, text string, x, y, w, h int) uintptr {
	b := create("BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, x, y, w, h, mainWnd, uintptr(id))
	setFont(b, fontNormal); addPage(b); return b
}
func checkbox(id int, text string, x, y, w, h int, checked bool) uintptr {
	b := create("BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, x, y, w, h, mainWnd, uintptr(id))
	setFont(b, fontSmall)
	if checked { pSendMessage.Call(b, BM_SETCHECK, BST_CHECKED, 0) }
	addPage(b); return b
}
func checked(h uintptr) bool {
	r, _, _ := pSendMessage.Call(h, BM_GETCHECK, 0, 0)
	return r == BST_CHECKED
}
func combo(id int, items []string, selected string, x, y, w, h int) uintptr {
	c := create("COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST, x, y, w, h, mainWnd, uintptr(id))
	setFont(c, fontNormal)
	sel := 0
	for i, s := range items {
		pSendMessage.Call(c, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(s))))
		if strings.EqualFold(s, selected) { sel = i }
	}
	pSendMessage.Call(c, CB_SETCURSEL, uintptr(sel), 0)
	addPage(c); return c
}
func comboText(h uintptr) string {
	r, _, _ := pSendMessage.Call(h, CB_GETCURSEL, 0, 0)
	if int32(r) < 0 { return "" }
	n, _, _ := pSendMessage.Call(h, CB_GETLBTEXTLEN, r, 0)
	b := make([]uint16, n+2)
	pSendMessage.Call(h, CB_GETLBTEXT, r, uintptr(unsafe.Pointer(&b[0])))
	return syscall.UTF16ToString(b)
}

func addListColumn(lv uintptr, idx int, text string, width int) {
	c := LVCOLUMN{Mask: LVCF_FMT | LVCF_WIDTH | LVCF_TEXT, Fmt: LVCFMT_LEFT, Cx: int32(width), PszText: ptr(text)}
	pSendMessage.Call(lv, LVM_INSERTCOLUMNW, uintptr(idx), uintptr(unsafe.Pointer(&c)))
}
func addListRow(lv uintptr, row int, vals []string) {
	if len(vals) == 0 { return }
	it := LVITEM{Mask: LVIF_TEXT, IItem: int32(row), PszText: ptr(vals[0])}
	pSendMessage.Call(lv, LVM_INSERTITEMW, 0, uintptr(unsafe.Pointer(&it)))
	for i := 1; i < len(vals); i++ {
		it2 := LVITEM{ISubItem: int32(i), PszText: ptr(vals[i])}
		pSendMessage.Call(lv, LVM_SETITEMTEXTW, uintptr(row), uintptr(unsafe.Pointer(&it2)))
	}
}

func inferKind(h string) colKind {
	n := strings.ToLower(strings.TrimSpace(h))
	switch {
	case strings.Contains(n, "tuổi nghề") || strings.Contains(n, "tuoi nghe"):
		return kindTenure
	case strings.Contains(n, "tỉ lệ") || strings.Contains(n, "tiến độ") || strings.Contains(n, "%") || strings.Contains(n, "ty le"):
		return kindPercent
	case strings.Contains(n, "thời gian bắt đầu") || strings.Contains(n, "thời gian kết thúc đơn") && !strings.Contains(n, "cuối cùng"):
		return kindDate
	case strings.HasPrefix(n, "do ") || strings.HasPrefix(n, "sl ") || strings.Contains(n, "nsld") || strings.Contains(n, "mã nhân viên") || strings.Contains(n, "site") || strings.Contains(n, "tốc độ"):
		return kindNumber
	case strings.Contains(n, "bỏ qua") || strings.Contains(n, "kiểm tra 20"):
		return kindBool
	}
	return kindText
}
func formatCell(v any, k colKind) string {
	if v == nil { return "" }
	switch k {
	case kindPercent:
		f, ok := core.ToFloat(v); if ok { return fmt.Sprintf("%.1f%%", f*100) }
	case kindDate:
		switch x := v.(type) {
		case time.Time:
			if x.IsZero() { return "" }
			return x.Format("02/01/2006 15:04:05")
		}
	case kindBool:
		if b, ok := v.(bool); ok { if b { return "Có" }; return "Không" }
	case kindNumber:
		if f, ok := core.ToFloat(v); ok {
			if f == float64(int64(f)) { return strconv.FormatInt(int64(f), 10) }
			return strconv.FormatFloat(f, 'f', 2, 64)
		}
	}
	return fmt.Sprint(v)
}

func compare(a, b any, k colKind) int {
	switch k {
	case kindNumber, kindPercent:
		af, _ := core.ToFloat(a); bf, _ := core.ToFloat(b)
		if af < bf { return -1 }; if af > bf { return 1 }; return 0
	case kindTenure:
		ad := core.ParseTenureDays(fmt.Sprint(a)); bd := core.ParseTenureDays(fmt.Sprint(b))
		if ad < bd { return -1 }; if ad > bd { return 1 }; return 0
	case kindDate:
		at, _ := a.(time.Time); bt, _ := b.(time.Time)
		if at.Before(bt) { return -1 }; if at.After(bt) { return 1 }; return 0
	}
	return strings.Compare(strings.ToLower(fmt.Sprint(a)), strings.ToLower(fmt.Sprint(b)))
}
func renderTable(t core.Table, top int) {
	start := time.Now()
	currentTableTop = top
	var rc RECT
	pGetClientRect.Call(mainWnd, uintptr(unsafe.Pointer(&rc)))
	w := int(rc.Right) - 48
	h := int(rc.Bottom) - top - 82
	if w < 760 { w = 760 }
	if h < 240 { h = 240 }
	lv := create("SysListView32", "", WS_CHILD|WS_VISIBLE|WS_BORDER|LVS_REPORT|LVS_SHOWSELALWAYS|WS_HSCROLL|WS_VSCROLL, 24, top, w, h, mainWnd, 0)
	setFont(lv, fontNormal)
	pSendMessage.Call(lv, LVM_SETEXTENDEDLISTVIEWSTYLE, 0, LVS_EX_FULLROWSELECT|LVS_EX_GRIDLINES|LVS_EX_DOUBLEBUFFER)
	addPage(lv)
	kinds := make([]colKind, len(t.Headers))
	for i, header := range t.Headers {
		kinds[i] = inferKind(header)
		width := 115
		if strings.Contains(strings.ToLower(header), "họ") || strings.Contains(strings.ToLower(header), "thời gian") { width = 180 }
		addListColumn(lv, i, header, width)
	}
	currentTable = &tableModel{hwnd: lv, data: t, kinds: kinds, sortCol: -1, asc: true}
	tableWnd = lv
	fillTable()
	logEvent("INFO", "TABLE_RENDER_DONE",
		"page", strconv.Itoa(currentPage),
		"rows", strconv.Itoa(len(t.Rows)),
		"columns", strconv.Itoa(len(t.Headers)),
		"elapsed_ms", strconv.FormatInt(time.Since(start).Milliseconds(), 10))
}

func fillTable() {
	if currentTable == nil { return }
	start := time.Now()
	pSendMessage.Call(currentTable.hwnd, WM_SETREDRAW, 0, 0)
	defer func() {
		pSendMessage.Call(currentTable.hwnd, WM_SETREDRAW, 1, 0)
		pInvalidateRect.Call(currentTable.hwnd, 0, 1)
		elapsed := time.Since(start)
		if elapsed > 300*time.Millisecond {
			logEvent("WARN", "TABLE_FILL_SLOW",
				"page", strconv.Itoa(currentPage),
				"rows", strconv.Itoa(len(currentTable.data.Rows)),
				"elapsed_ms", strconv.FormatInt(elapsed.Milliseconds(), 10))
		}
	}()
	pSendMessage.Call(currentTable.hwnd, LVM_DELETEALLITEMS, 0, 0)
	for r, row := range currentTable.data.Rows {
		vals := make([]string, len(currentTable.data.Headers))
		for i := range vals {
			if i < len(row) { vals[i] = formatCell(row[i], currentTable.kinds[i]) }
		}
		addListRow(currentTable.hwnd, r, vals)
	}
}

func sortTable(col int) {
	if currentTable == nil || col < 0 || col >= len(currentTable.data.Headers) { return }
	if currentTable.sortCol == col { currentTable.asc = !currentTable.asc } else { currentTable.sortCol = col; currentTable.asc = true }
	k := currentTable.kinds[col]; asc := currentTable.asc
	sort.SliceStable(currentTable.data.Rows, func(i, j int) bool {
		var a, b any
		if col < len(currentTable.data.Rows[i]) { a = currentTable.data.Rows[i][col] }
		if col < len(currentTable.data.Rows[j]) { b = currentTable.data.Rows[j][col] }
		c := compare(a, b, k)
		if asc { return c < 0 }; return c > 0
	})
	fillTable()
	logEvent("INFO", "TABLE_SORT", "column", currentTable.data.Headers[col], "ascending", strconv.FormatBool(asc))
}
func selectedRow() ([]any, bool) {
	if currentTable == nil { return nil, false }
	r, _, _ := pSendMessage.Call(currentTable.hwnd, LVM_GETNEXTITEM, ^uintptr(0), LVNI_SELECTED)
	if int32(r) < 0 || int(r) >= len(currentTable.data.Rows) { return nil, false }
	return currentTable.data.Rows[int(r)], true
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) (ret uintptr) {
	defer func() {
		if v := recover(); v != nil {
			logEvent("ERROR", "UI_PANIC", "message", fmt.Sprint(v), "stack", string(debug.Stack()))
			setStatus("Ứng dụng vừa chặn một lỗi giao diện. Mở LOG để kiểm tra.")
			ret = 0
		}
	}()
	switch msg {
	case WM_CREATE:
		mainWnd = hwnd
		createShell()
		loadSettings()
		provisionRuntimeProfile()
		loadRuntimeProfile()
		loadCredentials()
		if sessionReady() && strings.TrimSpace(creds.URL) != "" {
			if err := ensureDashboardBindings(); err != nil {
				logEvent("WARN", "DASHBOARD_BINDING_MIGRATION_FAILED", "error", err.Error())
			} else {
				logEvent("INFO", "DASHBOARD_BINDINGS_READY", "binding_count", "2")
			}
		}
		lastTelemetry = time.Now()
		logEvent("INFO", "APP_START", "version", appVersion)
		go logRuntimeSnapshot("APP_STATE_START")
		renderPage(currentPage)
		pSetTimer.Call(hwnd, TIMER_METRICS, 2000, 0)
		go checkUpdateQuiet()
		return 0

	case WM_SYSCOMMAND:
		cmd := wParam & 0xFFF0
		// Never force SW_MAXIMIZE from inside WM_SYSCOMMAND. That can create an
		// unstable Win32 sizing/message loop on some Windows builds.
		// The app is instead pinned to the monitor work area (taskbar excluded).
		if cmd == SC_SIZE || cmd == SC_MOVE || cmd == SC_MAXIMIZE {
			return 0
		}

	case WM_SIZE:
		layout()
		if wParam == SIZE_RESTORED && !fittingWindow.Load() {
			pPostMessage.Call(hwnd, WM_APP_FIT_WORKAREA, 0, 0)
		}
		return 0

	case WM_SETTINGCHANGE, WM_DISPLAYCHANGE:
		pPostMessage.Call(hwnd, WM_APP_FIT_WORKAREA, 0, 0)
		return 0

	case WM_COMMAND:
		id := int(uint16(wParam & 0xffff))
		code := int(uint16((wParam >> 16) & 0xffff))
		start := time.Now()
		handleCommand(id, code, lParam)
		elapsed := time.Since(start)
		if elapsed > 250*time.Millisecond {
			logEvent("WARN", "UI_COMMAND_SLOW", "id", strconv.Itoa(id), "code", strconv.Itoa(code), "elapsed_ms", strconv.FormatInt(elapsed.Milliseconds(), 10))
		}
		return 0

	case WM_NOTIFY:
		return handleNotify(lParam)

	case WM_TIMER:
		if wParam == TIMER_METRICS {
			if currentPage == ID_NAV_OVERVIEW { requestOverviewMetrics() }
			if time.Since(lastTelemetry) >= 30*time.Second {
				lastTelemetry = time.Now()
				go logRuntimeSnapshot("PERF_SAMPLE")
			}
			return 0
		}

	case WM_APP_STATUS:
		statusMu.Lock()
		text := pendingStatus
		statusMu.Unlock()
		setText(statusText, text)
		return 0

	case WM_APP_REFRESH:
		renderPage(currentPage)
		return 0

	case WM_APP_FIT_WORKAREA:
		fitWindowToWorkArea()
		layout()
		return 0

	case WM_APP_PING:
		uiPongSeq.Store(uint64(wParam))
		return 0

	case WM_APP_METRICS:
		metricMu.Lock()
		text := metricText
		metricMu.Unlock()
		if currentPage == ID_NAV_OVERVIEW && overviewMetric != 0 { setText(overviewMetric, text) }
		return 0

	case WM_APP_LOGTEXT:
		pendingLogMu.Lock()
		text := pendingLogText
		pendingLogMu.Unlock()
		if currentPage == ID_NAV_LOG && logEdit != 0 { setText(logEdit, text) }
		return 0

	case WM_DESTROY:
		pKillTimer.Call(hwnd, TIMER_METRICS)
		logEvent("INFO", "APP_EXIT")
		flushLogs(1200 * time.Millisecond)
		pPostQuit.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func createShell() {
	fontNormal, _, _ = pCreateFont.Call(18, 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	fontSmall, _, _ = pCreateFont.Call(15, 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	fontBold, _, _ = pCreateFont.Call(21, 0, 0, 0, 600, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(ptr("Segoe UI Semibold"))))

	headerText = create("STATIC", "SUPRA PRODUCTIVITY   ·   "+appVersion, WS_CHILD|WS_VISIBLE|SS_LEFT, 24, 15, 720, 30, mainWnd, 0)
	setFont(headerText, fontBold)
	statusText = create("STATIC", "● Sẵn sàng", WS_CHILD|WS_VISIBLE|SS_LEFT, 760, 18, 420, 24, mainWnd, 0)
	setFont(statusText, fontNormal)

	updateButton = create("BUTTON", "CẬP NHẬT", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 0, 12, 110, 34, mainWnd, ID_UPDATE)
	setFont(updateButton, fontSmall)
	syncButton = create("BUTTON", "ĐỒNG BỘ", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 0, 12, 120, 34, mainWnd, ID_SYNC)
	setFont(syncButton, fontSmall)

	for i, id := range navOrder {
		style := uint32(WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTORADIOBUTTON|BS_PUSHLIKE)
		if i == 0 { style |= WS_GROUP }
		b := create("BUTTON", navLabels[id], style, 0, 0, 140, 40, mainWnd, uintptr(id))
		setFont(b, fontSmall)
		nav[id] = b
	}
}

func layout() {
	var rc RECT
	pGetClientRect.Call(mainWnd, uintptr(unsafe.Pointer(&rc)))
	w, h := int(rc.Right), int(rc.Bottom)
	if w <= 0 || h <= 0 { return }
	windowWidth.Store(int64(w))
	windowHeight.Store(int64(h))

	move(headerText, 24, 15, maxInt(420, w-790), 30)
	move(statusText, maxInt(520, w-740), 18, 430, 24)
	move(updateButton, maxInt(20, w-252), 12, 108, 34)
	move(syncButton, maxInt(132, w-136), 12, 120, 34)

	navY := h - 56
	left, gap := 12, 5
	avail := w - left*2 - gap*(len(navOrder)-1)
	bw := avail / len(navOrder)
	if bw < 96 { bw = 96 }
	for i, id := range navOrder {
		move(nav[id], left+i*(bw+gap), navY, bw, 40)
	}

	if tableWnd != 0 {
		move(tableWnd, 24, currentTableTop, maxInt(760, w-48), maxInt(220, navY-currentTableTop-14))
	}
	if logEdit != 0 {
		move(logEdit, 24, currentLogTop, maxInt(760, w-48), maxInt(220, navY-currentLogTop-14))
	}
}

func pageTitle(title, sub string) {
	static(title, 24, 62, 620, 30, true)
	static(sub, 24, 94, 1280, 22, false)
}

func renderPage(id int) {
	if !rendering.CompareAndSwap(false, true) {
		logEvent("WARN", "UI_RENDER_SKIPPED_REENTRANT", "requested_page", strconv.Itoa(id), "current_page", strconv.Itoa(currentPage))
		return
	}
	defer rendering.Store(false)
	start := time.Now()
	previous := currentPage
	currentPage = id
	for navID, h := range nav {
		state := uintptr(0)
		if navID == id { state = BST_CHECKED }
		pSendMessage.Call(h, BM_SETCHECK, state, 0)
	}
	logEvent("INFO", "UI_PAGE_BEGIN", "from", strconv.Itoa(previous), "to", strconv.Itoa(id))
	destroyPage()
	switch id {
	case ID_NAV_OVERVIEW: renderOverview()
	case ID_NAV_ACTIVE: renderActive()
	case ID_NAV_PICK: renderPick()
	case ID_NAV_PACK: renderPack()
	case ID_NAV_SHIFT: renderShift()
	case ID_NAV_USERPDA: renderUserPDA()
	case ID_NAV_LOG: renderLog()
	case ID_NAV_SETTINGS: renderSettings()
	}
	layout()
	elapsed := time.Since(start)
	level := "INFO"
	if elapsed > 500*time.Millisecond { level = "WARN" }
	rows, cols := 0, 0
	if currentTable != nil {
		rows, cols = len(currentTable.data.Rows), len(currentTable.data.Headers)
	}
	logEvent(level, "UI_PAGE_DONE", "page", strconv.Itoa(id), "rows", strconv.Itoa(rows), "columns", strconv.Itoa(cols), "elapsed_ms", strconv.FormatInt(elapsed.Milliseconds(), 10))
}

func renderOverview() {
	pageTitle("TỔNG QUAN", "Tình trạng vận hành, dữ liệu và hiệu năng ứng dụng.")
	liveMu.RLock()
	ls := live
	liveMu.RUnlock()
	w, _ := clientSize()
	cardW := (w - 72) / 2
	if cardW < 520 { cardW = 520 }

	groupBox("SẢN LƯỢNG HIỆN TẠI", 24, 126, cardW, 142)
	static(fmt.Sprintf("PICK   %d", len(ls.Pick.Rows)), 48, 164, 180, 30, true)
	static(fmt.Sprintf("PACK   %d", len(ls.Pack.Rows)), 250, 164, 180, 30, true)
	static(fmt.Sprintf("PHÂN CA   %d", len(ls.Shift.Rows)), 452, 164, 220, 30, true)
	static(fmt.Sprintf("ĐANG LẤY HÀNG   %d", len(ls.Active.Rows)), 48, 208, 260, 26, false)

	x2 := 48 + cardW
	groupBox("HỆ THỐNG", x2, 126, cardW, 142)
	overviewMetric = static("Đang đo hiệu năng…", x2+24, 164, cardW-48, 28, false)
	static("Dữ liệu hợp lệ được giữ lại khi mạng hoặc dịch vụ tạm thời lỗi.", x2+24, 208, cardW-48, 24, false)
	requestOverviewMetrics()

	groupBox("TRẠNG THÁI ĐỒNG BỘ", 24, 286, w-48, 104)
	lastSync := "Chưa đồng bộ trong phiên này"
	if !ls.LastSync.IsZero() { lastSync = ls.LastSync.Format("02/01/2006 15:04:05") }
	static("Lần đồng bộ: "+lastSync, 48, 322, 420, 24, false)
	route := ls.Route
	if route == "" { route = "Chưa xác định" }
	static("Kênh dữ liệu: "+route, 500, 322, 400, 24, false)
	if ls.LastError != "" {
		static("Lỗi gần nhất: "+ls.LastError, 48, 352, w-96, 24, false)
	} else {
		static("Trạng thái: Không ghi nhận lỗi đồng bộ.", 48, 352, w-96, 24, false)
	}
}

func renderOverviewMetrics() {
	requestOverviewMetrics()
}

func renderActive() {
	pageTitle("ĐANG LẤY HÀNG", "Theo dõi tiến độ hiện tại; nhấp đúp một dòng để xem chi tiết.")
	groupBox("BỘ LỌC", 24, 124, 520, 66)
	static("Trạng thái", 46, 151, 85, 22, false)
	activeCombo = combo(ID_ACTIVE_STATUS, []string{"Tất cả", "Đang lấy", "Quá thời gian", "Hoàn thành"}, settings.Business.ActiveStatus, 142, 144, 180, 200)
	liveMu.RLock()
	t := live.Active
	liveMu.RUnlock()
	renderTable(filterActive(t, settings.Business.ActiveStatus), 204)
}

func renderPick() {
	pageTitle("PICK", "Theo dõi năng suất, target và các điều kiện kiểm tra chẵn/lẻ.")
	w, _ := clientSize()
	groupBox("BỘ LỌC & QUY TẮC", 24, 124, w-48, 78)
	static("Ca", 46, 153, 28, 22, false)
	pickShiftCombo = combo(ID_PICK_SHIFT, []string{"Tất cả", "Ca 1", "Ca 2", "Ca HC"}, settings.Business.PickShift, 82, 146, 112, 180)
	checkbox(ID_PICK_DEDUCT, "Khấu trừ SKU", 220, 149, 125, 24, settings.Business.PickDeductSKU)
	checkbox(ID_PICK_REQUIRE, "Kiểm tra đủ chẵn", 360, 149, 145, 24, settings.Business.PickRequireEven)
	checkbox(ID_PICK_ALLSITE, "Tất cả Site", 520, 149, 105, 24, settings.Business.ShowAllSite)
	checkbox(ID_PICK_1C1L, "1 chẵn 1 lẻ", 640, 149, 115, 24, settings.Business.Enable1C1L)
	checkbox(ID_PICK_INCOMPLETE, "Chưa đủ chẵn", 770, 149, 130, 24, settings.Business.ShowIncompleteEven)
	checkbox(ID_PICK_1C1LERR, "Lỗi 1C1L", 915, 149, 110, 24, settings.Business.Show1C1LErrors)
	liveMu.RLock()
	t := live.Pick
	liveMu.RUnlock()
	renderTable(t, 216)
}

func renderPack() {
	pageTitle("PACK", "Theo dõi năng suất đóng gói theo ca.")
	groupBox("BỘ LỌC", 24, 124, 430, 66)
	static("Ca hiển thị", 46, 151, 92, 22, false)
	packShiftCombo = combo(ID_PACK_SHIFT, []string{"Tất cả", "Ca 1", "Ca 2", "Ca HC"}, settings.Business.PackShift, 148, 144, 132, 180)
	liveMu.RLock()
	t := live.Pack
	liveMu.RUnlock()
	renderTable(t, 204)
}

func renderShift() {
	pageTitle("PHÂN CA", "Phân ca tự động và điều chỉnh thủ công khi cần.")
	groupBox("ĐIỀU CHỈNH CA", 24, 124, 500, 66)
	static("Phân ca", 46, 151, 70, 22, false)
	shiftManualCombo = combo(ID_SHIFT_MANUAL, []string{"Tự động", "Ca 1", "Ca 2", "Ca HC"}, "Tự động", 128, 144, 150, 180)
	liveMu.RLock()
	t := live.Shift
	liveMu.RUnlock()
	renderTable(t, 204)
}

func renderUserPDA() {
	pageTitle("USER / PDA", "Danh sách user, nhân sự và thông tin phục vụ vận hành.")
	liveMu.RLock()
	t := live.UserPDA
	liveMu.RUnlock()
	if len(t.Headers) == 0 {
		t = core.Table{Headers: []string{"Họ và tên", "Mã nhân viên", "User", "Nhà cung cấp", "Site", "Tuổi nghề"}}
	}
	renderTable(t, 128)
}

func renderLog() {
	pageTitle("LOG", "Nhật ký vận hành và lỗi đã được loại thông tin nhạy cảm.")
	w, _ := clientSize()
	groupBox("CÔNG CỤ LOG", 24, 124, w-48, 72)
	button(ID_LOG_OPEN, "MỞ THƯ MỤC LOG", 46, 149, 178, 34)
	button(ID_LOG_EXPORT, "XUẤT LOG...", 238, 149, 138, 34)
	static("File xuất gồm toàn bộ log cần thiết để phân tích lỗi; màn hình chỉ tải phần gần nhất.", 398, 154, w-440, 22, false)
	currentLogTop = 210
	logEdit = create("EDIT", "Đang tải log gần nhất…", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY, 24, currentLogTop, w-48, 560, mainWnd, 0)
	setFont(logEdit, fontSmall)
	addPage(logEdit)
	loadLogTailAsync()
}

func renderSettings() {
	pageTitle("THIẾT LẬP", "Nhận một phiên Dashboard; ứng dụng tự đồng bộ Đang lấy hàng và Sản lượng.")
	w, _ := clientSize()
	leftW := (w - 72) * 3 / 5
	if leftW < 720 { leftW = 720 }
	rightX := 48 + leftW
	rightW := w - rightX - 24
	if rightW < 420 { rightW = 420 }

	groupBox("CẤU HÌNH DASHBOARD", 24, 124, leftW, 360)
	static("Dán cURL (bash)", 46, 158, 170, 22, true)
	static("Chỉ cần dán một cURL hợp lệ của Dashboard. Ứng dụng dùng chung phiên để tải Đang lấy hàng và Sản lượng.", 46, 188, leftW-44, 42, false)
	settingsCurl = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL, 46, 238, leftW-44, 132, mainWnd, 0)
	setFont(settingsCurl, fontNormal)
	addPage(settingsCurl)
	button(ID_CURL_IMPORT, "LƯU CẤU HÌNH", 46, 386, 154, 34)
	button(ID_NET_TEST, "KIỂM TRA ĐỒNG BỘ", 216, 386, 176, 34)

	groupBox("TRẠNG THÁI KẾT NỐI", rightX, 124, rightW, 360)
	settingsSummary = static(credentialSummary(), rightX+24, 160, rightW-48, 210, false)
	button(ID_SECRET_TOGGLE, "HIỆN / ẨN PHIÊN", rightX+24, 386, 156, 34)
	static("Phiên và cấu hình chỉ lưu cục bộ theo Windows user; dữ liệu nhạy cảm không nằm trong bản phát hành public.", rightX+24, 434, rightW-48, 42, false)

	groupBox("LUỒNG XỬ LÝ", 24, 504, w-48, 116)
	static("Đang lấy hàng → dữ liệu live.   Sản lượng → Mapping → Phân ca → Pick / Pack.   User / PDA được hợp nhất từ dữ liệu nhân sự có trong hai luồng.", 46, 540, w-92, 42, true)
	static("Sau khi lưu cấu hình một lần, vận hành hàng ngày chỉ cần bấm ĐỒNG BỘ.", 46, 584, w-92, 24, false)
}

func filterActive(t core.Table, status string) core.Table {
	if status == "" || status == "Tất cả" { return t }
	idx := -1
	for i, h := range t.Headers { if strings.EqualFold(strings.TrimSpace(h), "Trạng thái") { idx = i; break } }
	if idx < 0 { return t }
	out := core.Table{Headers: append([]string(nil), t.Headers...)}
	for _, r := range t.Rows { if idx < len(r) && strings.EqualFold(strings.TrimSpace(fmt.Sprint(r[idx])), status) { out.Rows = append(out.Rows, r) } }
	return out
}

func handleNotify(lParam uintptr) uintptr {
	if lParam == 0 { return 0 }
	h := (*NMHDR)(unsafe.Pointer(lParam))
	if h == nil || currentTable == nil || h.HwndFrom != currentTable.hwnd { return 0 }
	n := (*NMLISTVIEW)(unsafe.Pointer(lParam)); code := int32(h.Code)
	if code == int32(LVN_COLUMNCLICK) { sortTable(int(n.ISubItem)); return 0 }
	if code == NM_DBLCLK { handleDoubleClick(n); return 0 }
	return 0
}

func handleDoubleClick(n *NMLISTVIEW) {
	row, ok := selectedRow(); if !ok { return }
	if currentPage == ID_NAV_PICK && n != nil && int(n.ISubItem) >= 0 && int(n.ISubItem) < len(currentTable.data.Headers) && strings.Contains(strings.ToLower(currentTable.data.Headers[n.ISubItem]), "bỏ qua kiểm tra 20") {
		user := ""
		for i, h := range currentTable.data.Headers { if strings.EqualFold(strings.TrimSpace(h), "User") && i < len(row) { user = strings.TrimSpace(fmt.Sprint(row[i])); break } }
		if user != "" {
			if settings.Business.Skip20 == nil { settings.Business.Skip20 = map[string]bool{} }
			settings.Business.Skip20[user] = !settings.Business.Skip20[user]
			saveSettings(); recalculate()
			logEvent("INFO", "PICK_SKIP20_DBLCLICK", "user_hash", shortHash(user), "enabled", strconv.FormatBool(settings.Business.Skip20[user]))
			setStatus("Đã đổi kiểm tra 20 phút"); pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0); return
		}
	}
	var b strings.Builder
	for i, h := range currentTable.data.Headers { if i < len(row) { fmt.Fprintf(&b, "%s: %s\r\n", h, formatCell(row[i], currentTable.kinds[i])) } }
	pMessageBox.Call(mainWnd, uintptr(unsafe.Pointer(ptr(b.String()))), uintptr(unsafe.Pointer(ptr("Chi tiết"))), MB_OK|MB_ICONINFORMATION)
	logEvent("INFO", "DETAIL_OPEN", "page", strconv.Itoa(currentPage))
}

func handleCommand(id, code int, source uintptr) {
	if id == 0 { return }
	logEvent("INFO", "UI_COMMAND", "id", strconv.Itoa(id), "code", strconv.Itoa(code), "page", strconv.Itoa(currentPage))
	if id >= ID_NAV_OVERVIEW && id <= ID_NAV_SETTINGS {
		if id != currentPage { renderPage(id) }
		return
	}
	switch id {
	case ID_SYNC:
		startSync()
	case ID_UPDATE:
		go checkUpdateInteractive()
	case ID_CURL_IMPORT:
		importCurl()
	case ID_SECRET_TOGGLE:
		revealSecrets = !revealSecrets
		setText(settingsSummary, credentialSummary())
		logEvent("INFO", "SECRET_DISPLAY_TOGGLE", "visible", strconv.FormatBool(revealSecrets))
	case ID_NET_TEST:
		go networkTest()
	case ID_LOG_OPEN:
		openFolder(logDir())
	case ID_LOG_EXPORT:
		startLogExport()
	case ID_ACTIVE_STATUS:
		if code == 1 {
			settings.Business.ActiveStatus = comboText(activeCombo)
			saveSettings()
			pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
		}
	case ID_PICK_SHIFT:
		if code == 1 {
			settings.Business.PickShift = comboText(pickShiftCombo)
			saveSettings()
			recalculate()
			pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
		}
	case ID_PACK_SHIFT:
		if code == 1 {
			settings.Business.PackShift = comboText(packShiftCombo)
			saveSettings()
			recalculate()
			pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
		}
	case ID_PICK_DEDUCT:
		settings.Business.PickDeductSKU = checked(source); saveSettings(); recalculate(); pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
	case ID_PICK_REQUIRE:
		settings.Business.PickRequireEven = checked(source); saveSettings(); recalculate(); pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
	case ID_PICK_ALLSITE:
		settings.Business.ShowAllSite = checked(source); saveSettings(); recalculate(); pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
	case ID_PICK_1C1L:
		settings.Business.Enable1C1L = checked(source); saveSettings(); recalculate(); pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
	case ID_PICK_INCOMPLETE:
		settings.Business.ShowIncompleteEven = checked(source); saveSettings(); recalculate(); pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
	case ID_PICK_1C1LERR:
		settings.Business.Show1C1LErrors = checked(source); saveSettings(); recalculate(); pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)
	case ID_SHIFT_MANUAL:
		if code == 1 { applyManualShift() }
	}
}

func recalculate() {
	liveMu.Lock(); defer liveMu.Unlock()
	if len(live.Payroll) == 0 { return }
	p, pa, sh := core.BuildTables(live.Payroll, settings.Business, time.Now())
	live.Pick = p
	live.Pack = pa
	live.Shift = sh
	live.UserPDA = liveio.BuildPeopleTable(live.Payroll)
}
func applyManualShift() {
	row, ok := selectedRow()
	if !ok {
		return
	}
	user, job := "", ""
	for i, h := range currentTable.data.Headers {
		if i >= len(row) {
			continue
		}
		if strings.EqualFold(h, "User") {
			user = fmt.Sprint(row[i])
		}
		if strings.EqualFold(h, "Loại công việc") {
			job = fmt.Sprint(row[i])
		}
	}
	if user == "" {
		return
	}
	v := comboText(shiftManualCombo)
	if settings.Business.ManualShifts == nil {
		settings.Business.ManualShifts = map[string]string{}
	}
	k := core.ManualShiftKey(user, job)
	if v == "Tự động" {
		delete(settings.Business.ManualShifts, k)
	} else {
		settings.Business.ManualShifts[k] = v
	}
	saveSettings()
	recalculate()
	logEvent("INFO", "MANUAL_SHIFT_CHANGE", "user_hash", shortHash(user), "value", v)
	renderPage(ID_NAV_SHIFT)
}

type liveSyncResult struct {
	name    string
	payroll []core.PayrollRow
	table   core.Table
	meta    liveio.HTTPMeta
	route   string
	err     error
}

type clientCandidate struct {
	name   string
	client *http.Client
}

func sessionReady() bool {
	return creds.Authorization != "" || creds.Token != "" || creds.APISID != "" ||
		creds.USID != "" || creds.Signature != ""
}

func currentLiveSession() liveio.Session {
	return liveio.Session{
		Authorization: creds.Authorization,
		Token: creds.Token,
		APISID: creds.APISID,
		USID: creds.USID,
		Signature: creds.Signature,
		Nonce: creds.Nonce,
		UserAgent: creds.UserAgent,
		Headers: creds.OtherHeaders,
	}
}

func profileBinding(name string) (liveio.Binding, bool) {
	for k, b := range profile.Bindings {
		if strings.EqualFold(strings.TrimSpace(k), strings.TrimSpace(name)) {
			return b, true
		}
	}
	return liveio.Binding{}, false
}

func windowsUserProxy() *url.URL {
	out, err := exec.Command("reg.exe", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`, "/v", "ProxyEnable").CombinedOutput()
	if err != nil || !strings.Contains(strings.ToLower(string(out)), "0x1") {
		return nil
	}
	out, err = exec.Command("reg.exe", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`, "/v", "ProxyServer").CombinedOutput()
	if err != nil {
		return nil
	}
	text := strings.TrimSpace(string(out))
	lines := strings.Split(text, "\n")
	value := ""
	for _, line := range lines {
		if !strings.Contains(strings.ToLower(line), "proxyserver") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			value = fields[len(fields)-1]
		}
	}
	if value == "" {
		return nil
	}
	if strings.Contains(value, ";") {
		parts := strings.Split(value, ";")
		value = ""
		for _, p := range parts {
			kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
			if len(kv) == 2 && strings.EqualFold(kv[0], "https") {
				value = kv[1]
				break
			}
		}
		if value == "" {
			for _, p := range parts {
				kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
				if len(kv) == 2 && strings.EqualFold(kv[0], "http") {
					value = kv[1]
					break
				}
			}
		}
	}
	if value == "" {
		return nil
	}
	if !strings.Contains(value, "://") {
		value = "http://" + value
	}
	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return nil
	}
	return u
}

func liveClients() []clientCandidate {
	out := []clientCandidate{}
	if p := windowsUserProxy(); p != nil {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.Proxy = http.ProxyURL(p)
		out = append(out, clientCandidate{name: "windows-user-proxy", client: &http.Client{Transport: tr, Timeout: 45 * time.Second}})
	}
	envTr := http.DefaultTransport.(*http.Transport).Clone()
	out = append(out, clientCandidate{name: "system-auto", client: &http.Client{Transport: envTr, Timeout: 45 * time.Second}})
	directTr := http.DefaultTransport.(*http.Transport).Clone()
	directTr.Proxy = nil
	out = append(out, clientCandidate{name: "direct", client: &http.Client{Transport: directTr, Timeout: 45 * time.Second}})
	return out
}

func executeWithFallback(ctx context.Context, b liveio.Binding, s liveio.Session) ([]byte, liveio.HTTPMeta, string, error) {
	var last error
	var lastMeta liveio.HTTPMeta
	for _, candidate := range liveClients() {
		data, meta, err := liveio.Execute(ctx, candidate.client, b, s)
		if err == nil {
			return data, meta, candidate.name, nil
		}
		last, lastMeta = err, meta
		if meta.StatusCode == http.StatusUnauthorized || meta.StatusCode == http.StatusForbidden {
			break
		}
	}
	if last == nil {
		last = fmt.Errorf("no network route")
	}
	return nil, lastMeta, "", last
}

func syncOne(ctx context.Context, name string, b liveio.Binding, s liveio.Session, ch chan<- liveSyncResult) {
	b = liveio.ExpandBinding(b, time.Now())
	data, meta, route, err := executeWithFallback(ctx, b, s)
	if err != nil {
		ch <- liveSyncResult{name: name, meta: meta, route: route, err: err}
		return
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "payroll-productivity":
		rows, e := liveio.ParsePayrollXLSX(data, b)
		ch <- liveSyncResult{name: name, payroll: rows, meta: meta, route: route, err: e}
	case "active-picking":
		table, e := liveio.ParseActiveJSON(data, b)
		ch <- liveSyncResult{name: name, table: table, meta: meta, route: route, err: e}
	default:
		ch <- liveSyncResult{name: name, meta: meta, route: route, err: fmt.Errorf("unsupported live binding")}
	}
}

func startSync() {
	if !busy.CompareAndSwap(false, true) {
		setStatus("Đang đồng bộ; không tạo thêm tác vụ chồng nhau.")
		return
	}
	setStatus("Đang đồng bộ Dashboard…")
	logEvent("INFO", "SYNC_START")
	go func() {
		defer busy.Store(false)
		defer pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0)

		if !sessionReady() {
			setStatus("Chưa có phiên Dashboard. Vào Thiết lập và dán cURL một lần.")
			logEvent("WARN", "SYNC_NO_CREDENTIAL")
			return
		}
		activeBinding, activeOK := profileBinding("active-picking")
		payBinding, payOK := profileBinding("payroll-productivity")
		if !activeOK || !payOK {
			if err := ensureDashboardBindings(); err != nil {
				setStatus("Chưa thiết lập Dashboard: " + err.Error())
				logEvent("WARN", "SYNC_DASHBOARD_CONFIG_MISSING", "error", err.Error())
				return
			}
			activeBinding, activeOK = profileBinding("active-picking")
			payBinding, payOK = profileBinding("payroll-productivity")
		}
		if !activeOK || !payOK {
			setStatus("Cấu hình Dashboard chưa đầy đủ.")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
		defer cancel()
		ch := make(chan liveSyncResult, 2)
		session := currentLiveSession()
		go syncOne(ctx, "active-picking", activeBinding, session, ch)
		go syncOne(ctx, "payroll-productivity", payBinding, session, ch)

		var payroll []core.PayrollRow
		var active core.Table
		var problems []string
		success := 0
		route := ""
		for i := 0; i < 2; i++ {
			r := <-ch
			if r.err != nil {
				problems = append(problems, sourceLabel(r.name)+": "+r.err.Error())
				logEvent("ERROR", "SYNC_SOURCE_FAILED",
					"source", r.name,
					"http_status", strconv.Itoa(r.meta.StatusCode),
					"bytes", strconv.Itoa(r.meta.Bytes),
					"elapsed_ms", strconv.FormatInt(r.meta.Elapsed.Milliseconds(), 10),
					"error", r.err.Error())
				continue
			}
			success++
			if route == "" { route = r.route }
			if r.name == "payroll-productivity" { payroll = r.payroll }
			if r.name == "active-picking" { active = r.table }
			rows := len(r.payroll)
			if r.name == "active-picking" { rows = len(r.table.Rows) }
			logEvent("INFO", "SYNC_SOURCE_OK",
				"source", r.name,
				"http_status", strconv.Itoa(r.meta.StatusCode),
				"bytes", strconv.Itoa(r.meta.Bytes),
				"rows", strconv.Itoa(rows),
				"elapsed_ms", strconv.FormatInt(r.meta.Elapsed.Milliseconds(), 10),
				"route", r.route)
		}

		if success == 0 {
			liveMu.Lock()
			live.LastError = strings.Join(problems, " | ")
			liveMu.Unlock()
			setStatus("Đồng bộ lỗi; dữ liệu hợp lệ trước đó vẫn được giữ nguyên.")
			return
		}

		// Preserve the Excel behavior: production is processed through Mapping /
		// Phân ca, while current-picking profile fields enrich missing employee data.
		liveMu.RLock()
		effectivePayroll := append([]core.PayrollRow(nil), live.Payroll...)
		effectiveActive := live.Active
		liveMu.RUnlock()
		if len(active.Headers) > 0 { effectiveActive = active }
		if len(payroll) > 0 {
			effectivePayroll = liveio.EnrichPayrollFromActive(payroll, effectiveActive)
		}

		var pick, pack, shift core.Table
		if len(payroll) > 0 {
			pick, pack, shift = core.BuildTables(effectivePayroll, settings.Business, time.Now())
		}
		people := liveio.BuildPeopleTableCombined(effectivePayroll, effectiveActive)

		liveMu.Lock()
		if len(payroll) > 0 {
			live.Payroll = effectivePayroll
			live.Pick, live.Pack, live.Shift = pick, pack, shift
		}
		if len(active.Headers) > 0 { live.Active = active }
		if len(people.Headers) > 0 { live.UserPDA = people }
		live.LastSync = time.Now()
		live.LastError = strings.Join(problems, " | ")
		live.Route = route
		pickN := len(live.Pick.Rows)
		packN := len(live.Pack.Rows)
		shiftN := len(live.Shift.Rows)
		activeN := len(live.Active.Rows)
		peopleN := len(live.UserPDA.Rows)
		payrollN := len(live.Payroll)
		liveMu.Unlock()

		if len(problems) > 0 {
			setStatus(fmt.Sprintf("Đồng bộ một phần · User/PDA %d · Pick %d · Pack %d · Phân ca %d · Đang lấy %d", peopleN, pickN, packN, shiftN, activeN))
		} else {
			setStatus(fmt.Sprintf("Đồng bộ xong · User/PDA %d · Pick %d · Pack %d · Phân ca %d · Đang lấy %d", peopleN, pickN, packN, shiftN, activeN))
		}
		logEvent("INFO", "SYNC_DONE",
			"success_sources", strconv.Itoa(success),
			"failed_sources", strconv.Itoa(len(problems)),
			"payroll_rows", strconv.Itoa(payrollN),
			"user_pda_rows", strconv.Itoa(peopleN),
			"pick_rows", strconv.Itoa(pickN),
			"pack_rows", strconv.Itoa(packN),
			"shift_rows", strconv.Itoa(shiftN),
			"active_rows", strconv.Itoa(activeN))
	}()
}

var reHeader = regexp.MustCompile(`(?is)-H\s+(?:'([^']*)'|"([^"]*)")`)
var reURL = regexp.MustCompile(`(?is)(?:--url\s+)?(?:'(https?://[^']+)'|"(https?://[^"]+)"|(https?://\S+))`)
var reMethod = regexp.MustCompile(`(?i)(?:-X|--request)\s+['"]?([A-Z]+)`)
var reBody = regexp.MustCompile(`(?is)(?:--data-raw|--data-binary|--data)\s+(?:'([^']*)'|"([^"]*)")`)

func parseCurl(raw string) (credentials, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return credentials{}, fmt.Errorf("cURL trống")
	}
	m := reURL.FindStringSubmatch(raw)
	u := ""
	if len(m) > 1 {
		for _, x := range m[1:] {
			if x != "" {
				u = x
				break
			}
		}
	}
	if u == "" {
		return credentials{}, fmt.Errorf("không tìm thấy URL")
	}
	if _, e := url.ParseRequestURI(u); e != nil {
		return credentials{}, fmt.Errorf("URL không hợp lệ")
	}
	c := credentials{URL: u, Method: "GET", OtherHeaders: map[string]string{}, ImportedAt: time.Now().Format(time.RFC3339)}
	if mm := reMethod.FindStringSubmatch(raw); len(mm) > 1 {
		c.Method = strings.ToUpper(mm[1])
	}
	if bm := reBody.FindStringSubmatch(raw); len(bm) > 1 {
		for _, x := range bm[1:] {
			if x != "" {
				c.Body = x
				break
			}
		}
		if c.Method == "GET" {
			c.Method = "POST"
		}
	}
	for _, hm := range reHeader.FindAllStringSubmatch(raw, -1) {
		h := ""
		if hm[1] != "" {
			h = hm[1]
		} else {
			h = hm[2]
		}
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k, v := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		lk := strings.ToLower(k)
		switch lk {
		case "authorization":
			c.Authorization = v
			c.Token = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(v, "Bearer"), "bearer"))
		case "x-signature":
			c.Signature = v
		case "x-signature-nonce":
			c.Nonce = v
		case "user-agent":
			c.UserAgent = v
		case "cookie":
			for _, p := range strings.Split(v, ";") {
				q := strings.SplitN(strings.TrimSpace(p), "=", 2)
				if len(q) != 2 {
					continue
				}
				switch strings.ToUpper(q[0]) {
				case "APISID":
					c.APISID = q[1]
				case "USID":
					c.USID = q[1]
				}
			}
		default:
			c.OtherHeaders[k] = v
		}
	}
	return c, nil
}
func importCurl() {
	if busy.Load() {
		setStatus("Đang đồng bộ; chờ hoàn tất trước khi thay đổi cấu hình.")
		return
	}
	raw := getText(settingsCurl)
	c, err := parseCurl(raw)
	setText(settingsCurl, "")
	if err != nil {
		setStatus("Không nhận được cURL: " + err.Error())
		logEvent("WARN", "CURL_IMPORT_FAILED", "error", err.Error())
		return
	}
	bindings, err := dashboardBindingsFromCurl(c)
	if err != nil {
		setStatus("Không tạo được cấu hình Dashboard: " + err.Error())
		logEvent("WARN", "DASHBOARD_CONFIG_FAILED", "error", err.Error())
		return
	}

	creds = mergeCredentials(creds, c)
	if err = saveCredentials(); err != nil {
		setStatus("Không lưu được phiên Dashboard: " + err.Error())
		logEvent("ERROR", "CREDENTIAL_SAVE_FAILED", "error", err.Error())
		return
	}
	profile = runtimeProfile{
		SchemaVersion: 1,
		ProfileID: "dashboard-single-source",
		Bindings: bindings,
	}
	if err = saveRuntimeProfile(); err != nil {
		setStatus("Không lưu được cấu hình Dashboard: " + err.Error())
		logEvent("ERROR", "DASHBOARD_CONFIG_SAVE_FAILED", "error", err.Error())
		return
	}

	setText(settingsSummary, credentialSummary())
	setStatus("Đã lưu cấu hình Dashboard. Có thể Đồng bộ ngay.")
	logEvent("INFO", "DASHBOARD_CONFIGURED",
		"binding_count", strconv.Itoa(len(bindings)),
		"authorization_present", strconv.FormatBool(creds.Authorization != ""),
		"signature_present", strconv.FormatBool(creds.Signature != ""))
}

func mask(s string) string {
	if s == "" {
		return "—"
	}
	if revealSecrets {
		return s
	}
	r := []rune(s)
	if len(r) <= 8 {
		return "••••••••"
	}
	return string(r[:4]) + "••••••••" + string(r[len(r)-4:])
}
func credentialSummary() string {
	return "Phiên Dashboard: " + map[bool]string{true: "Đã nhận", false: "Chưa có"}[sessionReady()] +
		"\r\nAuthorization/Token: " + mask(firstNonEmpty(creds.Token, creds.Authorization)) +
		"\r\nAPISID: " + mask(creds.APISID) + "    USID: " + mask(creds.USID) +
		"\r\nx-signature: " + mask(creds.Signature) + "    nonce: " + mask(creds.Nonce) +
		"\r\n" + sourceSummary()
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func requestProbe(c credentials) error {
	client := &http.Client{Timeout: 15 * time.Second}
	var body io.Reader
	if c.Body != "" {
		body = strings.NewReader(c.Body)
	}
	req, e := http.NewRequest(c.Method, c.URL, body)
	if e != nil {
		return e
	}
	if c.Authorization != "" {
		req.Header.Set("Authorization", c.Authorization)
	}
	if c.Signature != "" {
		req.Header.Set("x-signature", c.Signature)
	}
	if c.Nonce != "" {
		req.Header.Set("x-signature-nonce", c.Nonce)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	for k, v := range c.OtherHeaders {
		if !isSensitiveHeader(k) {
			req.Header.Set(k, v)
		}
	}
	resp, e := client.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
func networkTest() {
	if !sessionReady() {
		setStatus("Chưa có phiên Dashboard.")
		return
	}
	activeBinding, aOK := profileBinding("active-picking")
	payBinding, pOK := profileBinding("payroll-productivity")
	if !aOK || !pOK {
		if err := ensureDashboardBindings(); err != nil {
			setStatus("Chưa có cấu hình Dashboard: " + err.Error())
			return
		}
		activeBinding, aOK = profileBinding("active-picking")
		payBinding, pOK = profileBinding("payroll-productivity")
	}
	if !aOK || !pOK {
		setStatus("Cấu hình Dashboard chưa đầy đủ.")
		return
	}

	setStatus("Đang kiểm tra Đang lấy hàng và Sản lượng…")
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	ch := make(chan liveSyncResult, 2)
	session := currentLiveSession()
	go syncOne(ctx, "active-picking", activeBinding, session, ch)
	go syncOne(ctx, "payroll-productivity", payBinding, session, ch)

	okCount := 0
	var problems []string
	for i := 0; i < 2; i++ {
		r := <-ch
		if r.err != nil {
			problems = append(problems, sourceLabel(r.name)+": "+r.err.Error())
			logEvent("WARN", "DASHBOARD_TEST_FAILED", "source", r.name, "http_status", strconv.Itoa(r.meta.StatusCode), "error", r.err.Error())
			continue
		}
		okCount++
		rows := 0
		if r.name == "active-picking" { rows = len(r.table.Rows) }
		if r.name == "payroll-productivity" { rows = len(r.payroll) }
		logEvent("INFO", "DASHBOARD_TEST_OK", "source", r.name, "rows", strconv.Itoa(rows), "http_status", strconv.Itoa(r.meta.StatusCode), "route", r.route)
	}
	if okCount == 2 {
		setStatus("Kiểm tra Dashboard OK: Đang lấy hàng + Sản lượng.")
		return
	}
	setStatus("Kiểm tra Dashboard chưa đạt: " + strings.Join(problems, " | "))
}

func isSensitiveHeader(k string) bool {
	n := strings.ToLower(k)
	return strings.Contains(n, "token") || strings.Contains(n, "authorization") || strings.Contains(n, "cookie") || strings.Contains(n, "signature") || n == "apisid" || n == "usid" || strings.Contains(n, "password")
}

func appDir() string {
	p, _ := os.UserConfigDir()
	if p == "" {
		p = os.TempDir()
	}
	return filepath.Join(p, "Supra", "Productivity")
}
func secureDir() string    { return filepath.Join(appDir(), "Secure") }
func settingsPath() string { return filepath.Join(appDir(), "settings.json") }
func logDir() string { return filepath.Join(appDir(), "Logs") }

func ensureDirs() {
	os.MkdirAll(appDir(), 0700)
	os.MkdirAll(secureDir(), 0700)
	os.MkdirAll(logDir(), 0700)
}
func loadSettings() {
	settings = appSettings{DataFolder: filepath.Join(appDir(), "Data"), Business: core.DefaultBusinessSettings()}
	b, e := os.ReadFile(settingsPath())
	if e == nil {
		_ = json.Unmarshal(b, &settings)
	}
	core.NormalizeBusinessSettings(&settings.Business)
	ensureDirs()
}
func saveSettings() {
	ensureDirs()
	b, _ := json.MarshalIndent(settings, "", "  ")
	_ = os.WriteFile(settingsPath(), b, 0600)
}
func protect(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}
	in := DATA_BLOB{CbData: uint32(len(data)), PbData: &data[0]}
	var out DATA_BLOB
	r, _, e := pCryptProtect.Call(uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, CRYPTPROTECT_UI_FORBIDDEN, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, e
	}
	defer pLocalFree.Call(uintptr(unsafe.Pointer(out.PbData)))
	return append([]byte(nil), unsafe.Slice(out.PbData, out.CbData)...), nil
}
func unprotect(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}
	in := DATA_BLOB{CbData: uint32(len(data)), PbData: &data[0]}
	var out DATA_BLOB
	r, _, e := pCryptUnprotect.Call(uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, CRYPTPROTECT_UI_FORBIDDEN, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, e
	}
	defer pLocalFree.Call(uintptr(unsafe.Pointer(out.PbData)))
	return append([]byte(nil), unsafe.Slice(out.PbData, out.CbData)...), nil
}
func saveCredentials() error {
	ensureDirs()
	b, e := json.Marshal(creds)
	if e != nil {
		return e
	}
	p, e := protect(b)
	if e != nil {
		return e
	}
	return os.WriteFile(filepath.Join(secureDir(), "credentials.dat"), []byte(base64.StdEncoding.EncodeToString(p)), 0600)
}
func loadCredentials() {
	raw, e := os.ReadFile(filepath.Join(secureDir(), "credentials.dat"))
	if e != nil {
		return
	}
	enc, e := base64.StdEncoding.DecodeString(string(raw))
	if e != nil {
		return
	}
	plain, e := unprotect(enc)
	if e != nil {
		return
	}
	if e = json.Unmarshal(plain, &creds); e != nil {
		creds = credentials{}
		return
	}
	if creds.OtherHeaders == nil {
		creds.OtherHeaders = map[string]string{}
	}
}

func runtimeProfilePath() string { return filepath.Join(secureDir(), "runtime-profile.dat") }

func saveRuntimeProfile() error {
	ensureDirs()
	if profile.SchemaVersion == 0 { profile.SchemaVersion = 1 }
	if profile.ProfileID == "" { profile.ProfileID = "local-ui" }
	if profile.Bindings == nil { profile.Bindings = map[string]runtimeBinding{} }
	if err := validateRuntimeProfile(profile); err != nil { return err }
	plain, err := json.Marshal(profile)
	if err != nil { return err }
	enc, err := protect(plain)
	if err != nil { return err }
	return os.WriteFile(runtimeProfilePath(), []byte(base64.StdEncoding.EncodeToString(enc)), 0600)
}

func sourceBindingName(label string) string {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "sản lượng", "san luong", "payroll", "payroll/productivity":
		return "payroll-productivity"
	case "đang lấy hàng", "dang lay hang", "active", "active picking":
		return "active-picking"
	default:
		return ""
	}
}

func sourceLabel(binding string) string {
	switch strings.ToLower(strings.TrimSpace(binding)) {
	case "payroll-productivity":
		return "Sản lượng"
	case "active-picking":
		return "Đang lấy hàng"
	default:
		return binding
	}
}

func sourceConfigured(name string) bool {
	_, ok := profileBinding(name)
	return ok
}

func sourceSummary() string {
	activeOK := sourceConfigured("active-picking")
	payOK := sourceConfigured("payroll-productivity")
	if activeOK && payOK {
		return "Dashboard: Đã thiết lập đầy đủ"
	}
	return "Dashboard: Chưa thiết lập đầy đủ"
}

func mergeCredentials(old, next credentials) credentials {
	out := old
	set := func(dst *string, v string) { if strings.TrimSpace(v) != "" { *dst = v } }
	set(&out.URL, next.URL)
	set(&out.Method, next.Method)
	set(&out.Authorization, next.Authorization)
	set(&out.Token, next.Token)
	set(&out.APISID, next.APISID)
	set(&out.USID, next.USID)
	set(&out.Signature, next.Signature)
	set(&out.Nonce, next.Nonce)
	set(&out.Body, next.Body)
	set(&out.UserAgent, next.UserAgent)
	set(&out.ImportedAt, next.ImportedAt)
	if out.OtherHeaders == nil { out.OtherHeaders = map[string]string{} }
	for k, v := range next.OtherHeaders {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" { out.OtherHeaders[k] = v }
	}
	return out
}

func localHeaderSensitive(k string) bool {
	n := strings.ToLower(strings.TrimSpace(k))
	return n == "authorization" || n == "cookie" || n == "token" ||
		n == "apisid" || n == "usid" || strings.Contains(n, "signature") ||
		strings.Contains(n, "password")
}

func templatizeRequestDates(raw string, now time.Time) string {
	if raw == "" { return raw }
	today := now
	tomorrow := now.AddDate(0, 0, 1)
	yesterday := now.AddDate(0, 0, -1)
	repl := []struct{ old, next string }{
		{today.Format("2006-01-02"), "{{TODAY_ISO}}"},
		{tomorrow.Format("2006-01-02"), "{{TOMORROW_ISO}}"},
		{yesterday.Format("2006-01-02"), "{{YESTERDAY_ISO}}"},
		{today.Format("02/01/2006"), "{{TODAY_DMY}}"},
		{tomorrow.Format("02/01/2006"), "{{TOMORROW_DMY}}"},
		{yesterday.Format("02/01/2006"), "{{YESTERDAY_DMY}}"},
	}
	out := raw
	for _, r := range repl { out = strings.ReplaceAll(out, r.old, r.next) }
	return out
}

func dashboardOrigin(rawURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("URL Dashboard không hợp lệ")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("Dashboard phải dùng HTTP/HTTPS")
	}
	return u.Scheme + "://" + u.Host, nil
}

func nonSensitiveCurlHeaders(c credentials) map[string]string {
	h := map[string]string{}
	for k, v := range c.OtherHeaders {
		if strings.TrimSpace(k) == "" || localHeaderSensitive(k) { continue }
		lk := strings.ToLower(strings.TrimSpace(k))
		if lk == "host" || lk == "content-length" || lk == "referer" || lk == "origin" { continue }
		h[k] = v
	}
	return h
}

func cloneHeaders(in map[string]string) map[string]string {
	out := make(map[string]string, len(in)+6)
	for k, v := range in { out[k] = v }
	return out
}

func dashboardBindingsFromCurl(c credentials) (map[string]runtimeBinding, error) {
	origin, err := dashboardOrigin(c.URL)
	if err != nil { return nil, err }
	base := nonSensitiveCurlHeaders(c)

	activeHeaders := cloneHeaders(base)
	activeHeaders["Accept"] = "application/json, text/plain, */*"
	activeHeaders["Content-Type"] = "application/json"
	activeHeaders["Origin"] = origin
	activeHeaders["Referer"] = origin + "/app/dashboard/picking"
	activeHeaders["withcredentials"] = "true"
	activeHeaders["Cache-Control"] = "no-cache"
	activeHeaders["Pragma"] = "no-cache"

	payHeaders := cloneHeaders(base)
	payHeaders["Accept"] = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet, application/octet-stream, */*"
	payHeaders["Origin"] = origin
	payHeaders["Referer"] = origin + "/app/payroll/list"
	payHeaders["withcredentials"] = "true"
	payHeaders["Cache-Control"] = "no-cache"
	payHeaders["Pragma"] = "no-cache"

	// Stable V1.3 uses one Dashboard session and two internal requests:
	// current picking JSON plus previous-day..today payroll export XLSX.
	active := runtimeBinding{
		Method: "POST",
		URL: origin + "/api/v1/performance/picking",
		Body: `{"data":{"WarehouseCode":"HY1","ClientCode":""}}`,
		Headers: activeHeaders,
		ResponseKind: "json",
	}
	payroll := runtimeBinding{
		Method: "GET",
		URL: origin + "/api/v1/payroll/export?keywords=&WarehouseCode=HY1&Employee=&JobType=&ToDate={{TODAY_ISO}}&FromDate={{YESTERDAY_ISO}}",
		Headers: payHeaders,
		ResponseKind: "xlsx",
	}
	return map[string]runtimeBinding{
		"active-picking": active,
		"payroll-productivity": payroll,
	}, nil
}

func ensureDashboardBindings() error {
	if strings.TrimSpace(creds.URL) == "" { return fmt.Errorf("chưa có cURL Dashboard") }
	bindings, err := dashboardBindingsFromCurl(creds)
	if err != nil { return err }
	profile = runtimeProfile{
		SchemaVersion: 1,
		ProfileID: "dashboard-single-source",
		Bindings: bindings,
	}
	if err := saveRuntimeProfile(); err != nil { return err }
	return nil
}

func bindingFromCurl(c credentials, source string) (runtimeBinding, error) {
	name := sourceBindingName(source)
	if name == "" { return runtimeBinding{}, fmt.Errorf("chưa chọn nguồn dữ liệu") }
	headers := map[string]string{}
	for k, v := range c.OtherHeaders {
		if strings.TrimSpace(k) == "" || localHeaderSensitive(k) { continue }
		headers[k] = v
	}
	now := time.Now()
	for k, v := range headers { headers[k] = templatizeRequestDates(v, now) }
	b := runtimeBinding{
		Method: c.Method,
		URL: templatizeRequestDates(c.URL, now),
		Body: templatizeRequestDates(c.Body, now),
		Headers: headers,
	}
	if b.Method == "" { b.Method = "GET" }
	if name == "payroll-productivity" { b.ResponseKind = "xlsx" } else { b.ResponseKind = "json" }
	if _, err := url.ParseRequestURI(strings.TrimSpace(b.URL)); err != nil {
		return runtimeBinding{}, fmt.Errorf("URL nguồn không hợp lệ")
	}
	return b, nil
}

func removeSource(name string) error {
	if profile.Bindings == nil { profile.Bindings = map[string]runtimeBinding{} }
	delete(profile.Bindings, name)
	if len(profile.Bindings) == 0 {
		_ = os.Remove(runtimeProfilePath())
		profile = runtimeProfile{SchemaVersion: 1, ProfileID: "local-ui", Bindings: map[string]runtimeBinding{}}
		return nil
	}
	return saveRuntimeProfile()
}
func runtimeProfileProvisionPath() string {
	exe, e := os.Executable()
	if e != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(exe), "SupraProductivity.profile.json")
}
func validateRuntimeProfile(p runtimeProfile) error {
	if p.SchemaVersion != 1 {
		return fmt.Errorf("schema_version phải = 1")
	}
	if len(p.Bindings) == 0 {
		return fmt.Errorf("profile không có binding")
	}
	for name, b := range p.Bindings {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("binding không có tên")
		}
		if _, e := url.ParseRequestURI(strings.TrimSpace(b.URL)); e != nil {
			return fmt.Errorf("binding %s có URL không hợp lệ", name)
		}
		m := strings.ToUpper(strings.TrimSpace(b.Method))
		if m == "" {
			m = "GET"
		}
		switch m {
		case "GET", "POST", "PUT", "PATCH", "DELETE":
		default:
			return fmt.Errorf("binding %s có method không hợp lệ", name)
		}
	}
	return nil
}
func provisionRuntimeProfile() {
	path := runtimeProfileProvisionPath()
	if path == "" {
		return
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		return
	}
	var p runtimeProfile
	if e = json.Unmarshal(raw, &p); e != nil || validateRuntimeProfile(p) != nil {
		logEvent("WARN", "RUNTIME_PROFILE_IMPORT_FAILED", "file", filepath.Base(path))
		return
	}
	plain, _ := json.Marshal(p)
	enc, e := protect(plain)
	if e != nil {
		logEvent("ERROR", "RUNTIME_PROFILE_PROTECT_FAILED")
		return
	}
	if e = os.WriteFile(runtimeProfilePath(), []byte(base64.StdEncoding.EncodeToString(enc)), 0600); e != nil {
		logEvent("ERROR", "RUNTIME_PROFILE_SAVE_FAILED")
		return
	}
	_ = os.Remove(path)
	logEvent("INFO", "LEGACY_SOURCE_PROFILE_IMPORTED", "profile_id", sanitizeProfileID(p.ProfileID), "binding_count", strconv.Itoa(len(p.Bindings)))
}
func loadRuntimeProfile() {
	profile = runtimeProfile{}
	raw, e := os.ReadFile(runtimeProfilePath())
	if e != nil {
		return
	}
	enc, e := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if e != nil {
		return
	}
	plain, e := unprotect(enc)
	if e != nil {
		return
	}
	var p runtimeProfile
	if json.Unmarshal(plain, &p) != nil || validateRuntimeProfile(p) != nil {
		return
	}
	profile = p
}
func sanitizeProfileID(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 64 {
		v = v[:64]
	}
	return regexp.MustCompile("[^A-Za-z0-9._-]+").ReplaceAllString(v, "_")
}

func sanitizeValue(k, v string) string {
	if isSensitiveHeader(k) {
		return "[REDACTED]"
	}
	lv := strings.ToLower(v)
	if strings.Contains(lv, "authorization:") || strings.Contains(lv, "cookie:") || strings.Contains(lv, "x-signature:") {
		return "[REDACTED]"
	}
	v = strings.ReplaceAll(strings.ReplaceAll(v, "\r", " "), "\n", " ")
	if len(v) > 4000 {
		v = v[:4000] + "…"
	}
	return v
}
func writeLogLine(line string) {
	ensureDirs()
	path := filepath.Join(logDir(), "supra_"+time.Now().Format("20060102")+".log")
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil { return }
	_, _ = f.WriteString(line)
	_ = f.Close()
}

func logWriter() {
	ensureDirs()
	for {
		select {
		case line := <-logQueue:
			writeLogLine(line)
		case ack := <-logFlush:
			for {
				select {
				case line := <-logQueue:
					writeLogLine(line)
				default:
					close(ack)
					goto flushed
				}
			}
		flushed:
		}
	}
}

func logEvent(level, event string, kv ...string) {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s %s", time.Now().Format(time.RFC3339Nano), level, event)
	if dropped := logDropped.Swap(0); dropped > 0 {
		fmt.Fprintf(&b, " previous_log_dropped=%q", strconv.FormatUint(dropped, 10))
	}
	for i := 0; i+1 < len(kv); i += 2 {
		fmt.Fprintf(&b, " %s=%q", kv[i], sanitizeValue(kv[i], kv[i+1]))
	}
	b.WriteString("\r\n")
	select {
	case logQueue <- b.String():
	default:
		logDropped.Add(1)
	}
}

func readLogTail(max int) string {
	f, e := os.Open(filepath.Join(logDir(), "supra_"+time.Now().Format("20060102")+".log"))
	if e != nil { return "Chưa có log." }
	defer f.Close()
	const maxBytes int64 = 512 * 1024
	if st, err := f.Stat(); err == nil && st.Size() > maxBytes {
		_, _ = f.Seek(-maxBytes, io.SeekEnd)
	}
	b, e := io.ReadAll(io.LimitReader(f, maxBytes))
	if e != nil { return "Không đọc được log cục bộ." }
	lines := strings.Split(string(b), "\n")
	if len(lines) > max { lines = lines[len(lines)-max:] }
	return strings.Join(lines, "\r\n")
}

func shortHash(s string) string { h := sha256.Sum256([]byte(s)); return fmt.Sprintf("%x", h[:4]) }
func flushLogs(timeout time.Duration) bool {
	ack := make(chan struct{})
	select {
	case logFlush <- ack:
	case <-time.After(timeout):
		return false
	}
	select {
	case <-ack:
		return true
	case <-time.After(timeout):
		return false
	}
}

func logRuntimeSnapshot(event string) {
	var ms MEMORYSTATUSEX
	ms.Length = uint32(unsafe.Sizeof(ms))
	pGlobalMemory.Call(uintptr(unsafe.Pointer(&ms)))
	var rm runtime.MemStats
	runtime.ReadMemStats(&rm)
	liveMu.RLock()
	ls := live
	liveMu.RUnlock()
	logEvent("INFO", event,
		"version", appVersion,
		"go", runtime.Version(),
		"arch", runtime.GOARCH,
		"cpu_count", strconv.Itoa(runtime.NumCPU()),
		"system_ram_total_mb", fmt.Sprintf("%.0f", float64(ms.TotalPhys)/1024/1024),
		"system_ram_used_pct", strconv.Itoa(int(ms.MemoryLoad)),
		"heap_alloc_mb", fmt.Sprintf("%.2f", float64(rm.Alloc)/1024/1024),
		"heap_sys_mb", fmt.Sprintf("%.2f", float64(rm.HeapSys)/1024/1024),
		"goroutines", strconv.Itoa(runtime.NumGoroutine()),
		"page", strconv.Itoa(currentPage),
		"window_w", strconv.FormatInt(windowWidth.Load(), 10),
		"window_h", strconv.FormatInt(windowHeight.Load(), 10),
		"credential_present", strconv.FormatBool(sessionReady()),
		"runtime_profile_present", strconv.FormatBool(len(profile.Bindings) > 0),
		"profile_binding_count", strconv.Itoa(len(profile.Bindings)),
		"sync_busy", strconv.FormatBool(busy.Load()),
		"last_sync", func() string { if ls.LastSync.IsZero() { return "" }; return ls.LastSync.Format(time.RFC3339) }(),
		"sync_route", ls.Route,
		"sync_error", ls.LastError,
		"payroll_rows", strconv.Itoa(len(ls.Payroll)),
		"pick_rows", strconv.Itoa(len(ls.Pick.Rows)),
		"pack_rows", strconv.Itoa(len(ls.Pack.Rows)),
		"shift_rows", strconv.Itoa(len(ls.Shift.Rows)),
		"active_rows", strconv.Itoa(len(ls.Active.Rows)),
		"user_pda_rows", strconv.Itoa(len(ls.UserPDA.Rows)))
}

func chooseLogExportPath() (string, bool) {
	name := "SUPRA_LOG_" + time.Now().Format("20060102_150405") + ".txt"
	buf := make([]uint16, 1024)
	copy(buf, syscall.StringToUTF16(name))
	of := OPENFILENAME{
		LStructSize: uint32(unsafe.Sizeof(OPENFILENAME{})),
		HwndOwner: 0,
		LpstrFile: &buf[0],
		NMaxFile: uint32(len(buf)),
		LpstrTitle: ptr("Chọn nơi lưu file LOG"),
		Flags: OFN_OVERWRITEPROMPT | OFN_PATHMUSTEXIST,
		LpstrDefExt: ptr("txt"),
	}
	r, _, _ := pGetSaveFileName.Call(uintptr(unsafe.Pointer(&of)))
	if r == 0 { return "", false }
	return syscall.UTF16ToString(buf), true
}

func startLogExport() {
	if !logExportBusy.CompareAndSwap(false, true) {
		setStatus("Cửa sổ xuất LOG đang mở.")
		return
	}
	setStatus("Đang mở cửa sổ chọn nơi lưu LOG…")
	logEvent("INFO", "LOG_EXPORT_DIALOG_OPEN")
	go func() {
		defer logExportBusy.Store(false)
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		path, ok := chooseLogExportPath()
		if !ok {
			setStatus("Đã hủy xuất LOG.")
			logEvent("INFO", "LOG_EXPORT_DIALOG_CANCEL")
			return
		}
		exportLogTo(path)
	}()
}

func exportLogTo(path string) {
	logEvent("INFO", "LOG_EXPORT_REQUESTED")
	logRuntimeSnapshot("LOG_EXPORT_STATE")
	flushLogs(1500 * time.Millisecond)

	var b strings.Builder
	b.WriteString("SUPRA PRODUCTIVITY - FULL SANITIZED LOG\r\n")
	b.WriteString("Exported: " + time.Now().Format(time.RFC3339) + "\r\n")
	b.WriteString("Version: " + appVersion + "\r\n")
	b.WriteString("Sensitive credentials/endpoints: REDACTED / NOT EXPORTED\r\n")
	b.WriteString(strings.Repeat("=", 88) + "\r\n\r\n")

	entries, _ := os.ReadDir(logDir())
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() { continue }
		name := entry.Name()
		if strings.HasPrefix(name, "supra_") && strings.HasSuffix(strings.ToLower(name), ".log") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	const maxExportBytes int64 = 50 * 1024 * 1024
	var total int64
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(logDir(), name))
		if err != nil { continue }
		if total+int64(len(data)) > maxExportBytes {
			b.WriteString("\r\n[LOG_EXPORT_TRUNCATED_AT_50MB]\r\n")
			break
		}
		b.WriteString("\r\n--- " + name + " ---\r\n")
		b.Write(data)
		total += int64(len(data))
	}
	if err := os.WriteFile(path, []byte(b.String()), 0600); err != nil {
		setStatus("Xuất log lỗi: " + err.Error())
		logEvent("ERROR", "LOG_EXPORT_FAILED", "error", err.Error())
		return
	}
	setStatus("Đã xuất log: " + filepath.Base(path))
	logEvent("INFO", "LOG_EXPORT_DONE", "file", filepath.Base(path), "bytes", strconv.Itoa(len(b.String())))
}

func openFolder(path string) { os.MkdirAll(path, 0700); _ = exec.Command("explorer.exe", path).Start() }
func setStatus(s string) {
	statusMu.Lock()
	pendingStatus = s
	statusMu.Unlock()
	if mainWnd != 0 {
		pPostMessage.Call(mainWnd, WM_APP_STATUS, 0, 0)
	}
}

// GitHub update check is non-fatal. It never blocks app startup or internal operations.
type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}
type releaseInfo struct {
	TagName    string         `json:"tag_name"`
	HTMLURL    string         `json:"html_url"`
	Draft      bool           `json:"draft"`
	Prerelease bool           `json:"prerelease"`
	Assets     []releaseAsset `json:"assets"`
}

func latestRelease() (releaseInfo, error) {
	var out releaseInfo
	client := http.Client{Timeout: 8 * time.Second}
	userAgent := "SupraProductivity/" + appVersion

	if strings.Contains(strings.ToLower(appVersion), "-test.") {
		req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+updateRepo+"/releases?per_page=20", nil)
		req.Header.Set("User-Agent", userAgent)
		resp, e := client.Do(req)
		if e != nil {
			return out, e
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return out, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		var releases []releaseInfo
		if e = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&releases); e != nil {
			return out, e
		}
		for _, r := range releases {
			if r.Draft {
				continue
			}
			if releaseIsNewer(r.TagName, appVersion) {
				return r, nil
			}
		}
		return releaseInfo{TagName: appVersion}, nil
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+updateRepo+"/releases/latest", nil)
	req.Header.Set("User-Agent", userAgent)
	resp, e := client.Do(req)
	if e != nil {
		return out, e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return out, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	e = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out)
	return out, e
}

type updateVersion struct {
	major, minor, patch int
	test                int
	prerelease          bool
	valid               bool
}

func parseUpdateVersion(v string) updateVersion {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	rx := regexp.MustCompile(`^([0-9]+).([0-9]+).([0-9]+)(?:-test.([0-9]+))?$`)
	m := rx.FindStringSubmatch(v)
	if len(m) == 0 {
		return updateVersion{}
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	out := updateVersion{major: major, minor: minor, patch: patch, valid: true}
	if len(m) > 4 && m[4] != "" {
		out.prerelease = true
		out.test, _ = strconv.Atoi(m[4])
	}
	return out
}

func releaseIsNewer(candidate, current string) bool {
	c := parseUpdateVersion(candidate)
	cur := parseUpdateVersion(current)
	if !c.valid || !cur.valid {
		return normalizeVersion(candidate) != normalizeVersion(current)
	}
	if c.major != cur.major {
		return c.major > cur.major
	}
	if c.minor != cur.minor {
		return c.minor > cur.minor
	}
	if c.patch != cur.patch {
		return c.patch > cur.patch
	}
	if c.prerelease != cur.prerelease {
		return !c.prerelease && cur.prerelease
	}
	if c.prerelease {
		return c.test > cur.test
	}
	return false
}

func normalizeVersion(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }
func checkUpdateQuiet() {
	r, e := latestRelease()
	if e != nil {
		logEvent("INFO", "UPDATE_CHECK_UNAVAILABLE", "error", e.Error())
		return
	}
	logEvent("INFO", "UPDATE_CHECK_OK", "latest", r.TagName)
	if appVersion != "dev" && releaseIsNewer(r.TagName, appVersion) {
		setStatus("Có bản cập nhật " + r.TagName + ". Bấm CẬP NHẬT để mở Release.")
	}
}

func assetByName(r releaseInfo, name string) (releaseAsset, bool) {
	for _, a := range r.Assets {
		if strings.EqualFold(a.Name, name) {
			return a, true
		}
	}
	return releaseAsset{}, false
}
func downloadLimited(rawURL string, max int64) ([]byte, error) {
	req, e := http.NewRequest("GET", rawURL, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("User-Agent", "SupraProductivity/"+appVersion)
	resp, e := (&http.Client{Timeout: 90 * time.Second}).Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	b, e := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if e != nil {
		return nil, e
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("asset vượt giới hạn")
	}
	return b, nil
}
func expectedSHA(text, filename string) (string, error) {
	for _, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 && strings.TrimPrefix(f[len(f)-1], "*") == filename {
			h := strings.ToLower(f[0])
			if len(h) == 64 {
				return h, nil
			}
		}
	}
	return "", fmt.Errorf("không tìm thấy SHA256 cho %s", filename)
}
func fileSHA256(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
func stageGitHubUpdate(r releaseInfo) error {
	exeAsset, ok := assetByName(r, "SupraProductivity.exe")
	if !ok {
		return fmt.Errorf("release thiếu SupraProductivity.exe")
	}
	shaAsset, ok := assetByName(r, "SHA256SUMS.txt")
	if !ok {
		return fmt.Errorf("release thiếu SHA256SUMS.txt")
	}
	shaBytes, e := downloadLimited(shaAsset.BrowserDownloadURL, 128*1024)
	if e != nil {
		return e
	}
	expected, e := expectedSHA(string(shaBytes), "SupraProductivity.exe")
	if e != nil {
		return e
	}
	exePath, e := os.Executable()
	if e != nil {
		return e
	}
	exePath, _ = filepath.Abs(exePath)
	stage := exePath + ".update"
	backup := exePath + ".bak"
	script := exePath + ".update.cmd"
	data, e := downloadLimited(exeAsset.BrowserDownloadURL, 150*1024*1024)
	if e != nil {
		return e
	}
	if e = os.WriteFile(stage, data, 0700); e != nil {
		return fmt.Errorf("thư mục EXE không ghi được: %w", e)
	}
	actual, e := fileSHA256(stage)
	if e != nil {
		_ = os.Remove(stage)
		return e
	}
	if actual != expected {
		_ = os.Remove(stage)
		return fmt.Errorf("SHA256 không khớp")
	}
	current, e := os.ReadFile(exePath)
	if e != nil {
		_ = os.Remove(stage)
		return e
	}
	if e = os.WriteFile(backup, current, 0700); e != nil {
		_ = os.Remove(stage)
		return e
	}
	cmd := "@echo off\r\nsetlocal\r\n" +
		"for /L %%I in (1,1,30) do (\r\n" +
		"  move /Y \"" + stage + "\" \"" + exePath + "\" >nul 2>nul && goto replaced\r\n" +
		"  ping 127.0.0.1 -n 2 >nul\r\n" +
		")\r\n" +
		"move /Y \"" + backup + "\" \"" + exePath + "\" >nul 2>nul\r\nexit /b 1\r\n" +
		":replaced\r\nstart \"\" \"" + exePath + "\"\r\ndel \"%~f0\"\r\n"
	if e = os.WriteFile(script, []byte(cmd), 0700); e != nil {
		return e
	}
	logEvent("INFO", "UPDATE_STAGED", "version", r.TagName, "sha256", expected[:12])
	if e = exec.Command("cmd.exe", "/C", script).Start(); e != nil {
		return e
	}
	setStatus("Đã xác minh SHA256. Ứng dụng sẽ đóng để cập nhật " + r.TagName + ".")
	time.Sleep(500 * time.Millisecond)
	pPostMessage.Call(mainWnd, WM_CLOSE, 0, 0)
	return nil
}
func checkUpdateInteractive() {
	setStatus("Đang kiểm tra cập nhật…")
	r, e := latestRelease()
	if e != nil {
		setStatus("Không thể kiểm tra cập nhật trên mạng hiện tại. Ứng dụng vẫn hoạt động bình thường.")
		logEvent("INFO", "UPDATE_CHECK_UNAVAILABLE", "error", e.Error())
		return
	}
	if appVersion == "dev" || !releaseIsNewer(r.TagName, appVersion) {
		setStatus("Đang dùng bản mới nhất: " + r.TagName)
		return
	}
	setStatus("Đang tải và xác minh " + r.TagName + "…")
	if e = stageGitHubUpdate(r); e != nil {
		setStatus("Không thể tự cập nhật: " + e.Error() + ". Có thể cập nhật thủ công từ Release.")
		logEvent("WARN", "UPDATE_STAGE_FAILED", "error", e.Error())
		if r.HTMLURL != "" {
			pShellExecute.Call(0, uintptr(unsafe.Pointer(ptr("open"))), uintptr(unsafe.Pointer(ptr(r.HTMLURL))), 0, 0, SW_SHOW)
		}
		return
	}
}

func main() {
	// Win32 windows, child controls and the message pump are thread-affine.
	// Keep the complete UI lifetime on one OS thread.
	runtime.LockOSThread()

	pInitCommon.Call()
	ensureDirs()
	go logWriter()

	hInst, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	cls := ptr("SupraProductivityWindow")
	wc := WNDCLASSEX{
		CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})),
		LpfnWndProc: syscall.NewCallback(wndProc),
		HInstance: hInst,
		HCursor: func() uintptr { r, _, _ := pLoadCursor.Call(0, 32512); return r }(),
		HbrBackground: uintptr(COLOR_WINDOW + 1),
		LpszClassName: cls,
	}
	if r, _, _ := pRegisterClass.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		panic("RegisterClassExW failed")
	}

	title := appName + " · " + appVersion
	style := uint32(WS_VISIBLE | WS_CAPTION | WS_SYSMENU | WS_MINIMIZEBOX | WS_CLIPCHILDREN)
	hwnd, _, _ := pCreateWindow.Call(
		0,
		uintptr(unsafe.Pointer(cls)),
		uintptr(unsafe.Pointer(ptr(title))),
		uintptr(style),
		CW_USEDEFAULT, CW_USEDEFAULT, 1280, 760,
		0, 0, hInst, 0,
	)
	if hwnd == 0 { panic("CreateWindowExW failed") }

	pShowWindow.Call(hwnd, SW_SHOW)
	fitWindowToWorkArea()
	pUpdateWindow.Call(hwnd)
	go uiWatchdog()

	var msg MSG
	for {
		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) == -1 {
			logEvent("ERROR", "GET_MESSAGE_FAILED")
			break
		}
		if r == 0 { break }
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

