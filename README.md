# x11draw

Finally, a simple way to see `image.Image`s without having to resort to dumping it to a PNG and opening it in another program.

x11draw makes an X11 window that implements the draw.Draw interface, meaning you can use any arbitrary go package that works with `image.Image`s or `draw.Image`s. For example, https://github.com/fogleman/gg.

## Status

It works for viewing image data. I'm still working on making the input events useful somehow. They are exposed via the CGO interface, but a better means is coming.

## Performance

The most basic implementation of draw.Draw is flexible, easy to use, and horrendously slow. the [gfx package](https://github.com/sparques/gfx) defines some interfaces that an image.Drawor image.Image might optionally support, permitting much improved performance.

For example:

```
	if fillable, ok := someImage.(gfx.Filler); ok {
		fillable.Fill(someImage.Bounds(), color.Black{})
	} else {
		draw.Draw(someImage, someImage.Bounds(), image.Black, image.Point{}, draw.Over)
	}
```

In the above example we cast someImage to gfx.Filler. If this succeeds, we call the Fill method. Otherwise, we use draw.Draw which depends on the single-pixel-addressing Set() method. 

Generally, `image.Image`s have some efficient means of storing pixels, usually in a slice, which means pixels can be more efficiently modified with direct memory operations like copying or even simply looping over each element of the slice and setting it--both of which will be faster than simply calling Set().

x11draw supports gfx.Scroller, gfx.RegionScroller, and gfx.Fill. Support for gfx.Blitter and gfx.VectorScroller are comng.