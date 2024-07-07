package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.6-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	initialWindowWidth  = 800
	initialWindowHeight = 600
	enableWireframe     = true
)

var (
	vertices = []float32{
		0.5, 0.5, 0.0, // top right
		0.5, -0.5, 0.0, // bottom right
		-0.5, -0.5, 0.0, // bottom left
		-0.5, 0.5, 0.0, // top left
	}
	indices = []uint32{ // note that we start from 0!
		0, 1, 3, // first triangle
		1, 2, 3, // second triangle
	}
	vertexShaderSource = `#version 460 compatibility
layout (location = 0) in vec3 aPos;
void main()
{
	gl_Position = vec4(aPos.x, aPos.y, aPos.z, 1.0);
}`
	fragmentShaderSource = `#version 460 compatibility
out vec4 FragColor;
void main()
{
	FragColor = vec4(1.0f, 0.5f, 0.2f, 1.0f);
}`
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()
}

func main() {
	// Initialize GLFW, which is used to manage windows, user input, opengl contexts, and related
	// events.
	err := glfw.Init()
	if err != nil {
		log.Fatal(err)
	}
	// Free resources used by GLFW when the program exits.
	defer glfw.Terminate()

	// Set various options for the window we're to create using hints.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	// Compatibility profile allows more deprecated function calls over core profile.
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCompatProfile)

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

	// Load the OS-specific OpenGL function pointers
	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	shaderProgram := compileShaderProgram()

	/* A graphics pipeline are the stages taken to determine how to paint 2D pixels
	according to 3D coordinates. The stages are individual programs called shaders.

	The vertex shader is the first shader. It takes as input a single vertex and transforms
	it to other 3D coordinates. OpenGL expects a vertex and fragment shader.
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
	/*
		param1: which vertex attribute to configure. We specified the location of the position
		        vertex attribute in the shader with `layout (location = 0)`.
		param2: size of vertex attribute. It's vec3, so 3 values
		param3: type of data (vec in GLSL is always float)
		param4: if the data should be normalized. Ours already is
		param5: the stride, or space between consecutive vertex attributes
		param6: offset of position data in buffer
	*/
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(3*unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.EnableVertexAttribArray(0)

	if enableWireframe {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	}

	// Run the render loop until the window is closed by the user.
	for !window.ShouldClose() {
		// Handle user input.
		processInput(window)

		// Perform render logic.
		// Clear the screen with a custom color
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		// Clear the color buffer (as opposed to the depth or stencil buffer)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Active new program object by instructing OpenGL to use it.
		gl.UseProgram(shaderProgram)
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

func compileShaderProgram() uint32 {
	// -------- Shaders ----------
	// Create a vertex shader object ID
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	// Put the shader source code into the vertex shader. It must be a null-terminated string
	// in C flavor.
	sourceString, vertexFreeFunc := gl.Strs(vertexShaderSource, "\x00")
	defer vertexFreeFunc()
	gl.ShaderSource(vertexShader, 1, sourceString, nil)
	// Compile the shader
	gl.CompileShader(vertexShader)

	// The next shader is the fragment shader, concerned with calculating color output
	// for each pixel.
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	sourceString, fragmentFreeFunc := gl.Strs(fragmentShaderSource, "\x00")
	defer fragmentFreeFunc()
	gl.ShaderSource(fragmentShader, 1, sourceString, nil)
	gl.CompileShader(fragmentShader)

	// Check shader compilation.
	var success int32
	infoLog := make([]uint8, 512)
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		gl.GetShaderInfoLog(vertexShader, 512, nil, (*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalf("Failed to compile vertex shader: %v", string(infoLog))
	}
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		gl.GetShaderInfoLog(fragmentShader, 512, nil, (*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalf("Failed to compile fragment shader: %v", string(infoLog))
	}

	// Link all shaders together to form a shader program, which is used during rendering.
	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)

	// Check program linking
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		gl.GetProgramInfoLog(shaderProgram, 512, nil, (*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalf("Failed to link shader program: %v", string(infoLog))
	}

	// Clean up shader objects
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return shaderProgram
}
