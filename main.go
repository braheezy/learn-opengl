package main

import (
	"fmt"
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
	windowWidth   = 800
	windowHeight  = 600
	SHADOW_WIDTH  = 1024
	SHADOW_HEIGHT = 1024
)

var (
	// Track time stats related to frame speed to account for different
	// computer performance
	deltaTime = 0.0 // time between current frame and last frame
	lastFrame = 0.0 // time of last frame
	// Last mouse positions, initially in the center of the window
	lastX = float64(windowWidth / 2)
	lastY = float64(windowHeight / 2)
	// Handle when mouse first enters window and has large offset to center
	firstMouse      = true
	camera          *Camera
	bloom           = true
	bloomKeyPressed = false
	exposure        = float32(1.0)
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()

	camera = NewDefaultCameraAtPosition(mgl32.Vec3{0.0, 0.0, 5.0})
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
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "LearnOpenGL", nil, nil)
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

	/*
	 * Build and compile our shader program
	 */
	shader, err := NewShader("shaders/shader.vs", "shaders/shader.fs", "")
	if err != nil {
		log.Fatal(err)
	}
	shaderLight, err := NewShader("shaders/shader.vs", "shaders/light_box.fs", "")
	if err != nil {
		log.Fatal(err)
	}
	shaderBlur, err := NewShader("shaders/blur.vs", "shaders/blur.fs", "")
	if err != nil {
		log.Fatal(err)
	}

	shaderBloom, err := NewShader("shaders/bloom.vs", "shaders/bloom.fs", "")
	if err != nil {
		log.Fatal(err)
	}

	woodTexture := loadTextures("assets/wood.png", true)
	containerTexture := loadTextures("assets/container2.png", true)

	// configure floating point framebuffer
	// ------------------------------------
	var hdrFBO uint32
	gl.GenFramebuffers(1, &hdrFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, hdrFBO)
	// create floating point color buffer
	colorBuffers := make([]uint32, 2)
	gl.GenTextures(2, &colorBuffers[0])
	for i := 0; i < 2; i++ {
		gl.BindTexture(gl.TEXTURE_2D, colorBuffers[i])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16, windowWidth, windowHeight, 0, gl.RGBA, gl.FLOAT, gl.Ptr(nil))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0+uint32(i), gl.TEXTURE_2D, colorBuffers[i], 0)
	}
	// create depth buffer (renderbuffer)
	var rboDepth uint32
	gl.GenRenderbuffers(1, &rboDepth)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rboDepth)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, windowWidth, windowHeight)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rboDepth)
	// tell OpenGL which color attachments we'll use (of this framebuffer) for rendering
	attachments := []uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1}
	gl.DrawBuffers(2, &attachments[0])
	// attach buffers
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		log.Fatal("Framebuffer not complete!")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// ping-pong-framebuffer for blurring
	pingpongFBO := make([]uint32, 2)
	pingpongColorbuffers := make([]uint32, 2)
	gl.GenFramebuffers(2, &pingpongFBO[0])
	gl.GenTextures(2, &pingpongColorbuffers[0])
	for i := 0; i < 2; i++ {
		gl.BindFramebuffer(gl.FRAMEBUFFER, pingpongFBO[i])
		gl.BindTexture(gl.TEXTURE_2D, pingpongColorbuffers[i])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16, windowWidth, windowHeight, 0, gl.RGBA, gl.FLOAT, gl.Ptr(nil))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // we clamp to the edge as the blur filter would otherwise sample repeated texture values!
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, pingpongColorbuffers[i], 0)
		// also check if framebuffers are complete (no need for depth buffer)
		if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
			log.Fatal("pingpong framebuffer not complete!")
		}
	}
	// lighting info
	// -------------
	// positions
	lightPositions := []mgl32.Vec3{
		{0.0, 0.0, 1.5},
		{-4.0, 0.5, -3.0},
		{3.0, 0.5, 1.0},
		{-0.8, 2.4, -1.0},
	}
	// colors
	lightColors := []mgl32.Vec3{
		{5.0, 5.0, 5.0},
		{10.0, 0.0, 0.0},
		{0.0, 0.0, 15.0},
		{0.0, 5.0, 0.0},
	}

	// shader configuration
	// --------------------
	shader.use()
	shader.setInt("diffuseTexture", 0)
	shaderBlur.use()
	shaderBlur.setInt("image", 0)
	shaderBloom.use()
	shaderBloom.setInt("scene", 0)
	shaderBloom.setInt("bloomBlur", 1)

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

		// 1. render scene into floating point framebuffer
		// -----------------------------------------------
		gl.BindFramebuffer(gl.FRAMEBUFFER, hdrFBO)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		projection := mgl32.Perspective(mgl32.DegToRad(camera.zoom), windowWidth/windowHeight, 0.1, 100.0)
		view := camera.getViewMatrix()
		shader.use()
		shader.setMat4("projection", projection)
		shader.setMat4("view", view)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, woodTexture)
		// set lighting uniforms
		for i, lightPosition := range lightPositions {
			shader.setVec3(fmt.Sprintf("lights[%d].Position", i), lightPosition)
			shader.setVec3(fmt.Sprintf("lights[%d].Color", i), lightColors[i])
		}
		shader.setVec3("viewPos", camera.position)
		// create one large cube that acts as the floor
		model := mgl32.Ident4().Mul4(mgl32.Translate3D(0.0, -1.0, 0.0))
		model = model.Mul4(mgl32.Scale3D(12.5, 0.5, 12.5))
		shader.setMat4("model", model)
		renderCube()
		// then create multiple cubes as the scenery
		gl.BindTexture(gl.TEXTURE_2D, containerTexture)
		model = mgl32.Ident4().Mul4(mgl32.Translate3D(0.0, 1.5, 0.0))
		model = model.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))
		shader.setMat4("model", model)
		renderCube()

		model = mgl32.Ident4().Mul4(mgl32.Translate3D(2.0, 0.0, 1.0))
		model = model.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))
		shader.setMat4("model", model)
		renderCube()

		model = mgl32.Ident4().Mul4(mgl32.Translate3D(-1.0, -1.0, 2.0))
		model = model.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(60.0), mgl32.Vec3{1.0, 0.0, 1.0}.Normalize()))
		shader.setMat4("model", model)
		renderCube()

		model = mgl32.Ident4().Mul4(mgl32.Translate3D(0.0, 2.7, 4.0))
		model = model.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(23.0), mgl32.Vec3{1.0, 0.0, 1.0}.Normalize()))
		model = model.Mul4(mgl32.Scale3D(1.25, 1.25, 1.25))
		shader.setMat4("model", model)
		renderCube()

		model = mgl32.Ident4().Mul4(mgl32.Translate3D(-2.0, 1.0, -3.0))
		model = model.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(124.0), mgl32.Vec3{1.0, 0.0, 1.0}.Normalize()))
		shader.setMat4("model", model)
		renderCube()

		model = mgl32.Ident4().Mul4(mgl32.Translate3D(-3.0, 0.0, 0.0))
		model = model.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))
		shader.setMat4("model", model)
		renderCube()

		// finally show all the light sources as bright cubes
		shaderLight.use()
		shaderLight.setMat4("projection", projection)
		shaderLight.setMat4("view", view)

		for i := 0; i < len(lightPositions); i++ {
			model = mgl32.Ident4().Mul4(mgl32.Translate3D(lightPositions[i].X(), lightPositions[i].Y(), lightPositions[i].Z()))
			model = model.Mul4(mgl32.Scale3D(0.25, 0.25, 0.25))
			shaderLight.setMat4("model", model)
			shaderLight.setVec3("lightColor", lightColors[i])
			renderCube()
		}
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		// 2. blur bright fragments with two-pass Gaussian Blur
		// --------------------------------------------------
		horizontal := true
		first_iteration := true
		amount := 10
		shaderBlur.use()
		for i := 0; i < amount; i++ {
			if horizontal {
				gl.BindFramebuffer(gl.FRAMEBUFFER, pingpongFBO[1])
			} else {
				gl.BindFramebuffer(gl.FRAMEBUFFER, pingpongFBO[0])
			}
			shaderBlur.setBool("horizontal", horizontal)
			if first_iteration {
				gl.BindTexture(gl.TEXTURE_2D, colorBuffers[1])
			} else {
				if horizontal {
					gl.BindTexture(gl.TEXTURE_2D, pingpongColorbuffers[0])
				} else {
					gl.BindTexture(gl.TEXTURE_2D, pingpongColorbuffers[1])
				}
			}
			renderQuad()
			horizontal = !horizontal
			if first_iteration {
				first_iteration = false
			}
		}
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		// 3. now render floating point color buffer to 2D quad and tonemap HDR colors to default framebuffer's (clamped) color range
		// --------------------------------------------------------------------------------------------------------------------------
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		shaderBloom.use()
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, colorBuffers[0])
		gl.ActiveTexture(gl.TEXTURE1)
		if horizontal {
			gl.BindTexture(gl.TEXTURE_2D, pingpongColorbuffers[0])
		} else {
			gl.BindTexture(gl.TEXTURE_2D, pingpongColorbuffers[1])
		}
		shaderBloom.setBool("bloom", bloom)
		shaderBloom.setFloat("exposure", exposure)
		renderQuad()

		bloomText := "off"
		if bloom {
			bloomText = "on"
		}
		fmt.Printf("\rbloom: %v | exposure: %v", bloomText, exposure)

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
	if w.GetKey(glfw.KeyBackspace) == glfw.Press {
		// reset view
		camera = NewDefaultCameraAtPosition(mgl32.Vec3{1.0, 0.5, 4.0})
	}

	if w.GetKey(glfw.KeySpace) == glfw.Press && !bloomKeyPressed {
		bloom = !bloom
		bloomKeyPressed = true
	}
	if w.GetKey(glfw.KeySpace) == glfw.Release {
		bloomKeyPressed = false
	}

	if w.GetKey(glfw.KeyQ) == glfw.Press {
		if exposure > 0.0 {
			exposure -= 0.01
		} else {
			exposure = 0.0
		}
	} else if w.GetKey(glfw.KeyE) == glfw.Press {
		exposure += 0.01
	}
}
func loadTextures(filePath string, gammaCorrection bool) (texture uint32) {
	// Load the texture data
	pixelData, format, width, height := loadPixels(filePath)

	var internalFormat int32
	switch format {
	case gl.RED:
		internalFormat = gl.RED
	case gl.RGB:
		if gammaCorrection {
			internalFormat = gl.SRGB
		} else {
			internalFormat = gl.RGB
		}
	case gl.RGBA:
		if gammaCorrection {
			internalFormat = gl.SRGB_ALPHA
		} else {
			internalFormat = gl.RGBA
		}
	}

	// Generate texture ID and bind it
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	// Specify texture parameters
	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat, int32(width), int32(height), 0, uint32(format), gl.UNSIGNED_BYTE, gl.Ptr(pixelData))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return texture
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

