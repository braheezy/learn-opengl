package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"
	"unsafe"

	_ "github.com/mdouchement/hdr/codec/rgbe"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Settings
const (
	windowWidth  = 800
	windowHeight = 600
)

// Character represents a glyph's texture and related data.
type Character struct {
	TextureID          uint32 // ID handle of the glyph texture
	width, height      int
	bearingH, bearingV int   // Offset from baseline to left/top of glyph
	Advance            int32 // Horizontal offset to advance to next glyph
}

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
	Characters map[rune]Character

	VAO, VBO uint32
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
	glfw.WindowHint(glfw.Samples, 4)
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
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	/*
	 * Build and compile our shader program
	 */
	shader, err := NewShader("shaders/text.vs", "shaders/text.fs", "")
	if err != nil {
		log.Fatal(err)
	}

	// shader configuration
	// --------------------

	projection := mgl32.Ortho2D(0.0, windowWidth, 0.0, windowHeight)
	shader.use()
	shader.setMat4("projection", projection)

	// Initialize freetype context and load the font
	fontBytes, err := os.ReadFile("assets/Antonio-Regular.ttf")
	if err != nil {
		fmt.Println("ERROR: Failed to load font file:", err)
		return
	}

	// Parse the font
	ttf, err := opentype.Parse(fontBytes)
	if err != nil {
		fmt.Println("ERROR: Failed to parse font:", err)
		return
	}

	// Load the font face with a specific size
	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    48,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		fmt.Println("ERROR: Failed to create font face:", err)
		return
	}

	// Initialize the characters map
	Characters = make(map[rune]Character)

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	// Load the first 128 ASCII characters
	for c := rune(32); c < 127; c++ {
		// Measure the bounding box and advance to determine the glyph size
		bounds, advance, ok := face.GlyphBounds(c)
		if !ok {
			log.Printf("failed glyph bounds for %v", c)
		}
		width := (bounds.Max.X - bounds.Min.X).Ceil()  // Calculate width
		height := (bounds.Max.Y - bounds.Min.Y).Ceil() // Calculate height

		if c == ' ' {
			// Manually set the width and height to zero for the space character, but use the advance
			Characters[c] = Character{
				TextureID: 0, // No texture needed for space
				width:     0,
				height:    0,
				bearingH:  0,
				bearingV:  0,
				Advance:   int32(advance), // Use the advance value to move the cursor forward
			}
			continue
		}

		if width <= 0 || height <= 0 {
			fmt.Printf("Skipping character %v: invalid glyph size (width: %v, height: %v)\n", c, width, height)
			continue
		}

		dst := image.NewGray(image.Rect(0, 0, width, height))
		d := font.Drawer{
			Dst:  dst,
			Src:  image.White,
			Face: face,
			Dot:  fixed.P(-bounds.Min.X.Ceil(), -bounds.Min.Y.Ceil()),
		}
		d.DrawString(string(c))

		// Create a texture in OpenGL
		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		// Upload the texture data to OpenGL
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, int32(width), int32(height), 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(dst.Pix))

		// Set texture parameters to ensure proper glyph rendering
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		// Store the character with relevant OpenGL and bounding information
		Characters[c] = Character{
			TextureID: texture,
			width:     width,
			height:    height,
			bearingH:  bounds.Min.X.Ceil(),
			bearingV:  bounds.Max.Y.Ceil(),
			Advance:   int32(advance),
		}

	}

	gl.BindTexture(gl.TEXTURE_2D, 0)
	// Cleanup resources
	face.Close()

	// configure VAO/VBO for texture quads
	// -----------------------------------
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.BindVertexArray(VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, int(6*4*unsafe.Sizeof(float32(0))), gl.Ptr(nil), gl.DYNAMIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, int32(4*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	// Run the render loop until the window is closed by the user.
	for !window.ShouldClose() {
		// calculate time stats
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Handle user input.
		processInput(window)

		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		renderText(shader, "This is sample text", 25.0, 25.0, 1.0, mgl32.Vec3{0.5, 0.8, 0.2})
		renderText(shader, "Learn OpenGL in Go!", 540.0, 570.0, 0.5, mgl32.Vec3{0.3, 0.7, 0.9})

		// Swap the color buffer and poll events
		window.SwapBuffers()
		glfw.PollEvents()
	}

}

// render line of text
func renderText(shader *Shader, text string, x, y, scale float32, color mgl32.Vec3) {
	// Activate corresponding render state
	shader.use()
	gl.Uniform3f(gl.GetUniformLocation(shader.id, gl.Str("textColor\x00")), color.X(), color.Y(), color.Z())
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(VAO)

	// Iterate through all characters in the text string
	for _, char := range text {
		ch, ok := Characters[char]
		if !ok {
			fmt.Printf("Skipping %v\n", ch)
			continue // Skip characters that are not loaded
		}

		// Calculate position and size of the character quad
		xpos := x + float32(ch.bearingH)*scale
		ypos := y - float32(ch.bearingV)*scale

		w := float32(ch.width) * scale
		h := float32(ch.height) * scale

		// Update VBO for each character
		vertices := []float32{
			xpos, ypos + h, 0.0, 0.0,
			xpos, ypos, 0.0, 1.0,
			xpos + w, ypos, 1.0, 1.0,

			xpos, ypos + h, 0.0, 0.0,
			xpos + w, ypos, 1.0, 1.0,
			xpos + w, ypos + h, 1.0, 0.0,
		}

		// Render glyph texture over quad
		gl.BindTexture(gl.TEXTURE_2D, ch.TextureID)

		// Update content of VBO memory
		gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*int(unsafe.Sizeof(vertices[0])), gl.Ptr(vertices))

		// Render quad
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)

		// Now advance cursors for next glyph
		newX := float32(ch.Advance>>6) * scale
		x += newX // Bitshift by 6 to get value in pixels
	}

	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
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

