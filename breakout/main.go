package main

import (
	"log"
	"math"
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

type Direction int

const (
	Up Direction = iota
	Right
	Down
	Left
)

type Collision struct {
	hit  bool
	dir  Direction
	diff mgl32.Vec2
}

type Game struct {
	state         GameState
	keys          [1024]bool
	width, height int
	levels        []GameLevel
	currentLevel  int
}

var game = Game{
	state:  GameActive,
	width:  windowWidth,
	height: windowHeight,
}
var renderer *SpriteRenderer

var (
	player         *GameObject
	playerSize     = mgl32.Vec2{100.0, 20.0}
	playerVelocity = 500.0

	ball                *Ball
	initialBallVelocity = mgl32.Vec2{100.0, -350.0}
	ballRadius          = float32(12.5)

	particles *ParticleGenerator

	effects   *PostProcessor
	shakeTime = float32(0)
)

func (g *Game) Init() {
	// load shaders
	LoadShader("shaders/sprite.vs", "shaders/sprite.fs", "", "sprite")
	LoadShader("shaders/particle.vs", "shaders/particle.fs", "", "particle")
	LoadShader("shaders/post_processing.vs", "shaders/post_processing.fs", "", "postprocessing")
	// configure shaders
	projection := mgl32.Ortho(0.0, float32(g.width), float32(g.height), 0.0, -1.0, 1.0)
	GetShader("sprite").use().setInt("image", 0)
	GetShader("sprite").use().setMat4("projection", projection)
	GetShader("particle").use().setInt("sprite", 0)
	GetShader("particle").use().setMat4("projection", projection)
	// load textures
	LoadTexture("textures/awesomeface.png", true, "face")
	LoadTexture("textures/block.png", false, "block")
	LoadTexture("textures/block_solid.png", false, "block_solid")
	LoadTexture("textures/background.jpg", false, "background")
	LoadTexture("textures/paddle.png", true, "paddle")
	LoadTexture("textures/particle.png", true, "particle")
	// set render-specific controls
	renderer = NewSpriteRenderer(GetShader("sprite"))
	particles = NewParticleGenerator(GetShader("particle"), GetTexture("particle"), 500)
	effects = NewPostProcessor(GetShader("postprocessing"), int32(g.width), int32(g.height))
	// load levels
	one := LoadLevel("levels/one.lvl", g.width, g.height/2)
	two := LoadLevel("levels/two.lvl", g.width, g.height/2)
	three := LoadLevel("levels/three.lvl", g.width, g.height/2)
	four := LoadLevel("levels/four.lvl", g.width, g.height/2)
	g.levels = append(g.levels, one, two, three, four)

	playerPosition := mgl32.Vec2{
		float32(g.width)/2.0 - playerSize.X()/2.0,
		float32(g.height) - playerSize.Y(),
	}
	player = &GameObject{
		position: playerPosition,
		size:     playerSize,
		sprite:   GetTexture("paddle"),
	}

	ballPos := playerPosition.Add(mgl32.Vec2{
		playerSize.X()/2.0 - ballRadius,
		-ballRadius * 2.0,
	})
	ball = NewBall(ballPos, ballRadius, initialBallVelocity, GetTexture("face"))

}
func (g *Game) Update(deltaTime float64) {
	// update objects
	ball.Move(float32(deltaTime), g.width)
	// check for collisions
	game.DoCollisions()
	// update particles
	particles.Update(float32(deltaTime), ball, 2, mgl32.Vec2{ball.radius / 2.0, ball.radius / 2.0})
	// did ball reach bottom edge?
	if ball.obj.position.Y() >= float32(g.height) {
		g.ResetLevel()
		g.ResetPlayer()
	}
	if shakeTime > 0.0 {
		shakeTime -= float32(deltaTime)
		if shakeTime <= 0.0 {
			effects.shake = false
		}
	}
}
func (g *Game) ProcessInput(deltaTime float64) {
	if g.state == GameActive {
		velocity := playerVelocity * deltaTime
		// move playerboard
		if g.keys[glfw.KeyA] || g.keys[glfw.KeyLeft] {
			if player.position.X() >= 0.0 {
				player.position[0] -= float32(velocity)
				if ball.stuck {
					ball.obj.position[0] -= float32(velocity)
				}
			}
		}
		if g.keys[glfw.KeyD] || g.keys[glfw.KeyRight] {
			if player.position.X() <= float32(g.width)-playerSize.X() {
				player.position[0] += float32(velocity)
				if ball.stuck {
					ball.obj.position[0] += float32(velocity)
				}
			}
		}
		// player start game
		if g.keys[glfw.KeySpace] {
			ball.stuck = false
		}
	}
}
func (g *Game) Render() {
	if g.state == GameActive {

		effects.BeginRender()

		// draw background
		renderer.DrawSprite(
			GetTexture("background"),
			mgl32.Vec2{0.0, 0.0},
			SpriteRendererOptions{size: mgl32.Vec2{float32(g.width), float32(g.height)}},
		)
		// draw level
		g.levels[g.currentLevel].Draw(renderer)
		// draw player
		player.Draw(renderer)
		// draw particles
		particles.Draw()
		// draw ball
		ball.obj.Draw(renderer)

		effects.EndRender()
		effects.Render(float32(glfw.GetTime()))
	}
}
func (g *Game) DoCollisions() {
	for i := range g.levels[g.currentLevel].bricks {
		box := &g.levels[g.currentLevel].bricks[i]
		if !box.destroyed {
			collision := CheckBallCollision(ball, box)
			if collision.hit {
				// destroy block if not solid
				if !box.isSolid {
					box.destroyed = true
				} else {
					// if block is solid, shake it
					shakeTime = 0.05
					effects.shake = true
				}
				// collision resolution
				dir := collision.dir
				diff := collision.diff
				if dir == Left || dir == Right {
					// horizontal collision
					// reverse
					ball.obj.velocity[0] = -ball.obj.velocity[0]
					// relocate
					penetration := ball.radius - float32(math.Abs(float64(diff.X())))
					if dir == Left {
						// move ball right
						ball.obj.position[0] += penetration
					} else {
						// move ball left
						ball.obj.position[0] -= penetration
					}
				} else {
					// vertical collision
					// reverse
					ball.obj.velocity[1] = -ball.obj.velocity[1]
					// relocate
					penetration := ball.radius - float32(math.Abs(float64(diff.Y())))
					if dir == Up {
						// move ball right
						ball.obj.position[1] -= penetration
					} else {
						// move ball left
						ball.obj.position[1] += penetration
					}
				}
			}
		}
	}

	collision := CheckBallCollision(ball, player)
	if !ball.stuck && collision.hit {
		// check where it hit the board, and change velocity accordingly
		centerBoard := player.position.X() + player.size.X()/2.0
		distance := (ball.obj.position.X() + ball.radius) - centerBoard
		percentage := distance / (player.size.X() / 2.0)
		// do the move
		strength := float32(2.0)
		oldVelocity := ball.obj.velocity
		ball.obj.velocity[0] = initialBallVelocity.X() * percentage * strength
		ball.obj.velocity[1] = -1.0 * float32(math.Max(1.5*math.Abs(float64(ball.obj.velocity[1])), 250.0))
		ball.obj.velocity = ball.obj.velocity.Normalize().Mul(oldVelocity.Len())
	}
}
func (g *Game) ResetLevel() {
	switch g.currentLevel {
	case 0:
		g.levels[0] = LoadLevel("levels/one.lvl", g.width, g.height/2)
	case 1:
		g.levels[1] = LoadLevel("levels/two.lvl", g.width, g.height/2)
	case 2:
		g.levels[2] = LoadLevel("levels/three.lvl", g.width, g.height/2)
	case 3:
		g.levels[3] = LoadLevel("levels/four.lvl", g.width, g.height/2)
	}
}
func (g *Game) ResetPlayer() {
	// reset player/ball stats
	player.size = playerSize
	player.position = mgl32.Vec2{float32(g.width)/2.0 - playerSize.X()/2.0, float32(g.height) - playerSize.Y()}
	ball.Reset(player.position.Add(mgl32.Vec2{playerSize.X()/2.0 - ballRadius, -(ballRadius * 2.0)}), initialBallVelocity)
}
func VectorDirection(target mgl32.Vec2) Direction {
	compass := []mgl32.Vec2{
		{0.0, 1.0},  // up
		{1.0, 0.0},  // right
		{0.0, -1.0}, // down
		{-1.0, 0.0}, // left
	}
	max := float32(0)
	bestMatch := -1
	for i := 0; i < 4; i++ {
		dotProduct := target.Normalize().Dot(compass[i])
		if dotProduct > max {
			max = dotProduct
			bestMatch = i
		}
	}
	return Direction(bestMatch)
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
