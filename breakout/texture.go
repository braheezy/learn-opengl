package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Texture2D struct {
	// holds the ID of the texture object, used for all texture operations to reference to this particular texture
	ID uint32
	// texture image dimensions
	// width and height of loaded image in pixels
	Width, Height int32
	// texture Format
	// format of texture object
	Internal_Format int32
	// format of loaded image
	Image_Format uint32
	// texture configuration
	// wrapping mode on S axis
	Wrap_S int32
	// wrapping mode on T axis
	Wrap_T int32
	// filtering mode if texture pixels < screen pixels
	Filter_Min int32
	// filtering mode if texture pixels > screen pixels
	Filter_Max int32
}

func NewTexture() *Texture2D {
	t := Texture2D{
		Internal_Format: gl.RGB,
		Image_Format:    gl.RGB,
		Wrap_S:          gl.REPEAT,
		Wrap_T:          gl.REPEAT,
		Filter_Min:      gl.LINEAR,
		Filter_Max:      gl.LINEAR,
	}
	gl.GenTextures(1, &t.ID)
	return &t
}

func (tex *Texture2D) Generate(width, height int32, pixelData []byte) {
	tex.Width = width
	tex.Height = height

	gl.BindTexture(gl.TEXTURE_2D, tex.ID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, tex.Internal_Format, width, height, 0, tex.Image_Format, gl.UNSIGNED_BYTE, gl.Ptr(pixelData))
	// set Texture wrap and filter modes
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, tex.Wrap_S)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, tex.Wrap_T)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, tex.Filter_Min)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, tex.Filter_Max)
	// unbind texture
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (tex *Texture2D) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, tex.ID)
}
