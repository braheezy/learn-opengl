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

func CheckCollision(one, two GameObject) bool {
	// collision x-axis?
	collisionX := one.position.X()+one.size.X() >= two.position.X() && two.position.X()+two.size.X() >= one.position.X()
	collisionY := one.position.Y()+one.size.Y() >= two.position.Y() && two.position.Y()+two.size.Y() >= one.position.Y()
	return collisionX && collisionY

}

func CheckBallCollision(one *Ball, two *GameObject) Collision {
	// get center point circle first
	center := one.obj.position.Add(mgl32.Vec2{one.radius, one.radius})

	// calculate AABB info (center, half-extents)
	aabbHalfExtents := mgl32.Vec2{two.size.X() / 2.0, two.size.Y() / 2.0}
	aabbCenter := mgl32.Vec2{two.position.X() + aabbHalfExtents.X(), two.position.Y() + aabbHalfExtents.Y()}
	// get difference vector between both centers
	difference := center.Sub(aabbCenter)
	clamped := ClampVec2(difference, aabbHalfExtents.Mul(-1.0), aabbHalfExtents)
	// add clamped value to AABB_center and we get the value of box closest to circle
	closest := aabbCenter.Add(clamped)
	// retrieve vector between center circle and closest point AABB and check if length <= radius
	difference = closest.Sub(center)

	if difference.Len() < one.radius {
		return Collision{true, VectorDirection(difference), difference}
	} else {
		return Collision{false, Up, mgl32.Vec2{0.0, 0.0}}
	}
}
func ClampVec2(v, min, max mgl32.Vec2) mgl32.Vec2 {
	return mgl32.Vec2{
		mgl32.Clamp(v.X(), min.X(), max.X()),
		mgl32.Clamp(v.Y(), min.Y(), max.Y()),
	}
}
