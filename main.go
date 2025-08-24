package main // VIMSWEEPER

import (
	"context"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"time"
	"unsafe"

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
	i int // j
	j int // i
}

type Dir Pos

type Cell struct {
	mine    bool
	flagged bool
	clicked bool
	number  uint8
}

type Game struct {
	state      GameState
	grid       [][]Cell
	cursor     Pos
	firstClick bool
	numFlags   int
	cfg        Cfg
	ux         UX
	startTime  time.Time
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

    cfgPath := "pkl/defaultOpts.pkl"
    var auto bool
	if len(os.Args) == 2 && (os.Args[1] == "-a") {
        auto = true
	} else {
        auto = false
    }
	opts, err := opts.LoadFromPath(context.Background(), cfgPath)
	if err != nil {
		panic(err) // alternatively - if user does not have pkl, just set default opts. require pkl for configuration tho
	}
	ux, cfg := initConfig(opts)

	// welcomeMsgs := []string{"mine time", "мое время"}
	// r := rand.Intn(len(welcomeMsgs))

	game := restartGame(!auto, ux, cfg)
	// game.render(welcomeMsgs[r])
    if !auto {
        go game.handleInput()
    }

	game.gameLoop(auto)
}

// Also for starting the game
func restartGame(start bool, ux UX, cfg Cfg) Game {
	var gs GameState
	if start {
		gs = StartScreen
	} else {
		gs = GameScreen
	}

	grid := make([][]Cell, cfg.rows)
	for i := range grid {
		grid[i] = make([]Cell, cfg.cols)
	}

	return Game{
		state:      gs,
		grid:       grid,
		cursor:     Pos{cfg.rows / 2, cfg.cols / 2},
		firstClick: true,
		numFlags:   0,
		cfg:        cfg,
		ux:         ux,
	}
}

func (g *Game) placeMines(firstClickPos Pos) {
	var numMines int
	switch g.cfg.difficulty {
	case Easy:
		numMines = 10
	case Medium:
		numMines = 40
	case Hard:
		numMines = 99
	}
	mines := make([]Pos, numMines)

    firstClickRegion := getNeighborPositions(firstClickPos.i, firstClickPos.j, g.cfg)
    firstClickRegion = append(firstClickRegion, firstClickPos)

	mineCellCandidates := make([]Pos, 0)
	for i := range g.cfg.rows {
		for j := range g.cfg.cols {
            if p := (Pos{i, j}); !slices.Contains(firstClickRegion, p) {
                mineCellCandidates = append(mineCellCandidates, p)
            }
		}
	}
	for i := range numMines {
		j := rand.Intn(len(mineCellCandidates))
		mines[i] = mineCellCandidates[j]

		// J* Syntax: emptyCells = emptyCells[:j] ++ emptyCells[j+1:]
		mineCellCandidates = append(mineCellCandidates[:j], mineCellCandidates[j+1:]...)
	}

	// Add mines to grid
	for _, mine := range mines {
		cell := &g.grid[mine.i][mine.j]
		cell.mine = true
	}

	// Count number of surrounding mines on non-mine squares
	for i := range g.cfg.rows {
		for j := range g.cfg.cols {
			cell := &g.grid[i][j] // & is important here
			if cell.mine == false {
				neighbors := getNeighbors(g.grid, i, j, g.cfg)
				for _, neighbor := range neighbors {
					if neighbor.mine == true {
						cell.number++
					}
				}
			}
		}
	}
}

func getNeighbors(grid [][]Cell, i, j int, cfg Cfg) []Cell {
	neighbors := make([]Cell, 0)
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			// in J* you could say `if di == dj == 0` :D
			if di == 0 && dj == 0 {
				continue
			}
			// ni === new i, nj === new j
			ni := i + di
			nj := j + dj

			if ni >= 0 && ni < cfg.rows && nj >= 0 && nj < cfg.cols {
				neighbors = append(neighbors, grid[ni][nj])
			}
		}
	}
	return neighbors
}
func getNeighborPositions(i, j int, cfg Cfg) []Pos {
	neighbors := make([]Pos, 0)
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			// in J* you could say `if di == dj == 0` :D
			if di == 0 && dj == 0 {
				continue
			}
			// ni === new i, nj === new j
			ni := i + di
			nj := j + dj

			if ni >= 0 && ni < cfg.rows && nj >= 0 && nj < cfg.cols {
				neighbors = append(neighbors, Pos{ni, nj})
			}
		}
	}
	return neighbors
}

