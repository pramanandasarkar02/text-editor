package main

import (
	"fmt"
	"os"
	"pramanandasarkar02/text-editor/editor"
	"pramanandasarkar02/text-editor/window"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var (

	// app instances
	running bool = true
	padding      = 10
	Font    *ttf.Font
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

	Font, err := ttf.OpenFont("/usr/share/fonts/TTF/DejaVuSansMono.ttf", int(editor.FONTSIZE))
	if err != nil {
		panic(err)
	}
	defer Font.Close()

	window, err := sdl.CreateWindow(window.APP_NAME, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, window.WINDOW_WIDTH, window.WINDOW_HEIGHT, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}
	defer renderer.Destroy()

	sdl.StartTextInput()
	defer sdl.StopTextInput()

	for running {

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quiting event received. exiting...")
				running = false
			case *sdl.KeyboardEvent:
				keyEvent := event.(*sdl.KeyboardEvent)
				if keyEvent.Type == sdl.KEYDOWN {
					fmt.Printf("Key pressed: %v\n", keyEvent.Keysym.Sym)
					if keyEvent.Keysym.Sym == sdl.K_ESCAPE {
						running = false
					}
				}
			case *sdl.MouseButtonEvent:
				mouseEvent := event.(*sdl.MouseButtonEvent)
				if mouseEvent.Type == sdl.MOUSEBUTTONDOWN {
					fmt.Printf("Mouse button clicked at (%d, %d)\n", mouseEvent.X, mouseEvent.Y)
				}
			}

		}

		// render
		// renderer.SetDrawColor(30, 30, 30, 255)
		// renderer.Clear()

		// // render cursor
		// cursorX := 20
		// cursorY := 20

		// renderer.SetDrawColor(255, 255, 255, 255)
		// renderer.FillRect(&sdl.Rect{X: int32(cursorX), Y: int32(cursorY), W: 2, H: int32(editor.FONTSIZE)})

		// renderer.Present()

		editor.DrawFrame(renderer, Font)

		sdl.Delay(16)

	}
}
