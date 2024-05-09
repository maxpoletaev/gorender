# GOXGL

A purely software (no OpenGL or DirectX) 3D rasterizer written, that I’m trying 
to build to learn more about 3D graphics programming. It uses raylib for window 
management and input handling, but all the rendering is done from scratch.

[Screenshot](screenshot.png)

## Building

For this, you may need a C compiler (gcc or clang) and additional dependencies 
required by raylib. See https://github.com/gen2brain/raylib-go#requirements
for more details.

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
 * [ ] Texture mapping
 * [ ] Shading
 * [ ] Z-buffering
 * [ ] View frustum clipping

## Resources

 * [Scratchapixel](https://www.scratchapixel.com) 
 * [tinyrenderer](https://github.com/ssloy/tinyrenderer)
 * [Math for Game Developers](https://www.youtube.com/playlist?list=PLW3Zl3wyJwWOpdhYedlD-yCB7WQoHf-My)