package main

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Shader struct {
	id uint32
}

func NewShader(vertexShaderSource string, fragmentShaderSource string, geometryShaderSource string) (*Shader, error) {
	var geometryShader uint32

	if geometryShaderSource != "" {
		geometryShader = gl.CreateShader(gl.GEOMETRY_SHADER)
		sourceString, geometryFreeFunc := gl.Strs(geometryShaderSource, "\x00")
		defer geometryFreeFunc()
		gl.ShaderSource(geometryShader, 1, sourceString, nil)
		gl.CompileShader(geometryShader)
		checkCompile(geometryShader, "GEOMETRY")
	}

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
	checkCompile(fragmentShader, "VERTEX")
	checkCompile(fragmentShader, "FRAGMENT")

	// Link all shaders together to form a shader program, which is used during rendering.
	ID := gl.CreateProgram()
	gl.AttachShader(ID, vertexShader)
	gl.AttachShader(ID, fragmentShader)
	if geometryShaderSource != "" {
		gl.AttachShader(ID, geometryShader)
	}
	gl.LinkProgram(ID)

	// Check program linking
	checkCompile(ID, "PROGRAM")

	// Clean up shader objects
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
	if geometryShaderSource != "" {
		gl.DeleteShader(geometryShader)
	}

	return &Shader{id: ID}, nil
}

func checkCompile(shader uint32, shaderType string) error {
	var success int32
	infoLog := make([]uint8, 512)
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		gl.GetShaderInfoLog(shader, 512, nil, (*uint8)(unsafe.Pointer(&infoLog)))
		return fmt.Errorf("failed to compile %v shader: %v", shaderType, string(infoLog))
	}
	return nil
}

func (s *Shader) use() *Shader {
	gl.UseProgram(s.id)
	return s
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
func (s *Shader) setMat3(name string, value mgl32.Mat3) {
	gl.UniformMatrix3fv(gl.GetUniformLocation(s.id, gl.Str(name+"\x00")), 1, false, &value[0])
}
func (s *Shader) setMat4(name string, value mgl32.Mat4) {
	gl.UniformMatrix4fv(gl.GetUniformLocation(s.id, gl.Str(name+"\x00")), 1, false, &value[0])
}
func (s *Shader) setVec3(name string, value mgl32.Vec3) {
	gl.Uniform3fv(gl.GetUniformLocation(s.id, gl.Str(name+"\x00")), 1, &value[0])
}
