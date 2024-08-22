package main

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type CameraMovement int

const (
	FORWARD CameraMovement = iota
	BACKWARD
	LEFT
	RIGHT
)

// Camera provides a fly camera to navigate a scene
type Camera struct {
	// camera attributes
	position mgl32.Vec3
	front    mgl32.Vec3
	up       mgl32.Vec3
	right    mgl32.Vec3
	worldUp  mgl32.Vec3
	// euler angles
	yaw, pitch float32
	// camera options
	movementSpeed    float32
	mouseSensitivity float32
	zoom             float32
}

func newCamera() *Camera {
	c := Camera{
		position:         mgl32.Vec3{0.0, 0.0, 0.0},
		worldUp:          mgl32.Vec3{0.0, 1.0, 0.0},
		up:               mgl32.Vec3{0.0, 1.0, 0.0},
		yaw:              -90.0,
		pitch:            0.0,
		zoom:             45.0,
		mouseSensitivity: 0.1,
		movementSpeed:    5,
	}
	return &c
}

func NewDefaultCamera() *Camera {
	c := newCamera()

	c.updateVectors()
	return c
}

func NewDefaultCameraAtPosition(position mgl32.Vec3) *Camera {
	c := newCamera()

	c.position = position

	c.updateVectors()
	return c
}

func NewCamera(position mgl32.Vec3, up mgl32.Vec3, yaw float32, pitch float32) *Camera {
	c := newCamera()

	c.position = position
	c.up = up
	c.yaw = yaw
	c.pitch = pitch

	c.updateVectors()
	return c
}

func NewCameraWithScalars(posX, posY, posZ, upX, upY, upZ, yaw, pitch float32) *Camera {
	c := newCamera()

	c.position = mgl32.Vec3{posX, posY, posZ}
	c.up = mgl32.Vec3{upX, upY, upZ}
	c.yaw = yaw
	c.pitch = pitch

	c.updateVectors()
	return c
}

func (c *Camera) processKeyboard(direction CameraMovement, deltaTime float32) {
	velocity := c.movementSpeed * deltaTime
	switch direction {
	case FORWARD:
		c.position = c.position.Add(c.front.Mul(velocity))
	case BACKWARD:
		c.position = c.position.Sub(c.front.Mul(velocity))
	case LEFT:
		c.position = c.position.Sub(c.right.Mul(velocity))
	case RIGHT:
		c.position = c.position.Add(c.right.Mul(velocity))
	}
}

func (c *Camera) processMouseMovement(xOffset, yOffset float32, constrainPitch bool) {
	xOffset *= c.mouseSensitivity
	yOffset *= c.mouseSensitivity

	c.yaw += xOffset
	c.pitch += yOffset

	if constrainPitch {
		if c.pitch > 89.0 {
			c.pitch = 89.0
		}
		if c.pitch < -89.0 {
			c.pitch = -89.0
		}
	}

	c.updateVectors()
}

func (c *Camera) processMouseScroll(yOffset float32) {
	c.zoom -= yOffset
	if c.zoom < 1.0 {
		c.zoom = 1.0
	}
	if c.zoom > 45.0 {
		c.zoom = 45.0
	}
}

func (c *Camera) getViewMatrix() mgl32.Mat4 {
	// The lookAt matrix makes the camera viewpoint look at the given target
	cameraDirection := c.position.Add(c.front)

	return lookAt(
		c.position,
		cameraDirection,
		c.up,
	)
}

func (c *Camera) updateVectors() {
	// calculate new front vector
	front := mgl32.Vec3{
		float32(math.Cos(float64(mgl32.DegToRad(c.yaw))) * math.Cos(float64(mgl32.DegToRad(c.pitch)))),
		float32(math.Sin(float64(mgl32.DegToRad(c.pitch)))),
		float32(math.Sin(float64(mgl32.DegToRad(c.yaw))) * math.Cos(float64(mgl32.DegToRad(c.pitch)))),
	}
	c.front = front.Normalize()
	// re-calc right and up too
	// normalize the vectors, because their length gets closer to 0 the more you look up or down which results in slower movement.
	c.right = c.front.Cross(c.worldUp).Normalize()
	c.up = c.right.Cross(c.front).Normalize()
}

func lookAt(cameraPosition, cameraDirection, cameraUp mgl32.Vec3) mgl32.Mat4 {
	forward := cameraDirection.Sub(cameraPosition).Normalize()
	right := forward.Cross(cameraUp.Normalize()).Normalize()
	up := right.Cross(forward)
	rotation := mgl32.Mat4{
		right.X(), up.X(), -forward.X(), 0,
		right.Y(), up.Y(), -forward.Y(), 0,
		right.Z(), up.Z(), -forward.Z(), 0,
		0, 0, 0, 1,
	}
	translation := mgl32.Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		-cameraPosition.X(), -cameraPosition.Y(), -cameraPosition.Z(), 1,
	}

	return rotation.Mul4(translation)
}
