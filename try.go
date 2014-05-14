package winq

import (
	"fmt"
	"reflect"
	"syscall"
)

var (
	procs = map[string]*syscall.Proc{}
	mods  = []*syscall.DLL{
		syscall.MustLoadDLL("kernel32.dll"),
		syscall.MustLoadDLL("user32.dll"),
		syscall.MustLoadDLL("gdi32.dll"),
	}
)

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

func (t *Try) F(proc string, args ...interface{}) (r uintptr, lastErr error) {
	if t.Failed {
		return
	}

	//FIXME: no synchronization!
	p := procs[proc]
	if p != nil {
		goto found
	}
	for _, m := range mods {
		for _, fullproc := range []string{proc, proc + "W"} {
			var err error
			p, err = m.FindProc(fullproc)
			if err != nil {
				continue
			}
			procs[proc] = p
			goto found
		}
	}
	//FIXME: for now, fail hard, but "in future" do something safer
	panic("WinAPI procedure '" + proc + "' not found in predefined DLLs")

found:
	raws := make([]uintptr, len(args))
	for i, arg := range args {
		v := reflect.ValueOf(arg)
		switch v.Kind() {
		case reflect.Ptr:
			raws[i] = v.Pointer()
		case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			raws[i] = uintptr(v.Uint())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			raws[i] = uintptr(v.Int())
		case reflect.Bool:
			if v.Bool() {
				raws[i] = uintptr(1)
			}
		case reflect.Invalid:
			raws[i] = uintptr(0)
		default:
			panic("unknown kind " + v.Kind().String() + " in argument #" + string([]byte{'0' + byte(i)}) + " to " + proc)
		}
	}

	r1, _, lastErr := p.Call(raws...)
	return r1, lastErr
}

func (t *Try) Failf(orig error, format string, args ...interface{}) {
	if t.Failed {
		return
	}
	t.Failed = true
	t.Err = Error{orig, fmt.Sprintf(format, args...)}
}

// N calls F, and checks if result is nonzero, otherwise remembers lastErr.
func (t *Try) N(proc string, args ...interface{}) uintptr {
	r, err := t.F(proc, args...)
	if r == 0 {
		t.Failf(err, "%s", proc)
	}
	return r
}

// Z calls F, and checks if result is zero, otherwise remembers lastErr.
func (t *Try) Z(proc string, args ...interface{}) uintptr {
	r, err := t.F(proc, args...)
	if r != 0 {
		t.Failf(err, "%s", proc)
	}
	return r
}

func (t *Try) A(proc string, args ...interface{}) uintptr {
	r, _ := t.F(proc, args...)
	return r
}

func (t *Try) X(isok func(r uintptr) bool, proc string, args ...interface{}) uintptr {
	r, err := t.F(proc, args...)
	if !isok(r) {
		t.Failf(err, "%s", proc)
	}
	return r
}
