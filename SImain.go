package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 640
	screenHeight = 480
	playerWidth  = 40
	playerHeight = 20
	alienWidth   = 30
	alienHeight  = 20
	bulletWidth  = 4
	bulletHeight = 10
	alienRows    = 5
	alienCols    = 11
)

type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StateLevelClear
	StateGameOver
	StateWon
)

type Entity struct {
	x, y          float64
	width, height float64
	active        bool
	animOffset    float64
}

type Game struct {
	state        GameState
	player       Entity
	bullets      []*Entity
	aliens       []*Entity
	alienDir     float64
	alienSpeed   float64
	score        int
	level        int
	lastShot     time.Time
	frameCounter int
	lives        int
}

func (g *Game) Update() error {
	g.frameCounter++
	switch g.state {
	case StateTitle:
		if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.state = StatePlaying
		}
	case StatePlaying:
		g.updatePlaying()
	case StateLevelClear:
		if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.nextLevel()
			g.state = StatePlaying
		}
	case StateGameOver, StateWon:
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.init()
			g.state = StatePlaying
		}
	}
	return nil
}

func (g *Game) updatePlaying() {
	// Player movement
	if (ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)) && g.player.x > 10 {
		g.player.x -= 5
	}
	if (ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)) && g.player.x < screenWidth-playerWidth-10 {
		g.player.x += 5
	}

	// Shooting
	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyW)) && time.Since(g.lastShot) > 350*time.Millisecond {
		g.bullets = append(g.bullets, &Entity{
			x:      g.player.x + playerWidth/2 - bulletWidth/2,
			y:      g.player.y - bulletHeight,
			width:  bulletWidth,
			height: bulletHeight,
			active: true,
		})
		g.lastShot = time.Now()
	}

	// Update bullets
	for _, b := range g.bullets {
		if b.active {
			b.y -= 8
			if b.y < 0 {
				b.active = false
			}

			// Collision with aliens
			for _, a := range g.aliens {
				if a.active && b.x < a.x+a.width && b.x+b.width > a.x && b.y < a.y+a.height && b.y+b.height > a.y {
					a.active = false
					b.active = false
					g.score += 10 * g.level
					break
				}
			}
		}
	}

	// Update aliens
	moveDown := false
	for _, a := range g.aliens {
		if a.active {
			a.x += g.alienDir * g.alienSpeed
			a.animOffset = math.Sin(float64(g.frameCounter)*0.1) * 2

			if a.x <= 10 || a.x >= screenWidth-alienWidth-10 {
				moveDown = true
			}
			if a.y+a.height >= g.player.y {
				g.lives--
				if g.lives <= 0 {
					g.state = StateGameOver
				} else {
					g.resetPositions()
				}
			}
		}
	}

	if moveDown {
		g.alienDir *= -1
		for _, a := range g.aliens {
			a.y += 15
		}
		g.alienSpeed += 0.1
	}

	// Check win condition
	allDead := true
	for _, a := range g.aliens {
		if a.active {
			allDead = false
			break
		}
	}
	if allDead {
		if g.level >= 5 {
			g.state = StateWon
		} else {
			g.state = StateLevelClear
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	// Background stars
	for i := 0; i < 50; i++ {
		sx := float32((i * 123) % screenWidth)
		sy := float32((i*456 + g.frameCounter/2) % screenHeight)
		vector.DrawFilledRect(screen, sx, sy, 1, 1, color.RGBA{100, 100, 100, 255}, true)
	}

	switch g.state {
	case StateTitle:
		ebitenutil.DebugPrintAt(screen, "SPACE INVADERS", screenWidth/2-50, screenHeight/2-40)
		ebitenutil.DebugPrintAt(screen, "Press ENTER to Start", screenWidth/2-60, screenHeight/2)
	case StatePlaying:
		g.drawPlaying(screen)
	case StateLevelClear:
		g.drawPlaying(screen)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("LEVEL %d CLEAR!", g.level), screenWidth/2-45, screenHeight/2-20)
		ebitenutil.DebugPrintAt(screen, "Press ENTER for Next Level", screenWidth/2-75, screenHeight/2+10)
	case StateGameOver:
		g.drawPlaying(screen)
		ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-30, screenHeight/2-20)
		ebitenutil.DebugPrintAt(screen, "Press R to Restart", screenWidth/2-55, screenHeight/2+10)
	case StateWon:
		g.drawPlaying(screen)
		ebitenutil.DebugPrintAt(screen, "CONGRATULATIONS! YOU SAVED EARTH!", screenWidth/2-110, screenHeight/2-20)
		ebitenutil.DebugPrintAt(screen, "Press R to Restart", screenWidth/2-55, screenHeight/2+10)
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Player
	vector.DrawFilledRect(screen, float32(g.player.x), float32(g.player.y+5), float32(g.player.width), float32(g.player.height-5), color.RGBA{0, 255, 0, 255}, true)
	vector.DrawFilledRect(screen, float32(g.player.x+15), float32(g.player.y), 10, 5, color.RGBA{0, 255, 0, 255}, true)

	// Aliens
	for _, a := range g.aliens {
		if a.active {
			c := color.RGBA{255, 50, 50, 255}
			if (g.frameCounter/20)%2 == 0 {
				c = color.RGBA{200, 0, 0, 255}
			}
			vector.DrawFilledRect(screen, float32(a.x), float32(a.y+a.animOffset), float32(a.width), float32(a.height), c, true)
		}
	}

	// Bullets
	for _, b := range g.bullets {
		if b.active {
			vector.DrawFilledRect(screen, float32(b.x), float32(b.y), float32(b.width), float32(b.height), color.White, true)
		}
	}

	// UI
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d | Level: %d | Lives: %d", g.score, g.level, g.lives))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) init() {
	g.score = 0
	g.level = 1
	g.lives = 3
	g.nextLevel()
}

func (g *Game) nextLevel() {
	g.resetPositions()
	g.aliens = nil
	for i := 0; i < alienRows; i++ {
		for j := 0; j < alienCols; j++ {
			g.aliens = append(g.aliens, &Entity{
				x:      float64(j*(alienWidth+15) + 50),
				y:      float64(i*(alienHeight+15) + 50),
				width:  alienWidth,
				height: alienHeight,
				active: true,
			})
		}
	}
	g.alienDir = 1
	g.alienSpeed = 1.0 + float64(g.level)*0.5
	if g.state == StateLevelClear {
		g.level++
	}
}

func (g *Game) resetPositions() {
	g.player = Entity{
		x:      screenWidth/2 - playerWidth/2,
		y:      screenHeight - 40,
		width:  playerWidth,
		height: playerHeight,
		active: true,
	}
	g.bullets = nil
	g.lastShot = time.Now()
	g.frameCounter = 0
}

func main() {
	rand.Seed(time.Now().UnixNano())
	game := &Game{}
	game.init()
	game.state = StateTitle

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Space Invaders - GoLang Edition")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
