package editor

import (
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var (
	APP_NAME       = "Jibon বদমাশ"
	WINDOW_WIDTH   int32 = 800
	WINDOW_HEIGHT  int32 = 600
	DELAY          uint32 = 16
	FONTSIZE       uint32 = 20
	PADDING_X      uint32 = 10
	PADDING_Y      uint32 = 5
	LINE_NUMBER_LEN uint32 = 40
	SCROLL_LEN     uint32 = 8
	INTER_LINE_SPACE uint32 = 12
	RUNNING        bool = true
	IS_SHOW_LINE_NUMBER = true
)

func DrawFrame(renderer *sdl.Renderer, font *ttf.Font) {

	// background
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	// draw line numbers
	if IS_SHOW_LINE_NUMBER {

		// how many lines can fit?
		lineCount := uint32(math.Floor(float64(WINDOW_HEIGHT) /
			(float64(FONTSIZE) + float64(INTER_LINE_SPACE))))

		for i := uint32(0); i < lineCount; i++ {

			y := int32(i*(FONTSIZE+INTER_LINE_SPACE) + PADDING_Y)

			// background bar
			renderer.SetDrawColor(40, 40, 40, 255)
			renderer.FillRect(&sdl.Rect{
				X: int32(PADDING_X),
				Y: y,
				W: int32(LINE_NUMBER_LEN),
				H: int32(FONTSIZE + INTER_LINE_SPACE),
			})

			// centered text inside bar
			DrawText(renderer, font,
				int32(PADDING_X+5),
				y,
				fmt.Sprintf("%d", i+1),
				200, 200, 200)
		}
	}

	renderer.Present()
}

func DrawText(renderer *sdl.Renderer, font *ttf.Font,
	x, y int32, text string, r, g, b byte) {

	surface, err := font.RenderUTF8Blended(text,
		sdl.Color{R: r, G: g, B: b, A: 255})
	if err != nil {
		panic(err)
	}
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	renderer.Copy(texture, nil, &sdl.Rect{
		X: x,
		Y: y,
		W: surface.W,
		H: surface.H,
	})
}
