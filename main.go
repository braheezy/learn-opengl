package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
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
	drawModel           = true
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
	faces = []string{
		"assets/skybox/right.jpg",
		"assets/skybox/left.jpg",
		"assets/skybox/top.jpg",
		"assets/skybox/bottom.jpg",
		"assets/skybox/front.jpg",
		"assets/skybox/back.jpg",
	}
	skyboxVertices = []float32{
		// positions
		-1.0, 1.0, -1.0,
		-1.0, -1.0, -1.0,
		1.0, -1.0, -1.0,
		1.0, -1.0, -1.0,
		1.0, 1.0, -1.0,
		-1.0, 1.0, -1.0,

		-1.0, -1.0, 1.0,
		-1.0, -1.0, -1.0,
		-1.0, 1.0, -1.0,
		-1.0, 1.0, -1.0,
		-1.0, 1.0, 1.0,
		-1.0, -1.0, 1.0,

		1.0, -1.0, -1.0,
		1.0, -1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, -1.0,
		1.0, -1.0, -1.0,

		-1.0, -1.0, 1.0,
		-1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, -1.0, 1.0,
		-1.0, -1.0, 1.0,

		-1.0, 1.0, -1.0,
		1.0, 1.0, -1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		-1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,

		-1.0, -1.0, -1.0,
		-1.0, -1.0, 1.0,
		1.0, -1.0, -1.0,
		1.0, -1.0, -1.0,
		-1.0, -1.0, 1.0,
		1.0, -1.0, 1.0,
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
	// window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
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

	/*
	 * Build and compile our shader program
	 */
	shaderProgram, err := NewShader("shaders/shader.vs", "shaders/shader.fs")
	if err != nil {
		log.Fatal(err)
	}
	skyboxShader, err := NewShader("shaders/skybox.vs", "shaders/skybox.fs")
	if err != nil {
		log.Fatal(err)
	}

	/*
	 * load models
	 */
	customModel := Model{}
	if drawModel {
		customModel.LoadModel("assets/backpack/backpack.obj")
	}

	// cube VAO
	var cubeVAO, cubeVBO uint32
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &cubeVBO)
	gl.BindVertexArray(cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVertices)*int(unsafe.Sizeof(cubeVertices[0])), gl.Ptr(cubeVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(6*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(6*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
	// skybox VAO
	var skyboxVAO, skyboxVBO uint32
	gl.GenVertexArrays(1, &skyboxVAO)
	gl.GenBuffers(1, &skyboxVBO)
	gl.BindVertexArray(skyboxVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, skyboxVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(skyboxVertices)*int(unsafe.Sizeof(skyboxVertices[0])), gl.Ptr(skyboxVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(3*unsafe.Sizeof(float32(0))), gl.Ptr(nil))

	// load textures
	cubeTexture := loadTextures("assets/container.jpg")
	cubemapTexture := loadCubemap(faces)

	// shader configuration
	shaderProgram.use()
	shaderProgram.setInt("texture1", 0)

	skyboxShader.use()
	skyboxShader.setInt("skybox", 0)

	// Run the render loop until the window is closed by the user.
	for !window.ShouldClose() {
		// calculate time stats
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Handle user input.
		processInput(window)

		// render
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		shaderProgram.use()
		model := mgl32.Ident4()
		view := camera.getViewMatrix()
		projection := mgl32.Perspective(mgl32.DegToRad(camera.zoom), initialWindowWidth/initialWindowHeight, 0.1, 100.0)
		shaderProgram.setMat4("view", view)
		shaderProgram.setMat4("projection", projection)
		shaderProgram.setMat4("model", model)
		shaderProgram.setVec3("cameraPos", camera.position)

		if drawModel {
			model = model.Mul4(mgl32.Translate3D(0.0, 0.0, 0.0))
			model = model.Mul4(mgl32.Scale3D(1.0, 1.0, 1.0))
			shaderProgram.setMat4("model", model)
			customModel.Draw(*shaderProgram)
		} else {
			// cubes
			gl.BindVertexArray(cubeVAO)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, cubeTexture)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
			gl.BindVertexArray(0)
		}

		// skybox last for efficiency
		gl.DepthFunc(gl.LEQUAL)
		skyboxShader.use()
		view = extract3x3From4x4(camera.getViewMatrix())
		skyboxShader.setMat4("view", view)
		skyboxShader.setMat4("projection", projection)
		// skybox cube
		gl.BindVertexArray(skyboxVAO)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, cubemapTexture)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		gl.BindVertexArray(0)
		gl.DepthFunc(gl.LESS)

		// Swap the color buffer and poll events
		window.SwapBuffers()
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
	// Generate and bind a new texture ID
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	pixelData, format, width, height := loadPixels(filePath)

	// Upload texture data to the GPU
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, int32(width), int32(height), 0, uint32(format), gl.UNSIGNED_BYTE, gl.Ptr(pixelData))

	// Generate mipmaps
	gl.GenerateMipmap(gl.TEXTURE_2D)

	// Set texture wrapping/filtering options
	// if format == gl.RGBA {
	// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	// } else {
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	// }
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return texture
}

func loadCubemap(faces []string) uint32 {
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, textureID)

	for i, face := range faces {
		pixelData, format, width, height := loadPixels(face)
		gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), 0, format, int32(width), int32(height), 0, uint32(format), gl.UNSIGNED_BYTE, gl.Ptr(pixelData))
	}

	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

	return textureID
}

