package main
import "github.com/nsf/termbox-go"

func setSquare(x, y int, ch rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x*2, y, ch, fg, bg)
	termbox.SetCell(x*2+1, y, ch, fg, bg)
}

func writeStr(x, y int, str string, fg, bg termbox.Attribute) {
	for i, c := range str {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}

