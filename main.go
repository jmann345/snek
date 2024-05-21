package main

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/jmann345/snek/opts"
	"github.com/nsf/termbox-go"
)

// IDEA: rewrite everything in lua
// TODO: Go through each TODO and delete them as you go

type GameState byte

const (
	StartScreen GameState = iota
	GameScreen
	GameOverScreen
    ConfigScreen
)

type Pos struct {
	x int
	y int
}

var (
	h = Pos{x: -1, y: 0}
	j = Pos{x: 0, y: 1}
	k = Pos{x: 0, y: -1}
	l = Pos{x: 1, y: 0}
)

type Snek struct {
	body    []Pos
	snekMap map[Pos]bool
	dir     Pos
	len     int
}

type Game struct {
	state     GameState
	positions []Pos
	food      []Pos
	snek      *Snek
	score     int
	cfg       Cfg
	ux        UX
}

// TODO::Refactor to only initialize parts that require logic
// (but see if logic can be handled within pkl file)
// Also TODO : Take rgb values for all colors and use RGBToAttribute

func main() {
	// TODO: path should be var determined by user input

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	// "go run . --config {filepath}"
    var cfgPath string
    if len(os.Args) == 3 && (os.Args[1] == "-c" || os.Args[1] == "--config") {
        cfgPath = os.Args[2]
    } else {
        cfgPath = "pkl/defaultOpts.pkl"
    }
	opts, err := opts.LoadFromPath(context.Background(), cfgPath)
	if err != nil {
		panic(err) // alternatively - if user does not have pkl, just set default opts. require pkl for configuration tho
	}
	ux, cfg := initConfig(opts)

    game := restartGame(true, ux, cfg)
	game.render()
	go game.handleInput()

	game.gameLoop()
}

func restartGame(start bool, ux UX, cfg Cfg) Game {
	i := 0
	positions := make([]Pos, cfg.rows*cfg.cols)
	for y := 0; y < cfg.rows; y++ {
		for x := 0; x < cfg.cols; x++ {
			positions[i] = Pos{x: x, y: y}
			i++
		}
	}
	snek := Snek{
		body:    []Pos{{x: 2, y: 7}, {x: 3, y: 7}, {x: 4, y: 7}},
		snekMap: map[Pos]bool{{x: 2, y: 7}: true, {x: 3, y: 7}: true, {x: 4, y: 7}: true},
		dir:     l,
		len:     3,
	}

	food := make([]Pos, cfg.snax)
	food[0] = Pos{x: 8, y: 7}

	emptyCells := make([]Pos, 0)
	for _, pos := range positions {
		if !snek.snekMap[pos] && indexOf(food, pos) == -1 {
			emptyCells = append(emptyCells, Pos{x: pos.x, y: pos.y})
		}
	}
	for i := 1; i < cfg.snax; i++ {
		j := rand.Intn(len(emptyCells))
		food[i] = emptyCells[j]

		emptyCells = append(emptyCells[:j], emptyCells[j+1:]...)
	}

    var gs GameState
    if start {
        gs = StartScreen
    } else {
        gs = GameScreen
    }

	return Game{
		state:     gs,
		positions: positions,
		food:      food,
		snek:      &snek,
		score:     0,
		cfg:       cfg,
		ux:        ux,
	}
}

func (g *Game) gameLoop() {
	for {
		g.render()

		if g.state == GameScreen {
			time.Sleep(g.cfg.speed * time.Millisecond)
			g.updateGameState()
		}
	}
}

func (g *Game) setBorder(offset int) {
	for x := 0; x <= g.cfg.cols*2+2; x++ {
		termbox.SetCell(x+offset, 0, ' ', g.ux.fg, g.ux.fg)
		termbox.SetCell(x+offset, g.cfg.rows+1, ' ', g.ux.fg, g.ux.fg)
	}
	for y := 0; y <= g.cfg.rows+1; y++ {
		termbox.SetCell(offset, y, ' ', g.ux.fg, g.ux.fg)
		termbox.SetCell(1+offset, y, ' ', g.ux.fg, g.ux.fg)

		termbox.SetCell(2*g.cfg.cols+2+offset, y, ' ', g.ux.fg, g.ux.fg)
		termbox.SetCell(2*g.cfg.cols+3+offset, y, ' ', g.ux.fg, g.ux.fg)
	}
	if g.state == GameScreen || g.state == GameOverScreen {
		writeStr(0, g.cfg.rows+2, "SCORE", g.ux.fg, g.ux.bg)

		scoreStr := strconv.Itoa(g.score)
		for i, digitCh := range scoreStr {
			termbox.SetCell(6+i, g.cfg.rows+2, digitCh, g.ux.fg, g.ux.bg)
		}
	}
}

