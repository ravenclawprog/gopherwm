package xwindowmgr

// WindowID identifies an X11 window; it is an ID, not a native pointer.
type WindowID uint64

type KeySym uint64
type KeyCode uint32
type EventMask int64
type ModifierMask uint32
type ConfigureMask uint32

// GraphicsContext is an opaque backend-owned resource ID. Zero is invalid.
// A context must be freed by the backend that created it.
type GraphicsContext uint64

type MapState int

// WindowAttributes contains the attributes used by the manager.
type WindowAttributes struct {
	X, Y             int
	Width, Height    int
	MapState         MapState
	OverrideRedirect bool
}

type WindowChanges struct {
	X, Y          int
	Width, Height int
	BorderWidth   int
	Sibling       WindowID
	StackMode     int
}