func (g *Game) gameLoop(auto bool) {

	welcomeMsgs := []string{"mine time", "ласкаво просимо в україну!", "скупой", "Время мин"}
	r := rand.Intn(len(welcomeMsgs))

	for {
		g.render(welcomeMsgs[r])

		if g.state == GameScreen {
			time.Sleep(25 * time.Millisecond)
            if auto {
                if g.firstClick {
                    g.placeMines(g.cursor)
                    g.startTime = time.Now()
                    g.firstClick = false

                    g.grid[g.cursor.i][g.cursor.j].clicked = true
                }
                g.autoplay()
                time.Sleep(100 * time.Millisecond)  //temp
            }
			g.updateGameState()
		}
	}
}

// TODO: Add the minesweeper faces lol (check logic for switching btwn the expressions)
// ( ° ᴗ°)  and  (╯°o°)ᕗ
func (g *Game) setBorder(offset Pos) {
	for x := 0; x <= g.cfg.cols*2+2; x++ {
		termbox.SetCell(x+offset.j, 0+offset.i, ' ', g.ux.fg, g.ux.fg)
		termbox.SetCell(x+offset.j, g.cfg.rows+1+offset.i, ' ', g.ux.fg, g.ux.fg)
	}
	for y := 0; y <= g.cfg.rows+1; y++ {
		termbox.SetCell(offset.j, y+offset.i, ' ', g.ux.fg, g.ux.fg)
		termbox.SetCell(1+offset.j, y+offset.i, ' ', g.ux.fg, g.ux.fg)

		termbox.SetCell(2*g.cfg.cols+2+offset.j, y+offset.i, ' ', g.ux.fg, g.ux.fg)
		termbox.SetCell(2*g.cfg.cols+3+offset.j, y+offset.i, ' ', g.ux.fg, g.ux.fg)
	}
	if g.state == GameScreen || g.state == GameOverScreen {
		fg, bg := g.ux.fg, g.ux.bg
		var numMines int
		switch g.cfg.difficulty {
		case Easy:
			numMines = 10
		case Medium:
			numMines = 40
		case Hard:
			numMines = 99
		}
		remainingMines := numMines - g.numFlags
		writeStr(offset.j, g.cfg.rows+2+offset.i, "MINES: "+strconv.Itoa(remainingMines), fg, bg)

        timeStr := strconv.Itoa(int(time.Since(g.startTime).Seconds()))
        boardWidth := 2*g.cfg.cols + 4          // total columns of the framed board
        x := boardWidth - len(timeStr) - 1      // ‑1 so we don’t overwrite the border
        writeStr(x+offset.j, g.cfg.rows+2+offset.i, timeStr, fg, bg)
	}
}

