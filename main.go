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
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/lafriks/go-tiled"
	"github.com/solarlune/resolv"
	camera "github.com/tducasse/ebiten-camera"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

// define a mappath constant of the tiled tmx file
const mapPath = "level1.tmx"

// mapgame struct
type mapGame struct {
	Level          *tiled.Map               //pointer to tmx file
	tileHash       map[uint32]*ebiten.Image //map tile ids
	cameraView     *camera.Camera           //pointer to camera for what will be visible
	player         player                   //sprite for squirel
	playerImage    *ebiten.Image            ///squirl sprite
	drawOps        ebiten.DrawImageOptions
	acorns         []*acorn     //slice for acorns
	chocolates     []*chocolate //slice for chocolates
	acornImage     *ebiten.Image
	chocolateImage *ebiten.Image
	gateImage      *ebiten.Image
	score          int // number of acorns collected
	showGate       bool
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

// chocolate struct
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

// create a new chocolate at a random location
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

	scale := 0.05 // scale used for player and acorn images

	// player center coordinates
	playerCenterX := float64(m.player.x) + float64(m.playerImage.Bounds().Dx())*scale/2
	playerCenterY := float64(m.player.y) + float64(m.playerImage.Bounds().Dy())*scale/2

	// check acorn collection
	newAcorns := m.acorns[:0]
	for _, a := range m.acorns {
		acornCenterX := a.xLoc + float64(a.pict.Bounds().Dx())*scale/2
		acornCenterY := a.yLoc + float64(a.pict.Bounds().Dy())*scale/2

		dx := playerCenterX - acornCenterX
		dy := playerCenterY - acornCenterY
		distanceSquared := dx*dx + dy*dy

		// collision radius (half width of player + half width of acorn)
		radius := (float64(m.playerImage.Bounds().Dx())*scale + float64(a.pict.Bounds().Dx())*scale) / 2

		if distanceSquared <= radius*radius {
			m.score++
		} else {
			newAcorns = append(newAcorns, a)
		}
	}
	m.acorns = newAcorns

	// show arrow/gate if 9 or more acorns collected
	if m.score >= 9 {
		m.showGate = true
	}

	// check collision with gate to move to next level
	if m.showGate && m.gateImage != nil {
		playerHB := m.PlayerHitbox()
		gateHB := m.GateHitbox()

		if len(playerHB.Intersection(gateHB).Intersections) > 0 {
			// load level2 tmx
			gameMap, err := tiled.LoadFile("level2.tmx")
			if err != nil {
				log.Fatal("Failed to load level2:", err)
			}
			m.Level = gameMap
			m.tileHash = makeEbitenImagesFromMap(*gameMap)

			// reset player position if needed
			m.player.x = 0
			m.player.y = 0

			// hide gate until next criteria
			m.showGate = false
		}
	}

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

	//draw the acorns
	for _, a := range m.acorns {
		acornOps := ebiten.DrawImageOptions{}
		acornOps.GeoM.Scale(0.01, 0.01)
		acornOps.GeoM.Translate(a.xLoc, a.yLoc)
		world.DrawImage(a.pict, &acornOps)
	}

	//draw the chocolates
	for _, c := range m.chocolates {
		chocolateOps := ebiten.DrawImageOptions{}
		chocolateOps.GeoM.Scale(0.01, 0.01)
		chocolateOps.GeoM.Translate(c.xLoc, c.yLoc)
		world.DrawImage(c.pict, &chocolateOps)
	}

	//draw gate in bottom right only when score is at 9 so the player can go to next level
	if m.score >= 9 && m.gateImage != nil {
		gateOps := ebiten.DrawImageOptions{}
		gateOps.GeoM.Scale(0.1, 0.1)
		gateOps.GeoM.Translate(1220, 1220)
		world.DrawImage(m.gateImage, &gateOps)
	}

	playerOps := ebiten.DrawImageOptions{}
	playerOps.GeoM.Scale(0.05, 0.05)                                   //scale player down to be reaosnable
	playerOps.GeoM.Translate(float64(m.player.x), float64(m.player.y)) //move player to its current x or y
	world.DrawImage(m.playerImage, &playerOps)                         //draw player into map

	m.cameraView.Draw(world, screen) //show visible part of screen on map

	// draw the score at top right of screen so player knows how many acrons they have
	drawFace := text.NewGoXFace(basicfont.Face7x13)
	textOpts := &text.DrawOptions{
		DrawImageOptions: ebiten.DrawImageOptions{},
		LayoutOptions:    text.LayoutOptions{},
	}
	textOpts.GeoM.Reset()
	textOpts.GeoM.Scale(3.0, 3.0)
	screenW, _ := ebiten.WindowSize()
	// place near top-right; adjust X offset so it is visible and not cut off
	textOpts.GeoM.Translate(float64(screenW-200), 30)
	textOpts.ColorScale.ScaleWithColor(colornames.Red)
	text.Draw(screen, fmt.Sprintf("Acorns: %d", m.score), drawFace, textOpts)
}

// define screen layout
func (m *mapGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

// hitbox for player
func (m *mapGame) PlayerHitbox() *resolv.ConvexPolygon {
	scale := 0.05
	w := float64(m.playerImage.Bounds().Dx()) * scale
	h := float64(m.playerImage.Bounds().Dy()) * scale
	x := float64(m.player.x) + w/2
	y := float64(m.player.y) + h/2
	return resolv.NewRectangle(x-w/2, y-h/2, w, h)
}

// hitbox for gate
func (m *mapGame) GateHitbox() *resolv.ConvexPolygon {
	scaleFactor := 0.05
	w := float64(m.gateImage.Bounds().Dx()) * scaleFactor
	h := float64(m.gateImage.Bounds().Dy()) * scaleFactor
	return resolv.NewRectangle(1250, 1250, w, h)
}

// hitbox for acorn
func (a *acorn) Hitbox() *resolv.ConvexPolygon {
	scale := 0.05
	w := float64(a.pict.Bounds().Dx()) * scale
	h := float64(a.pict.Bounds().Dy()) * scale
	x := a.xLoc + w/2
	y := a.yLoc + h/2
	return resolv.NewRectangle(x-w/2, y-h/2, w, h)
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

	// Load chocolate image
	chocolateImg, _, err := ebitenutil.NewImageFromFile("chocolate.png")
	if err != nil {
		log.Fatal("Failed to load chocolate image:", err)
	}

	// Load gate image
	gateImg, _, err := ebitenutil.NewImageFromFile("gate.png")
	if err != nil {
		log.Fatal("Failed to load gate image:", err)
	}

	// initialize acorns on map
	acornList := make([]*acorn, 0)
	for i := 0; i < 15; i++ {
		acornList = append(acornList, NewAcorn(gameMap.Width*gameMap.TileWidth, gameMap.Height*gameMap.TileHeight, acornImg))
	}
	//init chocolate on map
	chocolateList := make([]*chocolate, 0)
	for i := 0; i < 5; i++ {
		chocolateList = append(chocolateList, NewChocolate(gameMap.Width*gameMap.TileWidth, gameMap.Height*gameMap.TileHeight, chocolateImg))
	}

	// initialize the game
	game := &mapGame{
		Level:          gameMap,
		tileHash:       ebitenImageMap,
		cameraView:     ourCamera,
		player:         player{x: 0, y: 0},
		playerImage:    playerImg,
		acorns:         acornList,
		chocolates:     chocolateList,
		acornImage:     acornImg,
		chocolateImage: chocolateImg,
		gateImage:      gateImg,
		score:          0,
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
