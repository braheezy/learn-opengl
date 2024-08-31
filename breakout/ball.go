package main

import "github.com/go-gl/mathgl/mgl32"

type Ball struct {
	obj    GameObject
	radius float32
	stuck  bool
}

func NewDefaultBall() *Ball {
	return &Ball{
		obj:    NewDefaultGameObject(),
		radius: 12.5,
		stuck:  true,
	}
}

func NewBall(pos mgl32.Vec2, radius float32, velocity mgl32.Vec2, sprite *Texture2D) *Ball {
	return &Ball{
		obj: GameObject{
			position: pos,
			size:     mgl32.Vec2{radius * 2.0, radius * 2.0},
			sprite:   sprite,
			color:    mgl32.Vec3{1.0, 1.0, 1.0},
			velocity: velocity,
		},
		radius: radius,
		stuck:  true,
	}
}

func (b *Ball) Move(deltaTime float32, wWidth int) mgl32.Vec2 {
	// if not stuck to the player board
	if !b.stuck {
		// move the ball
		b.obj.position = b.obj.position.Add(b.obj.velocity.Mul(deltaTime))
		// check if outside window bounds; if so, reverse velocity and restore at correct position
		if b.obj.position.X() <= 0.0 {
			b.obj.velocity[0] = -b.obj.velocity[0]
			b.obj.position[0] = 0.0
		} else if b.obj.position[0]+b.obj.size[0] >= float32(wWidth) {
			b.obj.velocity[0] = -b.obj.velocity[0]
			b.obj.position[0] = float32(wWidth) - b.obj.size[0]
		}
		if b.obj.position.Y() <= 0.0 {
			b.obj.velocity[1] = -b.obj.velocity[1]
			b.obj.position[1] = 0.0
		}
	}
	return b.obj.position
}

func (b *Ball) Reset(pos mgl32.Vec2, velocity mgl32.Vec2) {
	b.obj.position = pos
	b.obj.velocity = velocity
	b.stuck = true
}