func (g *Game) render() {
	defer termbox.Flush()

	err := termbox.Clear(g.ux.bg, g.ux.bg)
	if err != nil {
		panic(err)
	}

	switch g.state {
	case StartScreen:
		/* Offset game screen to right for previews
		   [s]tart
		   [c]onfigure
		   [q]uit
		*/
		fg, bg := g.ux.fg|termbox.AttrBold, g.ux.bg
		writeStr(0, 0, "no step on snek", fg, bg)
		writeStr(0, 1, "[s]tart", fg, bg)
		writeStr(0, 2, "[c]onfigure", fg, bg)
		writeStr(0, 3, "[q]uit", fg, bg)

		// need x, y for cursor (in configuration screen)!

		// g.setBorder(offset)
		// TODO
	case GameScreen:
		g.setBorder(0)
		for _, foodPos := range g.food {
			x, y := foodPos.x+1, foodPos.y+1
			termbox.SetCell(x*2, y, g.ux.foodCh, g.ux.foodFg, g.ux.foodBg)
			termbox.SetCell(x*2+1, y, g.ux.foodCh, g.ux.foodFgAlt, g.ux.foodBgAlt)
		}

		for _, snekCell := range g.snek.body {
			if (snekCell.x+snekCell.y)%2 != 0 {
				setSquare(snekCell.x+1, snekCell.y+1, g.ux.snekCh, g.ux.snekFgAlt, g.ux.snekBgAlt)
			} else {
				setSquare(snekCell.x+1, snekCell.y+1, g.ux.snekCh, g.ux.snekFg, g.ux.snekBg)
			}
		}
	case GameOverScreen:
		g.setBorder(0)
		fg, bg := g.ux.fg|termbox.AttrBold, g.ux.bg
        var msg string
        if len(g.snek.body) == g.cfg.rows * g.cfg.cols {
            msg = "u won"
        } else {
            msg = "rip"
        }
		writeStr(2, 1, msg, fg, bg)
		writeStr(2, 2, "[r]estart", fg, bg)
		writeStr(2, 3, "[q]uit", fg, bg)
    case ConfigScreen:
        // var cfgPath string
        // if len(os.Args) == 3 && (os.Args[1] == "-c" || os.Args[1] == "--config") {
        //     cfgPath = os.Args[2]
        // } else {
        //     cfgPath = "pkl/defaultOpts.pkl"
        // }
        // opts, err := opts.LoadFromPath(context.Background(), cfgPath)
        optStrs := [14]string{
            "rows", "cols", "fg", "bg", "snek fg", "food fg", "alt snek fg", "alt apple fg", "snek skin", "food skin",
            "speed", "apples", "portals", "length gain",
        }
        for i, s := range optStrs {
            writeStr(0, i, s, g.ux.fg, g.ux.bg)
        }

        // calculate max length of opt strings for offset of game preview
        var offset int = 2
        for _, opt := range optStrs {
            if len(opt) + 2 > offset {
                offset = len(opt) + 2
            }
        }

        g.setBorder(offset)
        previewSnek := [7]Pos{{x: 2, y: 7}, {x: 3, y: 7}, {x: 4, y: 7}, {x: 5, y: 7}, {x: 6, y: 7}, {x: 7, y: 7}, {x: 8, y: 7}}
		for _, snekCell := range previewSnek {
			if (snekCell.x+snekCell.y)%2 == 0 {
				setSquare(snekCell.x+1 + offset/2, snekCell.y+1, g.ux.snekCh, g.ux.snekFgAlt, g.ux.snekBgAlt)
			} else {
				setSquare(snekCell.x+1 + offset/2, snekCell.y+1, g.ux.snekCh, g.ux.snekFg, g.ux.snekBg)
			}
		}
		for _, foodPos := range g.food {
			x, y := foodPos.x+1, foodPos.y+1
			termbox.SetCell(x*2 + offset, y, g.ux.foodCh, g.ux.foodFg, g.ux.foodBg)
			termbox.SetCell(x*2+1 + offset, y, g.ux.foodCh, g.ux.foodFgAlt, g.ux.foodBgAlt)
		}

        
	}

}

