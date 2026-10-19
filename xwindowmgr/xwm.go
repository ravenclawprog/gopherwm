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
			xm.clientManager.Add(client)
			C.XMapWindow(xm.backend.display, client.Frame)
			C.XMapSubwindows(xm.backend.display, client.Window)
		}
	}
}

// func NewXWindowManager() (*XWindowManager, error) {
// 	tmpDisplay := C.XOpenDisplay(nil)
// 	if tmpDisplay == nil {
// 		return nil, errors.New("Cannot open display")
// 	}
// 	tmpScreen := C.XDefaultScreen(tmpDisplay)

// 	ret := &XWindowManager{
// 		xWindowContextManager: *manager.NewObjectManager[xWindowContext](),
// 		display:               tmpDisplay,
// 		screen:                tmpScreen,
// 		frames:                make(map[C.Window]C.Window),
// 		loopDone:              make(chan struct{}),
// 	}
// 	C.SetupErrorHandler()

// 	rootWindow := C._DefaultRootWindow(tmpDisplay)
// 	fmt.Printf("RootWindow::%d\n", rootWindow)
// 	rootWindowContext := createXWindowContext(tmpDisplay, rootWindow)
// 	ret.rootID = ret.xWindowContextManager.AddObject(rootWindowContext)

// 	return ret, nil
// }

// func createXWindowContext(display *C.Display, window C.Window) xWindowContext {
// 	var ret xWindowContext
// 	ret.windowHandle = window

// 	var outWindow C.Window
// 	var outBorderWidth C.uint
// 	var outDepth C.uint

// 	C.XGetGeometry(display, ret.windowHandle, &outWindow, &(ret.rect.x), &(ret.rect.y), &(ret.rect.width), &(ret.rect.height), &outBorderWidth, &outDepth)
// 	return ret
// }

// func (xm *XWindowManager) createFrameForWindow(windowContext xWindowContext) xWindowContext {
// 	rootContext, _ := xm.xWindowContextManager.GetObject(xm.rootID)

// 	frameWidth := windowContext.rect.width + TitlebarWidth
// 	frameWin := C.XCreateSimpleWindow(
// 		xm.display, rootContext.windowHandle,
// 		windowContext.rect.x, windowContext.rect.y, frameWidth, windowContext.rect.height,
// 		1, C.XBlackPixel(xm.display, xm.screen), C.XWhitePixel(xm.display, xm.screen),
// 	)

// 	windowContext.frameWindowHandle = frameWin
// 	// adjascent for exposure
// 	C.XSelectInput(xm.display, frameWin, C.ExposureMask|C.ButtonPressMask|C.SubstructureNotifyMask)

// 	// reparent the window
// 	C.XReparentWindow(xm.display, windowContext.windowHandle, frameWin, TitlebarWidth, 0)

// 	return windowContext
// }
// func (xm *XWindowManager) ScanExistingWindows() {
// 	var rootReturn, parentReturn C.Window
// 	var children *C.Window
// 	var nchildren C.uint

// 	rootContext, _ := xm.xWindowContextManager.GetObject(xm.rootID)
// 	status := C.XQueryTree(xm.display, rootContext.windowHandle, &rootReturn, &parentReturn, &children, &nchildren)

// 	if status == 0 || nchildren == 0 {
// 		return
// 	}
// 	defer C.XFree(unsafe.Pointer(children))

// 	windowSlice := unsafe.Slice(children, int(nchildren))

// 	for _, window := range windowSlice {
// 		var attrs C.XWindowAttributes
// 		C.XGetWindowAttributes(xm.display, window, &attrs)

// 		if attrs.map_state == C.IsViewable && attrs.override_redirect == 0 {
// 			newWindowContext := createXWindowContext(xm.display, window)
// 			newWindowContext = xm.createFrameForWindow(newWindowContext)
// 			xm.xWindowContextManager.AddObject(newWindowContext)
// 			xm.frames[newWindowContext.frameWindowHandle] = newWindowContext.windowHandle
// 			C.XMapWindow(xm.display, newWindowContext.frameWindowHandle)
// 			C.XMapSubwindows(xm.display, newWindowContext.windowHandle)
// 		}
// 	}
// }
// func (xm *XWindowManager) GrabShortcuts() {
// 	rootContext, _ := xm.xWindowContextManager.GetObject(xm.rootID)
// 	rootWin := rootContext.windowHandle

