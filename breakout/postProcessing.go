package main

import (
	"log"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// PostProcessor hosts all PostProcessing effects for the Breakout
// Game. It renders the game on a textured quad after which one can
// enable specific effects by enabling either the Confuse, Chaos or
// Shake boolean.
// It is required to call BeginRender() before rendering the game
// and EndRender() after rendering the game for the class to work.
type PostProcessor struct {
	shader        *Shader
	texture       *Texture2D
	width, height int32
	// options
	confuse, chaos, shake bool
	// render state.
	// MSFBO = Multisampled FBO. FBO is regular, used for blitting MS color-buffer to texture
	// RBO is used for multisampled color buffer
	MSFBO, FBO, RBO, VAO uint32
}

func NewPostProcessor(shader *Shader, width, height int32) *PostProcessor {
	pp := PostProcessor{shader: shader, width: width, height: height, texture: NewTexture()}
	// initialize renderbuffer/framebuffer object
	gl.GenFramebuffers(1, &pp.MSFBO)
	gl.GenFramebuffers(1, &pp.FBO)
	gl.GenRenderbuffers(1, &pp.RBO)
	// initialize renderbuffer storage with a multisampled color buffer (don't need a depth/stencil buffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, pp.MSFBO)
	gl.BindRenderbuffer(gl.RENDERBUFFER, pp.RBO)
	// allocate storage for render buffer object
	gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, 4, gl.RGB, width, height)
	// attach MS render buffer object to framebuffer
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, pp.RBO)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("Failed to initialize MSFBO")
	}
	// also initialize the FBO/texture to blit multisampled color-buffer to; used for shader operations (for postprocessing effects)
	gl.BindFramebuffer(gl.FRAMEBUFFER, pp.FBO)
	pp.texture.Generate(width, height, nil)
	// attach texture to framebuffer as its color attachment
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, pp.texture.ID, 0)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("Failed to initialize FBO")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	// initialize render data and uniforms
	pp.initRenderData()
	pp.shader.use()
	pp.shader.setInt("scene", 0)
	offset := float32(1.0 / 300.0)
	offsets := [][]float32{
		{-offset, offset},  // top-left
		{0.0, offset},      // top-center
		{offset, offset},   // top-right
		{-offset, 0.0},     // center-left
		{0.0, 0.0},         // center-center
		{offset, 0.0},      // center - right
		{-offset, -offset}, // bottom-left
		{0.0, -offset},     // bottom-center
		{offset, -offset},  // bottom-right
	}
	gl.Uniform2fv(gl.GetUniformLocation(pp.shader.id, gl.Str("offsets\x00")), 9, &offsets[0][0])
	edge_kernel := []int32{
		-1, -1, -1,
		-1, 8, -1,
		-1, -1, -1,
	}
	gl.Uniform1iv(gl.GetUniformLocation(pp.shader.id, gl.Str("edge_kernel\x00")), 9, &edge_kernel[0])
	blur_kernel := []float32{
		1.0 / 16.0, 2.0 / 16.0, 1.0 / 16.0,
		2.0 / 16.0, 4.0 / 16.0, 2.0 / 16.0,
		1.0 / 16.0, 2.0 / 16.0, 1.0 / 16.0,
	}
	gl.Uniform1fv(gl.GetUniformLocation(pp.shader.id, gl.Str("blur_kernel\x00")), 9, &blur_kernel[0])

	return &pp
}

func (pp *PostProcessor) initRenderData() {
	// configure VAO/VBO
	var VBO uint32
	vertices := []float32{
		// pos        // tex
		-1.0, -1.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
		-1.0, 1.0, 0.0, 1.0,

		-1.0, -1.0, 0.0, 0.0,
		1.0, -1.0, 1.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
	}
	gl.GenVertexArrays(1, &pp.VAO)
	gl.GenBuffers(1, &VBO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*int(unsafe.Sizeof(vertices[0])), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindVertexArray(pp.VAO)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*int32(unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (pp *PostProcessor) Render(time float32) {
	// set uniforms/options
	pp.shader.use()
	pp.shader.setFloat("time", time)
	pp.shader.setBool("confuse", pp.confuse)
	pp.shader.setBool("chaos", pp.chaos)
	pp.shader.setBool("shake", pp.shake)
	// render textured quad
	gl.ActiveTexture(gl.TEXTURE0)
	pp.texture.Bind()
	gl.BindVertexArray(pp.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}
func (pp *PostProcessor) BeginRender() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, pp.MSFBO)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}
func (pp *PostProcessor) EndRender() {
	// now resolve multisampled color-buffer into intermediate FBO to store to texture
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, pp.MSFBO)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, pp.FBO)
	gl.BlitFramebuffer(0, 0, pp.width, pp.height, 0, 0, pp.width, pp.height, gl.COLOR_BUFFER_BIT, gl.NEAREST)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0) // binds both READ and WRITE framebuffer to default framebuffer
}
