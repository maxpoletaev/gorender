package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime/pprof"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	viewScale   = 2
	viewWidth   = 400
	viewHeight  = 280
	fovFactor   = 800.0
	windowTitle = "goxgl"
)

type Camera struct {
	Position Vec3
}

func drawText(x, y int32, text string) {
	rl.DrawText(text, x+1, y+1, 10, rl.Black)
	rl.DrawText(text, x, y, 10, rl.White)
}

func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

func loadMeshFromFile(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	return ReadObj(file)
}

type options struct {
	cpuProfile string
	memProfile string
}

func parseOptions() *options {
	opts := &options{}
	flag.StringVar(&opts.cpuProfile, "cpuprof", "", "write cpu profile to file")
	flag.StringVar(&opts.memProfile, "memprof", "", "write memory profile to file")
	flag.Parse()
	return opts
}

func main() {
	opts := parseOptions()

	if flag.NArg() != 1 {
		log.Fatalf("usage: %s <filename>", os.Args[0])
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

	filename := flag.Arg(0)
	fb := NewFrameBuffer(viewWidth, viewHeight)
	camera := Camera{Position: Vec3{0, 0, 0}}
	renderer := NewRenderer(fb)

	mesh, err := loadMeshFromFile(filename)
	if err != nil {
		log.Fatalf("failed to load mesh: %s", err)
	}

	// Move away from the camera
	mesh.Translation.Z = -10

	var (
		windowWidth  = int32(fb.Width * viewScale)
		windowHeight = int32(fb.Height * viewScale)
	)

	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	texture := rl.LoadRenderTexture(int32(fb.Width), int32(fb.Height))
	defer rl.UnloadRenderTexture(texture)

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
		}

		cursorX := rl.GetMouseX()
		cursorY := rl.GetMouseY()

		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			deltaX := cursorX - lastCursorX
			deltaY := cursorY - lastCursorY

			if deltaX != 0 || deltaY != 0 {
				if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
					mesh.Translation.X -= float32(deltaX) * 0.005
					mesh.Translation.Y -= float32(deltaY) * 0.005
				} else {
					mesh.Rotation.X += float32(deltaY) * 0.01
					mesh.Rotation.Y += -float32(deltaX) * 0.01
				}
			}
		}

		lastCursorX = cursorX
		lastCursorY = cursorY

		// Zoom in/out with the mouse wheel
		if wheelMove := rl.GetMouseWheelMove(); wheelMove != 0 {
			factor := float32(wheelMove) * 0.01
			mesh.Scale.X += factor
			mesh.Scale.Y += factor
			mesh.Scale.Z += factor
		}

		// Copy the frame buffer to the render texture
		rl.BeginTextureMode(texture)
		rl.UpdateTexture(texture.Texture, fb.Pixels)
		rl.EndTextureMode()

		// Draw the render texture to the screen
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.DrawTexturePro(
			texture.Texture,
			rl.Rectangle{0, 0, float32(fb.Width), float32(fb.Height)},
			rl.Rectangle{0, 0, float32(fb.Width * viewScale), float32(fb.Height * viewScale)},
			rl.Vector2{0, 0},
			0,
			rl.White,
		)

		drawText(5, 5, fmt.Sprintf("%d fps", rl.GetFPS()))
		drawText(5, 15, path.Base(filename))
		drawText(5, 25, fmt.Sprintf("vertices: %d", len(mesh.Vertices)))
		drawText(5, 35, fmt.Sprintf("faces: %d", len(mesh.Faces)))

		drawText(5, windowHeight-15, fmt.Sprintf("[V]erticies: %s [E]dges: %s [F]aces: %s, [B]ackface culling: %s",
			onOff(renderer.ShowVertices), onOff(renderer.ShowEdges), onOff(renderer.ShowFaces), onOff(renderer.BackfaceCulling)))

		rl.EndDrawing()
	}
}
