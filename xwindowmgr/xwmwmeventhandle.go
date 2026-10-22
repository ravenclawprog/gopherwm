package xwindowmgr

/*
#include <X11/Xlib.h>
*/
import "C"
import (
	"context"
	"fmt"
)

func (xm *XWindowManager) handleWMEvent(event WMEvent) {
	switch event.Type {
	case C.MapRequest:
		data, ok := event.Data.(C.XMapRequestEvent)
		if ok {
			xm.handleMapRequest(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XMapRequest.")

	case C.UnmapNotify:
		data, ok := event.Data.(C.XUnmapEvent)
		if ok {
			xm.handleUnmapNotify(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XUnmapEvent.")
	case C.KeyPress:
		data, ok := event.Data.(C.XKeyEvent)
		if ok {
			xm.handleKeyPress(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XKeyEvent.")
	case C.ConfigureRequest:
		data, ok := event.Data.(C.XConfigureRequestEvent)
		if ok {
			xm.handleConfigureRequest(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XConfigureRequestEvent.")
	case C.Expose:
		data, ok := event.Data.(C.XExposeEvent)
		if ok {
			xm.handleExpose(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XExposeEvent.")
	case C.ButtonPress:
		data, ok := event.Data.(C.XButtonEvent)
		if ok {
			xm.handleButtonPress(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XButtonEvent(ButtonPress).")
	case C.ButtonRelease:
		data, ok := event.Data.(C.XButtonEvent)
		if ok {
			xm.handleButtonRelease(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XButtonEvent(ButtonRelease).")
	case C.MotionNotify:
		data, ok := event.Data.(C.XMotionEvent)
		if ok {
			xm.handleMotionEvent(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XMotionEvent.")
	case C.DestroyNotify:
		data, ok := event.Data.(C.XDestroyWindowEvent)
		if ok {
			xm.handleDestroyNotify(event.Window, &data)
			return
		}
		fmt.Println("Cannot convert data XDestroyNotify.")
	default:
		fmt.Println("Unhandled message type:", event.Type)
	}
}

func (xm *XWindowManager) HandleWMEvent(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-xm.eventChan:
			xm.handleWMEvent(event)
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