var sphereVAO, indexCount uint32

func renderSphere() {
	if sphereVAO == 0 {
		gl.GenVertexArrays(1, &sphereVAO)

		var vbo, ebo uint32
		gl.GenBuffers(1, &vbo)
		gl.GenBuffers(1, &ebo)

		positions := []mgl32.Vec3{}
		uv := []mgl32.Vec2{}
		normals := []mgl32.Vec3{}
		indices := []uint32{}

		const X_SEGMENTS = 64
		const Y_SEGMENTS = 64
		const PI = math.Pi

		// Generate vertex data
		for x := 0; x <= X_SEGMENTS; x++ {
			for y := 0; y <= Y_SEGMENTS; y++ {
				xSegment := float32(x) / float32(X_SEGMENTS)
				ySegment := float32(y) / float32(Y_SEGMENTS)
				xPos := float32(math.Cos(float64(xSegment*2.0*PI)) * math.Sin(float64(ySegment*PI)))
				yPos := float32(math.Cos(float64(ySegment * PI)))
				zPos := float32(math.Sin(float64(xSegment*2.0*PI)) * math.Sin(float64(ySegment*PI)))

				positions = append(positions, mgl32.Vec3{xPos, yPos, zPos})
				uv = append(uv, mgl32.Vec2{xSegment, ySegment})
				normals = append(normals, mgl32.Vec3{xPos, yPos, zPos})
			}
		}

		// Generate index data
		oddRow := false
		for y := 0; y < Y_SEGMENTS; y++ {
			if !oddRow {
				for x := 0; x <= X_SEGMENTS; x++ {
					indices = append(indices, uint32(y*(X_SEGMENTS+1)+x))
					indices = append(indices, uint32((y+1)*(X_SEGMENTS+1)+x))
				}
			} else {
				for x := X_SEGMENTS; x >= 0; x-- {
					indices = append(indices, uint32((y+1)*(X_SEGMENTS+1)+x))
					indices = append(indices, uint32(y*(X_SEGMENTS+1)+x))
				}
			}
			oddRow = !oddRow
		}
		indexCount = uint32(len(indices))

		// Combine vertex data into a single array
		data := []float32{}
		for i := 0; i < len(positions); i++ {
			data = append(data, positions[i].X(), positions[i].Y(), positions[i].Z())
			if len(normals) > 0 {
				data = append(data, normals[i].X(), normals[i].Y(), normals[i].Z())
			}
			if len(uv) > 0 {
				data = append(data, uv[i].X(), uv[i].Y())
			}
		}

		// Set up VAO, VBO, and EBO
		gl.BindVertexArray(sphereVAO)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, gl.Ptr(data), gl.STATIC_DRAW)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
		stride := int32((3 + 3 + 2) * 4) // 3 for position, 3 for normal, 2 for UV, all float32 (4 bytes each)
		// Position attribute
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, unsafe.Pointer(nil))
		// Normal attribute
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.Ptr(3*unsafe.Sizeof(float32(0))))
		// UV attribute
		gl.EnableVertexAttribArray(2)
		gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.Ptr(6*unsafe.Sizeof(float32(0))))

	}

	gl.BindVertexArray(sphereVAO)
	gl.DrawElements(gl.TRIANGLE_STRIP, int32(indexCount), gl.UNSIGNED_INT, gl.Ptr(nil))
}
