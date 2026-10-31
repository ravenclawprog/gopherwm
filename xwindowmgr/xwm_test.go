package xwindowmgr_test

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	wm "gopherwm/xwindowmgr"
)

type command struct {
	name    string
	window  wm.WindowID
	frame   wm.WindowID
	rect    wm.WindowRect
	mask    wm.ConfigureMask
	changes wm.WindowChanges
	context wm.GraphicsContext
	color   uint64
	text    string
	length  int
}

// fakeBackend models windows and records commands at the public backend boundary.
// An unexpected operation fails loudly rather than silently succeeding.
type fakeBackend struct {
	read       chan struct{}
	mu         sync.Mutex
	events     chan wm.WMEvent
	commands   chan command
	rects      map[wm.WindowID]wm.WindowRect
	nextFrame  wm.WindowID
	children   []wm.WindowID
	attributes map[wm.WindowID]wm.WindowAttributes
}

var _ wm.Backend = (*fakeBackend)(nil)

func newFakeBackend() *fakeBackend {
	return &fakeBackend{
		events: make(chan wm.WMEvent, 512), commands: make(chan command, 512), read: make(chan struct{}, 512),
		rects: map[wm.WindowID]wm.WindowRect{
			1:  {Width: 1024, Height: 768},
			42: {X: 10, Y: 20, Width: 300, Height: 200},
		}, nextFrame: 100,
	}
}
func (b *fakeBackend) GetRootWindow() wm.WindowID { return 1 }
func (b *fakeBackend) GetRect(window wm.WindowID) *wm.WindowRect {
	b.mu.Lock()
	defer b.mu.Unlock()
	rect, ok := b.rects[window]
	if !ok {
		return nil
	}
	return &rect
}
func (b *fakeBackend) Pending() int { return len(b.events) }
func (b *fakeBackend) NextEvent() wm.WMEvent {
	event := <-b.events
	b.read <- struct{}{}
	return event
}
func (b *fakeBackend) SelectInput(wm.EventMask) {}
func (b *fakeBackend) Sync(bool)                {}
func (b *fakeBackend) Flush()                   {}
func (b *fakeBackend) QueryTree(wm.WindowID) []wm.WindowID {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]wm.WindowID(nil), b.children...)
}
func (b *fakeBackend) GetWindowName(wm.WindowID) string                 { return "Test window" }
func (b *fakeBackend) KeysymToKeycode(wm.KeySym) wm.KeyCode             { return 1 }
func (b *fakeBackend) StringToKeysym(string) wm.KeySym                  { return 1 }
func (b *fakeBackend) GrabKey(wm.WindowID, wm.KeyCode, wm.ModifierMask) {}
func (b *fakeBackend) CreateFrame(parent wm.WindowID, x, y, w, h int) wm.WindowID {
	b.mu.Lock()
	frame := b.nextFrame
	b.nextFrame++
	rect := wm.WindowRect{X: int(x), Y: int(y), Width: int(w), Height: int(h)}
	b.rects[frame] = rect
	b.mu.Unlock()
	b.commands <- command{name: "create", window: parent, frame: frame, rect: rect}
	return frame
}
func (b *fakeBackend) Reparent(window, frame wm.WindowID, x, y int) {
	b.commands <- command{name: "reparent", window: window, frame: frame, rect: wm.WindowRect{X: int(x), Y: int(y)}}
}
func (b *fakeBackend) MapWindow(window wm.WindowID) {
	b.commands <- command{name: "map", window: window}
}
func (b *fakeBackend) UnmapWindow(window wm.WindowID) {
	b.commands <- command{name: "unmap", window: window}
}
func (b *fakeBackend) UnmapSubwindows(window wm.WindowID) {
	b.commands <- command{name: "unmapChildren", window: window}
}
func (b *fakeBackend) DestroyWindow(window wm.WindowID) {
	b.mu.Lock()
	delete(b.rects, window)
	b.mu.Unlock()
	b.commands <- command{name: "destroy", window: window}
}
func (b *fakeBackend) GetWindowAttributes(window wm.WindowID) wm.WindowAttributes {
	b.mu.Lock()
	defer b.mu.Unlock()
	if attrs, ok := b.attributes[window]; ok {
		return attrs
	}
	rect := b.rects[window]
	return wm.WindowAttributes{X: rect.X, Y: rect.Y, Width: rect.Width, Height: rect.Height, MapState: wm.IsViewable}
}
func (b *fakeBackend) RaiseWindow(wm.WindowID) { panic("unexpected RaiseWindow") }
func (b *fakeBackend) MoveWindow(window wm.WindowID, x, y int) {
	b.mu.Lock()
	rect := b.rects[window]
	rect.X, rect.Y = x, y
	b.rects[window] = rect
	b.mu.Unlock()
	b.commands <- command{name: "move", window: window, rect: wm.WindowRect{X: x, Y: y}}
}
func (b *fakeBackend) MoveResizeWindow(window wm.WindowID, x, y int, width, height uint) {
	b.commands <- command{name: "moveResize", window: window, rect: wm.WindowRect{X: x, Y: y, Width: int(width), Height: int(height)}}
}
func (b *fakeBackend) ConfigureWindow(window wm.WindowID, mask wm.ConfigureMask, changes *wm.WindowChanges) {
	b.commands <- command{name: "configure", window: window, mask: mask, changes: *changes}
}
func (b *fakeBackend) GrabPointer(window wm.WindowID) {
	b.commands <- command{name: "grabPointer", window: window}
}
func (b *fakeBackend) UngrabPointer() { b.commands <- command{name: "ungrabPointer"} }
func (b *fakeBackend) TranslateCoordinates(source, destination wm.WindowID, x, y int) (int, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	rect := b.rects[source]
	return rect.X + x, rect.Y + y
}
func (b *fakeBackend) CreateGC(window wm.WindowID) wm.GraphicsContext {
	b.commands <- command{name: "createGC", window: window, context: 1}
	return 1
}
func (b *fakeBackend) FreeGC(gc wm.GraphicsContext) {
	b.commands <- command{name: "freeGC", context: gc}
}
func (b *fakeBackend) SetForeground(gc wm.GraphicsContext, color uint64) {
	b.commands <- command{name: "foreground", context: gc, color: color}
}
func (b *fakeBackend) FillRectangle(window wm.WindowID, gc wm.GraphicsContext, x, y int, width, height uint) {
	b.commands <- command{name: "fill", window: window, context: gc, rect: wm.WindowRect{X: x, Y: y, Width: int(width), Height: int(height)}}
}
func (b *fakeBackend) DrawString(window wm.WindowID, gc wm.GraphicsContext, x, y int, text string, length int) {
	b.commands <- command{name: "text", window: window, context: gc, rect: wm.WindowRect{X: x, Y: y}, text: text, length: length}
}

