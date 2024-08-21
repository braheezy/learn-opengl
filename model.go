package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/udhos/gwob"
)

type Model struct {
	texturesLoaded  map[string]Texture // to avoid loading the same texture more than once
	meshes          []Mesh
	directory       string
	gammaCorrection bool
}

// LoadModel loads the model from the given path.
func LoadModel(path string) *Model {
	m := &Model{}
	dirName := filepath.Base(path)
	objPath := filepath.Join(path, dirName+".obj")
	mtlPath := filepath.Join(path, dirName+".mtl")
	m.directory = path

	// Initialize texturesLoaded map if not already initialized
	if m.texturesLoaded == nil {
		m.texturesLoaded = make(map[string]Texture)
	}

	// Load the OBJ file
	options := &gwob.ObjParserOptions{}
	obj, err := gwob.NewObjFromFile(objPath, options)
	if err != nil {
		log.Fatalf("Failed to load OBJ model: %v", err)
	}

	// Load the MTL file
	mtlLib, err := gwob.ReadMaterialLibFromFile(mtlPath, nil)
	if err != nil {
		log.Fatalf("Failed to load MTL file: %v", err)
	}

	// Process each group (mesh) in the OBJ file
	for _, group := range obj.Groups {
		mesh := m.processMesh(group, obj, mtlLib)
		m.meshes = append(m.meshes, mesh)
	}

	return m
}

// processMesh processes a group in the OBJ file and converts it into a Mesh structure.
func (m *Model) processMesh(group *gwob.Group, obj *gwob.Obj, mtlLib gwob.MaterialLib) Mesh {
	var vertices []Vertex
	var indices []uint32
	var textures []Texture

	// Process vertices and indices
	for i := group.IndexBegin; i < group.IndexBegin+group.IndexCount; i++ {
		index := obj.Indices[i]

		// Calculate the positions of vertex attributes within the Coord array
		posOffset := index*obj.StrideSize/4 + obj.StrideOffsetPosition
		texOffset := index*obj.StrideSize/4 + obj.StrideOffsetTexture/4
		normOffset := index*obj.StrideSize/4 + obj.StrideOffsetNormal/4

		// Safely access vertex data, handle cases where data might not be available
		var position mgl32.Vec3
		if posOffset+2 < len(obj.Coord) {
			position = mgl32.Vec3{
				obj.Coord[posOffset],
				obj.Coord[posOffset+1],
				obj.Coord[posOffset+2],
			}
		}

		var texCoords mgl32.Vec2
		if obj.TextCoordFound && texOffset+1 < len(obj.Coord) {
			texCoords = mgl32.Vec2{
				obj.Coord[texOffset],
				1.0 - obj.Coord[texOffset+1],
			}
		}

		var normal mgl32.Vec3
		if obj.NormCoordFound && normOffset+2 < len(obj.Coord) {
			normal = mgl32.Vec3{
				obj.Coord[normOffset],
				obj.Coord[normOffset+1],
				obj.Coord[normOffset+2],
			}
		}

		vertex := Vertex{
			Position:  position,
			TexCoords: texCoords,
			Normal:    normal,
		}
		vertices = append(vertices, vertex)
		indices = append(indices, uint32(len(vertices)-1))
	}

	// Process materials (textures)
	if group.Usemtl != "" {
		if material, exists := mtlLib.Lib[group.Usemtl]; exists {
			textures = append(textures, m.loadMaterialTextures(material)...)
		}
	}

	mesh := Mesh{
		vertices: vertices,
		indices:  indices,
		textures: textures,
	}
	mesh.setupMesh()

	return mesh
}

// loadMaterialTextures loads textures for a given material.
func (m *Model) loadMaterialTextures(material *gwob.Material) []Texture {
	var textures []Texture

	// Map the material fields to the corresponding texture types
	textureMappings := map[string]string{
		material.MapKd: "texture_diffuse",
		material.MapKs: "texture_specular",
		material.Bump:  "texture_normal",
		material.MapD:  "texture_height",
	}

	for texPath, texType := range textureMappings {
		if texPath != "" { // Ensure the texture path is not empty
			if texture, loaded := m.texturesLoaded[texPath]; loaded {
				textures = append(textures, texture)
			} else {
				texture := m.loadTexture(texPath, texType)
				textures = append(textures, texture)
				m.texturesLoaded[texPath] = texture
			}
		}
	}

	return textures
}

