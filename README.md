# goxgl

A purely software (no OpenGL or DirectX) realtime 3D renderer, that I’m trying
to build to learn some math and magic behind 3D graphics. It uses raylib for window 
management and delivering pixels to the screen, but all the rendering is done
from scratch in Go with no external libraries.

https://github.com/maxpoletaev/goxgl/assets/1812128/fb309de4-3a43-47c5-a90b-879352eb89ab

*Result of loading a model of a Doom map converted to OBJ format*

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

Camera uses WASD + mouse to move around (like in most first-person games). ESC
key closes the window. There is also a bunch of keys to toggle different rendering
options like wireframe, texturing, backface culling, etc.

## Features

 * Wireframe rendering
 * Backface culling
 * Affine texture mapping
 * Perspective correct texture mapping
 * Flat shading
 * Z-buffering
 * View frustum clipping
 * OBJ file support (with MTL files) - only triangulated
 * Parallel tile-based rendering
 * Multi-object scenes support

## Resources

 * [Scratchapixel](https://www.scratchapixel.com)
 * [tinyrenderer](https://github.com/ssloy/tinyrenderer) by Dmitry V. Sokolov
 * [Math for Game Developers](https://www.youtube.com/playlist?list=PLW3Zl3wyJwWOpdhYedlD-yCB7WQoHf-My) series by Jorge Rodriguez
 * Code-It-Yourself! - 3D Graphics Engine series by javidx9: [part 1][CIY-1], [part 2][CIY-2], [part 3][CIY-3], [part 4][CIY-4]
 * [60005/70090 Computer Graphics lectures](https://wp.doc.ic.ac.uk/bkainz/teaching/60005-co317-computer-graphics/) by Bernhard Kainz at Imperial College London
 * [Optimizing Software Occlusion Culling](https://fgiesen.wordpress.com/2013/02/17/optimizing-sw-occlusion-culling-index/) series by Fabian Giesen
 * [Alias/WaveFront Object (.obj) File Format](https://people.computing.clemson.edu/~dhouse/courses/405/docs/brief-obj-file-format.html)
 * [Doom E1M1: Hangar - Map (.obj)][DOOM-MAP] by pancakesbassoondonut
 * [16x16 pixel textures](https://piiixl.itch.io/textures) by PiiiXL

[CIY-1]: https://www.youtube.com/watch?v=ih20l3pJoeU&list=PLrOv9FMX8xJE8NgepZR1etrsU63fDDGxO&index=22&t=1938s&pp=iAQB
[CIY-2]: https://www.youtube.com/watch?v=XgMWc6LumG4&list=PLrOv9FMX8xJE8NgepZR1etrsU63fDDGxO&index=23&pp=iAQB
[CIY-3]: https://www.youtube.com/watch?v=HXSuNxpCzdM&list=PLrOv9FMX8xJE8NgepZR1etrsU63fDDGxO&index=24&t=621s&pp=iAQB
[CIY-4]: https://www.youtube.com/watch?v=nBzCS-Y0FcY&list=PLrOv9FMX8xJE8NgepZR1etrsU63fDDGxO&index=25&pp=iAQB
[DOOM-MAP]: https://sketchfab.com/3d-models/doom-e1m1-hangar-map-2148fb6a3fe7454b901fcea67d70b318
