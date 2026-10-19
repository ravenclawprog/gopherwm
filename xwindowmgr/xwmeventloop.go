package xwindowmgr

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include "x11gohelper.hpp"

*/
import "C"
import (
	"context"
	"time"
)

type WMEvent struct {
	Type   int
	Window C.Window
	Data   interface{}
}

func (xm *XWindowManager) EventLoop(ctx context.Context, eventChan chan<- WMEvent) {
	var ev C.XEvent
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if C.XPending(xm.backend.display) > 0 {
				C.XNextEvent(xm.backend.display, &ev)

				evType := unionToInt(ev)
				wmEv := xm.parseXEvent(evType, ev)
				eventChan <- wmEv
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func (xm *XWindowManager) parseXEvent(evType int, ev C.XEvent) WMEvent {
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
	case C.ButtonPress:
		bp := unionToXButtonEvent(ev)
		wmEv.Window = bp.window
		wmEv.Data = *bp
	case C.ButtonRelease:
		br := unionToXButtonEvent(ev)
		wmEv.Window = br.window
		wmEv.Data = *br
	case C.MotionNotify:
		me := unionToXMotionEvent(ev)
		wmEv.Window = me.window
		wmEv.Data = *me
	}

	return wmEv
}
