package main

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
*/
import "C"
import (
	"context"
	"fmt"
	"gopherwm/xwindowmgr"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	xManager, err := xwindowmgr.NewXWindowManager()
	if err != nil {
		fmt.Println("Error at creation XWindowManager:", err)
		os.Exit(1)
	}

	xManager.GrabKey(uint64(C.XK_F1))
	xManager.GrabButton(uint32(C.Button1), uint32(C.ButtonPressMask|C.ButtonReleaseMask|C.PointerMotionMask))
	xManager.GrabButton(uint32(C.Button3), uint32(C.ButtonPressMask|C.ButtonReleaseMask|C.PointerMotionMask))
	xManager.SelectInput(int64(C.KeyPressMask | C.SubstructureNotifyMask))

	xManager.Start(ctx)

	// wait till Done
	<-ctx.Done()
}
