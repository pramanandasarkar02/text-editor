package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)


var(
	WINDOW_WIDTH int32 = 800
	WINDOW_HEIGHT int32 = 600
	DELAY uint32 = 16

	running bool = true
)



func main() {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil{
		panic(err)
	}
	defer sdl.Quit()

	window,  err := sdl.CreateWindow("Text-Editor", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WINDOW_WIDTH, WINDOW_HEIGHT, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
 
	for running {

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quiting event received. exiting...")
				running = false
			case *sdl.KeyboardEvent:
				keyEvent := event.(*sdl.KeyboardEvent)
				if keyEvent.Type == sdl.KEYDOWN {
					fmt.Printf("Key pressed: %s\n", keyEvent.Keysym.Sym)
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



		sdl.Delay(DELAY)
		
		
	}
}