func (g *Game) render(welcomeMsg string) {
	defer termbox.Flush()

	err := termbox.Clear(g.ux.bg, g.ux.bg)
	if err != nil {
		panic(err)
	}

	switch g.state {
	case StartScreen:
		fg, bg := g.ux.fg|termbox.AttrBold, g.ux.bg

		writeStr(0, 0, welcomeMsg, fg, bg)
		writeStr(0, 1, "[s]tart", fg, bg)
		writeStr(0, 2, "[c]onfigure", fg, bg)
		writeStr(0, 3, "[q]uit", fg, bg)
		// g.setBorder(offset)
	case GameScreen:
		// TODO: Make 1 mine = blue, 2 surrouning mines = green, 3 = red
		// 4 = purple (magenta), 5 = maroon (light red), 6 = cyan, 7 = black, 8 = pink
		g.setBorder(Pos{0,0})
		for i := range g.cfg.rows {
			for j := range g.cfg.cols {
				cell := g.grid[i][j]
				x, y := j+1, i+1

				// Determine the colors for this cell.
				clickedColor := termbox.ColorWhite
				var fg termbox.Attribute
				var bg termbox.Attribute

				if cell.clicked {
					// Determine the digit to show
					digitRune := ' '
					if cell.number > 0 {
						digitRune = rune('0' + cell.number)
					}
					// Choose the foreground based on cell.number...
					switch digitRune {
					case '1':
						fg = termbox.ColorBlue
					case '2':
						fg = termbox.ColorGreen
					case '3':
						fg = termbox.ColorRed
					case '4':
						fg = termbox.ColorMagenta
					case '5':
						fg = termbox.ColorLightRed
					case '6':
						fg = termbox.ColorCyan
					case '7':
						fg = termbox.ColorLightBlue
					case '8':
						fg = termbox.ColorLightGray
					default:
						fg = clickedColor
					}
					bg = clickedColor

					// If this cell is where the cursor is, adjust the colors
					if g.cursor.i == i && g.cursor.j == j {
						fg |= termbox.AttrReverse
						bg = termbox.ColorYellow
					}

					termbox.SetCell(x*2, y, ' ', clickedColor, bg)
					termbox.SetCell(x*2+1, y, digitRune, fg, bg)
				} else if cell.flagged {
					// Draw flagged cells
					fg = termbox.ColorRed
					bg = g.ux.bg
					if g.cursor.i == i && g.cursor.j == j {
						fg |= termbox.AttrReverse
						bg = termbox.ColorYellow
					}
					termbox.SetCell(x*2, y, ' ', bg, bg)
					termbox.SetCell(x*2+1, y, '󰈻', fg, bg)
				} else {
					// Draw unclicked cells.
					fg = g.ux.fg
					bg = g.ux.bg
					if g.cursor.i == i && g.cursor.j == j {
						fg |= termbox.AttrReverse
						bg = termbox.ColorYellow
					}
					termbox.SetCell(x*2, y, ' ', fg, bg)
					termbox.SetCell(x*2+1, y, ' ', fg, bg)
				}
			}
		}

		// cursorX := (g.cursor.j + 1)*2
		// cursorY := g.cursor.i + 1
		// termbox.SetCursor(cursorX, cursorY)
	case GameOverScreen: // TODO: Make the gameoverscreen display mine locations
		g.setBorder(Pos{2, 0})
		for i := range g.cfg.rows {
			for j := range g.cfg.cols {
				cell := g.grid[i][j]
				x, y := j+1, i+1+2

				// Determine the colors for this cell.
				var fg termbox.Attribute
				var bg termbox.Attribute = termbox.ColorWhite

                // Determine the digit to show
                cellRune := ' '
                if cell.number > 0 {
                    cellRune = rune('0' + cell.number)
                }
                // Choose the foreground based on cell.number...
                switch cellRune {
                case '1':
                    fg = termbox.ColorBlue
                case '2':
                    fg = termbox.ColorGreen
                case '3':
                    fg = termbox.ColorRed
                case '4':
                    fg = termbox.ColorMagenta
                case '5':
                    fg = termbox.ColorLightRed
                case '6':
                    fg = termbox.ColorCyan
                case '7':
                    fg = termbox.ColorLightBlue
                case '8':
                    fg = termbox.ColorLightGray
                default:
                    fg = bg
                }

				if cell.flagged {
					// Draw flagged cells
					fg = termbox.ColorRed
                    if cell.mine {
                        bg = g.ux.bg
                        cellRune = '󰈻'
                    } else { // Bad flag placement
                        bg = g.ux.fg
                        cellRune = '󱣮'
                    }
				} else if cell.mine {
					fg = termbox.ColorRed
                    bg = g.ux.bg
                    cellRune = '󰷚'
                    if g.cursor.i == i && g.cursor.j == j {
                        bg = termbox.ColorLightRed
                    }
                } 
                termbox.SetCell(x*2, y, cellRune, fg, bg)
                termbox.SetCell(x*2+1, y, ' ', fg, bg)
			}
		}
		fg, bg := g.ux.fg|termbox.AttrBold, g.ux.bg
		clickedCells := 0
		for i := range g.cfg.rows {
			for j := range g.cfg.cols {
				if g.grid[i][j].clicked {
					clickedCells++
				}
			}
		}

		var numMines int
		switch g.cfg.difficulty {
		case Easy:
			numMines = 10
		case Medium:
			numMines = 40
		case Hard:
			numMines = 99
		}
		var msg string
		if clickedCells == g.cfg.rows*g.cfg.cols-numMines {
			msg = "u won"
		} else {
			msg = "rip"
		}
		writeStr(0, 0, msg, fg, bg)
		writeStr(0, 1, "[r]estart | ", fg, bg)
		writeStr(12, 1, "[q]uit", fg, bg)
	case ConfigScreen:
		offset := Pos{0, 2}
		g.setBorder(offset)
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
                var (
                    h = Dir{i: 0, j: -1}
                    j = Dir{i: 1, j: 0}
                    k = Dir{i: -1, j: 0}
                    l = Dir{i: 0, j: 1}
                )
				switch {
				case ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC:
					termbox.Close()
					os.Exit(0)
				case ev.Ch == 'j' || ev.Key == termbox.KeyArrowDown:
					g.cursor.i += j.i
					g.cursor.j += j.j
				case ev.Ch == 'k' || ev.Key == termbox.KeyArrowUp:
					g.cursor.i += k.i
					g.cursor.j += k.j
				case ev.Ch == 'h' || ev.Key == termbox.KeyArrowLeft:
					g.cursor.i += h.i
					g.cursor.j += h.j
				case ev.Ch == 'l' || ev.Key == termbox.KeyArrowRight:
					g.cursor.i += l.i
					g.cursor.j += l.j
                // IDEA: If current cell has white bg, go to next empty cell and vice versa
                case ev.Ch == 'w':
                    moved := false
                    clicked := g.grid[g.cursor.i][g.cursor.j].clicked
                    for j := g.cursor.j+2; j < g.cfg.cols; j++ {
                        if cell := g.grid[g.cursor.i][j]; cell.clicked != clicked {
                            g.cursor.j = j
                            moved = true
                            break
                        }
                    }
                    if !moved {
                        g.cursor.j = g.cfg.cols - 1
                    }
                case ev.Ch == 'b':
                    moved := false
                    clicked := g.grid[g.cursor.i][g.cursor.j].clicked
                    for j := g.cursor.j-2; j >= 0; j-- {
                        if cell := g.grid[g.cursor.i][j]; cell.clicked != clicked {
                            g.cursor.j = j
                            moved = true
                            break
                        }
                    }
                    if !moved {
                        g.cursor.j = 0
                    }
                case ev.Ch == 'u':
                    moved := false
                    clicked := g.grid[g.cursor.i][g.cursor.j].clicked
                    for i := g.cursor.i-2; i >= 0; i-- {
                        if cell := g.grid[i][g.cursor.j]; cell.clicked != clicked {
                            g.cursor.i = i
                            moved = true
                            break
                        }
                    }
                    if !moved {
                        g.cursor.i = 0
                    }
                case ev.Ch == 'g':
                    moved := false
                    clicked := g.grid[g.cursor.i][g.cursor.j].clicked
                    for i := g.cursor.i+2; i < g.cfg.rows; i++ {
                        if cell := g.grid[i][g.cursor.j]; cell.clicked != clicked {
                            g.cursor.i = i
                            moved = true
                            break
                        }
                    }
                    if !moved {
                        g.cursor.i = g.cfg.rows-1
                    }
                case ev.Ch == '0':
                    g.cursor.j = 0
                case ev.Ch == '$':
                    g.cursor.j = g.cfg.cols - 1
				case ev.Ch == 'f' || ev.Ch == ';' || ev.Key == termbox.KeyPgup:
					cell := &g.grid[g.cursor.i][g.cursor.j]

                    if !cell.clicked {
                        cell.flagged = !cell.flagged
                        // if the cell is flagged, g.numFlags++, else g.numFlags--
                        g.numFlags += 2*int(*(*byte)(unsafe.Pointer(&cell.flagged))) - 1
                    }
				case ev.Ch == 'd' || ev.Key == termbox.KeyPgdn || ev.Key == termbox.KeySpace:
					// toggle flag
					if g.firstClick {
						g.placeMines(g.cursor)
                        g.startTime = time.Now()
						g.firstClick = false
					}
                    cell := &g.grid[g.cursor.i][g.cursor.j]
                    if cell.clicked {
                        neighbors := getNeighbors(g.grid, g.cursor.i, g.cursor.j, g.cfg)
                        flaggedNeighbors := []Cell{}

                        neighborPositions := getNeighborPositions(g.cursor.i, g.cursor.j, g.cfg)
                        unflaggedNeighborPositions := []Pos{}
                        for i, nb := range neighbors {
                            if nb.flagged {
                                flaggedNeighbors = append(flaggedNeighbors, nb)
                            } else {
                                unflaggedNeighborPositions = append(unflaggedNeighborPositions, neighborPositions[i])
                            }
                        }
                        if len(flaggedNeighbors) == int(cell.number) {
                            for _, nbPos := range unflaggedNeighborPositions {
                                ni, nj := nbPos.i, nbPos.j
                                g.grid[ni][nj].clicked = true
                            }
                        }

                    }
                    if !cell.flagged {
                        cell.clicked = true
                    } 
				}
				g.cursor.i = min(g.cursor.i, g.cfg.rows-1)
				g.cursor.i = max(g.cursor.i, 0)
				g.cursor.j = min(g.cursor.j, g.cfg.cols-1)
				g.cursor.j = max(g.cursor.j, 0)
				time.Sleep(25 * time.Millisecond)
			case GameOverScreen:
				switch {
				case ev.Ch == 'r':
					*g = restartGame(false, g.ux, g.cfg)
				case ev.Ch == 'q' || ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC:
					termbox.Close()
					os.Exit(0)
				}
			case ConfigScreen:
				// TODO:
			}
		}
	}
}

