package main

import (
	"math/rand"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Particle struct {
	position, velocity mgl32.Vec2
	color              mgl32.Vec4
	life               float32
}

func NewParticle() Particle {
	return Particle{color: mgl32.Vec4{1.0, 1.0, 1.0, 1.0}}
}

// ParticleGenerator acts as a container for rendering a large number of
// particles by repeatedly spawning and updating particles and killing
// them after a given amount of time.
type ParticleGenerator struct {
	particles []Particle
	amount    int
	shader    *Shader
	texture   *Texture2D
	VAO       uint32
}

func NewParticleGenerator(shader *Shader, texture *Texture2D, amount int) *ParticleGenerator {
	// setup mesh and attribute properties
	var VBO uint32
	particleQuad := []float32{
		0.0, 1.0, 0.0, 1.0,
		1.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 0.0,

		0.0, 1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0,
		1.0, 0.0, 1.0, 0.0,
	}
	pg := ParticleGenerator{shader: shader, texture: texture, amount: amount}
	gl.GenVertexArrays(1, &pg.VAO)
	gl.GenBuffers(1, &VBO)
	gl.BindVertexArray(pg.VAO)
	// fill mesh buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(particleQuad)*int(unsafe.Sizeof(float32(0))), gl.Ptr(particleQuad), gl.STATIC_DRAW)
	// set mesh attributes
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*int32(unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.BindVertexArray(0)

	// create default particle instances
	for i := 0; i < pg.amount; i++ {
		pg.particles = append(pg.particles, NewParticle())
	}
	return &pg
}

func (pg *ParticleGenerator) Update(deltaTime float32, object *Ball, newParticles int, offset mgl32.Vec2) {
	// add new particles
	for i := 0; i < newParticles; i++ {
		unusedParticle := pg.firstUnusedParticle()
		pg.respawnParticle(&pg.particles[unusedParticle], object, offset)
	}
	// update all particles
	for i := range pg.particles {
		p := &pg.particles[i]
		p.life -= deltaTime
		if p.life > 0.0 {
			// particle is alive, update it
			p.position = p.position.Sub(p.velocity.Mul(deltaTime))
			p.color[3] -= deltaTime * 2.5
		}
	}
}

func (pg *ParticleGenerator) Draw() {
	// use additive blending to give it a glow effect
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	pg.shader.use()
	for _, particle := range pg.particles {
		if particle.life > 0.0 {
			pg.shader.setVec2("offset", particle.position)
			pg.shader.setVec4("color", particle.color)
			pg.texture.Bind()
			gl.BindVertexArray(pg.VAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
			gl.BindVertexArray(0)
		}
	}
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

var lastUsedParticle int

func (pg *ParticleGenerator) firstUnusedParticle() int {
	// first search from last used particule, this is will usually return instantly
	for i := lastUsedParticle; i < pg.amount; i++ {
		if pg.particles[i].life <= 0.0 {
			lastUsedParticle = i
			return i
		}
	}
	// otherwise, do linear search
	for i := 0; i < lastUsedParticle; i++ {
		if pg.particles[i].life <= 0.0 {
			lastUsedParticle = i
			return i
		}
	}

	// all particles are taken, override the first one (note that if it repeatedly hits this case, more particles should be reserved)
	lastUsedParticle = 0
	return 0
}

func (pg *ParticleGenerator) respawnParticle(particle *Particle, ball *Ball, offset mgl32.Vec2) {
	random := float32((rand.Int()%100)-50) / 10.0
	rColor := 0.5 + float32(rand.Int()%100)/100.0
	particle.position = ball.obj.position.Add(mgl32.Vec2{random, random}).Add(offset)
	particle.color = mgl32.Vec4{rColor, rColor, rColor, 1.0}
	particle.life = 1.0
	particle.velocity = ball.obj.velocity.Mul(0.1)
}
