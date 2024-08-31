package main

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type SpriteRenderer struct {
	shader  *Shader
	quadVAO uint32
}

func (sp *SpriteRenderer) DrawSprite(texture *Texture2D, position mgl32.Vec2) {

	size := mgl32.Vec2{10.0, 10.0}
	rotate := float32(0)
	color := mgl32.Vec3{1.0}
	sp.drawSprite(texture, position, size, rotate, color)
}
func (sp *SpriteRenderer) drawSprite(texture *Texture2D, position mgl32.Vec2, size mgl32.Vec2, rotate float32, color mgl32.Vec3) {

	sp.shader.use()
	model := mgl32.Ident4().Mul4(mgl32.Translate3D(position.X(), position.Y(), 0.0))

	model = model.Mul4(mgl32.Translate3D(0.5*size.X(), 0.5*size.Y(), 0.0))
	model = model.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(rotate), mgl32.Vec3{0.0, 0.0, 1.0}))
	model = model.Mul4(mgl32.Translate3D(-0.5*size.X(), -0.5*size.Y(), 0.0))

	model = model.Mul4(mgl32.Scale3D(size.X(), size.Y(), 1.0))

	sp.shader.setMat4("model", model)
	sp.shader.setVec3("spriteColor", color)

	gl.ActiveTexture(gl.TEXTURE0)
	texture.Bind()

	gl.BindVertexArray(sp.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

func NewSpriteRenderer(shader *Shader) *SpriteRenderer {
	// configure VAO/VBO
	var VBO uint32
	vertices := []float32{
		// pos      // tex
		0.0, 1.0, 0.0, 1.0,
		1.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 0.0,

		0.0, 1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0,
		1.0, 0.0, 1.0, 0.0,
	}

	sp := &SpriteRenderer{
		shader: shader,
	}

	gl.GenVertexArrays(1, &sp.quadVAO)
	gl.GenBuffers(1, &VBO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*int(unsafe.Sizeof(vertices[0])), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindVertexArray(sp.quadVAO)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*int32(unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return sp
}