// TODO: Reimpl floodflow without crossref

// it's just bfs with extra steps :)
func floodFill(grid [][]Cell, start Pos, cfg Cfg) []Pos {
	visited := make([][]bool, cfg.rows)
	for i := 0; i < cfg.rows; i++ {
		visited[i] = make([]bool, cfg.cols)
	}

	// result will store all positions that should be revealed.
	result := []Pos{}
	q := []Pos{start}
	visited[start.i][start.j] = true
	result = append(result, start)

	for len(q) > 0 {
		cur := q[0]
		q = q[1:]

		for _, nb := range getNeighborPositions(cur.i, cur.j, cfg) {
			if !visited[nb.i][nb.j] {
				visited[nb.i][nb.j] = true
				cell := grid[nb.i][nb.j]
				if !cell.mine {
					result = append(result, nb) // fix: also include nonzero, nonmine cells just dont enq them
					// only enq zero cells
					if cell.number == 0 {
						q = append(q, nb)
					}
				}
			}
		}
	}

	return result
}

func (g *Game) updateGameState() {

    clickedCells := 0 // TODO: Move this to the state so we dont have to recalculate every time
	for i := range g.cfg.rows {
		for j := range g.cfg.cols {
			if g.grid[i][j].clicked {
				clickedCells++
			}
		}
	}
	var numMines int
	switch g.cfg.difficulty {
	case Easy:
		numMines = 10
	case Medium:
		numMines = 40
	case Hard:
		numMines = 99
	}
	if clickedCells == g.cfg.rows*g.cfg.cols-numMines {
		g.state = GameOverScreen
	}
	for i := range g.cfg.rows {
		for j := range g.cfg.cols {
			if g.grid[i][j].clicked {
				if g.grid[i][j].mine {
					g.state = GameOverScreen
				} else if g.grid[i][j].number == 0 {
					zeros := floodFill(g.grid, Pos{i, j}, g.cfg)
					for _, zeroPos := range zeros {
						g.grid[zeroPos.i][zeroPos.j].clicked = true
					}
				}
			}
		}
	}
}

