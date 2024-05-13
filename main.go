package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/jmann345/snek/opts"
	"github.com/nsf/termbox-go"
)

// IDEA: rewrite everything in lua
// TODO: Go through each TODO and delete them as you go
type Difficulty = byte

const (
	easy Difficulty = iota
	medium
	hard
)

type Pos struct {
	x int
	y int
}
type Snek struct {
	body    []Pos
	snekMap map[Pos]bool
	dir     Pos
	len     int
}

var (
	h = Pos{x: -1, y: 0}
	j = Pos{x: 0, y: 1}
	k = Pos{x: 0, y: -1}
	l = Pos{x: 1, y: 0}
)

var (
	food      = []Pos{{x: 8, y: 7}}
	positions []Pos
	snek      Snek
	score     int
)

var ( // TODO: Make this a struct (Ideally, everything should be ripped straight from the pkl)
	rows      int               //CHECK
	cols      int               //CHECK
	fg        termbox.Attribute //CHECK
	bg        termbox.Attribute //CHECK
	snekFg    termbox.Attribute //CHECK
	snekBg    termbox.Attribute //CHECK
	foodFg    termbox.Attribute //CHECK
	foodBg    termbox.Attribute //CHECK
	alt       bool              //CHECK
	snekFgAlt termbox.Attribute //CHECK
	snekBgAlt termbox.Attribute //CHECK
	foodFgAlt termbox.Attribute //CHECK
	foodBgAlt termbox.Attribute //CHECK
	snekCh    rune              //CHECK
	foodCh    rune              //CHECK
	speed     time.Duration     //CHECK
	portals   bool              //CHECK
	snax      int               // CHECK
	// TODO: uncomment line below
	// lenGain int
)

// TODO: Refactor to only initialize parts that require logic
// (but see if logic can be handled within pkl file)
func initConfig(cfg *opts.Opts) {
	rows = cfg.Rows
	cols = cfg.Cols

	fg = termbox.Attribute(cfg.Fg)
	bg = termbox.Attribute(cfg.Bg)

	snekFg = termbox.Attribute(cfg.SnekFg)
	snekFgAlt = termbox.Attribute(cfg.SnekFgAlt)
	snekBg = termbox.Attribute(cfg.Bg)
	snekBgAlt = termbox.Attribute(cfg.Bg)
	snekCh = func() rune {
		switch cfg.SnekSkin {
		case "python": //TODO: Add more snek skins
			return ''
		default:
			snekBg = snekFg
			snekBgAlt = snekFgAlt
			return ' '
		}
	}()

	foodFg = termbox.Attribute(cfg.FoodFg) // IDEA: If allow for randomized foodCh, foodFgAlt, and foodBgAlt
	foodFgAlt = termbox.Attribute(cfg.FoodFgAlt)
	foodBg = termbox.Attribute(cfg.Bg)
	foodBgAlt = termbox.Attribute(cfg.Bg)
	foodCh = func() rune {
		switch cfg.FoodSkin {
		case "gopher":
			return ''
		default: //TODO: Add more food skins
			foodBg = foodFg
			foodBgAlt = foodFgAlt
			return ' '
		}
	}()

	alt = false

	speed = time.Duration(cfg.Speed)

	snax = cfg.Snax // init extra food on map

	portals = cfg.Portals
}

func main() {
	// TODO: path should be var determined by user input
	cfg, err := opts.LoadFromPath(context.Background(), "pkl/defaultOpts.pkl")
	if err != nil {
		panic(err)
	}
	initConfig(cfg)
	// set config
	//init global vars
	score = 0
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			positions = append(positions, Pos{x: x, y: y})
		}
	}
	snek = Snek{
		body:    []Pos{{x: 2, y: 7}, {x: 3, y: 7}, {x: 4, y: 7}},
		snekMap: map[Pos]bool{{x: 2, y: 7}: true, {x: 3, y: 7}: true, {x: 4, y: 7}: true},
		dir:     l,
		len:     3,
	}

	emptyCells := make([]Pos, 0) // TODO: Refactor this into a global
	for _, pos := range positions {
		if !snek.snekMap[pos] && pos.in(food) == -1 {
			emptyCells = append(emptyCells, Pos{x: pos.x, y: pos.y})
		}
	}
	for i := 1; i < snax; i++ {
		fmt.Printf("%d", snax)
		j := rand.Intn(len(emptyCells))
		foodPos := emptyCells[j]
		food = append(food, foodPos)

		emptyCells = append(emptyCells[:j], emptyCells[j+1:]...)
	}

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	gameLoop()
}

func setSquare(x, y int, ch rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x*2, y, ch, fg, bg)
	termbox.SetCell(x*2+1, y, ch, fg, bg)
}