func renderScene(shader *Shader) {
	// room cube
	shader.setMat4("model", mgl32.Ident4().Mul4(mgl32.Scale3D(5.0, 5.0, 5.0)))
	gl.Disable(gl.CULL_FACE)
	shader.setInt("reverse_normals", 1)
	renderCube()
	shader.setInt("reverse_normals", 0)
	gl.Enable(gl.CULL_FACE)
	// cubes
	model := mgl32.Ident4().Mul4(mgl32.Translate3D(4.0, -3.5, 0.0))
	model = model.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))
	shader.setMat4("model", model)
	renderCube()
	model = mgl32.Ident4().Mul4(mgl32.Translate3D(2.0, 3.0, 1.0))
	model = model.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75))
	shader.setMat4("model", model)
	renderCube()
	model = mgl32.Ident4().Mul4(mgl32.Translate3D(-3.0, -1.0, 0.0))
	model = model.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))
	shader.setMat4("model", model)
	renderCube()
	model = mgl32.Ident4().Mul4(mgl32.Translate3D(-1.5, 1.0, 1.5))
	model = model.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))
	shader.setMat4("model", model)
	renderCube()
	model = mgl32.Ident4().Mul4(mgl32.Translate3D(-1.5, 2.0, -3.0))
	model = model.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(60.0), mgl32.Vec3{1.0, 0.0, 1.0}.Normalize()))
	model = model.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75))
	shader.setMat4("model", model)
	renderCube()

}

