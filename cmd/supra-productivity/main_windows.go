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
