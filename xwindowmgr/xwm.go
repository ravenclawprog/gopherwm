package xwindowmgr

/*
#include <X11/Xlib.h>
#include <X11/keysym.h>
*/
import "C"
import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type XWindowManager struct {
	backend       Backend
	clientManager *ClientManager
	rootWindow    C.Window
	rootGeom      WindowRect

	eventChan chan WMEvent
	loopDone  chan struct{}

	startMoving atomic.Bool
	lifecycleMu sync.Mutex
	started     bool
}

func NewXWindowManager() (*XWindowManager, error) {
	backend, err := NewXBackend()
	if err != nil {
		return nil, err
	}

	return NewXWindowManagerWithBackend(backend)
}

// NewXWindowManagerWithBackend initializes a manager without opening a display.
// The supplied backend must be non-nil and remains owned by the caller.
func NewXWindowManagerWithBackend(backend Backend) (*XWindowManager, error) {
	if backend == nil {
		return nil, fmt.Errorf("backend is required")
	}

	rootWin := backend.GetRootWindow()
	rootGeo := backend.GetRect(rootWin)
	if rootGeo == nil {
		return nil, fmt.Errorf("cannot get root window geometry")
	}

	return &XWindowManager{
		backend:       backend,
		clientManager: NewClientManager(),
		rootWindow:    rootWin,
		rootGeom:      *rootGeo,
		eventChan:     make(chan WMEvent, 200),
		loopDone:      make(chan struct{}),
	}, nil
}

// Start initializes window management and starts workers once. A stopped manager
// cannot be restarted; construct a new manager for another session.
func (xm *XWindowManager) Start(ctx context.Context) {
	xm.lifecycleMu.Lock()
	defer xm.lifecycleMu.Unlock()
	if xm.started {
		return
	}
	xm.started = true

	xm.GrabShortcuts()
	xm.backend.Sync(false)
	xm.backend.SelectInput(int64(C.KeyPressMask | C.SubstructureRedirectMask))
	xm.ScanExistingWindows()

	var workers sync.WaitGroup
	workers.Add(3)
	go func() { defer workers.Done(); xm.EventLoop(ctx, xm.eventChan) }()
	go func() { defer workers.Done(); xm.MovingWindowsLoop(ctx) }()
	go func() { defer workers.Done(); xm.HandleWMEvent(ctx) }()
	go func() { workers.Wait(); close(xm.loopDone) }()
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

// Wait blocks until all workers finish. Call it after Start.
func (xm *XWindowManager) Wait() { <-xm.loopDone }
