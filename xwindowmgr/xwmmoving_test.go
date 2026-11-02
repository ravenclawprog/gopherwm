package xwindowmgr

import (
	"context"
	"testing"
	"time"
)

// Only the backend operations exercised by movement are implemented.
type movementBackend struct {
	Backend
	x, y  int
	moves chan int
}

func (b *movementBackend) MoveWindow(_ WindowID, x, y int) {
	b.x, b.y = x, y
	if b.moves != nil {
		b.moves <- x
	}
}
func (b *movementBackend) Flush()                         {}
func (b *movementBackend) RaiseWindow(WindowID)           {}
func (b *movementBackend) SetInputFocus(WindowID, uint32) {}
func (b *movementBackend) GetWindowAttributes(WindowID) WindowAttributes {
	return WindowAttributes{X: b.x, Y: b.y}
}
func (b *movementBackend) GrabPointer(WindowID) {}
func (b *movementBackend) UngrabPointer()       {}
func (b *movementBackend) TranslateCoordinates(WindowID, WindowID, int, int) (int, int) {
	return b.x, b.y
}

func TestAdvanceAxis(t *testing.T) {
	for _, tc := range []struct {
		name                                                           string
		position, velocity, extent, screen, wantPosition, wantVelocity int
	}{
		{"right overshoot", 78, 5, 20, 100, 80, -5},
		{"left overshoot", 2, -5, 20, 100, 0, 5},
		{"exact edge", 75, 5, 20, 100, 80, -5},
		{"outside moving inward", -20, 5, 20, 100, 0, 5},
		{"oversized", 10, 5, 120, 100, 0, 0},
		{"fits exactly", 0, 5, 100, 100, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, v := advanceAxis(tc.position, tc.velocity, tc.extent, tc.screen)
			if p != tc.wantPosition || v != tc.wantVelocity {
				t.Fatalf("got (%d, %d), want (%d, %d)", p, v, tc.wantPosition, tc.wantVelocity)
			}
		})
	}
}

func TestDragInertiaAndFrameBounce(t *testing.T) {
	backend := &movementBackend{x: 10, y: 20}
	manager := &XWindowManager{backend: backend, clientManager: NewClientManager(), rootGeom: WindowRect{Width: 100, Height: 100}}
	client := &Client{Window: 1, Frame: 2, Geom: WindowRect{X: 10, Y: 20, Width: 30, Height: 20}}
	manager.clientManager.Add(client)
	manager.handleButtonPress(2, &ButtonEvent{Time: 100, XRoot: 10, YRoot: 20})
	manager.handleMotionEvent(2, &MotionEvent{Time: 120, XRoot: 30, YRoot: 20})
	manager.moveWindows() // Dragged windows must not be animated.
	if backend.x != 30 {
		t.Fatalf("animation moved dragged window to %d", backend.x)
	}
	// Equal speed sampled over a shorter interval gives the same inertia.
	manager.handleMotionEvent(2, &MotionEvent{Time: 130, XRoot: 40, YRoot: 20})
	if client.Motion.VelocityX != 5 {
		t.Fatalf("velocity = %d, want 5", client.Motion.VelocityX)
	}
	manager.handleButtonRelease(2, &ButtonEvent{Time: 135})
	manager.moveWindows()
	if backend.x != 45 {
		t.Fatalf("inertia: x=%d, want 45", backend.x)
	}
	manager.moveWindows()
	// Client width 30 plus titlebar 20 leaves x <= 50.
	if backend.x != 50 || client.Motion.VelocityX != -5 {
		t.Fatalf("bounce: x=%d velocity=%d", backend.x, client.Motion.VelocityX)
	}
	manager.moveWindows()
	if backend.x != 45 {
		t.Fatalf("inertia: x=%d, want 45", backend.x)
	}
}

func TestDragPauseCancelsThrow(t *testing.T) {
	backend := &movementBackend{}
	manager := &XWindowManager{backend: backend, clientManager: NewClientManager()}
	client := &Client{Window: 1, Frame: 2}
	manager.clientManager.Add(client)
	manager.handleButtonPress(2, &ButtonEvent{Time: 100})
	manager.handleMotionEvent(2, &MotionEvent{Time: 120, XRoot: 10})
	manager.handleButtonRelease(2, &ButtonEvent{Time: 300})
	if client.Motion.VelocityX != 0 || client.Motion.VelocityY != 0 {
		t.Fatal("throw persisted after a pause")
	}
}

func TestEventHandlerAnimatesWithoutSeparateWorker(t *testing.T) {
	backend := &movementBackend{moves: make(chan int, 64)}
	manager := &XWindowManager{
		backend: backend, clientManager: NewClientManager(),
		rootGeom: WindowRect{Width: 100, Height: 100}, eventChan: make(chan WMEvent, 4),
	}
	manager.clientManager.Add(&Client{Window: 1, Frame: 2, Geom: WindowRect{Width: 20, Height: 20}, Motion: MotionData{VelocityX: 5}})
	manager.startMoving.Store(true)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); manager.HandleWMEvent(ctx) }()
	defer func() { cancel(); <-done }()
	select {
	case x := <-backend.moves:
		if x != 5 {
			t.Fatalf("first animation position = %d, want 5", x)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("event handler did not animate")
	}
	// Exercise mouse handling while animation ticks are enabled under -race.
	manager.eventChan <- WMEvent{Type: ButtonPress, Window: 2, Data: ButtonEvent{Time: 100}}
	manager.eventChan <- WMEvent{Type: MotionNotify, Window: 2, Data: MotionEvent{Time: 120, XRoot: 10}}
	manager.eventChan <- WMEvent{Type: ButtonRelease, Window: 2, Data: ButtonEvent{Time: 130}}
	select {
	case <-backend.moves:
	case <-time.After(2 * time.Second):
		t.Fatal("event handler stopped processing movement")
	}
}
