package xwindowmgr

// Backend is the manager's window-system dependency. Implementations must allow
// concurrent event polling, window operations and animation. NextEvent is called
// by a single reader, and only when Pending reports an available event.
// Resource ownership remains with the caller that supplies the backend.
type Backend interface {
	EventSource
	GetRootWindow() WindowID
	GetRect(WindowID) *WindowRect
	SelectInput(EventMask)
	CreateFrame(WindowID, int, int, int, int) WindowID
	Reparent(WindowID, WindowID, int, int)
	DestroyWindow(WindowID)
	Flush()
	Sync(bool)
	GetWindowName(WindowID) string
	QueryTree(WindowID) []WindowID
	GetWindowAttributes(WindowID) WindowAttributes
	MapWindow(WindowID)
	UnmapWindow(WindowID)
	UnmapSubwindows(WindowID)
	RaiseWindow(WindowID)
	MoveWindow(WindowID, int, int)
	MoveResizeWindow(WindowID, int, int, uint, uint)
	ConfigureWindow(WindowID, ConfigureMask, *WindowChanges)
	KeysymToKeycode(KeySym) KeyCode
	StringToKeysym(string) KeySym
	GrabKey(WindowID, KeyCode, ModifierMask)
	GrabPointer(WindowID)
	UngrabPointer()
	// ReplayPointer releases a passive synchronous grab and reprocesses its click.
	ReplayPointer(uint32)
	TranslateCoordinates(WindowID, WindowID, int, int) (int, int)
	CreateGC(WindowID) GraphicsContext
	FreeGC(GraphicsContext)
	SetForeground(GraphicsContext, uint64)
	FillRectangle(WindowID, GraphicsContext, int, int, uint, uint)
	DrawString(WindowID, GraphicsContext, int, int, string, int)
	SetInputFocus(WindowID, uint32)
}
