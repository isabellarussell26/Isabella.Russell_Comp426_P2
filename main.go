package main

//import all necessary imports
import (
	"fmt"
	"log"
	"math/rand"
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
	acorns      []*acorn     //slice for acorns
	chocolate   []*chocolate // slice for chocolate
}

// player struct
type player struct {
	x, y int //players x and y locations
}

// acorn struct
type acorn struct {
	pict *ebiten.Image
	xLoc float64
	yLoc float64
}
type chocolate struct {
	pict *ebiten.Image
	xLoc float64
	yLoc float64
}

// create a new acorn at a random location
func NewAcorn(maxX, maxY int, image *ebiten.Image) *acorn {
	return &acorn{
		pict: image,
		xLoc: float64(rand.Intn(maxX)),
		yLoc: float64(rand.Intn(maxY)),
	}
}

// create new chocolate at rand location
func NewChocolate(maxX, maxY int, image *ebiten.Image) *chocolate {
	return &chocolate{
		pict: image,
		xLoc: float64(rand.Intn(maxX)),
		yLoc: float64(rand.Intn(maxY)),
	}
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
	//up til 0 to keep on map
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

	//create world image to draw map and objects
	world := ebiten.NewImage(m.Level.Width*m.Level.TileWidth, m.Level.Height*m.Level.TileHeight)
	tileDrawOps := ebiten.DrawImageOptions{}

	//draws the tiled map on the screen
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

	//draw the acorns on top of the map
	for _, a := range m.acorns {
		acornOps := ebiten.DrawImageOptions{}
		acornOps.GeoM.Scale(0.009, 0.009) //scale acorn down to be reasonable
		acornOps.GeoM.Translate(a.xLoc, a.yLoc)
		world.DrawImage(a.pict, &acornOps)
	}
	for _, c := range m.chocolate {
		chocolateOps := ebiten.DrawImageOptions{}
		chocolateOps.GeoM.Scale(0.005, 0.005) //scale acorn down to be reasonable
		chocolateOps.GeoM.Translate(c.xLoc, c.yLoc)
		world.DrawImage(c.pict, &chocolateOps)
	}

	//draw player on top of map and acorns
	playerOps := ebiten.DrawImageOptions{}
	playerOps.GeoM.Scale(0.08, 0.08)                                   //scale player down to be reasonable
	playerOps.GeoM.Translate(float64(m.player.x), float64(m.player.y)) //move player to its current x or y
	world.DrawImage(m.playerImage, &playerOps)                         //draw player into map

	//show visible part of screen on map
	m.cameraView.Draw(world, screen)
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
	playerImg, _, err := ebitenutil.NewImageFromFile("playerRight.png")
	if err != nil {
		log.Fatal("Failed to load player image:", err)
	}

	// Load acorn image
	acornImg, _, err := ebitenutil.NewImageFromFile("acorn.png")
	if err != nil {
		log.Fatal("Failed to load acorn image:", err)
	}

	//initialize 15 acorns at random positions
	acornList := make([]*acorn, 0)
	for i := 0; i < 15; i++ {
		x := rand.Intn(800)
		y := rand.Intn(800)
		acornList = append(acornList, &acorn{pict: acornImg, xLoc: float64(x), yLoc: float64(y)})
	}
	//load chcolate image
	chocolateImg, _, err := ebitenutil.NewImageFromFile("chocolate.png")
	if err != nil {
		log.Fatal("Failed to load chocolate image:", err)
	}
	//init 5 chocolate items
	chocolateList := make([]*chocolate, 0)
	for i := 0; i < 5; i++ {
		x := rand.Intn(800)
		y := rand.Intn(800)
		chocolateList = append(chocolateList, &chocolate{pict: chocolateImg, xLoc: float64(x), yLoc: float64(y)})
	}

	// initialize the game
	game := &mapGame{
		Level:       gameMap,
		tileHash:    ebitenImageMap,
		cameraView:  ourCamera,
		player:      player{x: 0, y: 0},
		playerImage: playerImg,
		acorns:      acornList,
		chocolate:   chocolateList,
	}

	fmt.Println("Tilesets loaded:", len(gameMap.Tilesets[0].Tiles))
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// load tiles from tilemap, create map, loop through all of the tiles and create the image
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
