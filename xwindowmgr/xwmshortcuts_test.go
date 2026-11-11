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

func TestCustomShortcutToggleMotion(t *testing.T) {
	backend := &shortcutBackend{}
	shortcuts, err := compileShortcuts(backend, Config{
		Bindings: []Binding{
			{Key: "m", Modifiers: []string{"Control"}, Action: "toggle_motion"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	manager := &XWindowManager{backend: backend, rootWindow: 1, shortcuts: shortcuts}
	manager.GrabShortcuts()
	if len(backend.grabs) != 4 {
		t.Fatalf("got %d grabs, want 4", len(backend.grabs))
	}
	for _, grab := range backend.grabs {
		if grab.window != 1 || grab.key != 58 || grab.modifiers&^(LockMask|Mod2Mask) != ControlMask {
			t.Fatalf("unexpected custom shortcut grab: %+v", grab)
		}
	}

	manager.handleKeyPress(1, &KeyEvent{Keycode: 58, State: Mod4Mask})
	if manager.startMoving.Load() {
		t.Fatal("default Super+m shortcut still toggles movement")
	}
	manager.handleKeyPress(1, &KeyEvent{Keycode: 58, State: ControlMask})
	if !manager.startMoving.Load() {
		t.Fatal("custom Control+m shortcut did not enable movement")
	}
	manager.handleKeyPress(1, &KeyEvent{Keycode: 58, State: ControlMask})
	if manager.startMoving.Load() {
		t.Fatal("custom Control+m shortcut did not disable movement")
	}
}

func TestShortcutsWithLockModifiersAfterFocus(t *testing.T) {
	backend := &shortcutBackend{}
	manager := &XWindowManager{
		backend:    backend,
		rootWindow: 1,
		shortcuts: map[shortcutKey]shortcutAction{
			{keycode: 36, modifiers: Mod4Mask}: {
				action:  "spawn",
				command: []string{"xterm"},
			},
			{keycode: 58, modifiers: Mod4Mask}: {
				action: "toggle_motion",
			},
		},
	}
	manager.GrabShortcuts()
	manager.FocusClient(&Client{Window: 42, Frame: 100}, 123)

	if len(backend.grabs) != 8 {
		t.Fatalf("got %d grabs, want 8", len(backend.grabs))
	}
	gotGrabs := make(map[shortcutGrab]bool)
	for _, grab := range backend.grabs {
		gotGrabs[grab] = true
	}
	for _, key := range []KeyCode{36, 58} {
		for _, locks := range []ModifierMask{0, LockMask, Mod2Mask, LockMask | Mod2Mask} {
			want := shortcutGrab{1, key, Mod4Mask | locks}
			if !gotGrabs[want] {
				t.Fatalf("missing grab: %+v", want)
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
