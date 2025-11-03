package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/lafriks/go-tiled"
)

const mapPath = "level1.tmx"

type mapGame struct {
	Level    *tiled.Map
	tileHash map[uint32]*ebiten.Image
}

func (m mapGame) Update() error {
	return nil
}

func (m mapGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return m.Level.Width * m.Level.TileWidth, m.Level.Height * m.Level.TileHeight
}

func main() {
	gameMap, err := tiled.LoadFile(mapPath)
	if err != nil {
		fmt.Printf("Error parsing map: %s\n", err.Error())
		os.Exit(2)
	}

	windowWidth := gameMap.Width * gameMap.TileWidth
	windowHeight := gameMap.Height * gameMap.TileHeight
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Tile Map Game")

	ebitenImageMap := makeEbitenImagesFromMap(*gameMap)

	oneLevelGame := mapGame{
		Level:    gameMap,
		tileHash: ebitenImageMap,
	}

	fmt.Println("Tilesets loaded:", len(gameMap.Tilesets[0].Tiles))

	if err := ebiten.RunGame(&oneLevelGame); err != nil {
		fmt.Println("Couldn't run game:", err)
	}
}

func makeEbitenImagesFromMap(tiledMap tiled.Map) map[uint32]*ebiten.Image {
	idToImage := make(map[uint32]*ebiten.Image)

	for _, tile := range tiledMap.Tilesets[0].Tiles {
		imgPath := tile.Image.Source

		if _, err := os.Stat(imgPath); os.IsNotExist(err) {
			imgPath = filepath.Join("tiles", filepath.Base(imgPath))
		}

		ebitenImageTile, _, err := ebitenutil.NewImageFromFile(imgPath)
		if err != nil {
			fmt.Println("Error loading tile image:", imgPath, err)
			continue
		}

		idToImage[tile.ID] = ebitenImageTile
	}

	return idToImage
}

func (game mapGame) Draw(screen *ebiten.Image) {
	drawOptions := ebiten.DrawImageOptions{}

	for tileY := 0; tileY < game.Level.Height; tileY++ {
		for tileX := 0; tileX < game.Level.Width; tileX++ {
			drawOptions.GeoM.Reset()

			TileXpos := float64(game.Level.TileWidth * tileX)
			TileYpos := float64(game.Level.TileHeight * tileY)
			drawOptions.GeoM.Translate(TileXpos, TileYpos)

			tileToDraw := game.Level.Layers[0].Tiles[tileY*game.Level.Width+tileX]
			ebitenTileToDraw := game.tileHash[tileToDraw.ID]

			if ebitenTileToDraw != nil {
				screen.DrawImage(ebitenTileToDraw, &drawOptions)
			}
		}
	}
}
