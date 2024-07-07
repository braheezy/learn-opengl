package main

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v4.6-compatibility/gl"
)

type Shader struct {
	id uint32
}

func NewShader(vertexPath string, fragmentPath string) (*Shader, error) {
	// Read shader programs from disk.
	data, err := os.ReadFile(vertexPath)
	if err != nil {
		return nil, err
	}
	vertexShaderSource := string(data)
	data, err = os.ReadFile(fragmentPath)
	if err != nil {
		return nil, err
	}
	fragmentShaderSource := string(data)

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
		return nil, fmt.Errorf("failed to compile vertex shader: %v", string(infoLog))
	}
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		gl.GetShaderInfoLog(fragmentShader, 512, nil, (*uint8)(unsafe.Pointer(&infoLog)))
		return nil, fmt.Errorf("failed to compile fragment shader: %v", string(infoLog))
	}

	// Link all shaders together to form a shader program, which is used during rendering.
	ID := gl.CreateProgram()
	gl.AttachShader(ID, vertexShader)
	gl.AttachShader(ID, fragmentShader)
	gl.LinkProgram(ID)

	// Check program linking
	gl.GetProgramiv(ID, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		gl.GetProgramInfoLog(ID, 512, nil, (*uint8)(unsafe.Pointer(&infoLog)))
		return nil, fmt.Errorf("failed to link shader program: %v", string(infoLog))
	}

	// Clean up shader objects
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return &Shader{id: ID}, nil
}

func (s *Shader) use() {
	gl.UseProgram(s.id)
}

func (s *Shader) setBool(name string, value bool) {
	var v0 int32
	if value {
		v0 = 1
	}
	gl.Uniform1i(gl.GetUniformLocation(s.id, gl.Str(name+"\x00")), v0)
}

func (s *Shader) setInt(name string, value int32) {
	gl.Uniform1i(gl.GetUniformLocation(s.id, gl.Str(name+"\x00")), value)
}

func (s *Shader) setFloat(name string, value float32) {
	gl.Uniform1f(gl.GetUniformLocation(s.id, gl.Str(name+"\x00")), value)
}
