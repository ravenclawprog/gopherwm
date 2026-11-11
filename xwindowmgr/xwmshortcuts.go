package xwindowmgr

import "fmt"

type shortcutKey struct {
	keycode   KeyCode
	modifiers ModifierMask
}

type shortcutAction struct {
	action  string
	command []string
}

func parseModifiers(modifiers []string) (ModifierMask, error) {
	var mask ModifierMask
	for _, modifier := range modifiers {
		switch modifier {
		case "Super":
			mask |= Mod4Mask
		case "Shift":
			mask |= ShiftMask
		case "Control":
			mask |= ControlMask
		case "Alt":
			mask |= AltMask
		default:
			return 0, fmt.Errorf("Unknown mask : %q", modifier)
		}
	}
	return mask, nil
}
func compileShortcuts(
	backend Backend,
	config Config,
) (map[shortcutKey]shortcutAction, error) {
	shortCuts := make(map[shortcutKey]shortcutAction)
	for i, binding := range config.Bindings {
		keysym := backend.StringToKeysym(binding.Key)
		if keysym == 0 {
			return nil, fmt.Errorf("binding %d: unknown key %q", i+1, binding.Key)
		}
		keycode := backend.KeysymToKeycode(keysym)
		if keycode == 0 {
			return nil, fmt.Errorf("binding %d: no keycode for %q", i+1, binding.Key)
		}
		modifiers, err := parseModifiers(binding.Modifiers)
		if err != nil {
			return nil, fmt.Errorf("binding %d: %w", i+1, err)
		}
		key := shortcutKey{
			keycode:   keycode,
			modifiers: modifiers,
		}
		if _, exists := shortCuts[key]; exists {
			return nil, fmt.Errorf("shortcut alredy mapped: %d, %q, %q", i+1, binding.Key, binding.Modifiers)
		}
		shortCuts[key] = shortcutAction{
			action:  binding.Action,
			command: append([]string(nil), binding.Command...),
		}

	}
	return shortCuts, nil
}

func (xm *XWindowManager) GrabShortcuts() {
	for key := range xm.shortcuts {
		xm.grabShortcut(key.keycode, key.modifiers)
	}
	xm.backend.Flush()
	fmt.Println("XWM:Hotkeys are registrated")
}

func (xm *XWindowManager) grabShortcut(keyCode KeyCode, modifiers ModifierMask) {
	for _, locks := range []ModifierMask{0, LockMask, Mod2Mask, LockMask | Mod2Mask} {
		xm.backend.GrabKey(xm.rootWindow, keyCode, modifiers|locks)
	}
}
