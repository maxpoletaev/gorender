package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"os"
	"path"
	"runtime/pprof"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	viewScale   = 2
	viewWidth   = 800 / viewScale
	viewHeight  = 600 / viewScale
	windowTitle = "goxgl"
)

type Camera struct {
	Position Vec3
	FOVAngle float64
}

func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

func loadMeshFile(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	mesh, err := ReadObj(file)
	if err != nil {
		return nil, err
	}

	mesh.Name = path.Base(filename)

	return mesh, nil
}

func loadTextureFile(filename string) (*ImageTexture, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	texture := &ImageTexture{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Pixels: make([]color.RGBA, bounds.Dx()*bounds.Dy()),
	}

	for y := 0; y < texture.Height; y++ {
		for x := 0; x < texture.Width; x++ {
			c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			texture.Pixels[y*texture.Width+x] = c
		}
	}

	return texture, nil
}

type options struct {
	cpuProfile string
	memProfile string
	texture    string
}

func parseOptions() *options {
	opts := &options{}
	flag.StringVar(&opts.cpuProfile, "cpuprof", "", "write cpu profile to file")
	flag.StringVar(&opts.memProfile, "memprof", "", "write memory profile to file")
	flag.StringVar(&opts.texture, "texture", "", "texture file")
	flag.Parse()
	return opts
}

func drawText(x, y int32, text string) {
	rl.DrawText(text, x+1, y+1, 10, rl.Black)
	rl.DrawText(text, x, y, 10, rl.White)
}

func main() {
	opts := parseOptions()

	if flag.NArg() == 0 {
		log.Fatalf("usage: %s [options] filename.obj", os.Args[0])
	}

	if opts.cpuProfile != "" {
		f, err := os.Create(opts.cpuProfile)
		if err != nil {
			log.Fatalf("failed to create CPU profile: %s", err)
		}

		defer func() {
			_ = f.Close()
		}()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Printf("[ERROR] failed to start CPU profile: %s", err)
		}

		defer pprof.StopCPUProfile()
	}

	if opts.memProfile != "" {
		f, err := os.Create(opts.memProfile)
		if err != nil {
			log.Fatalf("failed to create memory profile: %s", err)
		}

		defer func() {
			_ = f.Close()
		}()

		defer func() {
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Printf("[ERROR] failed to write memory profile: %s", err)
			}
		}()
	}

	fb := NewFrameBuffer(viewWidth, viewHeight)
	camera := Camera{Position: Vec3{0, 0, -1}, FOVAngle: 45}
	renderer := NewRenderer(fb)

	mesh, err := loadMeshFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("failed to load mesh: %s", err)
	}

	if opts.texture != "" {
		texture, err := loadTextureFile(opts.texture)
		if err != nil {
			log.Fatalf("failed to load texture: %s", err)
		}

		mesh.Texture = texture
	}

	mesh.Translation.Z = 5    // Move away from the camera
	mesh.Rotation.Y = math.Pi // Rotate 180 degrees

	var (
		windowWidth  = int32(fb.Width * viewScale)
		windowHeight = int32(fb.Height * viewScale)
	)

	rl.SetTraceLogLevel(rl.LogError) // Make raylib less verbose
	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	defer rl.CloseWindow()

	rl.SetTargetFPS(30)

	renderTexture := rl.LoadRenderTexture(int32(fb.Width), int32(fb.Height))
	defer rl.UnloadRenderTexture(renderTexture)

	var (
		lastCursorX = rl.GetMouseX()
		lastCursorY = rl.GetMouseY()
	)

	for !rl.WindowShouldClose() {
		renderer.Draw(mesh, &camera)

		switch {
		case rl.IsKeyPressed(rl.KeyB):
			renderer.BackfaceCulling = !renderer.BackfaceCulling
		case rl.IsKeyPressed(rl.KeyE):
			renderer.ShowEdges = !renderer.ShowEdges
		case rl.IsKeyPressed(rl.KeyF):
			renderer.ShowFaces = !renderer.ShowFaces
		case rl.IsKeyPressed(rl.KeyV):
			renderer.ShowVertices = !renderer.ShowVertices
		case rl.IsKeyPressed(rl.KeyL):
			renderer.Lighting = !renderer.Lighting
		case rl.IsKeyPressed(rl.KeyD):
			renderer.DebugEnabled = !renderer.DebugEnabled
		}

		cursorX := rl.GetMouseX()
		cursorY := rl.GetMouseY()

		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			deltaX := cursorX - lastCursorX
			deltaY := cursorY - lastCursorY

			if deltaX != 0 || deltaY != 0 {
				if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
					mesh.Translation.X -= float64(deltaX) * 0.005
					mesh.Translation.Y -= float64(deltaY) * 0.005
				} else {
					mesh.Rotation.X -= float64(deltaY) * 0.01
					mesh.Rotation.Y += float64(deltaX) * 0.01
				}
			}
		}

		lastCursorX = cursorX
		lastCursorY = cursorY

		// Zoom in/out with the mouse wheel
		if wheelMove := rl.GetMouseWheelMove(); wheelMove != 0 {
			factor := float64(wheelMove) * 0.01
			mesh.Scale.X += factor
			mesh.Scale.Y += factor
			mesh.Scale.Z += factor
		}

		// Copy the frame buffer to the render texture
		rl.BeginTextureMode(renderTexture)
		rl.UpdateTexture(renderTexture.Texture, fb.Pixels)
		rl.EndTextureMode()

		// Draw the render texture to the screen
		rl.BeginDrawing()
		rl.DrawTexturePro(
			renderTexture.Texture,
			rl.Rectangle{0, 0, float32(fb.Width), float32(fb.Height)},
			rl.Rectangle{0, 0, float32(fb.Width * viewScale), float32(fb.Height * viewScale)},
			rl.Vector2{0, 0},
			0,
			rl.White,
		)

		drawText(5, 5, fmt.Sprintf("%d fps", rl.GetFPS()))
		drawText(5, 15, path.Base(mesh.Name))
		drawText(5, 25, fmt.Sprintf("vertices: %d", len(mesh.Vertices)))
		drawText(5, 35, fmt.Sprintf("faces: %d", len(mesh.Faces)))

		drawText(5, windowHeight-15, fmt.Sprintf("[V]erticies: %s [E]dges: %s [F]aces: %s, [L]ights: %s, [B]ackface culling: %s",
			onOff(renderer.ShowVertices),
			onOff(renderer.ShowEdges),
			onOff(renderer.ShowFaces),
			onOff(renderer.Lighting),
			onOff(renderer.BackfaceCulling),
		))

		for _, info := range renderer.DebugInfo {
			rl.DrawText(info.Text, int32(info.X*viewScale)+1, int32(info.Y*viewScale)+1, 12, rl.Black)
			rl.DrawText(info.Text, int32(info.X*viewScale), int32(info.Y*viewScale), 12, rl.Yellow)
		}

		rl.EndDrawing()
	}
}
