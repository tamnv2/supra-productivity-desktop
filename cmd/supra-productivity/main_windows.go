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
		mainWnd = hwnd; createShell(); loadSettings(); loadCredentials()
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
