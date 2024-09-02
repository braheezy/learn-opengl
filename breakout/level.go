package main

import (
	"bufio"
	"embed"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

// Embed all level files from the levels/ directory
//
//go:embed levels/*
var levelFiles embed.FS

type GameLevel struct {
	bricks []GameObject
}

func LoadLevel(file string, levelWidth, levelHeight int) GameLevel {
	f, err := levelFiles.Open(file)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	var tileData [][]uint
	var lvl GameLevel

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		var row []uint

		// Split the line into words and convert to unsigned integers
		for _, value := range strings.Fields(line) {
			tileCode, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				fmt.Println("Error parsing tile code:", err)
				continue
			}
			row = append(row, uint(tileCode))
		}
		tileData = append(tileData, row)
	}
	if len(tileData) > 0 {
		lvl.initialize(tileData, levelWidth, levelHeight)
	}
	return lvl
}

func (lvl *GameLevel) initialize(tileData [][]uint, levelWidth, levelHeight int) {
	// calculate dimensions
	height := len(tileData)
	width := len(tileData[0])
	unitWidth := levelWidth / width
	unitHeight := levelHeight / height
	// initialize level tiles based on tileData
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// check block type from level data (2D level arryay)
			if tileData[y][x] == 1 {
				// solid
				pos := mgl32.Vec2{float32(unitWidth * x), float32(unitHeight * y)}
				size := mgl32.Vec2{float32(unitWidth), float32(unitHeight)}
				obj := GameObject{
					position: pos,
					size:     size,
					sprite:   GetTexture("block_solid"),
					color:    mgl32.Vec3{0.8, 0.8, 0.7},
				}
				obj.isSolid = true
				lvl.bricks = append(lvl.bricks, obj)
			} else if tileData[y][x] > 1 {
				color := white
				if tileData[y][x] == 2 {
					color = mgl32.Vec3{0.2, 0.6, 1.0}
				} else if tileData[y][x] == 3 {
					color = mgl32.Vec3{0.0, 0.7, 0.0}
				} else if tileData[y][x] == 4 {
					color = mgl32.Vec3{0.8, 0.8, 0.4}
				} else if tileData[y][x] == 5 {
					color = mgl32.Vec3{1.0, 0.5, 0.0}
				}
				pos := mgl32.Vec2{float32(unitWidth * x), float32(unitHeight * y)}
				size := mgl32.Vec2{float32(unitWidth), float32(unitHeight)}
				obj := GameObject{position: pos, size: size, sprite: GetTexture("block"), color: color}
				lvl.bricks = append(lvl.bricks, obj)
			}
		}
	}
}

func (lvl *GameLevel) Draw(renderer *SpriteRenderer) {
	for _, tile := range lvl.bricks {
		if !tile.destroyed {
			tile.Draw(renderer)
		}
	}
}

func (lvl *GameLevel) IsCompleted() bool {
	for _, tile := range lvl.bricks {
		if !tile.isSolid && !tile.destroyed {
			return false
		}
	}
	return true
}
