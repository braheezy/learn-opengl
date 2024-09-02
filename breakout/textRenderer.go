package main

import (
	"embed"
	"fmt"
	"image"
	"log"
	"unsafe"

	"golang.org/x/image/font"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Embed all font files from the fonts/ directory
//
//go:embed fonts/*
var fontFiles embed.FS

// Character represents a glyph's texture and related data.
type Character struct {
	TextureID          uint32 // ID handle of the glyph texture
	width, height      int
	bearingH, bearingV int   // Offset from baseline to left/top of glyph
	Advance            int32 // Horizontal offset to advance to next glyph
}

// A renderer for rendering text displayed by a font loaded using the
// FreeType library. A single font is loaded, processed into a list of Character
// items for later rendering.
type TextRenderer struct {
	// holds a list of pre-compiled Characters
	characters map[rune]Character
	shader     *Shader
	VAO, VBO   uint32
	metrics    font.Metrics
}

func NewTextRenderer(width, height int) *TextRenderer {
	tr := TextRenderer{}
	// load and configure shader
	tr.shader = LoadShader("shaders/text_2d.vs", "shaders/text_2d.fs", "", "text")
	tr.shader.use()
	tr.shader.setMat4("projection", mgl32.Ortho2D(0.0, float32(width), float32(height), 0.0))
	tr.shader.setInt("text", 0)
	// configure VAO/VBO for texture quads
	gl.GenVertexArrays(1, &tr.VAO)
	gl.GenBuffers(1, &tr.VBO)
	gl.BindVertexArray(tr.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, tr.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(float32(0)))*6*4, gl.Ptr(nil), gl.DYNAMIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*int32(unsafe.Sizeof(float32(0))), gl.Ptr(nil))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return &tr
}

func (tr *TextRenderer) Load(fontPath string, fontSize int) {
	// Initialize freetype context and load the font
	fontBytes, err := fontFiles.ReadFile(fontPath)
	if err != nil {
		log.Fatalf("ERROR: Failed to load font file: %v", err)
	}

	// Parse the font
	ttf, err := opentype.Parse(fontBytes)
	if err != nil {
		log.Fatalf("ERROR: Failed to parse font: %v", err)
	}

	// Load the font face with a specific size
	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    float64(fontSize),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatalf("ERROR: Failed to create font face: %v", err)
	}

	tr.metrics = face.Metrics()

	// Initialize the characters map
	tr.characters = make(map[rune]Character)

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	// Load the first 128 ASCII characters
	for c := rune(32); c < 127; c++ {
		// Measure the bounding box and advance to determine the glyph size
		bounds, advance, ok := face.GlyphBounds(c)
		if !ok {
			log.Printf("failed glyph bounds for %v", c)
		}
		width := (bounds.Max.X - bounds.Min.X).Ceil()
		height := (bounds.Max.Y - bounds.Min.Y).Ceil()

		if c == ' ' {
			// Manually set the width and height to zero for the space character, but use the advance
			tr.characters[c] = Character{
				TextureID: 0, // No texture needed for space
				width:     0,
				height:    0,
				bearingH:  0,
				bearingV:  0,
				Advance:   int32(advance), // Use the advance value to move the cursor forward
			}
			continue
		}

		if width <= 0 || height <= 0 {
			fmt.Printf("Skipping character %v: invalid glyph size (width: %v, height: %v)\n", c, width, height)
			continue
		}

		dst := image.NewGray(image.Rect(0, 0, width, height))
		d := font.Drawer{
			Dst:  dst,
			Src:  image.White,
			Face: face,
			Dot:  fixed.P(-bounds.Min.X.Ceil(), -bounds.Min.Y.Ceil()),
		}
		d.DrawString(string(c))

		// Create a texture in OpenGL
		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		// Upload the texture data to OpenGL
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, int32(width), int32(height), 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(dst.Pix))

		// Set texture parameters to ensure proper glyph rendering
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		// Store the character with relevant OpenGL and bounding information
		tr.characters[c] = Character{
			TextureID: texture,
			width:     width,
			height:    height,
			bearingH:  bounds.Min.X.Ceil(),
			bearingV:  (bounds.Max.Y - bounds.Min.Y).Ceil(),
			Advance:   int32(advance),
		}
	}

	gl.BindTexture(gl.TEXTURE_2D, 0)
	// Cleanup resources
	face.Close()
}

func (tr *TextRenderer) RenderText(text string, x, y, scale float32, color mgl32.Vec3) {
	// activate corresponding render state
	tr.shader.use()
	tr.shader.setVec3("textColor", color)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(tr.VAO)

	// Iterate through all characters in the text string
	for _, char := range text {
		ch, ok := tr.characters[char]
		if !ok {
			fmt.Printf("Skipping %v\n", ch)
			continue // Skip characters that are not loaded
		}

		// Calculate position and size of the character quad
		xpos := x + float32(ch.bearingH)*scale
		ypos := y + (float32(tr.characters['H'].bearingV)-float32(ch.bearingV))*scale
		if char == 'p' || char == 'q' || char == 'g' || char == 'j' || char == 'y' {
			// These letters render weird
			magicNumber := float32(4.8)
			ypos = y + (float32(tr.characters['H'].bearingV)-float32(ch.bearingV)+magicNumber)*scale
		}

		w := float32(ch.width) * scale
		h := float32(ch.height) * scale

		vertices := []float32{
			xpos, ypos + h, 0.0, 1.0,
			xpos + w, ypos, 1.0, 0.0,
			xpos, ypos, 0.0, 0.0,

			xpos, ypos + h, 0.0, 1.0,
			xpos + w, ypos + h, 1.0, 1.0,
			xpos + w, ypos, 1.0, 0.0,
		}

		// Render glyph texture over quad
		gl.BindTexture(gl.TEXTURE_2D, ch.TextureID)

		// Update content of VBO memory
		gl.BindBuffer(gl.ARRAY_BUFFER, tr.VBO)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*int(unsafe.Sizeof(vertices[0])), gl.Ptr(vertices))

		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)

		// Now advance cursors for next glyph
		newX := float32(ch.Advance>>6) * scale
		x += newX
	}

	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}