// 	keyCode := C.XKeysymToKeycode(xm.display, C.KeySym(C.XK_Return))

// 	// mask for win key
// 	modifier := C.uint(C.Mod4Mask)

// 	// win + enter grub
// 	C.XGrabKey(
// 		xm.display,
// 		C.int(keyCode),
// 		modifier,
// 		rootWin,
// 		C.True, // owner_events
// 		C.GrabModeAsync,
// 		C.GrabModeAsync,
// 	)

// 	C.XFlush(xm.display)
// 	fmt.Println("Hotkeys are registrated")
// }

// func (xm *XWindowManager) GrabKey(key uint64) int {
// 	rootCtx, err := xm.xWindowContextManager.GetObject(xm.rootID)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return 0
// 	}
// 	generatedKey := C.XKeysymToKeycode(xm.display, C.ulong(key))
// 	ret := C.XGrabKey(
// 		xm.display,
// 		C.int(generatedKey),
// 		C.Mod1Mask,
// 		rootCtx.windowHandle,
// 		1,
// 		C.GrabModeAsync,
// 		C.GrabModeAsync,
// 	)

// 	return int(ret)
// }

// func (xm *XWindowManager) GrabButton(button uint32, mask uint32) int {
// 	rootCtx, err := xm.xWindowContextManager.GetObject(xm.rootID)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return 0
// 	}
// 	ret := C.XGrabButton(
// 		xm.display,
// 		C.uint(button),
// 		C.Mod1Mask,
// 		rootCtx.windowHandle,
// 		1,
// 		C.uint(mask),
// 		C.GrabModeAsync,
// 		C.GrabModeAsync,
// 		C.None,
// 		C.None,
// 	)
// 	return int(ret)
// }

// func (xm *XWindowManager) SelectInput(inputMask int64) {
// 	rootCtx, err := xm.xWindowContextManager.GetObject(xm.rootID)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	C.XSelectInput(xm.display, rootCtx.windowHandle, C.long(inputMask))
// }

// func (xm *XWindowManager) Sync(discard bool) {
// 	if discard {
// 		C.XSync(xm.display, C.True)
// 	} else {
// 		C.XSync(xm.display, C.False)
// 	}
// }

// func (xm *XWindowManager) Spawn(command string, args ...string) {
// 	go func() {
// 		cmd := exec.Command(command, args...)

// 		cmd.Env = os.Environ()

// 		cmd.Stderr = os.Stderr
// 		cmd.Stdout = os.Stdout

// 		if err := cmd.Start(); err != nil {
// 			fmt.Println("Error at run command:", command, "; error:", err)
// 		}
// 	}()
// }

// func (xm *XWindowManager) Start(ctx context.Context) {

// 	go func() {
// 		defer close(xm.loopDone)
// 		xm.ScanExistingWindows()
// 		xm.GrabShortcuts()
// 		fmt.Println("XWindowManager: event loop run")

// 		var ev C.XEvent
// 		for {
// 			for C.XPending(xm.display) != 0 {
// 				C.XNextEvent(xm.display, &ev)
// 				evType := unionToInt(ev)
// 				fmt.Println("XWindowManager:Event:", evType)
// 				if evType == C.KeyPress {
// 					xkeyEvent := unionToXKeyEvent(ev)
// 					fmt.Println("XWindowManager:Event:Key pressed:", xkeyEvent.keycode)
// 					if xkeyEvent.subwindow != C.None {
// 						C.XRaiseWindow(xm.display, xkeyEvent.subwindow)
// 					}
// 					cleanModifiers := xkeyEvent.state &^ C.uint(C.Mod2Mask|C.LockMask)
// 					targetCode := C.XKeysymToKeycode(xm.display, C.KeySym(C.XK_Return))
// 					fmt.Println("Target code is ", targetCode)
// 					if xkeyEvent.keycode == C.uint(targetCode) && cleanModifiers == C.uint(C.Mod4Mask) {
// 						fmt.Println("Win + Enter! Running xterm...")
// 						xm.Spawn("xterm")
// 					}

