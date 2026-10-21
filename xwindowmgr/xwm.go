package xwindowmgr

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include "x11gohelper.hpp"

#define MAX(a, b) ((a) > (b) ? (a) : (b))
*/
import "C"
import (
	"context"
	"fmt"
	"unsafe"
)

type XWindowManager struct {
	backend       *XBackend
	clientManager *ClientManager
	rootWindow    C.Window

	eventChan chan WMEvent
	loopDone  chan struct{}
}

func NewXWindowManager() (*XWindowManager, error) {
	backend, err := NewXBackend()
	if err != nil {
		return nil, err
	}

	rootWin := backend.GetRootWindow()

	return &XWindowManager{
		backend:       backend,
		clientManager: NewClientManager(),
		rootWindow:    rootWin,
		eventChan:     make(chan WMEvent, 200),
		loopDone:      make(chan struct{}),
	}, nil
}

func (xm *XWindowManager) Start(ctx context.Context) {

	xm.GrabShortcuts()
	xm.backend.Sync(false)
	xm.backend.SelectInput(int64(C.KeyPressMask | C.SubstructureRedirectMask))
	xm.ScanExistingWindows()

	go xm.EventLoop(ctx, xm.eventChan)

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-xm.eventChan:
			xm.handleWMEvent(event)
		}
	}
}

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
		fmt.Println("Cannot convert data XExposeEvent.")
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

func (xm *XWindowManager) ScanExistingWindows() {
	var rootReturn, parentReturn C.Window
	var children *C.Window
	var nchildren C.uint

	status := C.XQueryTree(xm.backend.display, xm.rootWindow, &rootReturn, &parentReturn, &children, &nchildren)

	if status == 0 || nchildren == 0 {
		return
	}
	defer C.XFree(unsafe.Pointer(children))

	windowSlice := unsafe.Slice(children, int(nchildren))

	for _, window := range windowSlice {
		var attrs C.XWindowAttributes
		C.XGetWindowAttributes(xm.backend.display, window, &attrs)

		if attrs.map_state == C.IsViewable && attrs.override_redirect == 0 {
			client := generateClientContext(xm.backend, window)
			client.CreateFrame(xm.backend, xm.rootWindow)
			fmt.Println("XWM: adding exisiting window \"", client.Window, "\"; create frame window \"", client.Frame, "\"; Window Title:", client.Title)
			xm.clientManager.Add(client)
			C.XMapWindow(xm.backend.display, client.Frame)
			C.XMapWindow(xm.backend.display, client.Window)
		}
	}
}
