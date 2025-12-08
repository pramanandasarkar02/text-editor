package editor

import (
    "fmt"
    "os"
    "strings"

    "github.com/veandco/go-sdl2/sdl"
    "github.com/veandco/go-sdl2/ttf"
)

// Constants
const (
    FONT_SIZE       = 20
    TAB_SIZE        = 4
    CURSOR_BLINK    = 500 // ms
    WINDOW_WIDTH    = 1000
    WINDOW_HEIGHT   = 700
    BG_COLOR        = 30
    TEXT_COLOR      = 220
    CURSOR_COLOR    = 255
    LINE_NUMBER_COL = 120
)

type Editor struct {
    lines            [][]rune
    cursorX, cursorY int

    font       *ttf.Font
    charWidth  int
    charHeight int

    cursorVisible bool
    lastBlink     uint32
    Running       bool

    scrollOffset    int
    maxVisibleLines int

    currentFile string // Track current file
}

func NewEditor(font *ttf.Font) *Editor {
    e := &Editor{
        lines:         [][]rune{{}},
        cursorX:       0,
        cursorY:       0,
        font:          font,
        Running:       true,
        cursorVisible: true,
        lastBlink:     uint32(sdl.GetTicks64()),
        currentFile:   "",
    }

    e.charWidth, e.charHeight = e.measureChar()
    e.maxVisibleLines = (WINDOW_HEIGHT - 20) / e.charHeight
    return e
}

func (e *Editor) measureChar() (w, h int) {
    w, h, _ = e.font.SizeUTF8("M")
    return w, h + 5
}

func (e *Editor) HandleEvent(event sdl.Event) {
    switch ev := event.(type) {

    case *sdl.QuitEvent:
        e.Running = false

    case *sdl.MouseWheelEvent:
        if ev.Y > 0 {
            e.scrollUp()
        } else if ev.Y < 0 {
            e.scrollDown()
        }

    case *sdl.KeyboardEvent:
        if ev.Type == sdl.KEYDOWN {
            switch ev.Keysym.Sym {

            case sdl.K_ESCAPE:
                e.Running = false

            case sdl.K_PAGEUP:
                for i := 0; i < e.maxVisibleLines; i++ {
                    e.scrollUp()
                }

            case sdl.K_PAGEDOWN:
                for i := 0; i < e.maxVisibleLines; i++ {
                    e.scrollDown()
                }

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

            case sdl.K_o:
                if ev.Keysym.Mod&sdl.KMOD_CTRL != 0 {
                    fmt.Print("Enter file path to open: ")
                    var filePath string
                    fmt.Scanln(&filePath)
                    if filePath != "" {
                        fmt.Println("Loading", filePath, "...")
                        if err := e.LoadFromFile(filePath); err != nil {
                            fmt.Println("Error loading file:", err)
                        } else {
                            fmt.Println("File loaded successfully")
                        }
                    }
                }

            case sdl.K_s:
                if ev.Keysym.Mod&sdl.KMOD_CTRL != 0 {
                    if e.currentFile != "" {
                        fmt.Println("Saving", e.currentFile, "...")
                        if err := e.SaveToFile(e.currentFile); err != nil {
                            fmt.Println("Error saving file:", err)
                        } else {
                            fmt.Println("File saved successfully")
                        }
                    } else {
                        fmt.Println("No file loaded. Use Ctrl+O to open a file first.")
                    }
                }

            case sdl.K_TAB:
                for i := 0; i < TAB_SIZE; i++ {
                    e.insertRune(' ')
                }
            }
        }

    case *sdl.TextInputEvent:
        for _, r := range ev.GetText() {
            e.insertRune(r)
        }
    }
}

func (e *Editor) scrollUp() {
    if e.scrollOffset > 0 {
        e.scrollOffset--
    }
}

func (e *Editor) scrollDown() {
    if e.scrollOffset < len(e.lines)-e.maxVisibleLines {
        e.scrollOffset++
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
    e.ensureCursorVisible()
    e.resetBlink()
}

func (e *Editor) backspace() {
    if e.cursorX > 0 {
        line := e.lines[e.cursorY]
        e.lines[e.cursorY] = append(line[:e.cursorX-1], line[e.cursorX:]...)
        e.cursorX--
    } else if e.cursorY > 0 {
        prevLen := len(e.lines[e.cursorY-1])
        e.lines[e.cursorY-1] = append(e.lines[e.cursorY-1], e.lines[e.cursorY]...)
        e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY+1:]...)
        e.cursorY--
        e.cursorX = prevLen
        e.ensureCursorVisible()
    }
    e.resetBlink()
}

