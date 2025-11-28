package editor

import "github.com/veandco/go-sdl2/sdl"


func GetWhiteColor() *sdl.Color{
	return &sdl.Color{R: 255, G: 255, B: 255, A: 255}
}

func GetBlackColor() *sdl.Color{
	return &sdl.Color{R: 10, G: 10, B: 10, A: 255}
}
