package main

import (
	"log"
	"math"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.6-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Settings
const (
	initialWindowWidth  = 800
	initialWindowHeight = 600
	enableWireframe     = false
)

var (
	// 36 points for a cube
	vertices = []float32{
		// positions      // normals
		-0.5, -0.5, -0.5, 0.0, 0.0, -1.0,
		0.5, -0.5, -0.5, 0.0, 0.0, -1.0,
		0.5, 0.5, -0.5, 0.0, 0.0, -1.0,
		0.5, 0.5, -0.5, 0.0, 0.0, -1.0,
		-0.5, 0.5, -0.5, 0.0, 0.0, -1.0,
		-0.5, -0.5, -0.5, 0.0, 0.0, -1.0,

		-0.5, -0.5, 0.5, 0.0, 0.0, 1.0,
		0.5, -0.5, 0.5, 0.0, 0.0, 1.0,
		0.5, 0.5, 0.5, 0.0, 0.0, 1.0,
		0.5, 0.5, 0.5, 0.0, 0.0, 1.0,
		-0.5, 0.5, 0.5, 0.0, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0, 1.0,

		-0.5, 0.5, 0.5, -1.0, 0.0, 0.0,
		-0.5, 0.5, -0.5, -1.0, 0.0, 0.0,
		-0.5, -0.5, -0.5, -1.0, 0.0, 0.0,
		-0.5, -0.5, -0.5, -1.0, 0.0, 0.0,
		-0.5, -0.5, 0.5, -1.0, 0.0, 0.0,
		-0.5, 0.5, 0.5, -1.0, 0.0, 0.0,

		0.5, 0.5, 0.5, 1.0, 0.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 0.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 0.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 0.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 0.0, 0.0,

		-0.5, -0.5, -0.5, 0.0, -1.0, 0.0,
		0.5, -0.5, -0.5, 0.0, -1.0, 0.0,
		0.5, -0.5, 0.5, 0.0, -1.0, 0.0,
		0.5, -0.5, 0.5, 0.0, -1.0, 0.0,
		-0.5, -0.5, 0.5, 0.0, -1.0, 0.0,
		-0.5, -0.5, -0.5, 0.0, -1.0, 0.0,

		-0.5, 0.5, -0.5, 0.0, 1.0, 0.0,
		0.5, 0.5, -0.5, 0.0, 1.0, 0.0,
		0.5, 0.5, 0.5, 0.0, 1.0, 0.0,
		0.5, 0.5, 0.5, 0.0, 1.0, 0.0,
		-0.5, 0.5, 0.5, 0.0, 1.0, 0.0,
		-0.5, 0.5, -0.5, 0.0, 1.0, 0.0,
	}
	lightPosition = mgl32.Vec3{1.2, 1.0, 2.0}
	// Track time stats related to frame speed to account for different
	// computer performance
	deltaTime = 0.0 // time between current frame and last frame
	lastFrame = 0.0 // time of last frame
	// Last mouse positions, initially in the center of the window
	lastX = float64(initialWindowWidth / 2)
	lastY = float64(initialWindowHeight / 2)
	// Handle when mouse first enters window and has large offset to center
	firstMouse = true
	camera     *Camera
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()

	camera = NewDefaultCameraAtPosition(mgl32.Vec3{1.0, 0.5, 4.0})
}

func main() {
	/*
	 * GLFW init and configure
	 */
	// Initialize GLFW, which is used to manage windows, user input, opengl contexts, and related
	// events.
	err := glfw.Init()
	if err != nil {
		log.Fatal(err)
	}
	// Free resources used by GLFW when the program exits.
	defer glfw.Terminate()
	// Using hints, set various options for the window we're about to create.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	// Compatibility profile allows more deprecated function calls over core profile.
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCompatProfile)

	/*
	 * GLFW window creation
	 */
	// Create the window object, the required central object to most of GLFW's functions.
	window, err := glfw.CreateWindow(initialWindowWidth, initialWindowHeight, "LearnOpenGL", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make the context relating to the just-created window the current context on the main thread.
	// A context can only be current on one thread.
	// A thread can only have one current context.
	window.MakeContextCurrent()
	// Set the function that is run every time the viewport is resized by the user.
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	// Tell glfw to capture and hide the cursor
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	// Listen to mouse events
	window.SetCursorPosCallback(mouseCallback)
	// Listen to scroll events
	window.SetScrollCallback(scrollCallback)

	if enableWireframe {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	}

	/*
	 * Load OS-specific OpenGL function pointers
	 */
	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	// Allow OpenGL to perform depth testing, where it uses the z-buffer to know when (not) to
	// draw overlapping entities
	gl.Enable(gl.DEPTH_TEST)

	/*
	 * Build and compile our shader program
	 */
	shaderProgram, err := NewShader("shaders/shader.vs", "shaders/shader.fs")
	if err != nil {
		log.Fatal(err)
	}
	lightShaderProgram, err := NewShader("shaders/shader.vs", "shaders/light.fs")
	if err != nil {
		log.Fatal(err)
	}

	/*
	 * set up vertex data (and buffer(s)) and configure vertex attributes
	 *
	 * A graphics pipeline are the stages taken to determine how to paint 2D pixels
	 * according to 3D coordinates. The stages are individual programs called shaders.
	 *
	 * The vertex shader is the first shader. It takes as input a single vertex and transforms
	 * it to other 3D coordinates. OpenGL expects a vertex and fragment shader.
	 */
	// Manage the memory in the GPU with a vertex buffer object (VBO) where many
	// vertices can be give to the GPU at once.
	var VBO uint32
	// Store vertex attributes for each object in the VBO in a vertex array object (cubeVAO)
	var cubeVAO uint32
	// Get a unique ID for buffers.
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &VBO)
	// Then bind VBO(s) and attribute pointers
	// Bind it as a ARRAY_BUFFER, allowing different typed buffers to bind.
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	// Get the vertex data into the buffer
	// STATIC_DRAW: the data is set only once and used many times.
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*int(unsafe.Sizeof(vertices[0])), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindVertexArray(cubeVAO)
	// Set the position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(6*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.EnableVertexAttribArray(0)
	// Set the normal attribute
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(6*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
	gl.EnableVertexAttribArray(1)

	// For the light cube
	var lightVAO uint32
	gl.GenVertexArrays(1, &lightVAO)
	gl.BindVertexArray(lightVAO)
	// only need to bind VBO, not refill it with data
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	// Set the position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(6*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.EnableVertexAttribArray(0)

	aspectRatio := float32(initialWindowWidth) / float32(initialWindowHeight)

	// Run the render loop until the window is closed by the user.
	for !window.ShouldClose() {
		// calculate time stats
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Handle user input.
		processInput(window)

		// Perform render logic.
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		// Clear the color and depth buffer (as opposed to the stencil buffer)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Activate shaders
		shaderProgram.use()
		shaderProgram.setVec3("light.position", lightPosition)
		shaderProgram.setVec3("viewPos", camera.position)
		shaderProgram.setVec3("material.ambient", mgl32.Vec3{1.0, 0.5, 0.31})
		shaderProgram.setVec3("material.diffuse", mgl32.Vec3{1.0, 0.5, 0.31})
		shaderProgram.setVec3("material.specular", mgl32.Vec3{0.5, 0.5, 0.5})
		shaderProgram.setFloat("material.shininess", 32.0)

		lightColor := mgl32.Vec3{
			float32(math.Sin(glfw.GetTime() * 2.0)),
			float32(math.Sin(glfw.GetTime() * 0.7)),
			float32(math.Sin(glfw.GetTime() * 1.3)),
		}
		diffuseColor := lightColor.Mul(0.5)
		ambientColor := diffuseColor.Mul(0.2)
		shaderProgram.setVec3("light.ambient", ambientColor)
		shaderProgram.setVec3("light.diffuse", diffuseColor)
		shaderProgram.setVec3("light.specular", mgl32.Vec3{1.0, 1.0, 1.0})

		// view/projection transformations
		// Create the projection matrix to add perspective to the scene
		projection := mgl32.Perspective(mgl32.DegToRad(camera.zoom), aspectRatio, 0.1, 100.0)
		view := camera.getViewMatrix()
		shaderProgram.setMat4("projection", projection)
		shaderProgram.setMat4("view", view)

		// world transformation
		model := mgl32.Ident4()
		shaderProgram.setMat4("model", model)

		// Render cube
		gl.BindVertexArray(cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)

		// Draw light cube
		lightShaderProgram.use()
		lightShaderProgram.setMat4("projection", projection)
		lightShaderProgram.setMat4("view", view)
		model = mgl32.Ident4().Mul4(mgl32.Translate3D(lightPosition.X(), lightPosition.Y(), lightPosition.Z()))
		model = model.Mul4(mgl32.Scale3D(0.2, 0.2, 0.2))
		lightShaderProgram.setMat4("model", model)
		gl.BindVertexArray(lightVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)

		// Swap the color buffer (a large 2D buffer that contains color values for each pixel in GLFW's window) that is used to render to during this render iteration and show it as output to the screen.
		window.SwapBuffers()
		// Check if events are triggered, update window state, and invoke any registered callbacks.
		glfw.PollEvents()
	}

}

// framebufferSizeCallback is called when the gl viewport is resized.
func framebufferSizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

// mouseCallback is called every time the mouse is moved. x, y are current positions of the mouse
func mouseCallback(w *glfw.Window, x float64, y float64) {
	if firstMouse {
		// prevent large visual jump
		lastX = x
		lastY = y
		firstMouse = false
	}
	// calculate mouse offset since last frame
	xOffset := x - lastX
	yOffset := lastY - y
	lastX = x
	lastY = y

	camera.processMouseMovement(float32(xOffset), float32(yOffset), true)
}

// scrollCallback is called every time the mouse scroll is moved. The offset values are how far the wheel has moved
func scrollCallback(w *glfw.Window, xOffset float64, yOffset float64) {
	camera.processMouseScroll(float32(yOffset))
}

// processInput handles key presses.
func processInput(w *glfw.Window) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}

	if w.GetKey(glfw.KeyW) == glfw.Press {
		camera.processKeyboard(FORWARD, float32(deltaTime))
	}
	if w.GetKey(glfw.KeyS) == glfw.Press {
		camera.processKeyboard(BACKWARD, float32(deltaTime))
	}
	if w.GetKey(glfw.KeyA) == glfw.Press {
		camera.processKeyboard(LEFT, float32(deltaTime))
	}
	if w.GetKey(glfw.KeyD) == glfw.Press {
		camera.processKeyboard(RIGHT, float32(deltaTime))
	}

	if w.GetKey(glfw.KeyLeftShift) == glfw.Press {
		// Tell glfw to capture and hide the cursor
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	}
	if w.GetKey(glfw.KeyRightShift) == glfw.Press {
		// Tell glfw to show and stop capturing cursor
		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
	if w.GetKey(glfw.KeySpace) == glfw.Press {
		camera = NewDefaultCameraAtPosition(mgl32.Vec3{1.0, 0.5, 4.0})
	}
}
