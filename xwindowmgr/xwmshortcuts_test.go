package xwindowmgr

import "testing"

type shortcutGrab struct {
	window    WindowID
	key       KeyCode
	modifiers ModifierMask
}

type shortcutBackend struct {
	movementBackend
	grabs []shortcutGrab
}

func (b *shortcutBackend) KeysymToKeycode(key KeySym) KeyCode {
	if key == KeyReturn {
		return 36
	}
	return 58
}

func (b *shortcutBackend) StringToKeysym(string) KeySym { return 109 }

func (b *shortcutBackend) GrabKey(window WindowID, key KeyCode, modifiers ModifierMask) {
	b.grabs = append(b.grabs, shortcutGrab{window, key, modifiers})
}

func TestShortcutsWithLockModifiersAfterFocus(t *testing.T) {
	backend := &shortcutBackend{}
	manager := &XWindowManager{backend: backend, rootWindow: 1}
	manager.GrabShortcuts()
	manager.FocusClient(&Client{Window: 42, Frame: 100}, 123)

	if len(backend.grabs) != 8 {
		t.Fatalf("got %d grabs, want 8", len(backend.grabs))
	}
	for i, key := range []KeyCode{36, 58} {
		for j, locks := range []ModifierMask{0, LockMask, Mod2Mask, LockMask | Mod2Mask} {
			want := shortcutGrab{1, key, Mod4Mask | locks}
			if got := backend.grabs[i*4+j]; got != want {
				t.Fatalf("grab: got %+v, want %+v", got, want)
			}
			if key == 58 {
				manager.startMoving.Store(false)
				manager.handleKeyPress(1, &KeyEvent{Keycode: key, State: want.modifiers})
				if !manager.startMoving.Load() {
					t.Fatalf("movement shortcut ignored with modifiers %v", want.modifiers)
				}
			}
		}
	}
}
