package xwindowmgr_test

import (
	wm "gopherwm/xwindowmgr"
	"testing"
)

func TestClientManagerAddDeleteGracefull(t *testing.T) {
	clientMngr := wm.NewClientManager()
	wantWindowID := wm.WindowID(1)
	wantClient := wm.Client{Window: wantWindowID}
	clientMngr.Add(&wantClient)
	gotClient, err := clientMngr.Remove(wantWindowID)
	if &wantClient != gotClient {
		t.Errorf("Stored and removed client isn't the same. got: %v; want: %v", wantClient, gotClient)
	}
	if err != nil {
		t.Errorf("Remove return error: %v", err)
	}
}

func TestClientManagerAddDeleteFailUnknownWindowId(t *testing.T) {
	clientMngr := wm.NewClientManager()
	wantWindowID := wm.WindowID(1)
	fakeWindowID := wm.WindowID(2)
	wantClient := wm.Client{Window: wantWindowID}
	clientMngr.Add(&wantClient)
	gotClient, err := clientMngr.Remove(fakeWindowID)
	if err == nil {
		t.Errorf("There is must be an error, but got nothing")
	}
	if gotClient != nil {
		t.Errorf("There is must be a nil client return, but got: %v", gotClient)
	}
}

func TestClientManagerFindClientGracefull(t *testing.T) {
	clientMngr := wm.NewClientManager()
	wantWindowID := wm.WindowID(1)
	wantFrame := wm.WindowID(2)
	wantClient := wm.Client{Window: wantWindowID,
		Frame: wantFrame}
	clientMngr.Add(&wantClient)
	gotClient, err := clientMngr.FindByClient(wantWindowID)
	if err != nil {
		t.Errorf("Error when try to find client. want: %v; got:%v; error:%v", wantClient, gotClient, err)
	}
	if gotClient != &wantClient {
		t.Errorf("Error when try to find client. want: %v; got:%v", wantClient, gotClient)
	}
	gotClient, err = clientMngr.FindByFrame(wantFrame)
	if err != nil {
		t.Errorf("Error when try to find client by frame. want: %v; got:%v; error:%v", wantClient, gotClient, err)
	}
	if gotClient != &wantClient {
		t.Errorf("Error when try to find client by frame. want: %v; got:%v", wantClient, gotClient)
	}
}
