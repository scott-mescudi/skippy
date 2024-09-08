package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	kernel32                = windows.NewLazySystemDLL("kernel32.dll")
	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")

	hook HHOOK
)

var asciiMap = map[int]string{
	32:  "Space",
	8: "BSPACE",
	48:  "0",
	49:  "1",
	50:  "2",
	51:  "3",
	52:  "4",
	53:  "5",
	54:  "6",
	55:  "7",
	56:  "8",
	57:  "9",
	65:  "A",
	66:  "B",
	67:  "C",
	68:  "D",
	69:  "E",
	70:  "F",
	71:  "G",
	72:  "H",
	73:  "I",
	74:  "J",
	75:  "K",
	76:  "L",
	77:  "M",
	78:  "N",
	79:  "O",
	80:  "P",
	81:  "Q",
	82:  "R",
	83:  "S",
	84:  "T",
	85:  "U",
	86:  "V",
	87:  "W",
	88:  "X",
	89:  "Y",
	90:  "Z",
	97:  "a",
	98:  "b",
	99:  "c",
	100: "d",
	101: "e",
	102: "f",
	103: "g",
	104: "h",
	105: "i",
	106: "j",
	107: "k",
	108: "l",
	109: "m",
	110: "n",
	111: "o",
	112: "p",
	113: "q",
	114: "r",
	115: "s",
	116: "t",
	117: "u",
	118: "v",
	119: "w",
	120: "x",
	121: "y",
	122: "z",
	13: "Enter",
	27: "ESC",
	160: "LSHIFT",
	161: "RSHIFT",
	20: "Caps lock",
	162: "LCTRL",
	163: "RCTRL",
    165: "PAUSE",
    166: "SCROLL",
    167: "HOME",
	91: "WIN",
	9: "TAB",
}

type HHOOK uintptr
type WPARAM uintptr
type LPARAM uintptr
type LRESULT uintptr
type KBDLLHOOKSTRUCT struct {
	VkCode     uint32
	ScanCode   uint32
	Flags      uint32
	Time       uint32
	ExtraInfo  uintptr
}

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	HC_ACTION      = 0
)

type POINT struct {
	X, Y int32
}

type MSG struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  WPARAM
	LParam  LPARAM
	Time    uint32
	Pt      POINT
	LPrivate uint32
}

func main() {
	hInstance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	hook = SetWindowsHookEx(WH_KEYBOARD_LL, syscall.NewCallback(LowLevelKeyboardProc), hInstance, 0)

	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if r == 0 {
			break
		}
	}
}

// SetWindowsHookEx installs a hook procedure.
func SetWindowsHookEx(idHook int, lpfn uintptr, hmod uintptr, dwThreadId uint32) HHOOK {
	ret, _, _ := procSetWindowsHookExW.Call(uintptr(idHook), lpfn, hmod, uintptr(dwThreadId))
	return HHOOK(ret)
}

// CallNextHookEx passes the hook information to the next hook procedure in the chain.
func CallNextHookEx(hhk HHOOK, nCode int, wParam WPARAM, lParam LPARAM) LRESULT {
	ret, _, _ := procCallNextHookEx.Call(uintptr(hhk), uintptr(nCode), uintptr(wParam), uintptr(lParam))
	return LRESULT(ret)
}

// LowLevelKeyboardProc is the hook procedure for low-level keyboard input.
func LowLevelKeyboardProc(nCode int, wParam WPARAM, lParam LPARAM) LRESULT {
	if nCode == HC_ACTION {
		kbdStruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		if wParam == WM_KEYDOWN {
			vkCode := kbdStruct.VkCode
			
			value, exists := asciiMap[int(vkCode)]
			if exists {
                fmt.Printf("Keydown: %s (0x%04X)\n", value, vkCode)
            } else {
                fmt.Printf("Keydown: Unknown (%d)\n", vkCode)
            }
		}
	}
	return CallNextHookEx(hook, nCode, wParam, lParam)
}

// UnhookWindowsHookEx uninstalls the hook procedure.
func UnhookWindowsHookEx(hhk HHOOK) bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(uintptr(hhk))
	return ret != 0
}
