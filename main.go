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
			// Handle key press event (e.g., logging, remapping, etc.)
			fmt.Printf("Key pressed: %d\n", vkCode)
		}
	}
	return CallNextHookEx(hook, nCode, wParam, lParam)
}

// UnhookWindowsHookEx uninstalls the hook procedure.
func UnhookWindowsHookEx(hhk HHOOK) bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(uintptr(hhk))
	return ret != 0
}
