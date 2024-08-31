package main

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type ResourceManager struct {
	shaders  map[string]Shader
	textures map[string]Texture2D
}

// Global instance of the resource manager.
var manager = &ResourceManager{
	shaders:  make(map[string]Shader),
	textures: make(map[string]Texture2D),
}

func LoadShader(vShaderFile, fShaderFile, gShaderFile, name string) Shader {
	var err error
	manager.shaders[name], err = loadShaderFromFile(vShaderFile, fShaderFile, gShaderFile)
	if err != nil {
		log.Fatal(err)
	}
	return manager.shaders[name]
}

func GetShader(name string) Shader {
	return manager.shaders[name]
}

func LoadTexture(file string, alpha bool, name string) Texture2D {
	manager.textures[name] = loadTextureFromFile(file, alpha)
	return manager.textures[name]
}

func GetTexture(name string) Texture2D {
	return manager.textures[name]
}

func loadShaderFromFile(vShaderFile, fShaderFile, gShaderFile string) (Shader, error) {
	// * 1. Retrieve the vertex/fragment source code from file paths
	var vertexCode, fragmentCode, geometryCode string

	// Read vertex shader file
	vBytes, err := os.ReadFile(vShaderFile)
	if err != nil {
		fmt.Println("ERROR::SHADER: Failed to read vertex shader file:", err)
	}
	vertexCode = string(vBytes)

	// Read fragment shader file
	fBytes, err := os.ReadFile(fShaderFile)
	if err != nil {
		fmt.Println("ERROR::SHADER: Failed to read fragment shader file:", err)
	}
	fragmentCode = string(fBytes)

	// Read geometry shader file if provided
	if gShaderFile != "" {
		gBytes, err := os.ReadFile(gShaderFile)
		if err != nil {
			fmt.Println("ERROR::SHADER: Failed to read geometry shader file:", err)
		}
		geometryCode = string(gBytes)
	}

	// 2. Now create shader object from source code
	return NewShader(vertexCode, fragmentCode, geometryCode)
}

func loadTextureFromFile(file string, alpha bool) Texture2D {
	var texture Texture2D
	if alpha {
		texture.Internal_Format = gl.RGBA
		texture.Image_Format = gl.RGBA
	}
	pixelData, width, height := loadPixels(file)
	texture.Generate(width, height, pixelData)
	return texture
}

func loadPixels(filePath string) ([]byte, int32, int32) {

	// Open and decode the texture image file
	textureFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open texture file: %v", err)
	}
	defer textureFile.Close()

	textureImage, _, err := image.Decode(textureFile)
	if err != nil {
		log.Fatalf("Failed to decode texture file [%s]: %v", filePath, err)
	}

	// Determine the bounds and dimensions of the image
	bounds := textureImage.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	var pixelData []byte

	// Check for alpha channel presence
	hasAlpha := false
	if rgba, ok := textureImage.(*image.RGBA); ok {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, a := rgba.At(x, y).RGBA()
				if a != 0xFFFF {
					hasAlpha = true
					break
				}
			}
			if hasAlpha {
				break
			}
		}
	} else if _, ok := textureImage.(*image.NRGBA); ok {
		hasAlpha = true
	}

	// Parse pixel data based on whether alpha is present
	if hasAlpha {
		pixelData = make([]byte, width*height*4)
		index := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := textureImage.At(x, y).RGBA()
				pixelData[index] = byte(r >> 8)
				pixelData[index+1] = byte(g >> 8)
				pixelData[index+2] = byte(b >> 8)
				pixelData[index+3] = byte(a >> 8)
				index += 4
			}
		}
	} else {
		pixelData = make([]byte, width*height*3)
		index := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := textureImage.At(x, y).RGBA()
				pixelData[index] = byte(r >> 8)
				pixelData[index+1] = byte(g >> 8)
				pixelData[index+2] = byte(b >> 8)
				index += 3
			}
		}
	}

	return pixelData, int32(width), int32(height)
}
