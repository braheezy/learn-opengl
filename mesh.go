package main

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const MaxBoneInfluence = 4

type Vertex struct {
	Position  mgl32.Vec3
	Normal    mgl32.Vec3
	TexCoords mgl32.Vec2
	Tangent   mgl32.Vec3
	Bitangent mgl32.Vec3
	BoneIDs   [MaxBoneInfluence]int32
	Weights   [MaxBoneInfluence]float32
}

type Texture struct {
	ID   uint32
	Type string
	Path string
}

type Mesh struct {
	vertices []Vertex
	indices  []uint32
	textures []Texture
	VAO      uint32
	VBO      uint32
	EBO      uint32
}

func NewMesh(vertices []Vertex, indices []uint32, textures []Texture) *Mesh {
	mesh := &Mesh{
		vertices: vertices,
		indices:  indices,
		textures: textures,
	}
	mesh.setupMesh()
	return mesh
}

func (mesh *Mesh) Draw(shader Shader) {
	// Bind appropriate textures
	var diffuseNr, specularNr, normalNr, heightNr uint32 = 1, 1, 1, 1
	for i, texture := range mesh.textures {
		gl.ActiveTexture(gl.TEXTURE0 + uint32(i))
		var number string
		switch texture.Type {
		case "texture_diffuse":
			number = fmt.Sprintf("%v", diffuseNr)
			diffuseNr++
		case "texture_specular":
			number = fmt.Sprintf("%v", specularNr)
			specularNr++
		case "texture_normal":
			number = fmt.Sprintf("%v", normalNr)
			normalNr++
		case "texture_height":
			number = fmt.Sprintf("%v", heightNr)
			heightNr++
		}

		textureName := texture.Type + number

		gl.Uniform1i(gl.GetUniformLocation(shader.id, gl.Str(textureName+"\x00")), int32(i))
		gl.BindTexture(gl.TEXTURE_2D, texture.ID)
	}

	// Draw mesh
	gl.BindVertexArray(mesh.VAO)
	gl.DrawElements(gl.TRIANGLES, int32(len(mesh.indices)), gl.UNSIGNED_INT, unsafe.Pointer(nil))
	gl.BindVertexArray(0)

	// Set everything back to defaults
	gl.ActiveTexture(gl.TEXTURE0)
}

func (mesh *Mesh) setupMesh() {
	// Create buffers/arrays
	gl.GenVertexArrays(1, &mesh.VAO)
	gl.GenBuffers(1, &mesh.VBO)
	gl.GenBuffers(1, &mesh.EBO)

	gl.BindVertexArray(mesh.VAO)

	// Load data into vertex buffers
	gl.BindBuffer(gl.ARRAY_BUFFER, mesh.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(mesh.vertices)*int(unsafe.Sizeof(Vertex{})), unsafe.Pointer(&mesh.vertices[0]), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(mesh.indices)*int(unsafe.Sizeof(uint32(0))), unsafe.Pointer(&mesh.indices[0]), gl.STATIC_DRAW)

	// Set the vertex attribute pointers
	// Vertex Positions
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), gl.Ptr(nil))

	// Vertex Normals
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), gl.Ptr(unsafe.Offsetof(Vertex{}.Normal)))

	// Vertex Texture Coords
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), gl.Ptr(unsafe.Offsetof(Vertex{}.TexCoords)))

	// Vertex Tangent
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 3, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), gl.Ptr(unsafe.Offsetof(Vertex{}.Tangent)))

	// Vertex Bitangent
	gl.EnableVertexAttribArray(4)
	gl.VertexAttribPointer(4, 3, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), gl.Ptr(unsafe.Offsetof(Vertex{}.Bitangent)))

	// Bone IDs
	gl.EnableVertexAttribArray(5)
	gl.VertexAttribIPointer(5, MaxBoneInfluence, gl.INT, int32(unsafe.Sizeof(Vertex{})), gl.Ptr(unsafe.Offsetof(Vertex{}.BoneIDs)))

	// Weights
	gl.EnableVertexAttribArray(6)
	gl.VertexAttribPointer(6, MaxBoneInfluence, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), gl.Ptr(unsafe.Offsetof(Vertex{}.Weights)))

	gl.BindVertexArray(0)
}
