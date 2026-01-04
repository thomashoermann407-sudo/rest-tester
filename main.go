package main

import (
	"runtime"

	"hoermi.com/rest-test/win32"
)

func main() {
	runtime.LockOSThread()
	NewProjectWindow()
	win32.Run()
}
