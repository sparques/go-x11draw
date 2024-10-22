package x11draw // import "github.com/sparques/go-x11draw"
/*
x11draw is a very simple way to get a draw.Draw interface the is displayed in an X11 window.

Its only dependency is CGO and X11--no need for OpenGL, SDL, or some complicated build tool chain.

Example below opens a window and then writes random colored pixels in random locations.

	// Open an X11 window
	win := x11draw.NewWindow(800, 600)
	// We automatically write changes to the window by default.
	// If we wanted to manually have to sync updates we could uncomment below
	// and then call win.Sync() to write the update to screen.
	// win.UseAsync()

	go func() {
		for {
			time.Sleep(time.Second/10)
			win.Set(rand.Intn(win.Bounds().Dx()),
			 	rand.Intn(win.Bounds().Dy()), color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255})
		}
	}()

	for {
		win.NextEvent()
		// win.Sync() // do update
	}
*/

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

void change_image(X11Context *ctx, uint32_t *data, int width, int height) {
	ctx->width = width;
	ctx->height = height;
	ctx->image = XCreateImage(ctx->display, DefaultVisual(ctx->display, 0), 24, ZPixmap, 0,
                                  (char *)data, width, height, 32, 0);
}

void create_window(X11Context *ctx, int width, int height, uint32_t *data) {
    ctx->display = XOpenDisplay(NULL);
    if (ctx->display == NULL) {
        fprintf(stderr, "Cannot open display\n");
        exit(1);
    }

    int screen = DefaultScreen(ctx->display);
    ctx->window = XCreateSimpleWindow(ctx->display, RootWindow(ctx->display, screen), 10, 10, width, height, 1,
                                       BlackPixel(ctx->display, screen), WhitePixel(ctx->display, screen));

    XSelectInput(ctx->display, ctx->window, ExposureMask | KeyPressMask | KeyReleaseMask);
    XMapWindow(ctx->display, ctx->window);
    ctx->gc = XCreateGC(ctx->display, ctx->window, 0, NULL);

	change_image(ctx, data, width, height);
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

// Window implements draw.Draw (and image.Draw). Writes to it show up on screen.
type Window struct {
	*image.RGBA
	ctx C.X11Context
	set func(x, y int, c color.Color)
}

// NewWindow creates and attempts to open a new X11 window. A backing buffer in the form of a
// image.RGBA of the same size is also created.
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

func (w *Window) UseImage(img *image.RGBA) {
	w.RGBA = img
	C.change_image(&w.ctx, (*C.uint32_t)(unsafe.Pointer(&img.Pix[0])), C.int(img.Bounds().Dx()), C.int(img.Bounds().Dy()))
}

// UseAsync makes it so calls to (*Window).Set do not cause the X11 window to be updated.
func (w *Window) UseAsync() {
	w.set = w.setAsync
}

// UseSync makes it so calls to (*Window).Set cause the X11 window to be updated.
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

// Set implements draw.Draw interface.
func (w *Window) Set(x, y int, c color.Color) {
	w.set(x, y, c)
}

// NextEvent returns the next xlib event. This is a blocking call.
func (w *Window) NextEvent() (event C.XEvent) {
	C.XNextEvent(w.ctx.display, &event)
	return
}

// Close closes the window. Image operations may still happen, but you cannot get the window back.
// Calling Sync or using Sync will cause a panic.
func (w *Window) Close() {
	C.destroy_window(&w.ctx)
}
