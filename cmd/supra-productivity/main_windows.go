//go:build windows

package main

import (
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
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/tamnv2/supra-productivity-desktop/internal/core"
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
	WM_CREATE                    = 0x0001
	WM_DESTROY                   = 0x0002
	WM_CLOSE                     = 0x0010
	WM_SIZE                      = 0x0005
	WM_COMMAND                   = 0x0111
	WM_NOTIFY                    = 0x004E
	WM_TIMER                     = 0x0113
	WM_SETFONT                   = 0x0030
	WM_APP                       = 0x8000
	WM_APP_STATUS                = WM_APP + 1
	WM_APP_REFRESH               = WM_APP + 2
	SW_SHOW                      = 5
	CW_USEDEFAULT                = 0x80000000
	SS_LEFT                      = 0x00000000
	ES_MULTILINE                 = 0x0004
	ES_AUTOVSCROLL               = 0x0040
	ES_READONLY                  = 0x0800
	ES_AUTOHSCROLL               = 0x0080
	BS_PUSHBUTTON                = 0x00000000
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

type credentials struct {
	URL, Method, Authorization, Token, APISID, USID, Signature, Nonce, Body, UserAgent, ImportedAt string
	OtherHeaders                                                                                   map[string]string
}
type appSettings struct {
	DataFolder string                `json:"data_folder"`
	Business   core.BusinessSettings `json:"business"`
}
type runtimeBinding struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    string            `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}
type runtimeProfile struct {
	SchemaVersion int                       `json:"schema_version"`
	ProfileID     string                    `json:"profile_id"`
	Bindings      map[string]runtimeBinding `json:"bindings"`
}
type liveState struct {
	Pick, Pack, Shift, Active core.Table
	Payroll                   []core.PayrollRow
	LastSync                  time.Time
	LastError                 string
	Route                     string
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
	pCreateFont          = gdi32.NewProc("CreateFontW")
	pInitCommon          = comctl32.NewProc("InitCommonControls")
	pCryptProtect        = crypt32.NewProc("CryptProtectData")
	pCryptUnprotect      = crypt32.NewProc("CryptUnprotectData")
	pLocalFree           = kernel32.NewProc("LocalFree")
	pGlobalMemory        = kernel32.NewProc("GlobalMemoryStatusEx")
	pShellExecute        = shell32.NewProc("ShellExecuteW")
	pGetProcMem          = psapi.NewProc("GetProcessMemoryInfo")
	pGetCurrentProcess   = kernel32.NewProc("GetCurrentProcess")
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
)

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
	overviewMetric = 0
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
	var rc RECT
	pGetClientRect.Call(mainWnd, uintptr(unsafe.Pointer(&rc)))
	w := int(rc.Right) - 164; h := int(rc.Bottom) - top - 34
	if w < 600 { w = 600 }; if h < 260 { h = 260 }
	lv := create("SysListView32", "", WS_CHILD|WS_VISIBLE|WS_BORDER|LVS_REPORT|LVS_SHOWSELALWAYS|WS_HSCROLL|WS_VSCROLL, 152, top, w, h, mainWnd, 0)
	setFont(lv, fontNormal)
	pSendMessage.Call(lv, LVM_SETEXTENDEDLISTVIEWSTYLE, 0, LVS_EX_FULLROWSELECT|LVS_EX_GRIDLINES|LVS_EX_DOUBLEBUFFER)
	addPage(lv)
	kinds := make([]colKind, len(t.Headers))
	for i, h := range t.Headers {
		kinds[i] = inferKind(h)
		width := 115
		if strings.Contains(strings.ToLower(h), "họ") || strings.Contains(strings.ToLower(h), "thời gian") { width = 180 }
		addListColumn(lv, i, h, width)
	}
	currentTable = &tableModel{hwnd: lv, data: t, kinds: kinds, sortCol: -1, asc: true}
	tableWnd = lv
	fillTable()
}
func fillTable() {
	if currentTable == nil { return }
	pSendMessage.Call(currentTable.hwnd, LVM_DELETEALLITEMS, 0, 0)
	for r, row := range currentTable.data.Rows {
		vals := make([]string, len(currentTable.data.Headers))
		for i := range vals { if i < len(row) { vals[i] = formatCell(row[i], currentTable.kinds[i]) } }
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

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		mainWnd = hwnd; createShell(); loadSettings(); provisionRuntimeProfile(); loadRuntimeProfile(); loadCredentials()
		logEvent("INFO", "APP_START", "version", appVersion)
		renderPage(currentPage); pSetTimer.Call(hwnd, TIMER_METRICS, 2000, 0); go checkUpdateQuiet(); return 0
	case WM_SIZE:
		layout(); return 0
	case WM_COMMAND:
		id := int(uint16(wParam & 0xffff)); code := int(uint16((wParam >> 16) & 0xffff))
		handleCommand(id, code, lParam); return 0
	case WM_NOTIFY:
		return handleNotify(lParam)
	case WM_TIMER:
		if wParam == TIMER_METRICS { if currentPage == ID_NAV_OVERVIEW { renderOverviewMetrics() }; return 0 }
	case WM_APP_STATUS:
		statusMu.Lock(); s := pendingStatus; statusMu.Unlock(); setText(statusText, s); return 0
	case WM_APP_REFRESH:
		renderPage(currentPage); return 0
	case WM_DESTROY:
		pKillTimer.Call(hwnd, TIMER_METRICS); logEvent("INFO", "APP_EXIT"); pPostQuit.Call(0); return 0
	}
	r, _, _ := pDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam); return r
}

func createShell() {
	fontNormal, _, _ = pCreateFont.Call(18, 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	fontSmall, _, _ = pCreateFont.Call(16, 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	fontBold, _, _ = pCreateFont.Call(20, 0, 0, 0, 600, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(ptr("Segoe UI Semibold"))))
	headerText = create("STATIC", appName+"  ·  "+appVersion, WS_CHILD|WS_VISIBLE|SS_LEFT, 152, 10, 700, 30, mainWnd, 0); setFont(headerText, fontBold)
	statusText = create("STATIC", "Sẵn sàng", WS_CHILD|WS_VISIBLE|SS_LEFT, 760, 14, 550, 24, mainWnd, 0); setFont(statusText, fontSmall)
	buttonShell := func(id int, text string, y int) { b := create("BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 10, y, 130, 34, mainWnd, uintptr(id)); setFont(b, fontSmall); nav[id] = b }
	for i, item := range []struct{id int; name string}{{ID_NAV_OVERVIEW,"TỔNG QUAN"},{ID_NAV_ACTIVE,"ĐANG LẤY HÀNG"},{ID_NAV_PICK,"PICK"},{ID_NAV_PACK,"PACK"},{ID_NAV_SHIFT,"PHÂN CA"},{ID_NAV_USERPDA,"USER / PDA"},{ID_NAV_LOG,"NHẬT KÝ"},{ID_NAV_SETTINGS,"THIẾT LẬP"}} { buttonShell(item.id, item.name, 58+i*42) }
	b := create("BUTTON", "ĐỒNG BỘ", WS_CHILD|WS_VISIBLE|BS_PUSHBUTTON, 1340, 10, 100, 32, mainWnd, ID_SYNC); setFont(b, fontSmall)
	u := create("BUTTON", "CẬP NHẬT", WS_CHILD|WS_VISIBLE|BS_PUSHBUTTON, 1230, 10, 100, 32, mainWnd, ID_UPDATE); setFont(u, fontSmall)
}
func layout() {
	var rc RECT; pGetClientRect.Call(mainWnd, uintptr(unsafe.Pointer(&rc)))
	w, h := int(rc.Right), int(rc.Bottom)
	move(headerText, 152, 10, w-620, 30); move(statusText, w-690, 14, 430, 24)
	if tableWnd != 0 { top := 145; if currentPage == ID_NAV_PICK { top = 205 }; move(tableWnd, 152, top, w-164, h-top-34) }
	if logEdit != 0 { move(logEdit, 152, 145, w-164, h-180) }
}
func pageTitle(title, sub string) { static(title, 152, 52, 500, 28, true); static(sub, 152, 80, 1000, 24, false) }
func renderPage(id int) {
	currentPage = id; destroyPage()
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
	layout(); logEvent("INFO", "UI_PAGE", "page", strconv.Itoa(id))
}

func renderOverview() {
	pageTitle("TỔNG QUAN", "Vận hành, dữ liệu live và sức khoẻ laptop/ứng dụng theo thời gian thực")
	liveMu.RLock(); ls := live; liveMu.RUnlock()
	static(fmt.Sprintf("Pick: %d     Pack: %d     Phân ca: %d", len(ls.Pick.Rows), len(ls.Pack.Rows), len(ls.Shift.Rows)), 152, 120, 900, 30, true)
	static("HỆ THỐNG & HIỆU NĂNG", 152, 175, 400, 28, true)
	overviewMetric = static("Đang đo…", 152, 215, 1000, 28, false)
	renderOverviewMetrics()
	static("Dữ liệu được giữ cục bộ. GitHub/public Internet lỗi không làm dừng nghiệp vụ nội bộ.", 152, 345, 1000, 24, false)
}
func renderOverviewMetrics() {
	if currentPage != ID_NAV_OVERVIEW { return }
	var ms MEMORYSTATUSEX
	ms.Length = uint32(unsafe.Sizeof(ms))
	pGlobalMemory.Call(uintptr(unsafe.Pointer(&ms)))
	var rm runtime.MemStats; runtime.ReadMemStats(&rm)
	used := ms.TotalPhys - ms.AvailPhys
	txt := fmt.Sprintf("RAM laptop: %.1f / %.1f GB (%.0f%%)     RAM Go heap: %.1f MB     Goroutine: %d", float64(used)/1e9, float64(ms.TotalPhys)/1e9, float64(ms.MemoryLoad), float64(rm.Alloc)/1024/1024, runtime.NumGoroutine())
	if overviewMetric != 0 { setText(overviewMetric, txt) }
}
func renderActive() {
	pageTitle("ĐANG LẤY HÀNG", "Lọc trạng thái áp dụng ngay; nhấp đúp dòng để xem chi tiết.")
	static("Trạng thái", 152, 112, 75, 24, false)
	activeCombo = combo(ID_ACTIVE_STATUS, []string{"Tất cả", "Đang lấy", "Quá thời gian", "Hoàn thành"}, settings.Business.ActiveStatus, 230, 106, 150, 200)
	liveMu.RLock(); t := live.Active; liveMu.RUnlock()
	renderTable(filterActive(t, settings.Business.ActiveStatus), 145)
}
func renderPick() {
	pageTitle("PICK", "Target, tốc độ, khoán/chẵn-lẻ và kiểm tra 1C1L ở đúng màn nghiệp vụ.")
	static("Ca", 152, 112, 25, 22, false)
	pickShiftCombo = combo(ID_PICK_SHIFT, []string{"Tất cả", "Ca 1", "Ca 2", "Ca HC"}, settings.Business.PickShift, 180, 106, 100, 180)
	checkbox(ID_PICK_DEDUCT, "Khấu trừ SKU", 300, 108, 120, 24, settings.Business.PickDeductSKU)
	checkbox(ID_PICK_REQUIRE, "Kiểm tra đủ chẵn", 430, 108, 140, 24, settings.Business.PickRequireEven)
	checkbox(ID_PICK_ALLSITE, "Hiện tất cả Site", 580, 108, 130, 24, settings.Business.ShowAllSite)
	checkbox(ID_PICK_1C1L, "Bật 1 chẵn 1 lẻ", 720, 108, 140, 24, settings.Business.Enable1C1L)
	checkbox(ID_PICK_INCOMPLETE, "Chỉ hiện chưa đủ chẵn", 870, 108, 170, 24, settings.Business.ShowIncompleteEven)
	checkbox(ID_PICK_1C1LERR, "Chỉ hiện lỗi 1C1L", 1050, 108, 150, 24, settings.Business.Show1C1LErrors)
	liveMu.RLock(); t := live.Pick; liveMu.RUnlock()
	renderTable(t, 145)
}
func renderPack() {
	pageTitle("PACK", "Chỉ hiển thị theo ca; chọn là áp dụng ngay.")
	static("Hiển thị ca", 152, 112, 85, 22, false)
	packShiftCombo = combo(ID_PACK_SHIFT, []string{"Tất cả", "Ca 1", "Ca 2", "Ca HC"}, settings.Business.PackShift, 240, 106, 120, 180)
	liveMu.RLock(); t := live.Pack; liveMu.RUnlock()
	renderTable(t, 145)
}
func renderShift() {
	pageTitle("PHÂN CA", "Chọn một dòng, sau đó chọn Tự động / Ca 1 / Ca 2 / Ca HC.")
	static("Phân ca thủ công", 152, 112, 120, 22, false)
	shiftManualCombo = combo(ID_SHIFT_MANUAL, []string{"Tự động", "Ca 1", "Ca 2", "Ca HC"}, "Tự động", 280, 106, 130, 180)
	liveMu.RLock(); t := live.Shift; liveMu.RUnlock()
	renderTable(t, 145)
}
func renderUserPDA() {
	pageTitle("USER / PDA", "Dữ liệu nhân sự live/local; không có dữ liệu cá nhân nào được đóng gói trong bản public.")
	renderTable(core.Table{Headers: []string{"Họ và tên", "Mã nhân viên", "User", "Nhà cung cấp", "Site", "Vị trí công việc", "PDA", "Trạng thái"}}, 125)
}
func renderLog() {
	pageTitle("NHẬT KÝ", "Toàn bộ sự kiện kỹ thuật cần thiết, tự loại thông tin nhạy cảm.")
	button(ID_LOG_OPEN, "MỞ THƯ MỤC LOG", 152, 108, 160, 32)
	button(ID_LOG_EXPORT, "XUẤT CHẨN ĐOÁN", 322, 108, 170, 32)
	logEdit = create("EDIT", readLogTail(500), WS_CHILD|WS_VISIBLE|WS_BORDER|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY, 152, 145, 1100, 560, mainWnd, 0)
	setFont(logEdit, fontSmall); addPage(logEdit)
}
func renderSettings() {
	pageTitle("THIẾT LẬP", "Chỉ quản lý phiên/token Dashboard. Nghiệp vụ đặt tại đúng màn Pick/Pack/Phân ca.")
	static("PHIÊN DASHBOARD · DÁN cURL (BASH)", 152, 120, 420, 24, true)
	settingsCurl = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL, 152, 150, 720, 120, mainWnd, 0)
	setFont(settingsCurl, fontNormal); addPage(settingsCurl)
	button(ID_CURL_IMPORT, "NHẬN CẤU HÌNH", 152, 282, 150, 32)
	button(ID_SECRET_TOGGLE, "HIỆN / ẨN", 312, 282, 120, 32)
	button(ID_NET_TEST, "KIỂM TRA", 442, 282, 120, 32)
	settingsSummary = static(credentialSummary(), 152, 330, 900, 120, false)
	static("Repo public không chứa endpoint nội bộ, credential, log thô hoặc dữ liệu nhân sự. Runtime profile nằm cục bộ theo Windows user.", 152, 465, 1000, 44, false)
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
	if id >= ID_NAV_OVERVIEW && id <= ID_NAV_SETTINGS { renderPage(id); return }
	switch id {
	case ID_SYNC: startSync()
	case ID_UPDATE: go checkUpdateInteractive()
	case ID_CURL_IMPORT: importCurl()
	case ID_SECRET_TOGGLE:
		revealSecrets = !revealSecrets; setText(settingsSummary, credentialSummary()); logEvent("INFO", "SECRET_DISPLAY_TOGGLE", "visible", strconv.FormatBool(revealSecrets))
	case ID_NET_TEST: go networkTest()
	case ID_LOG_OPEN: openFolder(logDir())
	case ID_LOG_EXPORT: exportDiagnostic()
	case ID_ACTIVE_STATUS:
		if code == 1 { settings.Business.ActiveStatus = comboText(activeCombo); saveSettings(); renderPage(ID_NAV_ACTIVE) }
	case ID_PICK_SHIFT:
		if code == 1 { settings.Business.PickShift = comboText(pickShiftCombo); saveSettings(); recalculate(); renderPage(ID_NAV_PICK) }
	case ID_PACK_SHIFT:
		if code == 1 { settings.Business.PackShift = comboText(packShiftCombo); saveSettings(); recalculate(); renderPage(ID_NAV_PACK) }
	case ID_PICK_DEDUCT:
		settings.Business.PickDeductSKU = checked(source); saveSettings(); recalculate(); renderPage(ID_NAV_PICK)
	case ID_PICK_REQUIRE:
		settings.Business.PickRequireEven = checked(source); saveSettings(); recalculate(); renderPage(ID_NAV_PICK)
	case ID_PICK_ALLSITE:
		settings.Business.ShowAllSite = checked(source); saveSettings(); recalculate(); renderPage(ID_NAV_PICK)
	case ID_PICK_1C1L:
		settings.Business.Enable1C1L = checked(source); saveSettings(); recalculate(); renderPage(ID_NAV_PICK)
	case ID_PICK_INCOMPLETE:
		settings.Business.ShowIncompleteEven = checked(source); saveSettings(); recalculate(); renderPage(ID_NAV_PICK)
	case ID_PICK_1C1LERR:
		settings.Business.Show1C1LErrors = checked(source); saveSettings(); recalculate(); renderPage(ID_NAV_PICK)
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

func startSync() {
	if !busy.CompareAndSwap(false, true) {
		setStatus("Đang đồng bộ; không tạo thêm tác vụ chồng nhau.")
		return
	}
	setStatus("Đang đồng bộ…")
	logEvent("INFO", "SYNC_START")
	go func() {
		defer busy.Store(false)
		defer pPostMessage.Call(mainWnd, WM_APP_REFRESH, 0, 0) // Public source intentionally externalizes private operational endpoints.
		if creds.URL == "" {
			setStatus("Chưa có phiên Dashboard. Vào Thiết lập → dán cURL bash.")
			logEvent("WARN", "SYNC_NO_CREDENTIAL")
			return
		}
		if err := requestProbe(creds); err != nil {
			setStatus("Kết nối phiên lỗi; giữ dữ liệu cũ.")
			logEvent("ERROR", "SYNC_PROBE_FAILED", "error", err.Error())
			return
		}
		if len(profile.Bindings) == 0 {
			setStatus("Phiên hợp lệ; chưa có runtime profile live cục bộ.")
			logEvent("WARN", "SYNC_PROFILE_MISSING")
			return
		}
		setStatus(fmt.Sprintf("Phiên hợp lệ · runtime profile %s · %d binding(s).", profile.ProfileID, len(profile.Bindings)))
		logEvent("INFO", "SYNC_SESSION_OK", "profile_id", sanitizeProfileID(profile.ProfileID), "binding_count", strconv.Itoa(len(profile.Bindings)))
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
	raw := getText(settingsCurl)
	c, e := parseCurl(raw)
	setText(settingsCurl, "")
	if e != nil {
		setStatus("Không nhận được cURL: " + e.Error())
		logEvent("WARN", "CURL_IMPORT_FAILED", "error", e.Error())
		return
	}
	creds = c
	if e = saveCredentials(); e != nil {
		setStatus("Không lưu được phiên: " + e.Error())
		logEvent("ERROR", "CREDENTIAL_SAVE_FAILED", "error", e.Error())
		return
	}
	setText(settingsSummary, credentialSummary())
	setStatus("Đã nhận cấu hình; cURL gốc đã xóa khỏi ô nhập.")
	logEvent("INFO", "CURL_IMPORTED", "authorization_present", strconv.FormatBool(c.Authorization != ""), "signature_present", strconv.FormatBool(c.Signature != ""))
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
	return "Authorization/Token: " + mask(firstNonEmpty(creds.Token, creds.Authorization)) + "\r\nAPISID: " + mask(creds.APISID) + "    USID: " + mask(creds.USID) + "\r\nx-signature: " + mask(creds.Signature) + "    nonce: " + mask(creds.Nonce) + "\r\nRequest URL: " + map[bool]string{true: "Đã nhận", false: "Chưa có"}[creds.URL != ""]
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
	if creds.URL == "" {
		setStatus("Chưa có cURL để kiểm tra.")
		return
	}
	setStatus("Đang kiểm tra kết nối…")
	start := time.Now()
	e := requestProbe(creds)
	if e != nil {
		setStatus("Kiểm tra lỗi: " + e.Error())
		logEvent("WARN", "NETWORK_TEST_FAILED", "elapsed_ms", strconv.FormatInt(time.Since(start).Milliseconds(), 10), "error", e.Error())
		return
	}
	setStatus("Kết nối phiên hoạt động.")
	logEvent("INFO", "NETWORK_TEST_OK", "elapsed_ms", strconv.FormatInt(time.Since(start).Milliseconds(), 10))
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
func logDir() string {
	if settings.DataFolder != "" {
		return filepath.Join(settings.DataFolder, "Logs")
	}
	return filepath.Join(appDir(), "Logs")
}
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
	

func runtimeProfilePath() string { return filepath.Join(secureDir(), "runtime-profile.dat") }
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
	logEvent("INFO", "RUNTIME_PROFILE_IMPORTED", "profile_id", sanitizeProfileID(p.ProfileID), "binding_count", strconv.Itoa(len(p.Bindings)))
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
	if len(v) > 800 {
		v = v[:800] + "…"
	}
	return v
}
func logEvent(level, event string, kv ...string) {
	ensureDirs()
	f, e := os.OpenFile(filepath.Join(logDir(), "supra_"+time.Now().Format("20060102")+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s %s", time.Now().Format(time.RFC3339), level, event)
	for i := 0; i+1 < len(kv); i += 2 {
		fmt.Fprintf(f, " %s=%q", kv[i], sanitizeValue(kv[i], kv[i+1]))
	}
	fmt.Fprintln(f)
}
func readLogTail(max int) string {
	b, e := os.ReadFile(filepath.Join(logDir(), "supra_"+time.Now().Format("20060102")+".log"))
	if e != nil {
		return "Chưa có log."
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) > max {
		lines = lines[len(lines)-max:]
	}
	return strings.Join(lines, "\r\n")
}
func shortHash(s string) string { h := sha256.Sum256([]byte(s)); return fmt.Sprintf("%x", h[:4]) }
func exportDiagnostic() {
	ensureDirs()
	name := filepath.Join(settings.DataFolder, "Diagnostics", "SUPRA_DIAGNOSTIC_"+time.Now().Format("20060102_150405")+".txt")
	os.MkdirAll(filepath.Dir(name), 0700)
	var rm runtime.MemStats
	runtime.ReadMemStats(&rm)
	content := fmt.Sprintf("version=%s\ntime=%s\napp_heap_mb=%.1f\ngoroutines=%d\ncredential_present=%t\n\n%s", appVersion, time.Now().Format(time.RFC3339), float64(rm.Alloc)/1024/1024, runtime.NumGoroutine(), creds.URL != "", readLogTail(1000))
	if e := os.WriteFile(name, []byte(content), 0600); e != nil {
		setStatus("Xuất chẩn đoán lỗi: " + e.Error())
		return
	}
	setStatus("Đã tạo chẩn đoán: " + name)
	logEvent("INFO", "DIAGNOSTIC_EXPORT", "file", filepath.Base(name))
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
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+updateRepo+k"/releases/latest", nil)
	req.Header.Set("User-Agent", "SupraProductivity/"+appVersion)
	c := http.Client{Timeout: 8 * time.Second}
	resp, e := c.Do(req)
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
func normalizeVersion(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }
func checkUpdateQuiet() {
	r, e := latestRelease()
	if e != nil {
		logEvent("INFO", "UPDATE_CHECK_UNAVAILABLE", "error", e.Error())
		return
	}
	logEvent("INFO", "UPDATE_CHECK_OK", "latest", r.TagName)
	if normalizeVersion(r.TagName) != normalizeVersion(appVersion) && appVersion != "dev" {
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
	if normalizeVersion(r.TagName) == normalizeVersion(appVersion) {
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
	pInitCommon.Call()
	hInst, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	cls := ptr("SupraProductivityWindow")
	wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hInst, HCursor: func() uintptr { r, _, _ := pLoadCursor.Call(0, 32512); return r }(), LpszClassName: cls}
	if r, _, _ := pRegisterClass.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		panic("RegisterClassExW failed")
	}
	title := appName + " · " + appVersion
	hwnd, _, _ := pCreateWindow.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(ptr(title))), WS_OVERLAPPEDWINDOW|WS_VISIBLE, CW_USEDEFAULT, CW_USEDEFAULT, 1500, 880, 0, 0, hInst, 0)
	if hwnd == 0 {
		panic("CreateWindowExW failed")
	}
	pShowWindow.Call(hwnd, SW_SHOW)
	pUpdateWindow.Call(hwnd)
	var msg MSG
	for {
		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
