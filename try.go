/*
Package winq provides functions for quick & dirty WinAPI calls, with easy errors capturing.

Sample usage:

	var try winq.Try
	r := try.N("MessageBox", 0, syscall.StringToUTF16Ptr(msg), syscall.StringToUTF16Ptr(title), 0)
	if try.Err != nil {
		panic(try.Err.Error())
	}
	println("got:", r)

For more details, see description of function Try.F.
*/
package winq

import (
	"fmt"
	"reflect"
	"syscall"
)

var (
	procs = map[string]*syscall.Proc{}
	Dlls  = []*syscall.DLL{
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
	Err error
}

/*
Function F tries to call a WinAPI procedure with given name with specified args.
A common set of DLLs is searched (as listed in Dlls package variable) for a
function with given name, or name + "W". A call is then made with specified arguments,
which must be convertible to uintptr (allowed are: all ints, uints, pointers, bool, nil).
The return value is stored in r, and result of GetLastError call is stored in lastErr.

As WinAPI functions tend to signal error in the returned value, a convenient set of
wrappers is provided as shortcuts for common cases:

	try.N(name, args) - when Nonzero result means success
	try.Z(name, args) - when Zero result means success
	try.A(name, args) - when Any result means success
	try.X(isok, name, args) - for more complex cases, isok condition is used to detect success

Function F (and the wrappers) panic if they can't find the procedure by given name,
or cannot convert some of the arguments to uintptr.

The functions DON'T DO ANYTHING at all if any previous call with the same Try object
has failed (and thus t.Err is non-nil). This results in the following patterns:

	// You can chain multiple calls, and they'll do nothing after first failure
	a := try.N("Foo")
	b := try.N("Bar", a)
	try.N("Baz", b)
	if try.Err != nil {
		return try.Err
	}


	// For some functions, like EndPaint, you want them to get called even if something fails later
	dc := try.N("BeginPaint", hwnd, &paintstruct)
	defer try.Detach().N("EndPaint", hwnd, &paintstruct)
	memdc := try.N("CreateCompatibleDC", dc) // if this fails, above EndPaint will still run
	defer try.Detach().N("DeleteDC", memdc)
	if try.Err != nil {
		return try.Err
	}


IMPORTANT CAVEAT: F has package-level cache, mapping function names to their DLL
addresses; you MUST NOT call it from different threads/goroutines! The library is
intended as fairly Quick And Dirty; this particular behavior may get improved in future,
but no promises for now.
*/
func (t *Try) F(name string, args ...interface{}) (r uintptr, lastErr error) {
	if t.Err != nil {
		return
	}

	//FIXME: no synchronization!
	p := procs[name]
	if p != nil {
		goto found
	}
	for _, m := range Dlls {
		for _, fullname := range []string{name, name + "W"} {
			var err error
			p, err = m.FindProc(fullname)
			if err != nil {
				continue
			}
			procs[name] = p
			goto found
		}
	}
	//FIXME: for now, fail hard, but "in future" do something safer
	panic("WinAPI procedure '" + name + "' not found in predefined DLLs")

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
			panic("unknown kind " + v.Kind().String() + " in argument #" + string([]byte{'0' + byte(i)}) + " to " + name)
		}
	}

	r1, _, lastErr := p.Call(raws...)
	return r1, lastErr
}

func (t *Try) failf(orig error, format string, args ...interface{}) {
	if t.Err != nil {
		return
	}
	t.Err = Error{orig, fmt.Sprintf(format, args...)}
}

// Function N calls WinAPI procedure, and treats nonzero result as success.
// Otherwise, error information is stored in t.Err. For details, see function F.
func (t *Try) N(name string, args ...interface{}) uintptr {
	r, err := t.F(name, args...)
	if r == 0 {
		t.failf(err, "%s", name)
	}
	return r
}

// Function Z calls WinAPI procedure, and treats zero result as success.
// Otherwise, error information is stored in t.Err. For details, see function F.
func (t *Try) Z(name string, args ...interface{}) uintptr {
	r, err := t.F(name, args...)
	if r != 0 {
		t.failf(err, "%s", name)
	}
	return r
}

// Function A calls WinAPI procedure, and treats all results as success.
// For details, see function F.
func (t *Try) A(name string, args ...interface{}) uintptr {
	r, _ := t.F(name, args...)
	return r
}

// Function X calls WinAPI procedure, and assumes success if isok returns true for result.
// Otherwise, error information is stored in t.Err. For details, see function F.
func (t *Try) X(isok func(r uintptr) bool, name string, args ...interface{}) uintptr {
	r, err := t.F(name, args...)
	if !isok(r) {
		t.failf(err, "%s", name)
	}
	return r
}

type Tryer interface {
	N(name string, args ...interface{}) uintptr
	Z(name string, args ...interface{}) uintptr
	A(name string, args ...interface{}) uintptr
	X(isok func(r uintptr) bool, name string, args ...interface{}) uintptr
}

func (t *Try) Detach() Tryer {
	if t.Err != nil {
		return nop{}
	} else {
		return &Try{}
	}
}

type nop struct{}

func (nop) N(name string, args ...interface{}) uintptr                            { return 0 }
func (nop) Z(name string, args ...interface{}) uintptr                            { return 0 }
func (nop) A(name string, args ...interface{}) uintptr                            { return 0 }
func (nop) X(isok func(r uintptr) bool, name string, args ...interface{}) uintptr { return 0 }