func getUnflaggedNeighbors(grid [][]Cell, i, j int, cfg Cfg) []Pos {
	neighbors := make([]Pos, 0)
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			// in J* you could say `if di == dj == 0` :D
			if di == 0 && dj == 0 {
				continue
			}
			ni := i + di
			nj := j + dj

			if ni >= 0 && ni < cfg.rows && nj >= 0 && nj < cfg.cols {
                nb := grid[ni][nj]
                if !nb.flagged && !nb.clicked {
                    neighbors = append(neighbors, Pos{ni, nj})
                }
			}
		}
	}
	return neighbors
}

func getFlaggedNeighbors(grid [][]Cell, i, j int, cfg Cfg) []Pos {
	neighbors := make([]Pos, 0)
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			// in J* you could say `if di == dj == 0` :D
			if di == 0 && dj == 0 {
				continue
			}
			ni := i + di
			nj := j + dj

			if ni >= 0 && ni < cfg.rows && nj >= 0 && nj < cfg.cols {
                nb := grid[ni][nj]
                if nb.flagged /*&& !nb.clicked*/ {
                    neighbors = append(neighbors, Pos{ni, nj})
                }
			}
		}
	}
	return neighbors
}

func (g *Game) autoplay() {
    // need to click randomly if no action
    // more advance, calculate probabilities and compute action (basic combinatorics)
    // Idea: F: (Candidates) -> map[pos]float 
    // the floats are probs btwn 0 and 1
    // Click on everything with a probability of 1
    // If nothing has a probability of 1, click whatever has the highest probability
    for i := range g.cfg.rows {
        for j := range g.cfg.cols {
            cell := &g.grid[i][j]
            g.cursor = Pos{i, j}
            time.Sleep(time.Millisecond)

            if cell.clicked {
                freeNbs := getUnflaggedNeighbors(g.grid, i, j, g.cfg)
                flaggedNbs := getFlaggedNeighbors(g.grid, i, j , g.cfg)
                numFlaggedNbs := len(flaggedNbs)

                mineConfigCandidates := combinations(freeNbs, int(cell.number) - numFlaggedNbs)
                mineConfigCandidates = g.filterCandidates(mineConfigCandidates)
                // THM 1: If a cell is flagged in all possible configuraitons, 
                // ... therefore the intersection of all mineConfigCandidates are all flags
                // THM 2: If a cell is not flagged in any configuration
                // ... therefore any cells not in the union of all possible configurations can be clicked

                if len(mineConfigCandidates) != 0 {
                    for _, flagPos := range intersection(mineConfigCandidates) {
                        g.grid[flagPos.i][flagPos.j].flagged = true
                        g.numFlags++
                    }
                    // refresh neighbor data
                    freeNbs = getUnflaggedNeighbors(g.grid, i, j, g.cfg)
                    numFlaggedNbs = len(getFlaggedNeighbors(g.grid, i, j, g.cfg))

                    for _, safePos := range difference(freeNbs, union(mineConfigCandidates)) {
                        g.grid[safePos.i][safePos.j].clicked = true
                    }

                    freeNbs = getUnflaggedNeighbors(g.grid, i, j, g.cfg)
                    numFlaggedNbs = len(getFlaggedNeighbors(g.grid, i, j, g.cfg))
                }
                

                // If the effective number is 0, click remaining neighbors
                if int(cell.number) - numFlaggedNbs == 0 {
                    for _, nb := range freeNbs {
                        g.grid[nb.i][nb.j].clicked = true
                    }
                }
                // 
            }
        }
    }
}