// 				} else if evType == C.ConfigureRequest {
// 					fmt.Println("XWindowManager:Event:Configure request")
// 					cr := unionToXConfigureRequestEvent(ev)
// 					var changes C.XWindowChanges
// 					changes.x = cr.x
// 					changes.y = cr.y
// 					changes.width = cr.width
// 					changes.height = cr.height
// 					changes.border_width = cr.border_width

// 					var associatedFrame C.Window = C.None
// 					for frameWin, clientWin := range xm.frames {
// 						if clientWin == cr.window {
// 							associatedFrame = frameWin
// 							break
// 						}
// 					}

// 					if associatedFrame != C.None {
// 						changes.width = cr.width + TitlebarWidth
// 						C.XConfigureWindow(xm.display, associatedFrame, C.uint(cr.value_mask), &changes)

// 						C.XResizeWindow(xm.display, cr.window, C.uint(cr.width), C.uint(cr.height))
// 					} else {
// 						C.XConfigureWindow(xm.display, cr.window, C.uint(cr.value_mask), &changes)
// 					}
// 					C.XFlush(xm.display)
// 				} else if evType == C.MapRequest {
// 					fmt.Println("XWindowManager:Event:Map request")
// 					mq := unionToXMapRequestEvent(ev)
// 					var existingFrame C.Window = C.None
// 					for frameWin, clientWin := range xm.frames {
// 						if clientWin == mq.window {
// 							existingFrame = frameWin
// 							break
// 						}
// 					}

// 					if existingFrame != C.None {
// 						fmt.Println("XWindowManager:Frame is already exist", existingFrame)
// 						C.XMapWindow(xm.display, existingFrame)
// 						C.XMapSubwindows(xm.display, existingFrame)
// 					} else {
// 						// Новое окно — создаем фрейм с нуля (ваш старый код)
// 						newWindowContext := createXWindowContext(xm.display, mq.window)
// 						newWindowContext = xm.createFrameForWindow(newWindowContext)
// 						xm.xWindowContextManager.AddObject(newWindowContext)
// 						xm.frames[newWindowContext.frameWindowHandle] = newWindowContext.windowHandle

// 						C.XMapWindow(xm.display, newWindowContext.frameWindowHandle)
// 						C.XMapSubwindows(xm.display, newWindowContext.frameWindowHandle)
// 					}
// 					C.XFlush(xm.display)
// 				} else if evType == C.Expose {
// 					fmt.Println("XWindowManager:Event:expose")
// 					ex := unionToXExposeEvent(ev)
// 					frameWin := ex.window

// 					clientWin, exists := xm.frames[frameWin]
// 					if !exists {
// 						continue
// 					}
// 					var rootReturn C.Window
// 					var x, y C.int
// 					var width, height, border C.uint
// 					var depth C.uint
// 					C.XGetGeometry(xm.display, C.Drawable(frameWin), &rootReturn, &x, &y, &width, &height, &border, &depth)

// 					// create grafic context for drawing
// 					gc := C.XCreateGC(xm.display, C.Drawable(frameWin), 0, nil)

// 					// drawing rectable
// 					// using color 0x3b82f6
// 					C.XSetForeground(xm.display, gc, 0x3B82F6)
// 					C.XFillRectangle(xm.display, C.Drawable(frameWin), gc, 0, 0, TitlebarWidth, height)

// 					// get name of window and draw it
// 					cName := C.get_window_name(xm.display, clientWin)
// 					var title string = "Window"
// 					if cName != nil {
// 						title = C.GoString(cName)
// 						C.free(unsafe.Pointer(cName))
// 					}

// 					// draw a text
// 					C.XSetForeground(xm.display, gc, 0xFFFFFF)
// 					startY := C.int(20)
// 					charSpacing := C.int(14)
// 					textX := C.int(6)

// 					runes := []rune(title)

// 					for i, r := range runes {
// 						currentY := startY + C.int(i)*charSpacing

// 						if currentY > C.int(height)-10 {
// 							break
// 						}

// 						charStr := C.CString(string(r))
// 						C.XDrawString(xm.display, C.Drawable(frameWin), gc, textX, currentY, charStr, 1)
// 						C.free(unsafe.Pointer(charStr))
// 					}

// 					C.XFreeGC(xm.display, gc)

