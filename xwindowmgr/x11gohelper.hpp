#pragma once

#include <X11/Xlib.h>
#include <stdlib.h>
#include <string.h>

static inline Window _DefaultRootWindow(Display *display) {
    return DefaultRootWindow(display);
}

static inline void send_quit_event(Display *dpy, Window win) {
    XClientMessageEvent ev;
    memset(&ev, 0, sizeof(ev));
    ev.type = ClientMessage;
    ev.window = win;
    ev.format = 32;
    ev.data.l[0] = 999; // Уникальный маркер выхода

    XSendEvent(dpy, win, False, NoEventMask, (XEvent*)&ev);
    XFlush(dpy); // Срочно проталкиваем в X-сервер
}

static inline char* get_window_name(Display *dpy, Window win) {
    char *name = NULL;
    if (XFetchName(dpy, win, &name) && name) {
        return name;
    }
    return NULL;
}
