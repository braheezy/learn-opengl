package main

import (
	"fmt"
	"strconv"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Meshes contain complex vertexes, lighting characteristics, materials, etc.
// After drawing multiple meshes, it is the full model

type Vertex struct {
	position  mgl32.Vec3
	normal    mgl32.Vec3
	texCoords mgl32.Vec2
}

type Texture struct {
	id    uint32
	_type string
}

type Mesh struct {
	// mesh data
	vertices []Vertex
	indices  []uint32
	textures []Texture
	// render data
	vao, vbo, ebo uint32
}

func NewMesh(vertices []Vertex, indices []uint32, textures []Texture) *Mesh {
	m := &Mesh{vertices: vertices, indices: indices, textures: textures}
	m.setupMesh()
	return m
}

func (m *Mesh) setupMesh() {
	gl.GenVertexArrays(1, &m.vao)
	gl.GenBuffers(1, &m.vbo)
	gl.GenBuffers(1, &m.ebo)

	gl.BindVertexArray(m.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)

	gl.BufferData(gl.ARRAY_BUFFER, len(m.vertices)*int(unsafe.Sizeof(m.vertices[0])), gl.Ptr(m.vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(m.indices)*int(unsafe.Sizeof(uint(0))), gl.Ptr(m.indices), gl.STATIC_DRAW)

	// vertex positions
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(unsafe.Sizeof(m.vertices[0])), gl.Ptr(nil))
	// vertex normals
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, int32(unsafe.Sizeof(m.vertices[0])), unsafe.Offsetof(m.vertices[0].normal))
	// vertex texture coords
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, int32(unsafe.Sizeof(m.vertices[0])), unsafe.Offsetof(m.vertices[0].texCoords))

	gl.BindVertexArray(0)
}

func (m *Mesh) Draw(shader *Shader) {
	diffuseN := 1
	specularN := 1
	for i := 0; i < len(m.textures); i++ {
		// activate proper texture unit before binding
		gl.ActiveTexture(gl.TEXTURE0 + uint32(i))
		// retrieve texture number (the N in diffuse_textureN)
		name := m.textures[i]._type
		var number string
		if name == "texture_diffuse" {
			number = strconv.Itoa(diffuseN)
			diffuseN++
		} else if name == "texture_specular" {
			number = strconv.Itoa(specularN)
			specularN++
		}
		shader.setInt(fmt.Sprintf("material.%v%v", name, number), int32(i))
		gl.BindTexture(gl.TEXTURE_2D, m.textures[i].id)
	}
	gl.ActiveTexture(gl.TEXTURE0)

	// draw mesh
	gl.BindVertexArray(m.vao)
	gl.DrawElements(gl.TRIANGLES, int32(len(m.indices)), gl.UNSIGNED_INT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}