// renderCube() renders a 1x1 3D cube in NDC.
var cubeVAO, cubeVBO uint32

func renderCube() {
	if cubeVAO == 0 {
		// initialize
		vertices := []float32{
			// back face
			-1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 0.0, // bottom-left
			1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 1.0, // top-right
			1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 0.0, // bottom-right
			1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 1.0, // top-right
			-1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 0.0, // bottom-left
			-1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 1.0, // top-left
			// front face
			-1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom-left
			1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 0.0, // bottom-right
			1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 1.0, // top-right
			1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 1.0, // top-right
			-1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 1.0, // top-left
			-1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom-left
			// left face
			-1.0, 1.0, 1.0, -1.0, 0.0, 0.0, 1.0, 0.0, // top-right
			-1.0, 1.0, -1.0, -1.0, 0.0, 0.0, 1.0, 1.0, // top-left
			-1.0, -1.0, -1.0, -1.0, 0.0, 0.0, 0.0, 1.0, // bottom-left
			-1.0, -1.0, -1.0, -1.0, 0.0, 0.0, 0.0, 1.0, // bottom-left
			-1.0, -1.0, 1.0, -1.0, 0.0, 0.0, 0.0, 0.0, // bottom-right
			-1.0, 1.0, 1.0, -1.0, 0.0, 0.0, 1.0, 0.0, // top-right
			// right face
			1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, // top-left
			1.0, -1.0, -1.0, 1.0, 0.0, 0.0, 0.0, 1.0, // bottom-right
			1.0, 1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 1.0, // top-right
			1.0, -1.0, -1.0, 1.0, 0.0, 0.0, 0.0, 1.0, // bottom-right
			1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, // top-left
			1.0, -1.0, 1.0, 1.0, 0.0, 0.0, 0.0, 0.0, // bottom-left
			// bottom face
			-1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 0.0, 1.0, // top-right
			1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 1.0, 1.0, // top-left
			1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, // bottom-left
			1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, // bottom-left
			-1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 0.0, 0.0, // bottom-right
			-1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 0.0, 1.0, // top-right
			// top face
			-1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 0.0, 1.0, // top-left
			1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom-right
			1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 1.0, 1.0, // top-right
			1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom-right
			-1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 0.0, 1.0, // top-left
			-1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 0.0, // bottom-left
		}
		gl.GenVertexArrays(1, &cubeVAO)
		gl.GenBuffers(1, &cubeVBO)
		// fill buffer
		gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*int(unsafe.Sizeof(vertices[0])), gl.Ptr(vertices), gl.STATIC_DRAW)
		// link vertex attributes
		gl.BindVertexArray(cubeVAO)
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*int32(unsafe.Sizeof(float32(0))), gl.Ptr(nil))
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(8*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
		gl.EnableVertexAttribArray(2)
		gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(8*unsafe.Sizeof(float32(0))), gl.Ptr(6*unsafe.Sizeof(float32(0))))
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
	}
	// render cube
	gl.BindVertexArray(cubeVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 36)
	gl.BindVertexArray(0)
}

