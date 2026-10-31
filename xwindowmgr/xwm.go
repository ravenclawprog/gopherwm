package xwindowmgr

/*
#include <X11/Xlib.h>
#include <X11/keysym.h>
*/
import "C"
import (
	"context"
	"fmt"
	"sync/atomic"
)

type XWindowManager struct {
	backend       *XBackend
	clientManager *ClientManager
	rootWindow    C.Window
	rootGeom      WindowRect

	eventChan chan WMEvent
	loopDone  chan struct{}

	startMoving atomic.Bool
}

func NewXWindowManager() (*XWindowManager, error) {
	backend, err := NewXBackend()
	if err != nil {
		return nil, err
	}

	rootWin := backend.GetRootWindow()
	rootGeo := *backend.GetRect(rootWin)

	return &XWindowManager{
		backend:       backend,
		clientManager: NewClientManager(),
		rootWindow:    rootWin,
		rootGeom:      rootGeo,
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
	go xm.MovingWindowsLoop(ctx)
	go xm.HandleWMEvent(ctx)
}

func (xm *XWindowManager) ScanExistingWindows() {
	for _, window := range xm.backend.QueryTree(xm.rootWindow) {
		attrs := xm.backend.GetWindowAttributes(window)

		if attrs.map_state == C.IsViewable && attrs.override_redirect == 0 {
			client := generateClientContext(xm.backend, window)
			client.CreateFrame(xm.backend, xm.rootWindow)
			fmt.Println("XWM: adding exisiting window \"", client.Window, "\"; create frame window \"", client.Frame, "\"; Window Title:", client.Title)
			xm.clientManager.Add(client)
			xm.backend.MapWindow(client.Frame)
			xm.backend.MapWindow(client.Window)
		}
	}
}
