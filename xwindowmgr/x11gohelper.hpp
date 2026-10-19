#pragma once

#include <X11/Xlib.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

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

static inline int XErrorHandlerImpl(Display *display, XErrorEvent *error) {
    // Игнорируем BadWindow (код 3), BadMatch (код 8) и т.д., просто выводя в лог
    if (error->error_code == BadWindow) {
        fprintf(stderr, " Ignored BadWindow error (Serial: %lu)\n", error->serial);
        return 0; 
    }
    fprintf(stderr, "X11 Error: code %d, request %d, minor %d\n",
            error->error_code, error->request_code, error->minor_code);
    return 0; // Возвращаем 0, чтобы программа НЕ завершалась
}

static inline void SetupErrorHandler() {
    XSetErrorHandler(XErrorHandlerImpl);
}