func (e *Editor) delete() {
    if e.cursorX < len(e.lines[e.cursorY]) {
        line := e.lines[e.cursorY]
        e.lines[e.cursorY] = append(line[:e.cursorX], line[e.cursorX+1:]...)
    } else if e.cursorY < len(e.lines)-1 {
        e.lines[e.cursorY] = append(e.lines[e.cursorY], e.lines[e.cursorY+1]...)
        e.lines = append(e.lines[:e.cursorY+1], e.lines[e.cursorY+2:]...)
    }
    e.resetBlink()
}

func (e *Editor) moveCursorLeft(word bool) {
    if word {
        for e.cursorX > 0 && e.lines[e.cursorY][e.cursorX-1] == ' ' {
            e.cursorX--
        }
        for e.cursorX > 0 && e.lines[e.cursorY][e.cursorX-1] != ' ' {
            e.cursorX--
        }
    } else if e.cursorX > 0 {
        e.cursorX--
    } else if e.cursorY > 0 {
        e.cursorY--
        e.cursorX = len(e.lines[e.cursorY])
        e.ensureCursorVisible()
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
        e.ensureCursorVisible()
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
    e.ensureCursorVisible()
    e.resetBlink()
}

func (e *Editor) moveCursorDown() {
    if e.cursorY < len(e.lines)-1 {
        e.cursorY++
        if e.cursorX > len(e.lines[e.cursorY]) {
            e.cursorX = len(e.lines[e.cursorY])
        }
    }
    e.ensureCursorVisible()
    e.resetBlink()
}

func (e *Editor) ensureCursorVisible() {
    if e.cursorY < e.scrollOffset {
        e.scrollOffset = e.cursorY
    } else if e.cursorY >= e.scrollOffset+e.maxVisibleLines {
        e.scrollOffset = e.cursorY - e.maxVisibleLines + 1
    }
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

    x := int32(LINE_NUMBER_COL)
    y := int32(10)

    start := e.scrollOffset
    end := start + e.maxVisibleLines
    if end > len(e.lines) {
        end = len(e.lines)
    }

    for row := start; row < end; row++ {
        // ----- Line Numbers -----
        lineNum := fmt.Sprintf("%4d", row+1)
        surface, _ := e.font.RenderUTF8Solid(lineNum, sdl.Color{R: 150, G: 150, B: 150, A: 255})
        if surface != nil {
            texture, _ := renderer.CreateTextureFromSurface(surface)
            if texture != nil {
                renderer.Copy(texture, nil, &sdl.Rect{X: 10, Y: y, W: int32(surface.W), H: int32(surface.H)})
                texture.Destroy()
            }
            surface.Free()
        }

        // ----- Text -----
        lineStr := string(e.lines[row])
        if lineStr != "" {
            surface, err := e.font.RenderUTF8Solid(lineStr, sdl.Color{R: TEXT_COLOR, G: TEXT_COLOR, B: TEXT_COLOR, A: 255})
            if err == nil {
                texture, _ := renderer.CreateTextureFromSurface(surface)
                if texture != nil {
                    renderer.Copy(texture, nil, &sdl.Rect{X: x, Y: y, W: int32(surface.W), H: int32(surface.H)})
                    texture.Destroy()
                }
                surface.Free()
            }
        }

        // ----- Cursor -----
        if row == e.cursorY && e.cursorVisible {
            cursorX := x + int32(e.cursorX*e.charWidth)
            renderer.SetDrawColor(CURSOR_COLOR, CURSOR_COLOR, CURSOR_COLOR, 255)
            renderer.FillRect(&sdl.Rect{X: cursorX, Y: y, W: 2, H: int32(e.charHeight - 5)})
        }

        y += int32(e.charHeight)
    }

    renderer.Present()
}

func (e *Editor) LoadFromFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    // Split file into lines
    rawLines := strings.Split(string(data), "\n")

    e.lines = make([][]rune, len(rawLines))
    for i, line := range rawLines {
        e.lines[i] = []rune(line)
    }

    // Ensure at least one empty line
    if len(e.lines) == 0 {
        e.lines = [][]rune{{}}
    }

    // Reset cursor & scroll
    e.cursorX = 0
    e.cursorY = 0
    e.scrollOffset = 0
    e.currentFile = path

    e.resetBlink()

    return nil
}

func (e *Editor) SaveToFile(path string) error {
    var builder strings.Builder

    for i, line := range e.lines {
        builder.WriteString(string(line))
        if i < len(e.lines)-1 {
            builder.WriteRune('\n')
        }
    }

    err := os.WriteFile(path, []byte(builder.String()), 0644)
    if err != nil {
        return err
    }

    return nil
}