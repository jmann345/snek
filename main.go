package main
import (
    "os"
    "time"
    "math/rand"
    "github.com/nsf/termbox-go"
)
// TODO: Go through each TODO and delete them as you go

const (
    rows = 16
    cols = 16
)

type Pos struct {
    x int
    y int
}

type Dir = Pos

type Snek struct { //TODO: Make Snek a Queue represented by a channel
    body []Pos 
    dir  Dir
    len  int
}

type Cell = int

var (
    h = Pos{x: -1, y: 0}
    j = Pos{x: 0, y: 1}
    k = Pos{x: 0, y: -1}
    l = Pos{x: 1, y: 0}
)

var foodPos = Pos{x: 8, y: 7}
var positions []Pos

const (
    blue = termbox.ColorBlue
    green = termbox.ColorGreen
    black = termbox.ColorBlack
    white = termbox.ColorWhite
)

func main() {

    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    // HideCursor() (maybe?)
    defer termbox.Close()

    gameLoop()
}

func setSquare(x, y int, fg, bg termbox.Attribute) { //maybe pass in rune as arg
    rune := ''
    if fg == blue { rune = '' }
    termbox.SetCell(x * 2, y, rune, fg, bg)
    termbox.SetCell(x * 2 + 1, y, rune, fg, bg)
}

func setBorder() {
    for x := 0; x <= rows * 2 + 2; x++ {
        termbox.SetCell(x, 0, ' ', white, white)
        termbox.SetCell(x, rows + 1, ' ', white, white)
    }
    for y := 0; y <= rows + 1; y++ {
        termbox.SetCell(0, y, ' ', white, white)
        termbox.SetCell(1, y, ' ', white, white)

        termbox.SetCell(2 * cols + 2, y, ' ', white, white)
        termbox.SetCell(2 * cols + 3, y, ' ', white, white)
    }
}

func gameLoop() {
    // Initialize game state, snek, foodPos, etc.
    // init
    //TODO: Move eveything before the for { } loop into func init()
    for y := 0; y < rows; y++ {
        for x := 0; x < cols; x++ {
            positions = append(positions, Pos{x: x, y: y})
        }
    }
    snek := Snek{
        body: []Pos{{x: 2, y: 7}, {x: 3, y: 7}, {x: 4, y: 7}},
        dir:  l,
        len:  3,
    }
    //TODO: print borders
    termbox.Clear(black, black)
    setBorder()
    setSquare(foodPos.x+1, foodPos.y+1, blue, black) // TODO: Set square
    for _, snekCell := range snek.body {
        setSquare(snekCell.x+1, snekCell.y+1, green, black) // TODO: Set square
    }
    termbox.Flush()

    for {

        snek.render()

        go snek.handleInput()

        snek.updateGameState()
        // Check for game over conditions

        // Delay to control game speed
        time.Sleep(100 * time.Millisecond)
    }
}

func (snek *Snek) render() {
    defer termbox.Flush()
    termbox.Clear(black, black)
    setBorder()
    setSquare(foodPos.x + 1, foodPos.y + 1, blue, black) // TODO: Set square
    for _, snekCell := range snek.body {
        setSquare(snekCell.x + 1, snekCell.y + 1, green, black) // TODO: Set square
    }
}

func (snek *Snek) handleInput() {
    // TODO: Only read h/l when dir is j/k and vice versa
    if ev := termbox.PollEvent(); ev.Type == termbox.EventKey {
        switch {
        case ev.Key == termbox.KeyEsc: 
            termbox.Close()
            os.Exit(0)
        case snek.dir == h || snek.dir == l:
            switch ev.Ch {
            case 'j':
                snek.dir = j
            case 'k':
                snek.dir = k
            }
        default:
            switch ev.Ch {
            case 'h':
                snek.dir = h
            case 'l':
                snek.dir = l
            }
        }
    }

}

func (snek *Snek) updateGameState() {
    head := snek.body[snek.len - 1]
    dir := snek.dir
    newHead := Pos{head.x + dir.x, head.y + dir.y}

    snekMap := make(map[Pos]bool)
    for _, snekCell := range(snek.body) {
        snekMap[snekCell] = true
    }

    snek.body = append(snek.body, newHead)

    switch {
    case newHead == foodPos: //change this to switch newHead, check for collisions with (if its not empty or food, its death)
        snek.len++ // TODO: replace this field with a GPS (call it gps)
        snekMap[newHead] = true

        emptyCells := make([]Pos, 0)
        for _, pos := range positions {
            if !snekMap[pos] {
                emptyCells = append(emptyCells, Pos{x: pos.x, y: pos.y})
            }
        }

        newFoodIdx := rand.Intn(len(emptyCells))
        foodPos = emptyCells[newFoodIdx]
    case snekMap[newHead] == true:
        termbox.Close()
        os.Exit(0)
    case newHead.x < 0 || newHead.x >= cols || newHead.y < 0 || newHead.y >= rows:
        termbox.Close()
        os.Exit(0)
    default:
        snek.body = snek.body[1:]
    }
}