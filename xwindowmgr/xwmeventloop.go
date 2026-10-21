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
				time.Sleep(15 * time.Millisecond)
			}

		}
	}
}