// loadTexture loads a texture from the file.
func (m *Model) loadTexture(path, texType string) Texture {
	textureID := TextureFromFile(path, m.directory, m.gammaCorrection)

	return Texture{
		ID:   textureID,
		Type: texType,
		Path: path,
	}
}

// Draw renders the model using the provided shader.
func (m *Model) Draw(shader Shader) {
	for _, mesh := range m.meshes {
		mesh.Draw(shader)
	}
}

// TextureFromFile loads a texture from a file and returns the OpenGL texture ID.
func TextureFromFile(path, directory string, gamma bool) uint32 {
	filename := filepath.Join(directory, path)

	// Generate and bind a new texture ID
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	// Open and decode the texture image file
	textureFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open texture file: %v", err)
	}
	defer textureFile.Close()

	// Decode the image (JPEG, PNG, etc.)
	textureImage, _, err := image.Decode(textureFile)
	if err != nil {
		log.Fatalf("Failed to decode texture file [%s]: %v", filename, err)
	}

	// Determine the number of components and the OpenGL format
	bounds := textureImage.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	var format, internalFormat int32
	var pixelData []byte

	switch img := textureImage.(type) {
	case *image.Gray:
		format = gl.RED
		internalFormat = gl.RED
		pixelData = make([]byte, width*height)
		index := 0
		for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r := img.GrayAt(x, y).Y
				pixelData[index] = r
				index++
			}
		}
	case *image.RGBA:
		// Check if the alpha channel is used
		hasAlpha := false
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, a := img.At(x, y).RGBA()
				if a != 0xFFFF {
					hasAlpha = true
					break
				}
			}
			if hasAlpha {
				break
			}
		}

		if hasAlpha {
			format = gl.RGBA
			internalFormat = gl.RGBA
			pixelData = make([]byte, width*height*4)
			index := 0
			for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, a := img.At(x, y).RGBA()
					pixelData[index] = byte(r >> 8)
					pixelData[index+1] = byte(g >> 8)
					pixelData[index+2] = byte(b >> 8)
					pixelData[index+3] = byte(a >> 8)
					index += 4
				}
			}
		} else {
			// Treat as RGB if alpha is not used
			format = gl.RGB
			internalFormat = gl.RGB
			pixelData = make([]byte, width*height*3)
			index := 0
			for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					pixelData[index] = byte(r >> 8)
					pixelData[index+1] = byte(g >> 8)
					pixelData[index+2] = byte(b >> 8)
					index += 3
				}
			}
		}
	case *image.NRGBA:
		format = gl.RGBA
		internalFormat = gl.RGBA
		pixelData = make([]byte, width*height*4)
		index := 0
		for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				pixelData[index] = byte(r >> 8)
				pixelData[index+1] = byte(g >> 8)
				pixelData[index+2] = byte(b >> 8)
				pixelData[index+3] = byte(a >> 8)
				index += 4
			}
		}
	default:
		// Handle RGB images without alpha channel
		format = gl.RGB
		internalFormat = gl.RGB
		pixelData = make([]byte, width*height*3)
		index := 0
		for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := textureImage.At(x, y).RGBA()
				pixelData[index] = byte(r >> 8)
				pixelData[index+1] = byte(g >> 8)
				pixelData[index+2] = byte(b >> 8)
				index += 3
			}
		}
	}

	// Upload texture data to the GPU
	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat, int32(width), int32(height), 0, uint32(format), gl.UNSIGNED_BYTE, gl.Ptr(pixelData))

	// Generate mipmaps
	gl.GenerateMipmap(gl.TEXTURE_2D)

	// Set texture wrapping/filtering options
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return textureID
}
