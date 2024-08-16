package main

import (
	"image/png"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Settings
const (
	initialWindowWidth  = 800
	initialWindowHeight = 600
)

var (
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

	cubeVertices = []float32{
		// positions          // texture Coords
		-0.5, -0.5, -0.5, 0.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, 0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 0.0,

		-0.5, -0.5, 0.5, 0.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		-0.5, 0.5, 0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,

		-0.5, 0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,
		-0.5, 0.5, 0.5, 1.0, 0.0,

		0.5, 0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, 0.5, 0.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 0.0,

		-0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,

		-0.5, 0.5, -0.5, 0.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, 0.5, 0.0, 0.0,
		-0.5, 0.5, -0.5, 0.0, 1.0,
	}
	planeVertices = []float32{
		// positions          // texture Coords (note we set these higher than 1 (together with GL_REPEAT as texture wrapping mode). this will cause the floor texture to repeat)
		5.0, -0.5, 5.0, 2.0, 0.0,
		-5.0, -0.5, 5.0, 0.0, 0.0,
		-5.0, -0.5, -5.0, 0.0, 2.0,

		5.0, -0.5, 5.0, 2.0, 0.0,
		-5.0, -0.5, -5.0, 0.0, 2.0,
		5.0, -0.5, -5.0, 2.0, 2.0,
	}
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()

	camera = NewDefaultCameraAtPosition(mgl32.Vec3{0.0, 0.0, 3.0})
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
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	// Compatibility profile allows more deprecated function calls over core profile.
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

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
	// Listen to mouse events
	window.SetCursorPosCallback(mouseCallback)
	// Tell glfw to capture and hide the cursor
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	// Listen to scroll events
	window.SetScrollCallback(scrollCallback)

	/*
	 * Load OS-specific OpenGL function pointers
	 */
	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	// Allow OpenGL to perform depth testing, where it uses the z-buffer to know when (not) to
	// draw overlapping entities
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	// Enable stencil
	gl.Enable(gl.STENCIL_TEST)
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.REPLACE)
	gl.StencilFunc(gl.NOTEQUAL, 1, 0xFF)

	/*
	 * Build and compile our shader program
	 */
	shaderProgram, err := NewShader("shaders/shader.vs", "shaders/shader.fs")
	if err != nil {
		log.Fatal(err)
	}
	singleColorShader, err := NewShader("shaders/single_color.vs", "shaders/single_color.fs")
	if err != nil {
		log.Fatal(err)
	}

	// cube VAO
	var cubeVAO, cubeVBO uint32
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &cubeVBO)
	gl.BindVertexArray(cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVertices)*int(unsafe.Sizeof(cubeVertices[0])), gl.Ptr(cubeVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(5*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, int32(5*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
	gl.BindVertexArray(0)
	// plane VAO
	var planeVAO, planeVBO uint32
	gl.GenVertexArrays(1, &planeVAO)
	gl.GenBuffers(1, &planeVBO)
	gl.BindVertexArray(planeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, planeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(planeVertices)*int(unsafe.Sizeof(planeVertices[0])), gl.Ptr(planeVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(5*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, int32(5*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
	gl.BindVertexArray(0)

	// load textures
	cubeTexture := loadTextures("assets/marble.png")
	floorTexture := loadTextures("assets/metal.png")

	// shader configuration
	shaderProgram.use()
	shaderProgram.setInt("texture1", 0)

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
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)

		// Activate shaders
		singleColorShader.use()

		model := mgl32.Ident4()
		view := camera.getViewMatrix()
		projection := mgl32.Perspective(mgl32.DegToRad(camera.zoom), float32(initialWindowWidth)/float32(initialWindowHeight), 0.1, 100.0)
		singleColorShader.setMat4("view", view)
		singleColorShader.setMat4("projection", projection)

		shaderProgram.use()
		shaderProgram.setMat4("view", view)
		shaderProgram.setMat4("projection", projection)

		// floor
		gl.StencilMask(0x00)
		gl.BindVertexArray(planeVAO)
		gl.BindTexture(gl.TEXTURE_2D, floorTexture)
		shaderProgram.setMat4("model", mgl32.Ident4())
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		gl.BindVertexArray(0)

		// cubes: 1st render, draw objects normal, write to stencil buffer
		gl.StencilFunc(gl.ALWAYS, 1, 0xFF)
		gl.StencilMask(0xFF)
		gl.BindVertexArray(cubeVAO)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, cubeTexture)
		model = model.Mul4(mgl32.Translate3D(-1.0, 0.0, -1.0))
		shaderProgram.setMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		model = mgl32.Ident4()
		model = model.Mul4(mgl32.Translate3D(2.0, 0.0, 0.0))
		shaderProgram.setMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		// cubes: 2nd render: draw slightly scaled versions, this time disabling stencil writing. b/c the stencil
		// buffer is now fillled with several 1s, the parts of the buffer that are 1 are not drawn, thus only drawing
		// the object size differences, making it look like a border
		gl.StencilFunc(gl.NOTEQUAL, 1, 0xFF)
		gl.StencilMask(0x00)
		gl.Disable(gl.DEPTH_TEST)
		singleColorShader.use()
		scale := float32(1.1)
		gl.BindVertexArray(cubeVAO)
		gl.BindTexture(gl.TEXTURE_2D, cubeTexture)
		model = mgl32.Ident4()
		model = model.Mul4(mgl32.Translate3D(-1.0, 0.0, -1.0))
		model = model.Mul4(mgl32.Scale3D(scale, scale, scale))
		shaderProgram.setMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		model = mgl32.Ident4()
		model = model.Mul4(mgl32.Translate3D(2.0, 0.0, 0.0))
		model = model.Mul4(mgl32.Scale3D(scale, scale, scale))
		shaderProgram.setMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		gl.BindVertexArray(0)
		gl.StencilMask(0xFF)
		gl.StencilFunc(gl.ALWAYS, 0, 0xFF)
		gl.Enable(gl.DEPTH_TEST)
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

func loadTextures(filePath string) uint32 {
	/*
	 * load and create a texture
	 */
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	// Set texture options
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	textureFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	textureImage, err := png.Decode(textureFile)
	if err != nil {
		log.Fatal(err)
	}
	textureFile.Close()
	bounds := textureImage.Bounds()
	// Create a slice to hold the pixel data
	pixelData := make([]byte, bounds.Dx()*bounds.Dy()*4)

	// Convert the image to RGB and copy the pixel data to the slice
	index := 0
	for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := textureImage.At(x, y).RGBA()
			pixelData[index] = byte(r >> 8)   // Red
			pixelData[index+1] = byte(g >> 8) // Green
			pixelData[index+2] = byte(b >> 8) // Blue
			pixelData[index+3] = byte(b >> 8) // Alpha
			index += 4
		}
	}

	// Generate texture from image data
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(textureImage.Bounds().Dx()), int32(textureImage.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixelData))
	// Generate mipmap, handling various resolutions of the texture at different distances.
	gl.GenerateMipmap(gl.TEXTURE_2D)

	return texture
}