func setBorder() {
	for x := 0; x <= cols*2+2; x++ {
		termbox.SetCell(x, 0, ' ', fg, fg)
		termbox.SetCell(x, rows+1, ' ', fg, fg)
	}
	for y := 0; y <= rows+1; y++ {
		termbox.SetCell(0, y, ' ', fg, fg)
		termbox.SetCell(1, y, ' ', fg, fg)

		termbox.SetCell(2*cols+2, y, ' ', fg, fg)
		termbox.SetCell(2*cols+3, y, ' ', fg, fg)
	}
	termbox.SetCell(0, rows+2, 'S', fg, bg)
	termbox.SetCell(1, rows+2, 'C', fg, bg)
	termbox.SetCell(2, rows+2, 'O', fg, bg)
	termbox.SetCell(3, rows+2, 'R', fg, bg)
	termbox.SetCell(4, rows+2, 'E', fg, bg)

	scoreStr := strconv.Itoa(score)
	for i, digitCh := range scoreStr {
		termbox.SetCell(6+i, rows+2, digitCh, fg, bg)
	}
}

func gameLoop() {
	for {
		render()

		time.Sleep(speed * time.Millisecond)
		go handleInput()

		updateGameState()
		// TODO: Check for game over conditions
	}
}

func render() {
	defer termbox.Flush()

	err := termbox.Clear(bg, bg)
	if err != nil {
		panic(err)
	}
	setBorder()

	for _, foodPos := range food {
		x, y := foodPos.x+1, foodPos.y+1
		termbox.SetCell(x*2, y, foodCh, foodFg, foodBg)
		termbox.SetCell(x*2+1, y, foodCh, foodFgAlt, foodBgAlt)
	}

	if score%2 != 0 {
		alt = !alt
	}
	for _, snekCell := range snek.body {
		if alt {
			setSquare(snekCell.x+1, snekCell.y+1, snekCh, snekFgAlt, snekBgAlt)
			alt = false
		} else {
			setSquare(snekCell.x+1, snekCell.y+1, snekCh, snekFg, snekBg)
			alt = true
		}
	}
}

func handleInput() {
	if ev := termbox.PollEvent(); ev.Type == termbox.EventKey {
		switch {
		case ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC:
			termbox.Close()
			os.Exit(0)
		case snek.dir == h || snek.dir == l:
			switch {
			case ev.Ch == 'j' || ev.Ch == 's' || ev.Key == termbox.KeyArrowDown:
				snek.dir = j
			case ev.Ch == 'k' || ev.Ch == 'w' || ev.Key == termbox.KeyArrowUp:
				snek.dir = k
			}
		default:
			switch {
			case ev.Ch == 'h' || ev.Ch == 'a' || ev.Key == termbox.KeyArrowLeft:
				snek.dir = h
			case ev.Ch == 'l' || ev.Ch == 'd' || ev.Key == termbox.KeyArrowRight:
				snek.dir = l
			}
		}
	}

}

func (p Pos) in(ps []Pos) int { // can I anonymize this while keeping syntax?
	for i, pos := range ps {
		if p == pos {
			return i
		}
	}
	return -1
}

func updateGameState() {
	head := snek.body[snek.len-1]
	dir := snek.dir
	newHead := Pos{head.x + dir.x, head.y + dir.y}

portal:
	switch i := newHead.in(food); {
	case i != -1:
		score++
		snek.len++
		snek.snekMap[newHead] = true
		snek.body = append(snek.body, newHead)

		emptyCells := make([]Pos, 0)
		for _, pos := range positions {
			if !snek.snekMap[pos] && pos.in(food) == -1 {
				emptyCells = append(emptyCells, Pos{x: pos.x, y: pos.y})
			}
		}

		newFoodIdx := rand.Intn(len(emptyCells))
		foodPos := emptyCells[newFoodIdx]
		food[i] = foodPos
	case snek.snekMap[newHead] == true: //TODO: Game over screen, select restart/change setting/quit
		termbox.Close()
		os.Exit(0)
	case newHead.x < 0 || newHead.x >= cols || newHead.y < 0 || newHead.y >= rows:
		if portals {
			switch {
			case newHead.x < 0:
				newHead.x = cols - 1
			case newHead.x >= cols:
				newHead.x = 0
			case newHead.y < 0:
				newHead.y = rows - 1
			case newHead.y >= rows:
				newHead.y = 0
			}
			goto portal
		} else { //TODO: Game over screen, select restart/change setting/quit
			termbox.Close()
			os.Exit(0)
		}
	default: // TODO: case victory (game over screen but it doesnt say i lost)
		snek.snekMap[newHead] = true
		tail := snek.body[0]
		snek.snekMap[tail] = false
		snek.body = append(snek.body[1:], newHead)
	}
}
