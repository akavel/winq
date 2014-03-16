package windo

import (
	"fmt"
	"syscall"
)

var modCache = map[string]*syscall.DLL{}
var procCache = map[string]*syscall.Proc{}

type Error struct {
	Original error
	Msg      string
	//TODO: Stack []byte
}

func (e Error) Error() string { return e.Msg + ": " + e.Original.Error() }

type Try struct {
	Failed bool
	Err    error
}

func (t *Try) Failf(orig error, format string, args ...interface{}) {
	if t.Failed {
		return
	}
	t.Failed = true
	t.Err = Error{orig, fmt.Sprintf(format, args...)}
}

func (t *Try) Kernel32(proc string, args ...uintptr) (uintptr, error) {
	return t.Call("kernel32.dll", proc, args...)
}
func (t *Try) User32(proc string, args ...uintptr) (uintptr, error) {
	return t.Call("user32.dll", proc, args...)
}

func (t *Try) Kernel32nonzero(proc string, args ...uintptr) uintptr {
	return t.CallNonzero("kernel32.dll", proc, args...)
}
func (t *Try) User32nonzero(proc string, args ...uintptr) uintptr {
	return t.CallNonzero("user32.dll", proc, args...)
}

func (t *Try) User32any(proc string, args ...uintptr) uintptr {
	r, _ := t.User32(proc, args...)
	return r
}

func (t *Try) Call(dll, proc string, args ...uintptr) (uintptr, error) {
	if t.Failed {
		return 0, t.Err
	}

	p := procCache[proc]
	if p == nil {
		m := modCache[dll]
		if m == nil {
			m, t.Err = syscall.LoadDLL(dll)
			if t.Err != nil {
				t.Failf(t.Err, "LoadDLL(%s)", dll)
				return 0, t.Err
			}
			modCache[dll] = m
		}
		p, t.Err = m.FindProc(proc)
		if t.Err != nil {
			t.Failf(t.Err, "FindProc(%s, %s)", dll, proc)
			return 0, t.Err
		}
		procCache[proc] = p
	}

	r, _, err := p.Call(args...)
	return r, err
}

func (t *Try) CallNonzero(dll, proc string, args ...uintptr) uintptr {
	r, err := t.Call(dll, proc, args...)
	if r == 0 {
		t.Failf(err, "%s", proc)
	}
	return r
}