func newManager(t *testing.T, backend wm.Backend) *wm.XWindowManager {
	t.Helper()
	manager, err := wm.NewXWindowManagerWithBackend(backend)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func awaitDone(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("manager did not finish")
	}
}

func waitManager(t *testing.T, manager *wm.XWindowManager) {
	t.Helper()
	done := make(chan struct{})
	go func() { manager.Wait(); close(done) }()
	awaitDone(t, done)
}

func startManager(t *testing.T, backend wm.Backend) (*wm.XWindowManager, context.CancelFunc) {
	t.Helper()
	manager := newManager(t, backend)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { cancel(); waitManager(t, manager) })
	manager.Start(ctx)
	return manager, cancel
}

func expectCommands(t *testing.T, backend *fakeBackend, want ...command) {
	t.Helper()
	for _, expected := range want {
		select {
		case actual := <-backend.commands:
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("command: got %+v, want %+v", actual, expected)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("missing command: %+v", expected)
		}
	}
}

func openWindow(t *testing.T, backend *fakeBackend, frame wm.WindowID) {
	t.Helper()
	backend.events <- wm.WMEvent{Type: wm.MapRequest, Window: 42, Data: wm.MapRequestEvent{}}
	expectCommands(t, backend,
		command{name: "create", window: 1, frame: frame, rect: wm.WindowRect{X: 10, Y: 20, Width: 320, Height: 200}},
		command{name: "reparent", window: 42, frame: frame, rect: wm.WindowRect{X: 20}},
		command{name: "map", window: frame}, command{name: "map", window: 42},
	)
}

func TestXWindowManagerWindowLifecycle(t *testing.T) {
	backend := newFakeBackend()
	manager, cancel := startManager(t, backend)
	openWindow(t, backend, 100)

	// Mapping an existing client reuses its frame.
	backend.events <- wm.WMEvent{Type: wm.MapRequest, Window: 42, Data: wm.MapRequestEvent{}}
	expectCommands(t, backend, command{name: "map", window: 100}, command{name: "map", window: 42})

	backend.events <- wm.WMEvent{Type: wm.UnmapNotify, Window: 42, Data: wm.UnmapEvent{}}
	expectCommands(t, backend, command{name: "unmapChildren", window: 100}, command{name: "unmap", window: 100})

	backend.events <- wm.WMEvent{Type: wm.DestroyNotify, Window: 42, Data: wm.DestroyWindowEvent{}}
	expectCommands(t, backend, command{name: "unmapChildren", window: 100}, command{name: "unmap", window: 100}, command{name: "destroy", window: 100})

	// A new frame proves the old client was removed, without reading its registry.
	openWindow(t, backend, 101)
	cancel()
	waitManager(t, manager)
	if len(backend.commands) != 0 {
		t.Fatal("unexpected extra window commands")
	}
}

func TestXWindowManagerContinuesAfterInvalidEvents(t *testing.T) {
	backend := newFakeBackend()
	startManager(t, backend)
	backend.events <- wm.WMEvent{Type: -1}
	backend.events <- wm.WMEvent{Type: wm.MapRequest, Window: 42, Data: "invalid payload"}
	openWindow(t, backend, 100)
}

