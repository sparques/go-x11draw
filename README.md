# x11draw

Finally, a simple way to see `image.Image`s without having to resort to dumping it to a PNG and opening it in a nother program.

x11draw makes an X11 window that implements the draw.Draw interface, meaning you can use any arbitrary go package that works with `image.Image`s or `draw.Image`s. For example, https://github.com/fogleman/gg.

## Status

It works for viewing image data. I'm still working on making the input events useful somehow. They are exposed via the CGO interface, but a better means is coming.