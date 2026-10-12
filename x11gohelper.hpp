#pragma once

#include <X11/Xlib.h>

static inline Window _DefaultRootWindow(Display *display) {
    return DefaultRootWindow(display);
}