// renders a 1x1 quad in NDC with manually calculated tangent vectors
var quadVAO, quadVBO uint32

func renderQuad() {
	if quadVAO == 0 {
		// // positions
		// pos1 := mgl32.Vec3{-1.0, 1.0, 0.0}
		// pos2 := mgl32.Vec3{-1.0, -1.0, 0.0}
		// pos3 := mgl32.Vec3{1.0, -1.0, 0.0}
		// pos4 := mgl32.Vec3{1.0, 1.0, 0.0}
		// // texture coordinates
		// uv1 := mgl32.Vec2{0.0, 1.0}
		// uv2 := mgl32.Vec2{0.0, 0.0}
		// uv3 := mgl32.Vec2{1.0, 0.0}
		// uv4 := mgl32.Vec2{1.0, 1.0}
		// // normal vector
		// nm := mgl32.Vec3{0.0, 0.0, 1.0}

		// // calculate tangent/bitangent vectors of both triangles
		// // triangle 1
		// edge1 := pos2.Sub(pos1)
		// edge2 := pos3.Sub(pos1)
		// deltaUV1 := uv2.Sub(uv1)
		// deltaUV2 := uv3.Sub(uv1)

		// f := 1.0 / (deltaUV1.X()*deltaUV2.Y() - deltaUV2.X()*deltaUV1.Y())

		// var tangent1 mgl32.Vec3
		// tangent1[0] = f * (deltaUV2.Y()*edge1.X() - deltaUV1.Y()*edge2.X())
		// tangent1[1] = f * (deltaUV2.Y()*edge1.Y() - deltaUV1.Y()*edge2.Y())
		// tangent1[2] = f * (deltaUV2.Y()*edge1.Z() - deltaUV1.Y()*edge2.Z())
		// var bitangent1 mgl32.Vec3
		// bitangent1[0] = f * (-deltaUV2.X()*edge1.X() + deltaUV1.X()*edge2.X())
		// bitangent1[1] = f * (-deltaUV2.X()*edge1.Y() + deltaUV1.X()*edge2.Y())
		// bitangent1[2] = f * (-deltaUV2.X()*edge1.Z() + deltaUV1.X()*edge2.Z())

		// // triangle 2
		// edge1 = pos3.Sub(pos1)
		// edge2 = pos4.Sub(pos1)
		// deltaUV1 = uv3.Sub(uv1)
		// deltaUV2 = uv4.Sub(uv1)

		// f = 1.0 / (deltaUV1.X()*deltaUV2.Y() - deltaUV2.X()*deltaUV1.Y())

		// var tangent2 mgl32.Vec3
		// tangent2[0] = f * (deltaUV2.Y()*edge1.X() - deltaUV1.Y()*edge2.X())
		// tangent2[1] = f * (deltaUV2.Y()*edge1.Y() - deltaUV1.Y()*edge2.Y())
		// tangent2[2] = f * (deltaUV2.Y()*edge1.Z() - deltaUV1.Y()*edge2.Z())
		// var bitangent2 mgl32.Vec3
		// bitangent2[0] = f * (-deltaUV2.X()*edge1.X() + deltaUV1.X()*edge2.X())
		// bitangent2[1] = f * (-deltaUV2.X()*edge1.Y() + deltaUV1.X()*edge2.Y())
		// bitangent2[2] = f * (-deltaUV2.X()*edge1.Z() + deltaUV1.X()*edge2.Z())

		// quadVertices := []float32{
		// 	// positions            // normal         // texcoords  // tangent                          // bitangent
		// 	pos1.X(), pos1.Y(), pos1.Z(), nm.X(), nm.Y(), nm.Z(), uv1.X(), uv1.Y(), tangent1.X(), tangent1.Y(), tangent1.Z(), bitangent1.X(), bitangent1.Y(), bitangent1.Z(),
		// 	pos2.X(), pos2.Y(), pos2.Z(), nm.X(), nm.Y(), nm.Z(), uv2.X(), uv2.Y(), tangent1.X(), tangent1.Y(), tangent1.Z(), bitangent1.X(), bitangent1.Y(), bitangent1.Z(),
		// 	pos3.X(), pos3.Y(), pos3.Z(), nm.X(), nm.Y(), nm.Z(), uv3.X(), uv3.Y(), tangent1.X(), tangent1.Y(), tangent1.Z(), bitangent1.X(), bitangent1.Y(), bitangent1.Z(),

		// 	pos1.X(), pos1.Y(), pos1.Z(), nm.X(), nm.Y(), nm.Z(), uv1.X(), uv1.Y(), tangent2.X(), tangent2.Y(), tangent2.Z(), bitangent2.X(), bitangent2.Y(), bitangent2.Z(),
		// 	pos3.X(), pos3.Y(), pos3.Z(), nm.X(), nm.Y(), nm.Z(), uv3.X(), uv3.Y(), tangent2.X(), tangent2.Y(), tangent2.Z(), bitangent2.X(), bitangent2.Y(), bitangent2.Z(),
		// 	pos4.X(), pos4.Y(), pos4.Z(), nm.X(), nm.Y(), nm.Z(), uv4.X(), uv4.Y(), tangent2.X(), tangent2.Y(), tangent2.Z(), bitangent2.X(), bitangent2.Y(), bitangent2.Z(),
		// }

		quadVertices := []float32{
			// positions        // texture Coords
			-1.0, 1.0, 0.0, 0.0, 1.0,
			-1.0, -1.0, 0.0, 0.0, 0.0,
			1.0, 1.0, 0.0, 1.0, 1.0,
			1.0, -1.0, 0.0, 1.0, 0.0,
		}

		// configure plane VAO
		gl.GenVertexArrays(1, &quadVAO)
		gl.GenBuffers(1, &quadVBO)
		gl.BindVertexArray(quadVAO)
		gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*int(unsafe.Sizeof(quadVertices[0])), gl.Ptr(quadVertices), gl.STATIC_DRAW)
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(5*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, int32(5*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
	}
	// render
	gl.BindVertexArray(quadVAO)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindVertexArray(0)
}
