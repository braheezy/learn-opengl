package main

import "github.com/go-gl/mathgl/mgl32"

var (
	PowerupSize     = mgl32.Vec2{60.0, 20.0}
	PowerupVelocity = mgl32.Vec2{0.0, 150.0}
)

type Powerup struct {
	kind      string
	duration  float32
	activated bool
	obj       *GameObject
}

func NewPowerup(kind string, color mgl32.Vec3, duration float32, position mgl32.Vec2, texture *Texture2D) *Powerup {
	return &Powerup{
		kind:     kind,
		duration: duration,
		obj: &GameObject{
			position: position,
			color:    color,
			sprite:   texture,
			size:     PowerupSize,
			velocity: PowerupVelocity,
		},
	}
}
