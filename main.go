package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"maps"
	"math"
	"math/rand"
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
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()

	camera = NewDefaultCameraAtPosition(mgl32.Vec3{0.0, 0.0, 155.0})
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

	/*
	 * Build and compile our shader program
	 */
	asteroidShader, err := NewShader("shaders/asteroids.vs", "shaders/asteroids.fs", "")
	if err != nil {
		log.Fatal(err)
	}
	planetShader, err := NewShader("shaders/planet.vs", "shaders/planet.fs", "")
	if err != nil {
		log.Fatal(err)
	}

	/*
	 * load models
	 */
	// if drawModel {
	rock := LoadModel("assets/rock")
	planet := LoadModel("assets/planet")
	// }

	// generate a large list of semi-random model transformation matrices
	amount := 100000
	modelMatrices := make([]mgl32.Mat4, amount)
	radius := 150.0
	offset := 25.0
	for i := 0; i < amount; i++ {
		model := mgl32.Ident4()
		// 1. translation: displace along circle with 'radius' in range [-offset, offset]
		angle := float64(i) / float64(amount) * 360.0
		displacement := (rand.Int()%int(2*offset*100))/100.0 - int(offset)
		x := math.Sin(angle)*radius + float64(displacement)
		// keep height of asteroid field smaller compared to width of x and z
		displacement = (rand.Int()%int(2*offset*100))/100.0 - int(offset)
		y := float64(displacement) * 0.4
		displacement = (rand.Int()%int(2*offset*100))/100.0 - int(offset)
		z := math.Cos(angle)*radius + float64(displacement)
		model = model.Mul4(mgl32.Translate3D(float32(x), float32(y), float32(z)))

		// 2. scale: Scale between 0.05 and 0.25f
		scale := float32(rand.Int()%20)/100.0 + 0.05
		model = model.Mul4(mgl32.Scale3D(scale, scale, scale))

		// 3. rotation: add random rotation around a (semi)randomly picked rotation axis vector
		rotationAngle := rand.Int() % 360
		model = model.Mul4(mgl32.HomogRotate3D(float32(rotationAngle), mgl32.Vec3{0.4, 0.6, 0.8}))

		modelMatrices[i] = model
	}

	// configure instanced array
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.BufferData(gl.ARRAY_BUFFER, amount*int(unsafe.Sizeof(mgl32.Mat4{})), gl.Ptr(modelMatrices), gl.STATIC_DRAW)

	// set transformation matrices as an instance vertex attribute (with divisor 1)
	// note: we're cheating a little by taking the, now publicly declared, VAO of the model's mesh(es) and adding new vertexAttribPointers
	for _, mesh := range rock.meshes {
		VAO := mesh.VAO
		gl.BindVertexArray(VAO)
		gl.EnableVertexAttribArray(3)
		gl.VertexAttribPointer(3, 4, gl.FLOAT, false, int32(unsafe.Sizeof(mgl32.Mat4{})), gl.Ptr(nil))
		gl.EnableVertexAttribArray(4)
		gl.VertexAttribPointer(4, 4, gl.FLOAT, false, int32(unsafe.Sizeof(mgl32.Mat4{})), gl.Ptr(unsafe.Sizeof(mgl32.Vec4{})))
		gl.EnableVertexAttribArray(5)
		gl.VertexAttribPointer(5, 4, gl.FLOAT, false, int32(unsafe.Sizeof(mgl32.Mat4{})), gl.Ptr(2*unsafe.Sizeof(mgl32.Vec4{})))
		gl.EnableVertexAttribArray(6)
		gl.VertexAttribPointer(6, 4, gl.FLOAT, false, int32(unsafe.Sizeof(mgl32.Mat4{})), gl.Ptr(3*unsafe.Sizeof(mgl32.Vec4{})))

		gl.VertexAttribDivisor(3, 1)
		gl.VertexAttribDivisor(4, 1)
		gl.VertexAttribDivisor(5, 1)
		gl.VertexAttribDivisor(6, 1)

		gl.BindVertexArray(0)
	}

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

		// configure transformation matrices
		projection := mgl32.Perspective(mgl32.DegToRad(45.0), initialWindowWidth/initialWindowHeight, 0.1, 1000.0)
		view := camera.getViewMatrix()
		asteroidShader.use()
		asteroidShader.setMat4("projection", projection)
		asteroidShader.setMat4("view", view)

		planetShader.use()
		planetShader.setMat4("projection", projection)
		planetShader.setMat4("view", view)

		// draw planet
		model := mgl32.Ident4().Mul4(mgl32.Translate3D(0.0, -3.0, 0.0))
		model = model.Mul4(mgl32.Scale3D(4.0, 4.0, 4.0))
		planetShader.setMat4("model", model)
		planet.Draw(*planetShader)

		// draw meteorites
		asteroidShader.use()
		asteroidShader.setInt("texture_diffuse1", 0)
		gl.ActiveTexture(gl.TEXTURE0)
		var bindTexture Texture
		for texture := range maps.Values(rock.texturesLoaded) {
			bindTexture = texture
			break
		}
		gl.BindTexture(gl.TEXTURE_2D, bindTexture.ID)
		for _, mesh := range rock.meshes {
			gl.BindVertexArray(mesh.VAO)
			gl.DrawElementsInstanced(gl.TRIANGLES, int32(len(mesh.indices)), gl.UNSIGNED_INT, gl.Ptr(nil), int32(amount))
			gl.BindVertexArray(0)
		}

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
