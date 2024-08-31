package main

import (
	"log"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
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

func (g *Game) Init() {

}
func (g *Game) Update(deltaTime float64) {

}
func (g *Game) ProcessInput(deltaTime float64) {

}
func (g *Game) Render() {

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
	//* Load OS-specific OpenGL function pointers

	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	//* Callbacks
	// Set the function that is run every time the viewport is resized by the user.
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	// Listen to mouse events
	window.SetKeyCallback(keyCallback)

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
