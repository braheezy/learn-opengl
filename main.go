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
	firstMouse = true
	camera     *Camera
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

	diffuseMap := loadTextures("assets/brickwall.jpg", false)
	normalMap := loadTextures("assets/brickwall_normal.jpg", false)

	shader.use()
	shader.setInt("diffuseMap", 0)
	shader.setInt("normalMap", 1)

	lightPos := mgl32.Vec3{0.5, 1.0, 0.3}

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

		projection := mgl32.Perspective(mgl32.DegToRad(camera.zoom), windowWidth/windowHeight, 0.1, 100.0)
		shader.use()
		shader.setMat4("projection", projection)
		shader.setMat4("view", camera.getViewMatrix())
		// render normal-mapped quad
		model := mgl32.Ident4().Mul4(mgl32.HomogRotate3D(float32(glfw.GetTime()), mgl32.Vec3{1.0, 0.0, 1.0}.Normalize()))
		shader.setVec3("viewPos", camera.position)
		shader.setVec3("lightPos", lightPos)
		shader.setMat4("model", model)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, diffuseMap)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, normalMap)
		renderQuad()

		// render light source (simply re-renders a smaller plane at the light's position for debugging/visualization)
		model = mgl32.Ident4().Mul4(mgl32.Translate3D(lightPos[0], lightPos[1], lightPos[2]))
		model = model.Mul4(mgl32.Scale3D(0.1, 0.1, 0.1))
		shader.setMat4("model", model)
		renderQuad()

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
		// positions
		pos1 := mgl32.Vec3{-1.0, 1.0, 0.0}
		pos2 := mgl32.Vec3{-1.0, -1.0, 0.0}
		pos3 := mgl32.Vec3{1.0, -1.0, 0.0}
		pos4 := mgl32.Vec3{1.0, 1.0, 0.0}
		// texture coordinates
		uv1 := mgl32.Vec2{0.0, 1.0}
		uv2 := mgl32.Vec2{0.0, 0.0}
		uv3 := mgl32.Vec2{1.0, 0.0}
		uv4 := mgl32.Vec2{1.0, 1.0}
		// normal vector
		nm := mgl32.Vec3{0.0, 0.0, 1.0}

		// calculate tangent/bitangent vectors of both triangles
		// triangle 1
		edge1 := pos2.Sub(pos1)
		edge2 := pos3.Sub(pos1)
		deltaUV1 := uv2.Sub(uv1)
		deltaUV2 := uv3.Sub(uv1)

		f := 1.0 / (deltaUV1.X()*deltaUV2.Y() - deltaUV2.X()*deltaUV1.Y())

		var tangent1 mgl32.Vec3
		tangent1[0] = f * (deltaUV2.Y()*edge1.X() - deltaUV1.Y()*edge2.X())
		tangent1[1] = f * (deltaUV2.Y()*edge1.Y() - deltaUV1.Y()*edge2.Y())
		tangent1[2] = f * (deltaUV2.Y()*edge1.Z() - deltaUV1.Y()*edge2.Z())
		var bitangent1 mgl32.Vec3
		bitangent1[0] = f * (-deltaUV2.X()*edge1.X() + deltaUV1.X()*edge2.X())
		bitangent1[1] = f * (-deltaUV2.X()*edge1.Y() + deltaUV1.X()*edge2.Y())
		bitangent1[2] = f * (-deltaUV2.X()*edge1.Z() + deltaUV1.X()*edge2.Z())

		// triangle 2
		edge1 = pos3.Sub(pos1)
		edge2 = pos4.Sub(pos1)
		deltaUV1 = uv3.Sub(uv1)
		deltaUV2 = uv4.Sub(uv1)

		f = 1.0 / (deltaUV1.X()*deltaUV2.Y() - deltaUV2.X()*deltaUV1.Y())

		var tangent2 mgl32.Vec3
		tangent2[0] = f * (deltaUV2.Y()*edge1.X() - deltaUV1.Y()*edge2.X())
		tangent2[1] = f * (deltaUV2.Y()*edge1.Y() - deltaUV1.Y()*edge2.Y())
		tangent2[2] = f * (deltaUV2.Y()*edge1.Z() - deltaUV1.Y()*edge2.Z())
		var bitangent2 mgl32.Vec3
		bitangent2[0] = f * (-deltaUV2.X()*edge1.X() + deltaUV1.X()*edge2.X())
		bitangent2[1] = f * (-deltaUV2.X()*edge1.Y() + deltaUV1.X()*edge2.Y())
		bitangent2[2] = f * (-deltaUV2.X()*edge1.Z() + deltaUV1.X()*edge2.Z())

		quadVertices := []float32{
			// positions            // normal         // texcoords  // tangent                          // bitangent
			pos1.X(), pos1.Y(), pos1.Z(), nm.X(), nm.Y(), nm.Z(), uv1.X(), uv1.Y(), tangent1.X(), tangent1.Y(), tangent1.Z(), bitangent1.X(), bitangent1.Y(), bitangent1.Z(),
			pos2.X(), pos2.Y(), pos2.Z(), nm.X(), nm.Y(), nm.Z(), uv2.X(), uv2.Y(), tangent1.X(), tangent1.Y(), tangent1.Z(), bitangent1.X(), bitangent1.Y(), bitangent1.Z(),
			pos3.X(), pos3.Y(), pos3.Z(), nm.X(), nm.Y(), nm.Z(), uv3.X(), uv3.Y(), tangent1.X(), tangent1.Y(), tangent1.Z(), bitangent1.X(), bitangent1.Y(), bitangent1.Z(),

			pos1.X(), pos1.Y(), pos1.Z(), nm.X(), nm.Y(), nm.Z(), uv1.X(), uv1.Y(), tangent2.X(), tangent2.Y(), tangent2.Z(), bitangent2.X(), bitangent2.Y(), bitangent2.Z(),
			pos3.X(), pos3.Y(), pos3.Z(), nm.X(), nm.Y(), nm.Z(), uv3.X(), uv3.Y(), tangent2.X(), tangent2.Y(), tangent2.Z(), bitangent2.X(), bitangent2.Y(), bitangent2.Z(),
			pos4.X(), pos4.Y(), pos4.Z(), nm.X(), nm.Y(), nm.Z(), uv4.X(), uv4.Y(), tangent2.X(), tangent2.Y(), tangent2.Z(), bitangent2.X(), bitangent2.Y(), bitangent2.Z(),
		}

		// configure plane VAO
		gl.GenVertexArrays(1, &quadVAO)
		gl.GenBuffers(1, &quadVBO)
		gl.BindVertexArray(quadVAO)
		gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*int(unsafe.Sizeof(quadVertices[0])), gl.Ptr(quadVertices), gl.STATIC_DRAW)
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(14*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(14*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
		gl.EnableVertexAttribArray(2)
		gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(14*unsafe.Sizeof(float32(0))), gl.Ptr(6*unsafe.Sizeof(float32(0))))
		gl.EnableVertexAttribArray(3)
		gl.VertexAttribPointer(3, 3, gl.FLOAT, false, int32(14*unsafe.Sizeof(float32(0))), gl.Ptr(8*unsafe.Sizeof(float32(0))))
		gl.EnableVertexAttribArray(4)
		gl.VertexAttribPointer(4, 3, gl.FLOAT, false, int32(14*unsafe.Sizeof(float32(0))), gl.Ptr(11*unsafe.Sizeof(float32(0))))
	}
	// render
	gl.BindVertexArray(quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}
