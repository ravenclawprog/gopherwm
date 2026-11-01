package xwindowmgr

// EventType uses X11 protocol event numbers; unknown values are preserved.
type EventType int

// WMEvent contains Go-owned data and never references Xlib's event buffer.
// Window is the event's target; Data contains the payload matching Type.
type WMEvent struct {
	Type   EventType
	Window WindowID
	Data   any
}

type MapRequestEvent struct{ Parent WindowID }
type UnmapEvent struct{ FromConfigure bool }
type DestroyWindowEvent struct{}
type ExposeEvent struct{ X, Y, Width, Height, Count int }

type KeyEvent struct {
	Subwindow WindowID
	State     ModifierMask
	Keycode   KeyCode
}

type ConfigureRequestEvent struct {
	X, Y          int
	Width, Height int
	BorderWidth   int
	Above         WindowID
	Detail        int
	ValueMask     ConfigureMask
}

type ButtonEvent struct {
	Time         uint32 // X server timestamp in milliseconds.
	X, Y         int
	XRoot, YRoot int
	State        ModifierMask
	Button       uint
}

type MotionEvent struct {
	Time         uint32 // X server timestamp in milliseconds.
	X, Y         int
	XRoot, YRoot int
	State        ModifierMask
}
