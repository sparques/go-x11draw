package x11draw

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>

typedef struct {
    Display *display;
    Window window;
    GC gc;
    XImage *image;
    int width;
    int height;
} X11Context;

void create_window(X11Context *ctx, int width, int height, uint32_t *data) {
	ctx->width=width;
	ctx->height=height;

    ctx->display = XOpenDisplay(NULL);
    if (ctx->display == NULL) {
        fprintf(stderr, "Cannot open display\n");
        exit(1);
    }

    int screen = DefaultScreen(ctx->display);
    ctx->window = XCreateSimpleWindow(ctx->display, RootWindow(ctx->display, screen), 10, 10, width, height, 1,
                                       BlackPixel(ctx->display, screen), WhitePixel(ctx->display, screen));

    XSelectInput(ctx->display, ctx->window, ExposureMask | KeyPressMask);
    XMapWindow(ctx->display, ctx->window);
    ctx->gc = XCreateGC(ctx->display, ctx->window, 0, NULL);

    ctx->image = XCreateImage(ctx->display, DefaultVisual(ctx->display, 0), 24, ZPixmap, 0,
                                  (char *)data, width, height, 32, 0);
}

void destroy_window(X11Context *ctx) {
	ctx->image->data = NULL;
    XDestroyImage(ctx->image);
    XDestroyWindow(ctx->display, ctx->window);
    XCloseDisplay(ctx->display);
}

void draw_image(X11Context *ctx, uint32_t *data, int width, int height) {
    // Create a pixmap to hold the image
    //Pixmap pixmap = XCreatePixmap(ctx->display, ctx->window, width, height, 24);

    //XPutImage(ctx->display, ctx->window, ctx->gc, ctx->image, 0, 0, 0, 0, width, height);

    // Draw the pixmap onto the window
    //XCopyArea(ctx->display, pixmap, ctx->window, ctx->gc, 0, 0, width, height, 0, 0);

	XFlush(ctx->display);
}

void update_window(X11Context *ctx) {
	XPutImage(ctx->display, ctx->window, ctx->gc, ctx->image, 0, 0, 0, 0, ctx->width, ctx->height);
	XFlush(ctx->display);
}

void wait_expose(X11Context *ctx) {
	XEvent event;
	for (;;) {
		XNextEvent(ctx->display, &event);
		if (event.type == Expose) {
			return;
		}
	}
}

KeySym getKeySym(Display *disp) {
	XEvent evt;
	KeySym sym;
	XNextEvent(disp, &evt);
	sym = XLookupKeysym((XKeyEvent *)(&evt),0);
	if (((XKeyEvent *)(&evt))->type == 2) {
		return sym;
	}
	return 0;
}

*/
import "C"

import (
	"image"
	"image/color"
	"unsafe"
)

type Window struct {
	*image.RGBA
	ctx C.X11Context
	set func(x, y int, c color.Color)
}

func NewWindow(width, height int) *Window {
	buf := image.NewRGBA(image.Rect(0, 0, width, height))
	ctx := C.X11Context{}
	C.create_window(&ctx, C.int(width), C.int(height), (*C.uint32_t)(unsafe.Pointer(&buf.Pix[0])))
	C.wait_expose(&ctx)
	w := &Window{
		RGBA: buf,
		ctx:  ctx,
	}
	w.set = w.setSync
	return w
}

func (w *Window) UseAsync() {
	w.set = w.setAsync
}

func (w *Window) UseSync() {
	w.set = w.setSync
}

// Sync syncs the window to the internal buffer.
func (w *Window) Sync() {
	C.update_window(&w.ctx)
}

func (w *Window) setAsync(x, y int, c color.Color) {
	w.RGBA.Set(x, y, c)
}

func (w *Window) setSync(x, y int, c color.Color) {
	w.RGBA.Set(x, y, c)
	C.update_window(&w.ctx)
}

func (w *Window) Set(x, y int, c color.Color) {
	w.set(x, y, c)
}

func (w *Window) NextEvent() (event C.XEvent) {
	C.XNextEvent(w.ctx.display, &event)
	return
}

func (w *Window) Close() {
	C.destroy_window(&w.ctx)
}
