package main

import "github.com/go-gl/mathgl/mgl32"

type GameObject struct {
	position, size, velocity mgl32.Vec2
	color                    mgl32.Vec3
	rotation                 float32
	isSolid                  bool
	destroyed                bool
	sprite                   *Texture2D
}

func NewDefaultGameObject() GameObject {
	return GameObject{
		size:  mgl32.Vec2{1.0, 1.0},
		color: mgl32.Vec3{1.0, 1.0, 1.0},
	}
}

func (g *GameObject) Draw(rd *SpriteRenderer) {
	rd.DrawSprite(g.sprite, g.position, SpriteRendererOptions{g.size, g.rotation, g.color})
}
