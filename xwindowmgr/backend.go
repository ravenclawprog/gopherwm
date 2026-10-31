package xwindowmgr

// Backend is the manager's window-system dependency. Implementations must allow
// concurrent event polling, window operations and animation. NextEvent is called
// by a single reader, and only when Pending reports an available event.
// Resource ownership remains with the caller that supplies the backend.
type Backend interface {
	EventSource
	GetRootWindow() WindowID
	GetRect(WindowID) *WindowRect
	SelectInput(int64)
	CreateFrame(WindowID, Coordinate, Coordinate, Coordinate, Coordinate) WindowID
	Reparent(WindowID, WindowID, Coordinate, Coordinate)
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
	MoveWindow(WindowID, Coordinate, Coordinate)
	MoveResizeWindow(WindowID, Coordinate, Coordinate, Dimension, Dimension)
	ConfigureWindow(WindowID, Dimension, *WindowChanges)
	KeysymToKeycode(KeySym) KeyCode
	StringToKeysym(string) KeySym
	GrabKey(WindowID, Coordinate, Dimension)
	GrabPointer(WindowID)
	UngrabPointer()
	TranslateCoordinates(WindowID, WindowID, Coordinate, Coordinate) (Coordinate, Coordinate)
	CreateGC(WindowID) GraphicsContext
	FreeGC(GraphicsContext)
	SetForeground(GraphicsContext, uint64)
	FillRectangle(WindowID, GraphicsContext, Coordinate, Coordinate, Dimension, Dimension)
	DrawString(WindowID, GraphicsContext, Coordinate, Coordinate, string, int)
}