func TestXWindowManagerStopsWithEmptySource(t *testing.T) {
	backend := newFakeBackend()
	manager, cancel := startManager(t, backend)
	cancel()
	waitManager(t, manager)
	// Waiting again after completion returns immediately.
	waitManager(t, manager)
}

func TestXWindowManagerWaitsForContextCancellation(t *testing.T) {
	backend := newFakeBackend()
	manager, cancel := startManager(t, backend)
	cancel()
	done := make(chan struct{})
	go func() { manager.Wait(); close(done) }()
	awaitDone(t, done)
}

func TestXWindowManagerStartIsIdempotent(t *testing.T) {
	backend := newFakeBackend()
	manager, _ := startManager(t, backend)
	manager.Start(context.Background())
	openWindow(t, backend, 100)
}

func TestXWindowManagerRejectsMissingBackendOrGeometry(t *testing.T) {
	if _, err := wm.NewXWindowManagerWithBackend(nil); err == nil {
		t.Fatal("expected missing backend error")
	}
	backend := newFakeBackend()
	delete(backend.rects, 1)
	if _, err := wm.NewXWindowManagerWithBackend(backend); err == nil {
		t.Fatal("expected missing geometry error")
	}
}

func TestXWindowManagerEventLoopStopsWithFullQueue(t *testing.T) {
	backend := newFakeBackend()
	manager := newManager(t, backend)
	output := make(chan wm.WMEvent, 1)
	output <- wm.WMEvent{Type: 1}
	backend.events <- wm.WMEvent{Type: 2}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { manager.EventLoop(ctx, output); close(done) }()
	// Observe consumption at the backend boundary before cancelling.
	awaitDone(t, backend.read)
	cancel()
	awaitDone(t, done)
	if event := <-output; event.Type != 1 {
		t.Fatalf("existing queued event changed: %+v", event)
	}
}

func TestXWindowManagerConfigureRequests(t *testing.T) {
	backend := newFakeBackend()
	startManager(t, backend)
	openWindow(t, backend, 100)
	mask := wm.CWX | wm.CWY | wm.CWWidth | wm.CWHeight | wm.CWBorderWidth | wm.CWSibling | wm.CWStackMode
	data := wm.ConfigureRequestEvent{X: -15, Y: 30, Width: 500, Height: 400, BorderWidth: 2, Above: 77, Detail: 1, ValueMask: mask}
	changes := wm.WindowChanges{X: -15, Y: 30, Width: 520, Height: 400, BorderWidth: 2, Sibling: 77, StackMode: 1}
	backend.events <- wm.WMEvent{Type: wm.ConfigureRequest, Window: 42, Data: data}
	expectCommands(t, backend,
		command{name: "configure", window: 100, mask: mask, changes: changes},
		command{name: "moveResize", window: 42, rect: wm.WindowRect{X: 20, Width: 500, Height: 400}},
	)
	// Unmanaged windows receive the application's changes without frame offsets.
	changes.Width = 500
	backend.events <- wm.WMEvent{Type: wm.ConfigureRequest, Window: 43, Data: data}
	expectCommands(t, backend, command{name: "configure", window: 43, mask: mask, changes: changes})
}

func TestXWindowManagerDragUsesSignedCoordinates(t *testing.T) {
	backend := newFakeBackend()
	startManager(t, backend)
	openWindow(t, backend, 100)
	backend.events <- wm.WMEvent{Type: wm.ButtonPress, Window: 100, Data: wm.ButtonEvent{X: 5, XRoot: 40, YRoot: 60, Button: 1}}
	expectCommands(t, backend, command{name: "grabPointer", window: 100})
	backend.events <- wm.WMEvent{Type: wm.MotionNotify, Window: 100, Data: wm.MotionEvent{XRoot: 0, YRoot: 10}}
	expectCommands(t, backend, command{name: "move", window: 100, rect: wm.WindowRect{X: -30, Y: -30}})
	backend.events <- wm.WMEvent{Type: wm.ButtonRelease, Window: 100, Data: wm.ButtonEvent{Button: 1}}
	expectCommands(t, backend, command{name: "ungrabPointer"})
}

func TestXWindowManagerScansOnlyManagedVisibleWindows(t *testing.T) {
	backend := newFakeBackend()
	backend.children = []wm.WindowID{42, 43, 44}
	backend.attributes = map[wm.WindowID]wm.WindowAttributes{
		42: {MapState: wm.IsViewable},
		43: {MapState: wm.IsViewable, OverrideRedirect: true},
		44: {MapState: wm.IsUnmapped},
	}
	manager, cancel := startManager(t, backend)
	expectCommands(t, backend,
		command{name: "create", window: 1, frame: 100, rect: wm.WindowRect{X: 10, Y: 20, Width: 320, Height: 200}},
		command{name: "reparent", window: 42, frame: 100, rect: wm.WindowRect{X: 20}},
		command{name: "map", window: 100}, command{name: "map", window: 42},
	)
	cancel()
	waitManager(t, manager)
	if len(backend.commands) != 0 {
		t.Fatal("unexpected commands for hidden or override-redirect windows")
	}
}
