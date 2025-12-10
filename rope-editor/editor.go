package editor

import (
    "fmt"
    "os"
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
    
    ROPE_LEAF_SIZE  = 512 
)


type RopeNode struct {
    left   *RopeNode
    right  *RopeNode
    text   []rune      
    weight int         
}

func newLeaf(text []rune) *RopeNode {
    return &RopeNode{
        text:   text,
        weight: len(text),
    }
}

func newInternal(left, right *RopeNode) *RopeNode {
    weight := left.weight
    if left.left == nil && left.right == nil {
        // left is a leaf
        weight = len(left.text)
    } else {
        weight = left.totalLength()
    }
    
    return &RopeNode{
        left:   left,
        right:  right,
        weight: weight,
    }
}

func (n *RopeNode) isLeaf() bool {
    return n.left == nil && n.right == nil
}

func (n *RopeNode) totalLength() int {
    if n == nil {
        return 0
    }
    if n.isLeaf() {
        return len(n.text)
    }
    return n.weight + n.right.totalLength()
}

// Get character at index
func (n *RopeNode) charAt(idx int) rune {
    if n.isLeaf() {
        return n.text[idx]
    }
    
    if idx < n.weight {
        return n.left.charAt(idx)
    }
    return n.right.charAt(idx - n.weight)
}

// Insert text at index
func (n *RopeNode) insert(idx int, text []rune) *RopeNode {
    if n.isLeaf() {
        // Split leaf and insert
        newText := make([]rune, len(n.text)+len(text))
        copy(newText[:idx], n.text[:idx])
        copy(newText[idx:idx+len(text)], text)
        copy(newText[idx+len(text):], n.text[idx:])
        
        // If result is too large, split into multiple leaves
        if len(newText) <= ROPE_LEAF_SIZE {
            return newLeaf(newText)
        }
        
        mid := len(newText) / 2
        return newInternal(newLeaf(newText[:mid]), newLeaf(newText[mid:]))
    }
    
    if idx <= n.weight {
        return newInternal(n.left.insert(idx, text), n.right)
    }
    return newInternal(n.left, n.right.insert(idx-n.weight, text))
}

// Delete range [start, end)
func (n *RopeNode) delete(start, end int) *RopeNode {
    if n.isLeaf() {
        newText := make([]rune, 0, len(n.text)-(end-start))
        newText = append(newText, n.text[:start]...)
        newText = append(newText, n.text[end:]...)
        return newLeaf(newText)
    }
    
    if end <= n.weight {
        return newInternal(n.left.delete(start, end), n.right)
    } else if start >= n.weight {
        return newInternal(n.left, n.right.delete(start-n.weight, end-n.weight))
    } else {
        // Range spans both children
        newLeft := n.left.delete(start, n.weight)
        newRight := n.right.delete(0, end-n.weight)
        return newInternal(newLeft, newRight)
    }
}

// Convert rope to string
func (n *RopeNode) toString() string {
    if n == nil {
        return ""
    }
    if n.isLeaf() {
        return string(n.text)
    }
    return n.left.toString() + n.right.toString()
}



type Editor struct {

    rope *RopeNode
    
    cursorX, cursorY int

    font       *ttf.Font
    charWidth  int
    charHeight int

    cursorVisible bool
    lastBlink     uint32
    Running       bool

    scrollOffset    int
    maxVisibleLines int

    currentFile string
    
    // Cache for line boundaries (updated on text change)
    lineStarts []int  // Start index of each line in rope
}

func NewEditor(font *ttf.Font) *Editor {
    e := &Editor{
        rope:          newLeaf([]rune{}),
        cursorX:       0,
        cursorY:       0,
        font:          font,
        Running:       true,
        cursorVisible: true,
        lastBlink:     uint32(sdl.GetTicks64()),
        currentFile:   "",
        lineStarts:    []int{0},
    }

    e.charWidth, e.charHeight = e.measureChar()
    e.maxVisibleLines = (WINDOW_HEIGHT - 20) / e.charHeight
    return e
}

func (e *Editor) measureChar() (w, h int) {
    w, h, _ = e.font.SizeUTF8("M")
    return w, h + 5
}

// Update line boundary cache
func (e *Editor) updateLineCache() {
    text := e.rope.toString()
    e.lineStarts = []int{0}
    
    for i, r := range text {
        if r == '\n' {
            e.lineStarts = append(e.lineStarts, i+1)
        }
    }
}

// Get absolute position in rope from cursor position
func (e *Editor) cursorToRopeIndex() int {
    if e.cursorY >= len(e.lineStarts) {
        return e.rope.totalLength()
    }
    return e.lineStarts[e.cursorY] + e.cursorX
}

// Get line text for rendering
func (e *Editor) getLine(lineNum int) []rune {
    if lineNum >= len(e.lineStarts) {
        return []rune{}
    }
    
    start := e.lineStarts[lineNum]
    end := e.rope.totalLength()
    if lineNum+1 < len(e.lineStarts) {
        end = e.lineStarts[lineNum+1] - 1 
    }
    
    if start >= end {
        return []rune{}
    }
    
    line := make([]rune, end-start)
    for i := start; i < end; i++ {
        line[i-start] = e.rope.charAt(i)
    }
    return line
}

