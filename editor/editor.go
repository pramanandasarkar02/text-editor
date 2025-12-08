package editor


import (
	"fmt"
	"os"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Constants
const (
	FONT_SIZE     = 20
	TAB_SIZE      = 4
	CURSOR_BLINK  = 500 // ms
	WINDOW_WIDTH  = 1000
	WINDOW_HEIGHT = 700
	BG_COLOR      = 30
	TEXT_COLOR    = 220
	CURSOR_COLOR  = 255
)

type Editor struct {
	lines          [][]rune       // 2D array of runes (each line)
	cursorX, cursorY int          // Cursor position in characters
	        
	font           *ttf.Font
	charWidth      int
	charHeight     int
	cursorVisible  bool
	lastBlink      uint32
	Running        bool
}

func NewEditor(font *ttf.Font) *Editor {
	e := &Editor{
		lines:         [][]rune{[]rune{}},
		cursorX:       0,
		cursorY:       0,
		font:          font,
		Running:       true,
		cursorVisible: true,
		lastBlink:     uint32(sdl.GetTicks64()),
	}
	e.charWidth, e.charHeight = e.measureChar()
	e.lines = append(e.lines, []rune{}) // Start with one empty line
	return e
}

func (e *Editor) measureChar() (w, h int) {
	w, h, _ = e.font.SizeUTF8("M")
	return w, h + 5 // little line spacing
}

func (e *Editor) HandleEvent(event sdl.Event) {
	switch ev := event.(type) {
	case *sdl.QuitEvent:
		e.Running = false

	case *sdl.KeyboardEvent:
		if ev.Type == sdl.KEYDOWN {
			switch ev.Keysym.Sym {
			case sdl.K_ESCAPE:
				e.Running = false
			case sdl.K_RETURN:
				e.insertNewline()
			case sdl.K_BACKSPACE:
				e.backspace()
			case sdl.K_DELETE:
				e.delete()
			case sdl.K_LEFT:
				e.moveCursorLeft(ev.Keysym.Mod&sdl.KMOD_CTRL != 0)
			case sdl.K_RIGHT:
				e.moveCursorRight(ev.Keysym.Mod&sdl.KMOD_CTRL != 0)
			case sdl.K_UP:
				e.moveCursorUp()
			case sdl.K_DOWN:
				e.moveCursorDown()
			case sdl.K_HOME:
				e.cursorX = 0
			case sdl.K_END:
				e.cursorX = len(e.lines[e.cursorY])
			}
		}

	case *sdl.TextInputEvent:
		text := ev.GetText()
		for _, r := range text {
			e.insertRune(r)
		}
	}
}

func (e *Editor) insertRune(r rune) {
	line := e.lines[e.cursorY]
	newLine := make([]rune, len(line)+1)
	copy(newLine[:e.cursorX], line[:e.cursorX])
	newLine[e.cursorX] = r
	copy(newLine[e.cursorX+1:], line[e.cursorX:])
	e.lines[e.cursorY] = newLine
	e.cursorX++
	e.resetBlink()
}

func (e *Editor) insertNewline() {
	current := e.lines[e.cursorY]
	left := current[:e.cursorX]
	right := current[e.cursorX:]

	e.lines[e.cursorY] = left
	e.lines = append(e.lines[:e.cursorY+1], append([][]rune{right}, e.lines[e.cursorY+1:]...)...)
	e.cursorY++
	e.cursorX = 0
	e.resetBlink()
}

func (e *Editor) backspace() {
	if e.cursorX > 0 {
		line := e.lines[e.cursorY]
		newLine := append(line[:e.cursorX-1], line[e.cursorX:]...)
		e.lines[e.cursorY] = newLine
		e.cursorX--
	} else if e.cursorY > 0 {
		// Join with previous line
		prevLen := len(e.lines[e.cursorY-1])
		e.lines[e.cursorY-1] = append(e.lines[e.cursorY-1], e.lines[e.cursorY]...)
		e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY+1:]...)
		e.cursorY--
		e.cursorX = prevLen
	}
	e.resetBlink()
}

