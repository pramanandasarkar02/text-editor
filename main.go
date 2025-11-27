package main

import "github.com/veandco/go-sdl2/sdl"


var(
	WINDOW_WIDTH int32 = 800
	WINDOW_HEIGHT int32 = 600
	DELAY uint32 = 16
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


	

	running := true 
	for running {



		sdl.Delay(DELAY)
		
	}
}