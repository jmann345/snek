package main

import (
    "context"
	"math/rand"
	"os"
	"time"
    "strconv"

	"github.com/nsf/termbox-go"
    "github.com/jmann345/snek/opts"
)

// IDEA: rewrite everything in lua
// TODO: Go through each TODO and delete them as you go
type Difficulty = byte
const (
    easy Difficulty = iota
    medium
    hard
)

type Speed = int
const (
    slow Speed = 75
    mid Speed = 100
    fast Speed = 150
)

type Options struct {
    rows     int
    cols     int
    fg       termbox.Attribute
    bg       termbox.Attribute
    snekFg   termbox.Attribute
    foodFg   termbox.Attribute
    snekSkin rune
    foodSkin rune
    speed    Speed
    snax     int
    portals  bool
}
// TODO: Make selection screen, offer initial opts for fun game modes 

const (
	rows = 16
	cols = 16
)

const (
	blue  = termbox.ColorBlue
	green = termbox.ColorGreen
	black = termbox.ColorBlack
	white = termbox.ColorWhite
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
    foodPos = Pos{x: 8, y: 7}
    positions []Pos
    snek Snek
    score int
    cfg Options
)

func init() {
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
}

func main() {

    pklCfg, err := opts.LoadFromPath(context.Background(), "pkl/defaultOpts.pkl") // TODO: path should be var determined by user input
    if err != nil {
        panic(err)
    }
    // set config

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	gameLoop()
}

func setSquare(x, y int, ch rune, fg, bg termbox.Attribute) { //maybe pass in rune as arg
	termbox.SetCell(x*2, y, ch, fg, bg)
	termbox.SetCell(x*2+1, y, ch, fg, bg)
}

func setBorder() {
	for x := 0; x <= cols*2+2; x++ {
		termbox.SetCell(x, 0, ' ', white, white)
		termbox.SetCell(x, rows+1, ' ', white, white)
	}
	for y := 0; y <= rows+1; y++ {
		termbox.SetCell(0, y, ' ', white, white)
		termbox.SetCell(1, y, ' ', white, white)

		termbox.SetCell(2*cols+2, y, ' ', white, white)
		termbox.SetCell(2*cols+3, y, ' ', white, white)
	}
    termbox.SetCell(0, rows+2, 'S', white, black)
    termbox.SetCell(1, rows+2, 'C', white, black)
    termbox.SetCell(2, rows+2, 'O', white, black)
    termbox.SetCell(3, rows+2, 'R', white, black)
    termbox.SetCell(4, rows+2, 'E', white, black)

    scoreStr := strconv.Itoa(score)
    for i, digitCh := range scoreStr {
        termbox.SetCell(6 + i, rows+2, digitCh, white, black)
    }
}

func gameLoop() {
	for {
		render()

		time.Sleep(100 * time.Millisecond)
		go handleInput()

		updateGameState()
		// TODO: Check for game over conditions
	}
}

func render() {
	defer termbox.Flush()
	termbox.Clear(black, black)
	setBorder()
	setSquare(foodPos.x+1, foodPos.y+1, '', blue, black)
	for _, snekCell := range snek.body {
		setSquare(snekCell.x+1, snekCell.y+1, ' ', green, green)
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
func updateGameState() {
	head := snek.body[snek.len - 1]
	dir := snek.dir
	newHead := Pos{head.x + dir.x, head.y + dir.y}

	snek.body = append(snek.body, newHead)

	switch {
	case newHead == foodPos:
        score++
        snek.len++
		snek.snekMap[newHead] = true

		emptyCells := make([]Pos, 0)
		for _, pos := range positions {
			if !snek.snekMap[pos] {
				emptyCells = append(emptyCells, Pos{x: pos.x, y: pos.y})
			}
		}

		newFoodIdx := rand.Intn(len(emptyCells))
		foodPos = emptyCells[newFoodIdx]
    case snek.snekMap[newHead] == true: //TODO: Game over screen, select restart/change setting/quit
		termbox.Close()
		os.Exit(0)
	case newHead.x < 0 || newHead.x >= cols || newHead.y < 0 || newHead.y >= rows://TODO: Game over screen, select restart/change setting/quit
		termbox.Close()
		os.Exit(0)
    default: // TODO: case victory (game over screen but it doesnt say i lost)
		tail := snek.body[0]
		snek.snekMap[tail] = false
		snek.snekMap[newHead] = true
		snek.body = snek.body[1:]
	}
}