func (e *Editor) delete() {
	if e.cursorX < len(e.lines[e.cursorY]) {
		line := e.lines[e.cursorY]
		newLine := append(line[:e.cursorX], line[e.cursorX+1:]...)
		e.lines[e.cursorY] = newLine
	} else if e.cursorY < len(e.lines)-1 {
		e.lines[e.cursorY] = append(e.lines[e.cursorY], e.lines[e.cursorY+1]...)
		e.lines = append(e.lines[:e.cursorY+1], e.lines[e.cursorY+2:]...)
	}
	e.resetBlink()
}

func (e *Editor) moveCursorLeft(word bool) {
	if word {
		for e.cursorX > 0 {
			e.cursorX--
			if e.cursorX > 0 && e.lines[e.cursorY][e.cursorX-1] == ' ' {
				break
			}
		}
		for e.cursorX > 0 && e.lines[e.cursorY][e.cursorX-1] != ' ' {
			e.cursorX--
		}
	} else if e.cursorX > 0 {
		e.cursorX--
	} else if e.cursorY > 0 {
		e.cursorY--
		e.cursorX = len(e.lines[e.cursorY])
	}
	e.resetBlink()
}

func (e *Editor) moveCursorRight(word bool) {
	lineLen := len(e.lines[e.cursorY])
	if word {
		for e.cursorX < lineLen && e.lines[e.cursorY][e.cursorX] == ' ' {
			e.cursorX++
		}
		for e.cursorX < lineLen && e.lines[e.cursorY][e.cursorX] != ' ' {
			e.cursorX++
		}
	} else if e.cursorX < lineLen {
		e.cursorX++
	} else if e.cursorY < len(e.lines)-1 {
		e.cursorY++
		e.cursorX = 0
	}
	e.resetBlink()
}

func (e *Editor) moveCursorUp() {
	if e.cursorY > 0 {
		e.cursorY--
		if e.cursorX > len(e.lines[e.cursorY]) {
			e.cursorX = len(e.lines[e.cursorY])
		}
	}
	e.resetBlink()
}

func (e *Editor) moveCursorDown() {
	if e.cursorY < len(e.lines)-1 {
		e.cursorY++
		if e.cursorX > len(e.lines[e.cursorY]) {
			e.cursorX = len(e.lines[e.cursorY])
		}
	}
	e.resetBlink()
}

func (e *Editor) resetBlink() {
	e.cursorVisible = true
	e.lastBlink = uint32(sdl.GetTicks64())
}

func (e *Editor) UpdateBlink() {
	now := uint32(sdl.GetTicks64())
	if now-e.lastBlink > CURSOR_BLINK {
		e.cursorVisible = !e.cursorVisible
		e.lastBlink = now
	}
}

func (e *Editor) Render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(BG_COLOR, BG_COLOR, BG_COLOR, 255)
	renderer.Clear()

	x := int32(10)
	y := int32(10)

	renderer.SetDrawColor(TEXT_COLOR, TEXT_COLOR, TEXT_COLOR, 255)

	for row := 0; row < len(e.lines); row++ {
		line := e.lines[row]
		lineStr := string(line)

		if lineStr != "" {
			surface, err := e.font.RenderUTF8Solid(lineStr, sdl.Color{R: TEXT_COLOR, G: TEXT_COLOR, B: TEXT_COLOR, A: 255})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to render text: %v\n", err)
				continue
			}
			defer surface.Free()

			texture, err := renderer.CreateTextureFromSurface(surface)
			if err != nil {
				continue
			}
			defer texture.Destroy()

			_, _, w, h, _ := texture.Query()
			renderer.Copy(texture, nil, &sdl.Rect{X: x, Y: y, W: w, H: h})
		}

		// Draw cursor if on this line
		if row == e.cursorY && e.cursorVisible {
			cursorX := x + int32(e.cursorX*e.charWidth)
			renderer.SetDrawColor(CURSOR_COLOR, CURSOR_COLOR, CURSOR_COLOR, 255)
			renderer.FillRect(&sdl.Rect{X: cursorX, Y: y, W: 2, H: int32(e.charHeight - 5)})
		}

		y += int32(e.charHeight)
	}

	renderer.Present()
}
