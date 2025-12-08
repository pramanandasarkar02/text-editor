package main

import (
	"pramanandasarkar02/text-editor/editor"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Constants
const (
	FONT_SIZE     = 12
	TAB_SIZE      = 4
	CURSOR_BLINK  = 500 
	WINDOW_WIDTH  = 800
	WINDOW_HEIGHT = 600
	BG_COLOR      = 0
	TEXT_COLOR    = 200
	CURSOR_COLOR  = 255
)



func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()

	font, err := ttf.OpenFont("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf", FONT_SIZE)
	if err != nil {
		// Fallback paths
		fontPaths := []string{
			"/System/Library/Fonts/Menlo.ttc",
			"C:\\Windows\\Fonts\\consola.ttf",
			"/usr/share/fonts/TTF/DejaVuSansMono.ttf",
		}
		for _, path := range fontPaths {
			font, err = ttf.OpenFont(path, FONT_SIZE)
			if err == nil {
				break
			}
		}
		if err != nil {
			panic("Could not load any monospace font: " + err.Error())
		}
	}
	defer font.Close()

	window, err := sdl.CreateWindow("Go Text Editor", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		WINDOW_WIDTH, WINDOW_HEIGHT, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	sdl.StartTextInput()

	edr := editor.NewEditor(font)

	for edr.Running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			edr.HandleEvent(event)
		}

		edr.UpdateBlink()
		edr.Render(renderer)
		sdl.Delay(16)
	}

	sdl.StopTextInput()
}