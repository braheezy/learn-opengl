package main

import (
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.6-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	initialWindowWidth  = 800
	initialWindowHeight = 600
	enableWireframe     = false
)

var (
	vertices = []float32{
		// positions  // texture coords
		0.5, 0.5, 0.0, 1.0, 1.0, // top right
		0.5, -0.5, 0.0, 1.0, 0.0, // bottom right
		-0.5, -0.5, 0.0, 0.0, 0.0, // bottom left
		-0.5, 0.5, 0.0, 0.0, 1.0, // top left
	}
	indices = []uint32{ // note that we start from 0!
		0, 1, 3, // first triangle
		1, 2, 3, // second triangle
	}
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()
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

	/*
	 * Load OS-specific OpenGL function pointers
	 */
	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	/*
	 * Build and compile our shader program
	 */
	shaderProgram, err := NewShader("shaders/shader.vs", "shaders/shader.fs")
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
	// Store vertex attributes for each object in the VBO in a vertex array object (VAO)
	var VAO uint32
	// Store indices that tell OpenGL in what order to draw vertices. This helps reduce number of vertices
	// by not having to provide individual triangles with overlapping sides e.g. for 2 triangles, we need
	// 4 vertices instead of 6
	var EBO uint32
	// Get a unique ID for buffers.
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &EBO)

	// Bind VAO first
	gl.BindVertexArray(VAO)
	// Then bind VBO(s) and attribute pointers
	// Bind it as a ARRAY_BUFFER, allowing different typed buffers to bind.
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	// Get the vertex data into the buffer
	// STATIC_DRAW: the data is set only once and used many times.
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*int(unsafe.Sizeof(vertices[0])), gl.Ptr(vertices), gl.STATIC_DRAW)
	// Get the EBO data into a buffer too
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*int(unsafe.Sizeof(indices[0])), gl.Ptr(indices), gl.STATIC_DRAW)
	// Describe how OpenGL should interpret the vertex data by setting attributes.
	// Set the position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(5*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.EnableVertexAttribArray(0)
	// Set the texture attribute.
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, int32(5*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
	gl.EnableVertexAttribArray(1)

	/*
	 * load and create a texture
	 */
	var texture1 uint32
	gl.GenTextures(1, &texture1)
	gl.BindTexture(gl.TEXTURE_2D, texture1)
	// Set texture options
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	textureFile, err := os.Open("assets/container.jpg")
	if err != nil {
		log.Fatal(err)
	}
	textureImage, err := jpeg.Decode(textureFile)
	if err != nil {
		log.Fatal(err)
	}
	textureFile.Close()
	bounds := textureImage.Bounds()
	// Create a slice to hold the pixel data
	pixelData := make([]byte, bounds.Dx()*bounds.Dy()*3)
	// Convert the image to RGB and copy the pixel data to the slice
	index := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := textureImage.At(x, y).RGBA()
			pixelData[index] = byte(r >> 8)   // Red
			pixelData[index+1] = byte(g >> 8) // Green
			pixelData[index+2] = byte(b >> 8) // Blue
			index += 3
		}
	}
	// Generate texture from image data
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(textureImage.Bounds().Dx()), int32(textureImage.Bounds().Dy()), 0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(pixelData))
	// Generate mipmap, handling various resolutions of the texture at different distances.
	gl.GenerateMipmap(gl.TEXTURE_2D)

	var texture2 uint32
	gl.GenTextures(1, &texture2)
	gl.BindTexture(gl.TEXTURE_2D, texture2)
	// Set texture options
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	textureFile, err = os.Open("assets/awesomeface.png")
	if err != nil {
		log.Fatal(err)
	}
	textureImage, err = png.Decode(textureFile)
	if err != nil {
		log.Fatal(err)
	}
	textureFile.Close()
	bounds = textureImage.Bounds()
	// Create a slice to hold the pixel data
	pixelData = make([]byte, bounds.Dx()*bounds.Dy()*4)

	// Convert the image to RGB and copy the pixel data to the slice
	index = 0
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

	if enableWireframe {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	}

	// Activate shaders
	shaderProgram.use()
	// tell opengl for each sampler to which texture unit it belongs to
	shaderProgram.setInt("texture1", 0)
	shaderProgram.setInt("texture2", 1)

	// Run the render loop until the window is closed by the user.
	for !window.ShouldClose() {
		// Handle user input.
		processInput(window)

		// Perform render logic.
		// Clear the screen with a custom color
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		// Clear the color buffer (as opposed to the depth or stencil buffer)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Do transform maths
		// Create an identity matrix
		trans := mgl32.Ident4()
		// Translate to bottom-right side
		translation := mgl32.Translate3D(0.5, -0.5, 0.0)
		trans = trans.Mul4(translation)
		// Rotate the matrix continuously
		// angle := mgl32.DegToRad(float32(glfw.GetTime()))
		rotation := mgl32.HomogRotate3DZ(float32(glfw.GetTime()))
		trans = trans.Mul4(rotation)
		transSlice := trans[:]

		// Bind texture
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture1)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, texture2)

		// Render
		shaderProgram.use()
		transformLoc := gl.GetUniformLocation(shaderProgram.id, gl.Str("transform"+"\x00"))
		gl.UniformMatrix4fv(transformLoc, 1, false, &transSlice[0])
		gl.BindVertexArray(VAO)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.Ptr(nil))

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

// processInput handles key presses.
func processInput(w *glfw.Window) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}
}
