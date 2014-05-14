package main

import (
	"github.com/akavel/winq"
	"syscall"
)

var msg, title = "Hello, world", "Example"

func main() {
	var try winq.Try
	r := try.N("MessageBox", 0, syscall.StringToUTF16Ptr(msg), syscall.StringToUTF16Ptr(title), 0)
	if try.Err != nil {
		panic(try.Err.Error())
	}
	println("got:", r)
}
