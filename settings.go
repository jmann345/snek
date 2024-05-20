package main

import (
	"github.com/jmann345/snek/opts"
	"github.com/nsf/termbox-go"
	"time"
)

type UX struct {
	fg        termbox.Attribute // = iota = TODO : make these rgb/Color structs instead (can handle within pkl)
	bg        termbox.Attribute
	snekFg    termbox.Attribute
	snekBg    termbox.Attribute
	foodFg    termbox.Attribute
	foodBg    termbox.Attribute
	snekFgAlt termbox.Attribute
	snekBgAlt termbox.Attribute
	foodFgAlt termbox.Attribute
	foodBgAlt termbox.Attribute

	snekCh rune
	foodCh rune
}

type Cfg struct {
	rows    int
	cols    int
	speed   time.Duration
	portals bool
	snax    int
	lenGain int
}

func initConfig(opts *opts.Opts) (UX, Cfg) {
	ux := UX{
		fg: termbox.Attribute(opts.Fg),
		bg: termbox.Attribute(opts.Bg),

		// TODO : use termbox.RGBToAttribute
		snekFg:    termbox.Attribute(opts.SnekFg),
		snekFgAlt: termbox.Attribute(opts.SnekFgAlt),
		snekBg:    termbox.Attribute(opts.Bg),
		snekBgAlt: termbox.Attribute(opts.Bg),
		snekCh: func() rune { // IDEA: Add skin that prints hex value of each index in body on snek (goes from 0 to FF)
			switch opts.SnekSkin {
			case "python": // TODO: Add more snek skins
				return ''
			default:
				return ' '
			}
		}(),

		foodFg:    termbox.Attribute(opts.FoodFg), // IDEA: If allow for randomized foodCh, foodFgAlt, and foodBgAlt
		foodFgAlt: termbox.Attribute(opts.FoodFgAlt),
		foodBg:    termbox.Attribute(opts.Bg),
		foodBgAlt: termbox.Attribute(opts.Bg),
		foodCh: func() rune { // TODO :  add random food option too
			switch opts.FoodSkin {
			case "gopher":
				return ''
			default: //TODO: Add more food skins
				return ' '
			}
		}(),
	}
	if ux.snekCh == ' ' {
		ux.snekBg = ux.snekFg
		ux.snekBgAlt = ux.snekFgAlt
	}
	if ux.foodCh == ' ' {
		ux.foodBg = ux.foodFg
		ux.foodBgAlt = ux.foodFgAlt
	}

	cfg := Cfg{
		rows:    opts.Rows,
		cols:    opts.Cols,
		speed:   time.Duration(opts.Speed),
		snax:    opts.Snax, // init extra food on map
		portals: opts.Portals,

		// TODO: uncomment line below
		// lenGain int
	}

	return ux, cfg
}
