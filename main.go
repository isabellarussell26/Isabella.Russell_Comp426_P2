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
	score          int           // number of acorns collected
	showGate       bool          // whether gate should show
	gameOver       bool          // flag for game over state
	npc1           *npc          //pointer to npc1
	npc2           *npc          //pointer to npc2
	npc1Image      *ebiten.Image //to draw npcs
	npc2Image      *ebiten.Image
	showNPCs       bool //for if the NPCs will be on screen
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

// npc struct
type npc struct {
	pict  *ebiten.Image //pointer to image
	x, y  float64       //current position
	dir   int           // left or right
	speed float64
	minX  float64 // min movement
	maxX  float64 // max movement
	time  int     //for npc movement - used ai
}

// update NPC movement- used ai
func (n *npc) Update() {
	n.time++ //increment time
	if n.minX != n.maxX {
		n.x += n.speed * float64(n.dir) //move npc on x axis
		//when npc hits max or min bound, will go other way
		if n.x < n.minX || n.x > n.maxX {
			n.dir *= -1
		}
	}
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
	// stop all updates if game over
	if m.gameOver {
		return nil
	}

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
		acornCenterX := a.xLoc + float64(a.pict.Bounds().Dx())*0.01/2 //center of acorn so hitbox when its actually touching the acron
		acornCenterY := a.yLoc + float64(a.pict.Bounds().Dy())*0.01/2
		dx := playerCenterX - acornCenterX //x distance from acorn
		dy := playerCenterY - acornCenterY //y distance form acorn
		distanceSquared := dx*dx + dy*dy   //used ai- "squared distance between player and acron"

		// collision radius (half width of player + half width of acorn)
		radius := (float64(m.playerImage.Bounds().Dx())*scale + float64(a.pict.Bounds().Dx())*0.01) / 2

		if distanceSquared <= radius*radius {
			m.score++ //if the distance player is from acron is a coliision- increase score
		} else {
			newAcorns = append(newAcorns, a) //if not collision keep it in same spot
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
				log.Fatal("Failed to load level 2", err)
			}
			m.Level = gameMap                              //assign new map
			m.tileHash = makeEbitenImagesFromMap(*gameMap) //convert tiles in new map

			// reset player position if needed
			m.player.x = 0
			m.player.y = 0

			// hide gate until next criteria
			m.showGate = false

			// enable NPCs
			m.showNPCs = true
		}
	}

	// check collision with chocolate to trigger game over
	for _, c := range m.chocolates {
		if len(m.PlayerHitbox().Intersection(c.Hitbox()).Intersections) > 0 {
			m.gameOver = true
			break
		}
	}

	// update NPCs if visible
	if m.showNPCs {
		if m.npc1 != nil {
			m.npc1.Update()
		}
		if m.npc2 != nil {
			m.npc2.Update()
		}
	}

	return nil
}

func (m *mapGame) Draw(screen *ebiten.Image) {
	// draw game over screen if triggered
	if m.gameOver {
		screen.Fill(colornames.Black) //backgroudn black
		drawFace := text.NewGoXFace(basicfont.Face7x13)
		textOpts := &text.DrawOptions{}
		textOpts.GeoM.Reset()
		textOpts.GeoM.Scale(6.0, 6.0) //text size
		screenW, screenH := ebiten.WindowSize()
		textOpts.GeoM.Translate(float64(screenW)/2-150, float64(screenH)/2-40)
		textOpts.ColorScale.ScaleWithColor(colornames.Red) //red text
		text.Draw(screen, "Game Over", drawFace, textOpts)
		return
	}

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

	// draw NPCs on level 2
	if m.showNPCs {
		if m.npc1 != nil {
			n1Ops := ebiten.DrawImageOptions{}
			n1Ops.GeoM.Scale(0.40, 0.40) //size up to be visible
			n1Ops.GeoM.Translate(m.npc1.x, m.npc1.y)
			world.DrawImage(m.npc1.pict, &n1Ops)
		}
		if m.npc2 != nil {
			n2Ops := ebiten.DrawImageOptions{}
			n2Ops.GeoM.Scale(0.12, 0.12) //size up to be visible
			n2Ops.GeoM.Translate(m.npc2.x, m.npc2.y)
			world.DrawImage(m.npc2.pict, &n2Ops)
		}
	}

	playerOps := ebiten.DrawImageOptions{}
	playerOps.GeoM.Scale(0.05, 0.05)                                   //scale player down to be reaosnable
	playerOps.GeoM.Translate(float64(m.player.x), float64(m.player.y)) //move player to its current x or y
	world.DrawImage(m.playerImage, &playerOps)                         //draw player into map

	m.cameraView.Draw(world, screen) //show visible part of screen on map

	// draw the score at top right of screen so player knows how many acrons they have
	drawFace := text.NewGoXFace(basicfont.Face7x13)
	textOpts := &text.DrawOptions{}
	textOpts.GeoM.Reset()
	textOpts.GeoM.Scale(3.0, 3.0)
	screenW, _ := ebiten.WindowSize()
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

// hitbox for chocolate
func (c *chocolate) Hitbox() *resolv.ConvexPolygon {
	scale := 0.01 // same as chocolate draw scale
	w := float64(c.pict.Bounds().Dx()) * scale
	h := float64(c.pict.Bounds().Dy()) * scale
	x := c.xLoc + w/2
	y := c.yLoc + h/2
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

	// Load NPC images
	npc1Img, _, err := ebitenutil.NewImageFromFile("NPC1.png")
	if err != nil {
		log.Fatal("Failed to load NPC1 image:", err)
	}
	npc2Img, _, err := ebitenutil.NewImageFromFile("NPC2.png")
	if err != nil {
		log.Fatal("Failed to load NPC2 image:", err)
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

	// create NPCs to only show on level2- used AI
	npc1 := &npc{pict: npc1Img, x: 400, y: 400, dir: 1, speed: 1.5, minX: 350, maxX: 700}
	npc2 := &npc{pict: npc2Img, x: 800, y: 600, dir: -1, speed: 1.2, minX: 750, maxX: 1000}

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
		npc1:           npc1,
		npc2:           npc2,
		npc1Image:      npc1Img,
		npc2Image:      npc2Img,
		showNPCs:       false, //npc not displayed until level2
	}

	fmt.Println("Tilesets loaded:", len(gameMap.Tilesets[0].Tiles))
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
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

	return idToImage
}