func (e *Editor) lineCount() int {
    return len(e.lineStarts)
}

func (e *Editor) lineLength(lineNum int) int {
    return len(e.getLine(lineNum))
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
                if ev.Keysym.Mod&sdl.KMOD_CTRL != 0 {
                    e.cursorY = 0
                    e.cursorX = 0
                    e.scrollOffset = 0
                } else {
                    e.cursorX = 0
                }

            case sdl.K_END:
                if ev.Keysym.Mod&sdl.KMOD_CTRL != 0 {
                    e.cursorY = e.lineCount() - 1
                    e.cursorX = e.lineLength(e.cursorY)
                    e.ensureCursorVisible()
                } else {
                    e.cursorX = e.lineLength(e.cursorY)
                }

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
                    filePath := e.currentFile
                    if filePath == "" {
                        filePath = "demo.txt"
                        e.currentFile = filePath
                        fmt.Println("No file loaded. Saving to default:", filePath)
                    } else {
                        fmt.Println("Saving", filePath, "...")
                    }
                    
                    if err := e.SaveToFile(filePath); err != nil {
                        fmt.Println("Error saving file:", err)
                    } else {
                        fmt.Println("File saved successfully to", filePath)
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
    if e.scrollOffset < e.lineCount()-e.maxVisibleLines {
        e.scrollOffset++
    }
}

func (e *Editor) insertRune(r rune) {
    // ROPE APPROACH
    idx := e.cursorToRopeIndex()
    e.rope = e.rope.insert(idx, []rune{r})
    e.updateLineCache()
    e.cursorX++
    e.resetBlink()
}

func (e *Editor) insertNewline() {
    // ROPE APPROACH
    idx := e.cursorToRopeIndex()
    e.rope = e.rope.insert(idx, []rune{'\n'})
    e.updateLineCache()
    e.cursorY++
    e.cursorX = 0
    e.ensureCursorVisible()
    e.resetBlink()
}

func (e *Editor) backspace() {
    // ROPE APPROACH
    if e.cursorX > 0 {
        idx := e.cursorToRopeIndex()
        e.rope = e.rope.delete(idx-1, idx)
        e.updateLineCache()
        e.cursorX--
    } else if e.cursorY > 0 {
        prevLen := e.lineLength(e.cursorY - 1)
        idx := e.cursorToRopeIndex()
        e.rope = e.rope.delete(idx-1, idx)
        e.updateLineCache()
        e.cursorY--
        e.cursorX = prevLen
        e.ensureCursorVisible()
    }
    e.resetBlink()
}

func (e *Editor) delete() {
    // ROPE APPROACH
    idx := e.cursorToRopeIndex()
    if idx < e.rope.totalLength() {
        e.rope = e.rope.delete(idx, idx+1)
        e.updateLineCache()
    }
    e.resetBlink()
}

func (e *Editor) moveCursorLeft(word bool) {
    line := e.getLine(e.cursorY)
    
    if word {
        for e.cursorX > 0 && line[e.cursorX-1] == ' ' {
            e.cursorX--
        }
        for e.cursorX > 0 && line[e.cursorX-1] != ' ' {
            e.cursorX--
        }
    } else if e.cursorX > 0 {
        e.cursorX--
    } else if e.cursorY > 0 {
        e.cursorY--
        e.cursorX = e.lineLength(e.cursorY)
        e.ensureCursorVisible()
    }
    e.resetBlink()
}

func (e *Editor) moveCursorRight(word bool) {
    line := e.getLine(e.cursorY)
    lineLen := len(line)
    
    if word {
        for e.cursorX < lineLen && line[e.cursorX] == ' ' {
            e.cursorX++
        }
        for e.cursorX < lineLen && line[e.cursorX] != ' ' {
            e.cursorX++
        }
    } else if e.cursorX < lineLen {
        e.cursorX++
    } else if e.cursorY < e.lineCount()-1 {
        e.cursorY++
        e.cursorX = 0
        e.ensureCursorVisible()
    }
    e.resetBlink()
}

func (e *Editor) moveCursorUp() {
    if e.cursorY > 0 {
        e.cursorY--
        if e.cursorX > e.lineLength(e.cursorY) {
            e.cursorX = e.lineLength(e.cursorY)
        }
    }
    e.ensureCursorVisible()
    e.resetBlink()
}

func (e *Editor) moveCursorDown() {
    if e.cursorY < e.lineCount()-1 {
        e.cursorY++
        if e.cursorX > e.lineLength(e.cursorY) {
            e.cursorX = e.lineLength(e.cursorY)
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
    if end > e.lineCount() {
        end = e.lineCount()
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
        line := e.getLine(row)
        lineStr := string(line)
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

    // ROPE APPROACH
    text := []rune(string(data))
    e.rope = newLeaf(text)
    e.updateLineCache()
    e.cursorX = 0
    e.cursorY = 0
    e.scrollOffset = 0
    e.currentFile = path
    e.resetBlink()

    return nil
}

func (e *Editor) SaveToFile(path string) error {
    // ROPE APPROACH
    text := e.rope.toString()
    err := os.WriteFile(path, []byte(text), 0644)
    if err != nil {
        return err
    }

    return nil
}