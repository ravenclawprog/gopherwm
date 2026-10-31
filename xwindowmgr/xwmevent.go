package xwindowmgr

/*
#include <X11/Xlib.h>
*/
import "C"
import "unsafe"

// WMEvent owns a copy of the event payload; it does not reference the Xlib buffer.
type WMEvent struct {
	Type   int
	Window C.Window
	Data   any
}

func parseXEvent(ev C.XEvent) WMEvent {
	evType := unionToInt(ev)
	var wmEv WMEvent
	wmEv.Type = evType
	switch evType {
	case C.MapRequest:
		mq := unionToXMapRequestEvent(ev)
		wmEv.Window = mq.window
		wmEv.Data = *mq

	case C.UnmapNotify:
		unmap := unionToXUnmapEvent(ev)
		wmEv.Window = unmap.window
		wmEv.Data = *unmap

	case C.DestroyNotify:
		destroy := unionToXDestroyWindowEvent(ev)
		wmEv.Window = destroy.window
		wmEv.Data = *destroy

	case C.ConfigureRequest:
		cr := unionToXConfigureRequestEvent(ev)
		wmEv.Window = cr.window
		wmEv.Data = *cr

	case C.KeyPress:
		kp := unionToXKeyEvent(ev)
		wmEv.Window = kp.window
		wmEv.Data = *kp
	case C.Expose:
		ee := unionToXExposeEvent(ev)
		wmEv.Window = ee.window
		wmEv.Data = *ee
	case C.ButtonPress, C.ButtonRelease:
		button := unionToXButtonEvent(ev)
		wmEv.Window = button.window
		wmEv.Data = *button
	case C.MotionNotify:
		me := unionToXMotionEvent(ev)
		wmEv.Window = me.window
		wmEv.Data = *me
	}

	return wmEv
}

func unionToInt(cbytes C.XEvent) (result int) {
	var addr *byte = &cbytes[0]
	var cast *C.int = (*C.int)(unsafe.Pointer(addr))
	return (int)(*cast)
}

func unionToXKeyEvent(cbytes C.XEvent) (result *C.XKeyEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XKeyEvent = (*C.XKeyEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXDestroyWindowEvent(cbytes C.XEvent) (result *C.XDestroyWindowEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XDestroyWindowEvent = (*C.XDestroyWindowEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXUnmapEvent(cbytes C.XEvent) (result *C.XUnmapEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XUnmapEvent = (*C.XUnmapEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXMapRequestEvent(cbytes C.XEvent) (result *C.XMapRequestEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XMapRequestEvent = (*C.XMapRequestEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXConfigureRequestEvent(cbytes C.XEvent) (result *C.XConfigureRequestEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XConfigureRequestEvent = (*C.XConfigureRequestEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXExposeEvent(cbytes C.XEvent) (result *C.XExposeEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XExposeEvent = (*C.XExposeEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXButtonEvent(cbytes C.XEvent) (result *C.XButtonEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XButtonEvent = (*C.XButtonEvent)(unsafe.Pointer(addr))
	return cast
}
func unionToXMotionEvent(cbytes C.XEvent) (result *C.XMotionEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XMotionEvent = (*C.XMotionEvent)(unsafe.Pointer(addr))
	return cast
}
