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
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// Model is a structure representing a 3D model
type Model struct {
	meshes       []*Mesh
	textureCache map[string]uint32
}

// LoadModel loads a GLTF model from the specified file path
func (m *Model) LoadModel(path string) {
	doc, err := gltf.Open(path)
	if err != nil {
		log.Fatalf("failed to load model: %v", err)
	}

	m.textureCache = make(map[string]uint32)
	directory := filepath.Dir(path)

	for _, mesh := range doc.Meshes {
		for _, prim := range mesh.Primitives {
			vertices, indices, textures := m.processPrimitive(doc, prim, directory)
			m.meshes = append(m.meshes, NewMesh(vertices, indices, textures))
		}
	}
}

func (m *Model) processPrimitive(doc *gltf.Document, prim *gltf.Primitive, directory string) ([]Vertex, []uint32, []Texture) {
	var vertices []Vertex
	var indices []uint32

	// Process vertices
	positions, _ := modeler.ReadPosition(doc, doc.Accessors[prim.Attributes["POSITION"]], nil)
	normals, _ := modeler.ReadNormal(doc, doc.Accessors[prim.Attributes["NORMAL"]], nil)
	texCoords, _ := modeler.ReadTextureCoord(doc, doc.Accessors[prim.Attributes["TEXCOORD_0"]], nil)

	for i := 0; i < len(positions); i++ {
		vertex := Vertex{
			position:  mgl32.Vec3{positions[i][0], positions[i][1], positions[i][2]},
			normal:    mgl32.Vec3{normals[i][0], normals[i][1], normals[i][2]},
			texCoords: mgl32.Vec2{texCoords[i][0], texCoords[i][1]},
		}
		vertices = append(vertices, vertex)
	}

	// Process indices
	indices, _ = modeler.ReadIndices(doc, doc.Accessors[*prim.Indices], nil)

	// Process textures
	var textures []Texture
	for _, material := range doc.Materials {
		if material.PBRMetallicRoughness != nil && material.PBRMetallicRoughness.BaseColorTexture != nil {
			textureID := m.loadTextureFromGLTF(doc, &material.PBRMetallicRoughness.BaseColorTexture.Index, directory)
			textures = append(textures, Texture{id: textureID, _type: "texture_diffuse"})
		}
		if material.NormalTexture != nil {
			textureID := m.loadTextureFromGLTF(doc, material.NormalTexture.Index, directory)
			textures = append(textures, Texture{id: textureID, _type: "texture_normal"})
		}
		if material.EmissiveTexture != nil {
			textureID := m.loadTextureFromGLTF(doc, &material.EmissiveTexture.Index, directory)
			textures = append(textures, Texture{id: textureID, _type: "texture_emissive"})
		}
		if material.OcclusionTexture != nil {
			textureID := m.loadTextureFromGLTF(doc, material.OcclusionTexture.Index, directory)
			textures = append(textures, Texture{id: textureID, _type: "texture_occlusion"})
		}
	}

	return vertices, indices, textures
}

func (m *Model) loadTextureFromGLTF(doc *gltf.Document, texIndex *uint32, directory string) uint32 {
	texture := doc.Textures[*texIndex]
	gltfImage := doc.Images[*texture.Source]

	// Check if texture is already loaded
	if textureID, found := m.textureCache[gltfImage.URI]; found {
		return textureID
	}

	// Load image file
	file, err := os.Open(filepath.Join(directory, gltfImage.URI))
	if err != nil {
		log.Fatalf("failed to open texture file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("failed to decode texture image: %v", err)
	}

	textureID := loadTextureFromImage(img)

	// Store the texture ID in the cache
	m.textureCache[gltfImage.URI] = textureID

	return textureID
}

func loadTextureFromImage(img image.Image) uint32 {
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	width := int32(img.Bounds().Dx())
	height := int32(img.Bounds().Dy())
	rgba := make([]uint8, width*height*4)
	index := 0
	for y := img.Bounds().Max.Y - 1; y >= img.Bounds().Min.Y; y-- { // Flip the image vertically
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rgba[index] = uint8(r >> 8)
			rgba[index+1] = uint8(g >> 8)
			rgba[index+2] = uint8(b >> 8)
			rgba[index+3] = uint8(a >> 8)
			index += 4
		}
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return textureID
}

// Draw renders the model using the given shader
func (m *Model) Draw(shader *Shader) {
	for _, mesh := range m.meshes {
		mesh.Draw(shader)
	}
}
