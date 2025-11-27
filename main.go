package main

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)


var(
	APP_NAME = "Jibon বদমাশ"
	WINDOW_WIDTH int32 = 800
	WINDOW_HEIGHT int32 = 600
	DELAY uint32 = 16


	// app instances
	running bool = true
	padding = 10

)





func main() {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil{
		panic(err)
	}
	defer sdl.Quit()

	window,  err := sdl.CreateWindow(APP_NAME, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WINDOW_WIDTH, WINDOW_HEIGHT, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil{
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


		renderer.SetDrawColor(255, 255, 255, 255)
		renderer.Clear()
		
		renderer.Present()



		sdl.Delay(DELAY)
		
		
	}
}