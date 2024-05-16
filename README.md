# goxgl

A purely software (no OpenGL or DirectX) 3D rasterizer, that I’m trying to build 
to learn some math and magic behind 3D graphics. It uses raylib for window 
management and delivering pixels to the screen, but all the rendering is done 
from scratch in Go with no external libraries.

![screenshot](screenshot.png)

## Building

For this, you may need a C compiler and additional dependencies required by 
raylib. See https://github.com/gen2brain/raylib-go#requirements for details.

```
$ make build
```

## Running

```
$ ./goxgl models/suzanne.obj
```

## Features

 * [x] Wireframe rendering
 * [x] OBJ file loading
 * [x] Backface culling
 * [x] Transformations
 * [x] Texture mapping
 * [x] Flat shading
 * [ ] Gouraud shading
 * [x] Z-buffering
 * [x] View frustum clipping

## Resources

 * [Scratchapixel](https://www.scratchapixel.com)
 * [tinyrenderer](https://github.com/ssloy/tinyrenderer) by Dmitry V. Sokolov
 * [Math for Game Developers](https://www.youtube.com/playlist?list=PLW3Zl3wyJwWOpdhYedlD-yCB7WQoHf-My) series by Jorge Rodriguez
 * Code-It-Yourself! - 3D Graphics Engine series by javidx9: [part 1][CIY-1], [part 2][CIY-2], [part 3][CIY-3], [part 4][CIY-4]
 * [60005/70090 Computer Graphics lectures](https://wp.doc.ic.ac.uk/bkainz/teaching/60005-co317-computer-graphics/) by Bernhard Kainz at Imperial College London
 * [Alias/WaveFront Object (.obj) File Format](https://people.computing.clemson.edu/~dhouse/courses/405/docs/brief-obj-file-format.html)

[CIY-1]: https://www.youtube.com/watch?v=ih20l3pJoeU&list=PLrOv9FMX8xJE8NgepZR1etrsU63fDDGxO&index=22&t=1938s&pp=iAQB
[CIY-2]: https://www.youtube.com/watch?v=XgMWc6LumG4&list=PLrOv9FMX8xJE8NgepZR1etrsU63fDDGxO&index=23&pp=iAQB
[CIY-3]: https://www.youtube.com/watch?v=HXSuNxpCzdM&list=PLrOv9FMX8xJE8NgepZR1etrsU63fDDGxO&index=24&t=621s&pp=iAQB
[CIY-4]: https://www.youtube.com/watch?v=nBzCS-Y0FcY&list=PLrOv9FMX8xJE8NgepZR1etrsU63fDDGxO&index=25&pp=iAQB