func flagRegion(grid *[][]Cell, region []Pos) {
    for _, flagPos := range region {
        (*grid)[flagPos.i][flagPos.j].flagged = true
    }
}
func unflagRegion(grid *[][]Cell, region []Pos) {
    for _, flagPos := range region {
        (*grid)[flagPos.i][flagPos.j].flagged = false
    }
}

func (g *Game) filterCandidates(candidates [][]Pos) [][]Pos {
    possibleCandidates := make([][]Pos, 0)
    for _, candidate := range candidates {
        if g.tryCandidate(candidate) {
            possibleCandidates = append(possibleCandidates, candidate)
        }
    }
    return possibleCandidates
}

// If any flag added in a candidate configuration violates the number of any of its neighbors,
// then that configuration is impossible
func (g *Game) tryCandidate(candidate []Pos) bool {
    flagRegion(&g.grid, candidate)
    defer unflagRegion(&g.grid, candidate)

    for i := range g.cfg.rows {
        for j := range g.cfg.cols {
            if g.grid[i][j].clicked {
                num := int(g.grid[i][j].number)
                // Check for overflagging
                numFlaggedNbs := len(getFlaggedNeighbors(g.grid, i, j, g.cfg))
                if  numFlaggedNbs > num {
                    return false
                }
                // Check for underflagging
                // I.e. if the effective number (cell.number - flaggedNbs) is greater than the number of freeNbs
                numUnflaggedNbNbs := len(getUnflaggedNeighbors(g.grid, i, j, g.cfg))
                if num - numFlaggedNbs > numUnflaggedNbNbs {
                    return false
                }
            }
        }
    }
    return true
}

