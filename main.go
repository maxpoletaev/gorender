package main

import (
	"flag"
	"fmt"
	_ "image/png"
	"log"
	"os"
	"path"
	"runtime/pprof"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	viewScale   = 2
	viewWidth   = 800 / viewScale
	viewHeight  = 600 / viewScale
	frameRate   = 60
	windowTitle = "goxgl"
)

func onOff(b bool) string {
	if b {
		return "ON"
	}

	return "OFF"
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
	renderer := NewRenderer(fb)
	filename := flag.Arg(0)

	var (
		scene *Scene
		err   error
	)

	switch path.Ext(filename) {
	case ".obj":
		mesh, err := LoadMeshFile(filename)
		if err != nil {
			log.Fatalf("failed to load mesh file: %s", err)
		}

		object := NewObject(mesh)
		scene = &Scene{Objects: []*Object{object}}
	case ".json":
		scene, err = LoadSceneFile(filename)
		if err != nil {
			log.Fatalf("failed to load scene file: %s", err)
		}
	default:
		log.Fatalf("unsupported file format: %s", path.Ext(filename))
	}

	var (
		windowWidth  = int32(fb.Width * viewScale)
		windowHeight = int32(fb.Height * viewScale)
		numVertices  = scene.NumVertices()
		numTriangles = scene.NumTriangles()
		oumObjects   = scene.NumObjects()
	)

	rl.SetTraceLogLevel(rl.LogError) // Make raylib less verbose
	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	defer rl.CloseWindow()

	rl.SetTargetFPS(frameRate)

	renderTexture := rl.LoadRenderTexture(int32(fb.Width), int32(fb.Height))
	defer rl.UnloadRenderTexture(renderTexture)

	var (
		lastCursorX = rl.GetMouseX()
		lastCursorY = rl.GetMouseY()
	)

	camera := &Camera{
		Direction: Vec3{0, 0, 1},
		Position:  Vec3{10, 5, -2},
		Up:        Vec3{0, 1, 0},
	}

	triggerDraw := make(chan struct{})
	frameReady := make(chan struct{})

	go func() {
		for {
			<-triggerDraw
			cameraCopy := *camera // to prevent updating camera mid-frame
			renderer.Draw(scene.Objects, &cameraCopy)
			frameReady <- struct{}{}
		}
	}()

	rl.DisableCursor()
	triggerDraw <- struct{}{}

	for !rl.WindowShouldClose() {
		<-frameReady
		fb.SwapBuffers()
		triggerDraw <- struct{}{}

		forward := camera.Direction.Normalize()
		//forward.Y = 0 // Only move in the XZ plane
		right := forward.CrossProduct(camera.Up).Normalize()

		switch {
		// WASD keys to move the camera
		case rl.IsKeyDown(rl.KeyW):
			camera.Position = camera.Position.Add(forward.Multiply(0.15))
		case rl.IsKeyDown(rl.KeyS):
			camera.Position = camera.Position.Sub(forward.Multiply(0.15))
		case rl.IsKeyDown(rl.KeyA):
			camera.Position = camera.Position.Sub(right.Multiply(0.15))
		case rl.IsKeyDown(rl.KeyD):
			camera.Position = camera.Position.Add(right.Multiply(0.15))

		// Render options
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
		case rl.IsKeyPressed(rl.KeyX):
			renderer.DebugEnabled = !renderer.DebugEnabled
		case rl.IsKeyPressed(rl.KeyC):
			renderer.FrustumClipping = !renderer.FrustumClipping
		case rl.IsKeyPressed(rl.KeyT):
			renderer.ShowTextures = !renderer.ShowTextures
		}

		cursorX := rl.GetMouseX()
		cursorY := rl.GetMouseY()

		deltaX := cursorX - lastCursorX
		deltaY := cursorY - lastCursorY

		if deltaX != 0 || deltaY != 0 {
			yaw := -float64(deltaX) * 0.002
			pitch := -float64(deltaY) * 0.002
			yawQuaternion := NewQuaternionFromAxisAngle(camera.Up, yaw)
			pitchQuaternion := NewQuaternionFromAxisAngle(right, pitch)
			camera.Direction = yawQuaternion.Rotate(camera.Direction).Normalize()
			camera.Direction = pitchQuaternion.Rotate(camera.Direction).Normalize()
		}

		lastCursorX = cursorX
		lastCursorY = cursorY

		// Copy the frame buffer to the render texture
		rl.BeginTextureMode(renderTexture)
		rl.UpdateTexture(renderTexture.Texture, fb.Pixels2)
		rl.EndTextureMode()

		// Draw the render texture to the screen
		rl.BeginDrawing()
		rl.DrawTexturePro(
			renderTexture.Texture,
			rl.NewRectangle(0, 0, float32(fb.Width), float32(fb.Height)),
			rl.NewRectangle(0, 0, float32(fb.Width*viewScale), float32(fb.Height*viewScale)),
			rl.NewVector2(0, 0),
			0,
			rl.White,
		)

		drawText(5, 5, fmt.Sprintf("%d fps", rl.GetFPS()))
		drawText(5, 15, fmt.Sprintf("objects: %d", oumObjects))
		drawText(5, 25, fmt.Sprintf("vertices: %d", numVertices))
		drawText(5, 35, fmt.Sprintf("triangles: %d", numTriangles))

		drawText(
			5,
			windowHeight-15,
			fmt.Sprintf(
				"[V]erticies: %s [E]dges: %s [F]aces: %s, [L]ights: %s, [B]ackface culling: %s, [C]lipping: %s, [T]extures: %s",
				onOff(renderer.ShowVertices),
				onOff(renderer.ShowEdges),
				onOff(renderer.ShowFaces),
				onOff(renderer.Lighting),
				onOff(renderer.BackfaceCulling),
				onOff(renderer.FrustumClipping),
				onOff(renderer.ShowTextures),
			),
		)

		for _, info := range renderer.DebugInfo {
			rl.DrawText(info.Text, int32(info.X*viewScale)+1, int32(info.Y*viewScale)+1, 12, rl.Black)
			rl.DrawText(info.Text, int32(info.X*viewScale), int32(info.Y*viewScale), 12, rl.Yellow)
		}

		drawText(
			5,
			windowHeight-35,
			fmt.Sprintf(
				"X=%.2f Y=%.2f Z=%.2f RX=%.2f RY=%.2f RZ=%.2f",
				camera.Position.X,
				camera.Position.Y,
				camera.Position.Z,
				camera.Direction.X,
				camera.Direction.Y,
				camera.Direction.Z,
			),
		)

		rl.EndDrawing()
	}
}