func loadPixels(filePath string) ([]byte, int32, int, int) {

	// Open and decode the texture image file
	textureFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open texture file: %v", err)
	}
	defer textureFile.Close()

	textureImage, _, err := image.Decode(textureFile)
	if err != nil {
		log.Fatalf("Failed to decode texture file [%s]: %v", filePath, err)
	}

	// Determine the number of components and the OpenGL format
	bounds := textureImage.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	var format int32
	var pixelData []byte

	switch img := textureImage.(type) {
	case *image.Gray:
		format = gl.RED
		pixelData = make([]byte, width*height)
		index := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r := img.GrayAt(x, y).Y
				pixelData[index] = r
				index++
			}
		}
	case *image.RGBA:
		// Check if the alpha channel is used
		hasAlpha := false
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, a := img.At(x, y).RGBA()
				if a != 0xFFFF {
					hasAlpha = true
					break
				}
			}
			if hasAlpha {
				break
			}
		}

		if hasAlpha {
			format = gl.RGBA
			pixelData = make([]byte, width*height*4)
			index := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, a := img.At(x, y).RGBA()
					pixelData[index] = byte(r >> 8)
					pixelData[index+1] = byte(g >> 8)
					pixelData[index+2] = byte(b >> 8)
					pixelData[index+3] = byte(a >> 8)
					index += 4
				}
			}
		} else {
			// Treat as RGB if alpha is not used
			format = gl.RGB
			pixelData = make([]byte, width*height*3)
			index := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					pixelData[index] = byte(r >> 8)
					pixelData[index+1] = byte(g >> 8)
					pixelData[index+2] = byte(b >> 8)
					index += 3
				}
			}
		}
	case *image.NRGBA:
		format = gl.RGBA
		pixelData = make([]byte, width*height*4)
		index := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				pixelData[index] = byte(r >> 8)
				pixelData[index+1] = byte(g >> 8)
				pixelData[index+2] = byte(b >> 8)
				pixelData[index+3] = byte(a >> 8)
				index += 4
			}
		}
	default:
		// Handle RGB images without alpha channel
		format = gl.RGB
		pixelData = make([]byte, width*height*3)
		index := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := textureImage.At(x, y).RGBA()
				pixelData[index] = byte(r >> 8)
				pixelData[index+1] = byte(g >> 8)
				pixelData[index+2] = byte(b >> 8)
				index += 3
			}
		}
	}

	return pixelData, format, width, height
}

func extract3x3From4x4(mat mgl32.Mat4) mgl32.Mat4 {
	return mgl32.Mat4{
		mat[0], mat[1], mat[2], 0, // First column
		mat[4], mat[5], mat[6], 0, // Second column
		mat[8], mat[9], mat[10], 0, // Third column
		0, 0, 0, // Fourth column
	}
}