// 				} else if evType == C.ButtonPress {
// 					fmt.Println("XWindowManager:Event: ButtonPress")
// 					bp := (*C.XButtonEvent)(unsafe.Pointer(&ev))

// 					if _, exists := xm.frames[bp.window]; exists {
// 						if bp.x > TitlebarWidth {
// 							fmt.Println("Client zone click, skip")
// 							continue
// 						}
// 						xm.activeFrame = bp.window
// 						xm.startMouseX = bp.x_root
// 						xm.startMouseY = bp.y_root

// 						C.XGetWindowAttributes(xm.display, xm.activeFrame, &xm.startAttr)

// 						C.XGrabPointer(
// 							xm.display,
// 							bp.window,
// 							C.False,
// 							C.uint(C.PointerMotionMask|C.ButtonReleaseMask),
// 							C.GrabModeAsync,
// 							C.GrabModeAsync,
// 							C.None,
// 							C.None,
// 							C.CurrentTime,
// 						)
// 					}
// 				} else if evType == C.MotionNotify {
// 					fmt.Println("XWindowManager:Event: Motion Notify")
// 					if xm.activeFrame != C.None {
// 						mn := (*C.XMotionEvent)(unsafe.Pointer(&ev))

// 						deltaX := mn.x_root - xm.startMouseX
// 						deltaY := mn.y_root - xm.startMouseY

// 						newX := xm.startAttr.x + deltaX
// 						newY := xm.startAttr.y + deltaY

// 						C.XMoveWindow(xm.display, xm.activeFrame, newX, newY)
// 						C.XFlush(xm.display)
// 					}
// 				} else if evType == C.ButtonRelease {
// 					fmt.Println("XWindowManager:Event: ButtonRelease")
// 					if xm.activeFrame != C.None {
// 						C.XUngrabPointer(xm.display, C.CurrentTime)
// 						xm.activeFrame = C.None
// 						C.XFlush(xm.display)
// 					}
// 				} else if evType == C.UnmapNotify {
// 					unmapEvent := (*C.XUnmapEvent)(unsafe.Pointer(&ev))
// 					fmt.Println("XWindowManager:Event: UnmapNotify for window:", unmapEvent.window)

// 					var associatedFrame C.Window = C.None
// 					for frameWin, clientWin := range xm.frames {
// 						if clientWin == unmapEvent.window {
// 							associatedFrame = frameWin
// 							break
// 						}
// 					}
// 					if associatedFrame != C.None {
// 						C.XUnmapWindow(xm.display, associatedFrame)
// 						C.XDestroyWindow(xm.display, associatedFrame)
// 						delete(xm.frames, associatedFrame)
// 						C.XFlush(xm.display)
// 					}

// 				} else if evType == C.DestroyNotify {
// 					destroyEvent := unionToXDestroyWindowEvent(ev)
// 					fmt.Println("XWindowManager:Event: DestroyNotify for window:", destroyEvent.window)

// 					for frameWin, clientWin := range xm.frames {
// 						if clientWin == destroyEvent.window {
// 							C.XDestroyWindow(xm.display, frameWin)
// 							delete(xm.frames, frameWin)
// 							break
// 						}
// 					}
// 				}
// 			}
// 			if xm.isClosing.Load() {
// 				fmt.Println("XWindowManager: event loop is interrupted")
// 				return
// 			}
// 			time.Sleep(30 * time.Millisecond)
// 		}
// 	}()
// 	// for exit
// 	go func() {
// 		<-ctx.Done()
// 		fmt.Println("XWindowManager: context is canceled")
// 		xm.Close()
// 	}()
// }

// func (xm *XWindowManager) Close() {
// 	if xm.isClosing.Swap(true) {
// 		return
// 	}

// 	fmt.Println("Freeing XWindowManager resources")

// 	// wake up C.XNextEvent
// 	rootCtx, _ := xm.xWindowContextManager.GetObject(xm.rootID)
// 	C.send_quit_event(xm.display, rootCtx.windowHandle)

// 	// wait for goroutine with event loop is done
// 	<-xm.loopDone

// 	// closing display
// 	if xm.display != nil {
// 		C.XCloseDisplay(xm.display)
// 		xm.display = nil
// 		fmt.Println("XWindowManager resources freed")
// 	}
// }
