package windo

import (
	"unsafe"

	"github.com/lxn/win"
)

func (t *Try) GetModuleHandle(name *uint16) win.HANDLE {
	return win.HANDLE(t.Kernel32nonzero("GetModuleHandleW",
		uintptr(unsafe.Pointer(name)),
	))
}

func (t *Try) DefWindowProc(hwnd win.HANDLE, msg uint32, wparam, lparam uintptr) uintptr {
	return t.User32any("DefWindowProcW",
		uintptr(hwnd),
		uintptr(msg),
		wparam,
		lparam,
	)
}

func (t *Try) RegisterClassEx(wc *win.WNDCLASSEX) win.ATOM {
	return win.ATOM(t.User32nonzero("RegisterClassExW",
		uintptr(unsafe.Pointer(wc)),
	))
}

func (t *Try) CreateWindowEx(exstyle uint32, classname, windowname *uint16, style uint32, x, y, w, h int32, parent, menu, hinst win.HANDLE, param uintptr) win.HANDLE {
	return win.HANDLE(t.User32nonzero("CreateWindowExW",
		uintptr(exstyle),
		uintptr(unsafe.Pointer(classname)),
		uintptr(unsafe.Pointer(windowname)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(w),
		uintptr(h),
		uintptr(parent),
		uintptr(menu),
		uintptr(hinst),
		uintptr(param),
	))
}

func (t *Try) ShowWindow(hwnd win.HANDLE, cmdshow int32) bool {
	return 0 != t.User32any("ShowWindow",
		uintptr(hwnd),
		uintptr(cmdshow),
	)
}

func (t *Try) UpdateWindow(hwnd win.HANDLE) {
	t.User32nonzero("UpdateWindow",
		uintptr(hwnd),
	)
}

func (t *Try) GetMessage(msg *win.MSG, hwnd win.HANDLE, msgfiltermin, msgfiltermax uint32) bool {
	r, err := t.User32("GetMessageW",
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgfiltermin),
		uintptr(msgfiltermax),
	)
	if r == ^uintptr(0) {
		t.Failf(err, "GetMessage")
	}
	return r != 0
}

func (t *Try) TranslateMessage(msg *win.MSG) bool {
	return 0 != t.User32any("TranslateMessage",
		uintptr(unsafe.Pointer(msg)),
	)
}

func (t *Try) DispatchMessage(msg *win.MSG) uint32 {
	return uint32(t.User32any("DispatchMessageW",
		uintptr(unsafe.Pointer(msg)),
	))
}

func (t *Try) DestroyWindow(hwnd win.HANDLE) {
	t.User32nonzero("DestroyWindow",
		uintptr(hwnd),
	)
}

func (t *Try) PostQuitMessage(exitcode uint32) {
	t.User32any("PostQuitMessage",
		uintptr(exitcode),
	)
}
