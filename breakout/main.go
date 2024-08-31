package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Settings
const (
	windowWidth  = 800
	windowHeight = 600
)

type GameState int

const (
	GameActive GameState = iota
	GameMenu
	GameWin
)

type Game struct {
	state         GameState
	keys          [1024]bool
	width, height int
}

var game = Game{
	state:  GameActive,
	width:  windowWidth,
	height: windowHeight,
}
var renderer *SpriteRenderer

func (g *Game) Init() {
	// load shaders
	LoadShader("shaders/sprite.vs", "shaders/sprite.fs", "", "sprite")
	// configure shaders
	projection := mgl32.Ortho(0.0, float32(g.width), float32(g.height), 0.0, -1.0, 1.0)
	GetShader("sprite").use().setInt("image", 0)
	GetShader("sprite").use().setMat4("projection", projection)
	// set render-specific controls
	renderer = NewSpriteRenderer(GetShader("sprite"))
	// load textures
	LoadTexture("textures/awesomeface.png", true, "face")
}
func (g *Game) Update(deltaTime float64) {

}
func (g *Game) ProcessInput(deltaTime float64) {

}
func (g *Game) Render() {
	renderer.drawSprite(GetTexture("face"), mgl32.Vec2{200.0, 200.0}, mgl32.Vec2{300.0, 400.0}, 45.0, mgl32.Vec3{0.0, 1.0, 0.0})
}

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()
}

func main() {
	//* GLFW init and configure
	err := glfw.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()
	// Using hints, set various options for the window we're about to create.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	// Compatibility profile allows more deprecated function calls over core profile.
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	//* GLFW window creation
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Breakout", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()
	//* Callbacks
	// Set the function that is run every time the viewport is resized by the user.
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	// Listen to mouse events
	window.SetKeyCallback(keyCallback)

	//* Load OS-specific OpenGL function pointers
	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	//* OpenGL configuration
	gl.Viewport(0, 0, windowWidth, windowHeight)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	//* initialize game
	game.Init()

	//* deltaTime variables
	deltaTime := 0.0
	lastFrame := 0.0

	for !window.ShouldClose() {
		// calculate time stats
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame
		glfw.PollEvents()

		//* manage user input
		game.ProcessInput(deltaTime)

		//* update game state
		game.Update(deltaTime)

		//* render
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		game.Render()

		window.SwapBuffers()
	}
}

// framebufferSizeCallback is called when the gl viewport is resized.
func framebufferSizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

// keyCallback is called when the gl viewport is resized.
func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}
	if key >= 0 && key < 1024 {
		if action == glfw.Press {
			game.keys[key] = true
		} else if action == glfw.Release {
			game.keys[key] = false
		}
	}
}
