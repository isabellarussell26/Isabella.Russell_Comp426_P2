package main

//import all necessary imports
import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/lafriks/go-tiled"
	camera "github.com/tducasse/ebiten-camera"
)

// define a mappath constant of the tiled tmx file
const mapPath = "level1.tmx"

// mapgame struct
type mapGame struct {
	Level       *tiled.Map               //pointer to tmx file
	tileHash    map[uint32]*ebiten.Image //map tile ids
	cameraView  *camera.Camera           //pointer to camera for what will be visible
	player      player                   //sprite for squirel
	playerImage *ebiten.Image            ///squirl sprite
	drawOps     ebiten.DrawImageOptions
}

// player struct
type player struct {
	x, y int //players x and y locations
}

func (m *mapGame) Update() error {
	//moving left so player doesnt go off screen
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && m.player.x > 0 {
		m.player.x -= 5
	}
	//moving right if under 1250 so player doesnt go off map
	if ebiten.IsKeyPressed(ebiten.KeyRight) && m.player.x < 1250 {
		m.player.x += 5
	}
	//up til 0 to keeo on map
	if ebiten.IsKeyPressed(ebiten.KeyUp) && m.player.y > 0 {
		m.player.y -= 5
	}
	//down til 1250 to keep on map
	if ebiten.IsKeyPressed(ebiten.KeyDown) && m.player.y < 1250 {
		m.player.y += 5
	}

	//camera follows the player when player moves camera moves
	m.cameraView.Follow.W = m.player.x
	m.cameraView.Follow.H = m.player.y

	return nil
}

func (m *mapGame) Draw(screen *ebiten.Image) {
	m.drawOps.GeoM.Reset()

	world := ebiten.NewImage(m.Level.Width*m.Level.TileWidth, m.Level.Height*m.Level.TileHeight)
	tileDrawOps := ebiten.DrawImageOptions{}

	//draws the tiledmpa on the screen
	for tileY := 0; tileY < m.Level.Height; tileY++ {
		for tileX := 0; tileX < m.Level.Width; tileX++ {
			tileDrawOps.GeoM.Reset()
			tileDrawOps.GeoM.Translate(float64(tileX*m.Level.TileWidth), float64(tileY*m.Level.TileHeight))

			tile := m.Level.Layers[0].Tiles[tileY*m.Level.Width+tileX]
			img := m.tileHash[tile.ID]
			if img != nil {
				world.DrawImage(img, &tileDrawOps)
			}
		}
	}

	playerOps := ebiten.DrawImageOptions{}
	playerOps.GeoM.Scale(0.015, 0.015)                                 //scale player down to be reaosnable
	playerOps.GeoM.Translate(float64(m.player.x), float64(m.player.y)) //move player to its current x or y
	world.DrawImage(m.playerImage, &playerOps)                         //draw player into map

	m.cameraView.Draw(world, screen) //show visible part of screen on map
}

// define screen layout
func (m *mapGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	//load tile map
	gameMap, err := tiled.LoadFile(mapPath)
	if err != nil {
		fmt.Printf("Error parsing map: %s\n", err.Error())
		os.Exit(2)
	}
	//set window size
	ebiten.SetWindowSize(1000, 1000)
	ebiten.SetWindowTitle("Squirrel Game")

	//convert tile images and set up camera
	ebitenImageMap := makeEbitenImagesFromMap(*gameMap)
	ourCamera := camera.Init(0, 0)

	// Load the squirel image
	playerImg, _, err := ebitenutil.NewImageFromFile("player.png")
	if err != nil {
		log.Fatal("Failed to load player image:", err)
	}

	// initialize the game
	game := &mapGame{
		Level:       gameMap,
		tileHash:    ebitenImageMap,
		cameraView:  ourCamera,
		player:      player{x: 0, y: 0},
		playerImage: playerImg,
	}

	fmt.Println("Tilesets loaded:", len(gameMap.Tilesets[0].Tiles))
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// load tiles from tilemap, creatre map.loop through all of the riles and create the image
func makeEbitenImagesFromMap(tiledMap tiled.Map) map[uint32]*ebiten.Image {
	idToImage := make(map[uint32]*ebiten.Image)

	for _, tile := range tiledMap.Tilesets[0].Tiles {
		imgPath := tile.Image.Source
		if _, err := os.Stat(imgPath); os.IsNotExist(err) {
			imgPath = filepath.Join("tiles", filepath.Base(imgPath))
		}

		ebitenTile, _, err := ebitenutil.NewImageFromFile(imgPath)
		if err != nil {
			fmt.Println("Error loading tile image:", imgPath, err)
			continue
		}
		idToImage[tile.ID] = ebitenTile
	}
	return idToImage //returns the map
}
