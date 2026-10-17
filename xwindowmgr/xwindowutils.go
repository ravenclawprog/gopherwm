package xwindowmgr

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
*/
import "C"
import "unsafe"

func unionToInt(cbytes [192]byte) (result int) {
	var addr *byte = &cbytes[0]
	var cast *C.int = (*C.int)(unsafe.Pointer(addr))
	return (int)(*cast)
}

func unionToXKeyEvent(cbytes [192]byte) (result *C.XKeyEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XKeyEvent = (*C.XKeyEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXMapRequestEvent(cbytes [192]byte) (result *C.XMapRequestEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XMapRequestEvent = (*C.XMapRequestEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXConfigureRequestEvent(cbytes [192]byte) (result *C.XConfigureRequestEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XConfigureRequestEvent = (*C.XConfigureRequestEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXExposeEvent(cbytes [192]byte) (result *C.XExposeEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XExposeEvent = (*C.XExposeEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXMapEvent(cbytes [192]byte) (result *C.XMapEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XMapEvent = (*C.XMapEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXButtonEvent(cbytes [192]byte) (result *C.XButtonEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XButtonEvent = (*C.XButtonEvent)(unsafe.Pointer(addr))
	return cast
}

func max(a, b C.int) C.uint {
	if a > b {
		return C.uint(a)
	}

	return C.uint(b)
}