func (g *Game) handleInput() {
	inputQueue := make(chan termbox.Event)
	go func() {
		for {
			inputQueue <- termbox.PollEvent()
		}
	}()

	for {
		if ev := <-inputQueue; ev.Type == termbox.EventKey {
			switch g.state {
			case StartScreen:
				switch {
				case ev.Ch == 's':
					g.state = GameScreen
                case ev.Ch == 'c':
                    g.state = ConfigScreen
				case ev.Ch == 'q' || ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC:
                    // TODO : change to return and see what happens
					termbox.Close()
					os.Exit(0)
				}
			case GameScreen:
				switch {
				case ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC:
					termbox.Close()
					os.Exit(0)
				case g.snek.dir == h || g.snek.dir == l:
					switch {
					case ev.Ch == 'j' || ev.Ch == 's' || ev.Key == termbox.KeyArrowDown:
						g.snek.dir = j
					case ev.Ch == 'k' || ev.Ch == 'w' || ev.Key == termbox.KeyArrowUp:
						g.snek.dir = k
					}
				default:
					switch {
					case ev.Ch == 'h' || ev.Ch == 'a' || ev.Key == termbox.KeyArrowLeft:
						g.snek.dir = h
					case ev.Ch == 'l' || ev.Ch == 'd' || ev.Key == termbox.KeyArrowRight:
						g.snek.dir = l
					}
				}
				time.Sleep(g.cfg.speed * time.Millisecond)
			case GameOverScreen:
				switch {
				case ev.Ch == 'r':
					*g = restartGame(false, g.ux, g.cfg)
				case ev.Ch == 'q' || ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC:
					termbox.Close()
					os.Exit(0)
				}
            case ConfigScreen:
                //TODO
			}
		}
	}
}

func indexOf[T comparable](xs []T, x T) int {
	for i, v := range xs {
		if v == x {
			return i
		}
	}
	return -1
}

func (g *Game) updateGameState() {
	snek := g.snek
	head := snek.body[snek.len-1]
	tail := snek.body[0]

	dir := snek.dir
	newHead := Pos{head.x + dir.x, head.y + dir.y}
    // refactor below to updatePos() and eatFood() (maybe no)
    if newHead.x < 0 || newHead.x >= g.cfg.cols || newHead.y < 0 || newHead.y >= g.cfg.rows {
		if g.cfg.portals {
			switch {
			case newHead.x < 0:
				newHead.x = g.cfg.cols - 1
			case newHead.x >= g.cfg.cols:
				newHead.x = 0
			case newHead.y < 0:
				newHead.y = g.cfg.rows - 1
			case newHead.y >= g.cfg.rows:
				newHead.y = 0
			}
        } else {
			g.state = GameOverScreen
        }
    }
    if snek.snekMap[newHead] == true {
        g.state = GameOverScreen
        return
    }

    if i := indexOf(g.food, newHead); i != -1 {
		g.score++
		snek.len++
        // generate new food location
		emptyCells := make([]Pos, g.cfg.rows * g.cfg.cols - snek.len)
        j := 0
        for _, pos := range g.positions {
            if !snek.snekMap[pos] && indexOf(g.food, pos) == -1 {
                emptyCells[j] = Pos{x: pos.x, y: pos.y}
                j++
            }
        }

		numEmptyCells := len(emptyCells)
		if numEmptyCells != 0 {
			newFoodIdx := rand.Intn(numEmptyCells)
			foodPos := emptyCells[newFoodIdx]
			g.food[i] = foodPos
		}
    } else {
        snek.snekMap[tail] = false
        snek.body = snek.body[1:]
    } 

    snek.snekMap[newHead] = true 
    snek.body = append(snek.body, newHead)
}
