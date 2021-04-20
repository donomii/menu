// +build !VR

package main

import (
	"io"

	"unicode"

	"github.com/donomii/glim"

	//"github.com/mitchellh/go-homedir"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/atotto/clipboard"
)

type GlobalConfig struct {
	ActiveBuffer   *Buffer
	ActiveBufferId int
	BufferList     []*Buffer
	LogBuffer      *Buffer
	StatusBuffer   *Buffer
	CommandBuffer  *Buffer
}

type BufferData struct {
	Text     string //FIXME rename Buffer to View, have proper text buffer manager
	FileName string
}

type Buffer struct {
	Data      *BufferData
	InputMode bool
	Formatter *glim.FormatParams
}

var fname string

func NewBuffer() *Buffer {
	buf := Buffer{}
	buf.Data = &BufferData{}
	buf.Formatter = glim.NewFormatter()
	buf.Data.Text = ""
	buf.Data.FileName = ""
	return &buf
}

//Create a new buffer, make it active and set its contents.  file name is required for a unique key to index it
//If a buffer called fileName already exists, its data will be replaced with the new data
func AddActiveBuffer(gc *GlobalConfig, text string, fileName string) {
	buff := NewBuffer()
	_, fbuff := FindByFileName(gc, fileName)
	if fbuff == nil {
		gc.BufferList = append(gc.BufferList, buff)
	} else {
		buff = fbuff
	}
	buff.Data.Text = text
	gc.ActiveBuffer = buff
}

func FindByFileName(gc *GlobalConfig, fileName string) (int, *Buffer) {
	for i, v := range gc.BufferList {
		//fmt.Println("Comparing ", fileName, v.Data.FileName)
		if v.Data.FileName == fileName {
			return i, v
		}
	}
	return -1, nil
}

func NewEditor() *GlobalConfig {
	var gc GlobalConfig
	gc.ActiveBuffer = NewBuffer()
	gc.ActiveBuffer.Formatter = glim.NewFormatter()
	gc.ActiveBuffer.Data.Text = ``
	gc.ActiveBuffer.Data.FileName = "Welcome"
	gc.StatusBuffer = NewBuffer()
	gc.StatusBuffer.Formatter = glim.NewFormatter()
	gc.StatusBuffer.Data.Text = `Status window`
	gc.StatusBuffer.Data.FileName = "Status"
	gc.BufferList = []*Buffer{NewBuffer(), gc.ActiveBuffer, gc.StatusBuffer, NewBuffer()}
	return &gc

}

func Log2Buff(gc *GlobalConfig, s string) {
	gc.StatusBuffer.Data.Text = s
}

//Does a page up, by searching backwards util the old top line is off the bottom of the screen
func SearchBackPage(txtBuf string, orig_f *glim.FormatParams, screenWidth, screenHeight int) int {
	input := *orig_f
	x := input.StartLinePos
	newLastDrawn := input.LastDrawnCharPos
	for x = input.Cursor; x > 0 && input.FirstDrawnCharPos < newLastDrawn; x = ScanToPrevLine(txtBuf, x) {
		f := input
		f.FirstDrawnCharPos = x

		glim.RenderPara(&f, 0, 0, 0, 0, screenWidth/2, screenHeight, screenWidth/2, screenHeight, 0, 0, nil, txtBuf, false, false, false)
		newLastDrawn = f.LastDrawnCharPos
	}
	return x
}

func DumpBuffer(gc *GlobalConfig, b *Buffer) {
	Log2Buff(gc, fmt.Sprintf(`
FileName: %v,
Active Buffer: %v,
StartChar: %v,
LastChar: %v,
Cursor: %v,
Tail: %v,
Font Size: %v,
`, b.Data.FileName, gc.ActiveBufferId, b.Formatter.FirstDrawnCharPos, b.Formatter.LastDrawnCharPos, b.Formatter.Cursor, b.Formatter.TailBuffer, b.Formatter.FontSize))
}

func ScanToPrevPara(txt string, c int) int {
	log.Println("To Previous Line")
	letters := strings.Split(txt, "")
	x := c
	for x = c - 1; x > 1 && x < len(txt) && !(letters[x-1] == "\n" && letters[x] != "\n"); x-- {
	}
	return x
}

func ScanToPrevLine(txt string, c int) int {
	log.Println("To Previous Line")
	letters := strings.Split(txt, "")
	x := c
	for x = c - 1; x > 1 && x < len(txt) && !(letters[x-1] == "\n"); x-- {
	}
	return x
}

func Is_space(l string) bool {
	if (l == " ") ||
		(l == "\n") ||
		(l == "\t") {
		return true
	}
	return false
}

func SOL(txt string, c int) int {
	if c == 0 {
		return c
	}
	letters := strings.Split(txt, "")
	if letters[c-1] == "\n" {
		return c
	}
	s := ScanToPrevLine(txt, c)
	return s
}
func SOT(txt string, c int) int { //Start of Text
	s := SOL(txt, c)
	letters := strings.Split(txt, "")
	x := c
	for x = s; x > 1 && x < len(txt) && (unicode.IsSpace([]rune(letters[x])[0])); x++ {
	}
	return x
}

func ScanToNextPara(txt string, c int) int {
	letters := strings.Split(txt, "")
	x := c
	for x = c + 1; x > 1 && x < len(txt) && !(letters[x-1] == "\n" && letters[x] != "\n"); x++ {
	}
	return x
}

func ScanToNextLine(txt string, c int) int {
	letters := strings.Split(txt, "")
	x := c
	for x = c + 1; x > 1 && x < len(txt) && !(letters[x-1] == "\n"); x++ {
	}
	if x == len(txt) {
		return c
	} else {
		return x
	}
}

func ScanToEndOfLine(txt string, c int) int {
	log.Println("To EOL")
	letters := strings.Split(txt, "")
	x := c
	for x = c + 1; x > 0 && x < len(txt) && !(letters[x] == "\n"); x++ {
	}
	return x
}

func DeleteLeft(t string, p int) string {
	log.Println("Delete left")
	if p > 0 {
		return strings.Join([]string{t[:p-1], t[p:]}, "")
	}
	return t
}

func SaveFile(gc *GlobalConfig, fname string, txt string) {
	Log2Buff(gc, fmt.Sprintf("Saving: %v", fname))
	err := ioutil.WriteFile(fname, []byte(txt), 0644)
	check(err, "saving file")
	Log2Buff(gc, fmt.Sprintf("File saved: %v", fname))
}

func check(e error, msg string) {
	if e != nil {
		log.Println("Error while ", msg, " : ", e)
	}
}

func ProcessPort(gc *GlobalConfig, r io.Reader) {
	for {
		buf := make([]byte, 1)
		if _, err := io.ReadAtLeast(r, buf, 5); err != nil {
			//log.Fatal(err)
		}
		ActiveBufferAppend(gc, string(buf))
		gc.ActiveBuffer.Formatter.Cursor = len(gc.ActiveBuffer.Data.Text)
	}
}

func LoadFileIfNotLoaded(gc *GlobalConfig, fileName string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("Could not load", fileName, "because", err)
	}

	buff := NewBuffer()
	_, fbuff := FindByFileName(gc, fileName)
	if fbuff == nil {
		//fmt.Printf("Loading file from disk: %v\n", fileName)
		gc.ActiveBuffer = buff
		gc.BufferList = append(gc.BufferList, buff)
		gc.ActiveBuffer.Data.Text = string(data)
		gc.ActiveBuffer.Data.FileName = fileName
		gc.ActiveBuffer.Formatter.Cursor = len(gc.ActiveBuffer.Data.Text)
	} else {
		//fmt.Printf("Reusing buffer for %v\n", fileName)
		gc.ActiveBuffer = fbuff
	}

}

func BuffIdAppend(gc *GlobalConfig, buffId int, txt string) {
	gc.BufferList[buffId].Data.Text = strings.Join([]string{gc.BufferList[buffId].Data.Text, txt}, "")
}

func BuffAppend(buff *Buffer, txt string) {
	buff.Data.Text = strings.Join([]string{buff.Data.Text, txt}, "")
}

func ActiveBufferAppend(gc *GlobalConfig, txt string) {
	gc.ActiveBuffer.Data.Text = strings.Join([]string{gc.ActiveBuffer.Data.Text, txt}, "")
}

func ActiveBufferInsert(gc *GlobalConfig, txt string) {
	if gc.ActiveBuffer.Formatter.Cursor < 0 {
		log.Println("Warning: cursor position < 0")
		gc.ActiveBuffer.Formatter.Cursor = 0
	}
	log.Printf("Inserting at %v, length %v\n", gc.ActiveBuffer.Formatter.Cursor, len(txt))
	gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", gc.ActiveBuffer.Data.Text[:ed.ActiveBuffer.Formatter.Cursor], txt, gc.ActiveBuffer.Data.Text[ed.ActiveBuffer.Formatter.Cursor:])
	gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor + len(txt)
	log.Printf("Cursor now %v\n", gc.ActiveBuffer.Formatter.Cursor)
}

func PageDown(buf *Buffer) {
	log.Println("Scanning to start of next page from ", buf.Formatter.LastDrawnCharPos)
	buf.Formatter.FirstDrawnCharPos = ScanToPrevLine(buf.Data.Text, buf.Formatter.LastDrawnCharPos)
	buf.Formatter.Cursor = buf.Formatter.FirstDrawnCharPos
}

func ScrollToCursor(buf *Buffer) {
	buf.Formatter.FirstDrawnCharPos = buf.Formatter.Cursor
}

func ExciseSelection(buf *Buffer) {
	if buf.Formatter.SelectStart >= 0 && buf.Formatter.SelectStart < len(buf.Data.Text) {
		if buf.Formatter.SelectEnd > 0 && buf.Formatter.SelectEnd < len(buf.Data.Text) {
			//log.Println("Clipping from ", buf.Formatter.SelectStart, " to ", buf.Formatter.SelectEnd)
			buf.Data.Text = fmt.Sprintf("%s%s",
				buf.Data.Text[:buf.Formatter.SelectStart],
				buf.Data.Text[buf.Formatter.SelectEnd+1:])
			buf.Formatter.SelectStart = 0
			buf.Formatter.SelectEnd = 0
		}
	}
}

func DecreaseFont(buf *Buffer) {
	buf.Formatter.FontSize -= 1
	glim.ClearAllCaches()

}

func SetFont(buf *Buffer, size float64) {
	buf.Formatter.FontSize = size
	//fmt.Println("Font size", buf.Formatter.FontSize)
	glim.ClearAllCaches()
}

func IncreaseFont(buf *Buffer) {
	buf.Formatter.FontSize += 1
	//fmt.Println("Font size", buf.Formatter.FontSize)
	glim.ClearAllCaches()
}
func ClearCaches(gc *GlobalConfig) {
	glim.ClearAllCaches()
}

func DoPageDown(gc *GlobalConfig, buf *Buffer) {
	PageDown(gc.ActiveBuffer)
}

func PreviousCharacter(buf *Buffer) {
	buf.Formatter.Cursor = buf.Formatter.Cursor - 1
}

func NextBuffer(gc *GlobalConfig) {
	gc.ActiveBufferId++
	if gc.ActiveBufferId > len(gc.BufferList)-1 {
		gc.ActiveBufferId = 0
	}
	gc.ActiveBuffer = gc.BufferList[gc.ActiveBufferId]
	log.Printf("Next buffer: %v, %v", gc.ActiveBufferId, gc.ActiveBuffer.Data.FileName)
}

func PreviousBuffer(gc *GlobalConfig) {
	gc.ActiveBufferId--
	if gc.ActiveBufferId < 0 {
		gc.ActiveBufferId = len(gc.BufferList) - 1
	}
	gc.ActiveBuffer = gc.BufferList[gc.ActiveBufferId]
	Log2Buff(gc, fmt.Sprintf("Previous buffer: %v, %v", gc.ActiveBufferId, gc.ActiveBuffer.Data.FileName))
}

func ToggleVerticalMode(gc *GlobalConfig) {
	if gc.ActiveBuffer.Formatter.Vertical {
		dispatch("HORIZONTAL-MODE", gc)
	} else {
		dispatch("VERTICAL-MODE", gc)
	}
}

func ClearBuffer(buff *Buffer) {
	buff.Data.Text = ""
	buff.Formatter.Cursor = 0
}

func ClearActiveBuffer(gc *GlobalConfig) {
	ClearBuffer(gc.ActiveBuffer)
}

func SetBuffer(buff *Buffer, text string) {
	buff.Data.Text = text
	buff.Formatter.Cursor = 0
}

func PasteFromClipBoard(gc *GlobalConfig) {
	text, _ := clipboard.ReadAll()
	dispatch("EXCISE-SELECTION", gc)

	if gc.ActiveBuffer.Formatter.Cursor < 0 {
		gc.ActiveBuffer.Formatter.Cursor = 0
	}

	ActiveBufferInsert(gc, text)
}

//This function carries out commands.  It is the interface between your scripting, and the actual engine operation
func dispatch(command string, gc *GlobalConfig) {
	switch command {
	case "CLEAR-CACHES":
		ClearCaches(gc)
	case "DELETE-LEFT":
		if gc.ActiveBuffer.Formatter.Cursor > 0 {
			gc.ActiveBuffer.Data.Text = DeleteLeft(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
			gc.ActiveBuffer.Formatter.Cursor--
		}
	case "WHEEL-UP":
		gc.ActiveBuffer.Formatter.Cursor = ScanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "WHEEL-DOWN":
		gc.ActiveBuffer.Formatter.Cursor = ScanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "EXCISE-SELECTION": //Cut
		ExciseSelection(gc.ActiveBuffer)
	case "DECREASE-FONT":
		DecreaseFont(gc.ActiveBuffer)
	case "INCREASE-FONT":
		IncreaseFont(gc.ActiveBuffer)
		//	case "PAGEDOWN":
	//	DoPageDown(gc.ActiveBuffer)
	//case "PAGEUP":
	//PageUp(gc.ActiveBuffer, screenWidth, screenHeight)
	case "PREVIOUS-CHARACTER":
		PreviousCharacter(gc.ActiveBuffer)
	case "NEXT-CHARACTER":
		gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor + 1
	case "PREVIOUS-LINE":
		gc.ActiveBuffer.Formatter.Cursor = ScanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "NEXT-LINE":
		gc.ActiveBuffer.Formatter.Cursor = ScanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "NEXT-BUFFER":
		NextBuffer(gc)
	case "PREVIOUS-BUFFER":
		PreviousBuffer(gc)
	case "CLEAR-BUFFER":
		ClearActiveBuffer(gc)
	case "INPUT-MODE":
		gc.ActiveBuffer.InputMode = true
	case "START-OF-LINE":
		gc.ActiveBuffer.Formatter.Cursor = SOL(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "HORIZONTAL-MODE":
		gc.ActiveBuffer.Formatter.Vertical = false
	case "VERTICAL-MODE":
		gc.ActiveBuffer.Formatter.Vertical = true
	case "TOGGLE-VERTICAL-MODE":
		ToggleVerticalMode(gc)
	case "PASTE-FROM-CLIPBOARD":
		PasteFromClipBoard(ed)
	case "COPY-TO-CLIPBOARD":
		clipboard.WriteAll(gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.SelectStart : gc.ActiveBuffer.Formatter.SelectEnd+1])
	case "CUT-TO-CLIPBOARD":
		dispatch("COPY-TO-CLIPBOARD", gc)
		dispatch("EXCISE-SELECTION", gc)
		gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.SelectStart
		gc.ActiveBuffer.Formatter.SelectStart = -1
		gc.ActiveBuffer.Formatter.SelectEnd = -1
	case "SAVE-FILE":
		SaveFile(gc, gc.ActiveBuffer.Data.FileName, gc.ActiveBuffer.Data.Text)
	case "SEEK-EOL":
		gc.ActiveBuffer.Formatter.Cursor = ScanToEndOfLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "END-OF-LINE":
		gc.ActiveBuffer.Formatter.Cursor = ScanToEndOfLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "TAIL":
		gc.ActiveBuffer.Formatter.TailBuffer = true
	case "START-OF-TEXT-ON-LINE":
		gc.ActiveBuffer.Formatter.Cursor = SOT(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	}
}

func PageUp(buf *Buffer, w, h int) {
	log.Println("Page up")
	start := SearchBackPage(buf.Data.Text, buf.Formatter, w, h)
	log.Println("New start at ", start)
	buf.Formatter.FirstDrawnCharPos = start
	buf.Formatter.Cursor = buf.Formatter.FirstDrawnCharPos
}

/*
func handleEvent(a app.App, i interface{}) {
	log.Println(i)
	DumpBuffer(gc.ActiveBuffer)
	switch e := a.Filter(i).(type) {
	case key.Event:
		switch e.Code {
		case key.CodeDeleteBackspace:
			if gc.ActiveBuffer.Formatter.Cursor > 0 {
				gc.ActiveBuffer.Data.Text = deleteLeft(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
				gc.ActiveBuffer.Formatter.Cursor--
			}
		case key.CodeHome:
			gc.ActiveBuffer.Formatter.Cursor = SOL(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
		case key.CodeEnd:
			dispatch("SEEK-EOL", gc)
		case key.CodeLeftArrow:
			dispatch("PREVIOUS-CHARACTER", gc)
		case key.CodeRightArrow:
			dispatch("NEXT-CHARACTER", gc)
		case key.CodeUpArrow:
			dispatch("PREVIOUS-LINE", gc)
		case key.CodeDownArrow:
			dispatch("NEXT-LINE", gc)
		case key.CodePageDown:
			dispatch("PAGEDOWN", gc)
		case key.CodePageUp:
			dispatch("PAGEUP", gc)
		case key.CodeA:
			if e.Modifiers > 0 {
				gc.ActiveBuffer.Formatter.SelectStart = 0
				gc.ActiveBuffer.Formatter.SelectEnd = len(gc.ActiveBuffer.Data.Text) - 1
				return
			}
		case key.CodeC:
			if e.Modifiers > 0 {
				dispatch("COPY-TO-CLIPBOARD", gc)
				return
			}
		case key.CodeX:
			if e.Modifiers > 0 {
				dispatch("COPY-TO-CLIPBOARD", gc)
				dispatch("EXCISE-SELECTION", gc)
				gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.SelectStart
				gc.ActiveBuffer.Formatter.SelectStart = -1
				gc.ActiveBuffer.Formatter.SelectEnd = -1
				return
			}
		case key.CodeV:
			if e.Modifiers > 0 {
				dispatch("EXCISE-SELECTION", gc)
				dispatch("PASTE-FROM-CLIPBOARD", gc)
			}
		default:
			if gc.ActiveBuffer.InputMode {
				switch e.Code {
				case key.CodeLeftShift:
				case key.CodeRightShift:
				case key.CodeReturnEnter:
					gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", gc.ActiveBuffer.Data.Text[:gc.ActiveBuffer.Formatter.Cursor], "\n", gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.Cursor:])
					gc.ActiveBuffer.Formatter.Cursor++
				default:
					switch e.Rune {
					case '`':
						gc.ActiveBuffer.InputMode = false
					default:
						if gc.ActiveBuffer.Formatter.SelectEnd > 0 {
							dispatch("EXCISE-SELECTION", gc)
						}
						if gc.ActiveBuffer.Formatter.Cursor < 0 {
							gc.ActiveBuffer.Formatter.Cursor = 0
						}
						fmt.Printf("Inserting at %v, length %v\n", gc.ActiveBuffer.Formatter.Cursor, len(gc.ActiveBuffer.Data.Text))
						gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", gc.ActiveBuffer.Data.Text[:gc.ActiveBuffer.Formatter.Cursor], string(e.Rune), gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.Cursor:])
						gc.ActiveBuffer.Formatter.Cursor++
					}

				}
			} else {
				switch e.Code {
				case key.CodeX:
					if e.Modifiers > 0 {
						dispatch("EXCISE-SELECTION", gc)
					}

				case key.CodeA:
					if e.Modifiers > 0 {
						gc.ActiveBuffer.Formatter.SelectStart = 0
						gc.ActiveBuffer.Formatter.SelectEnd = len(gc.ActiveBuffer.Data.Text)
					}
					gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor - 1
				case key.CodeD:
					gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor + 1
				case key.CodeQ:
					gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor + 1
				case key.CodeE:
					gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor - 1
				}
				switch e.Rune {
				case 'L':
					go startSshConn(1, "", "", "")
				case 'N':
					dispatch("NEXT-BUFFER", gc)
				case 'p':
					dispatch("PASTE-FROM-CLIPBOARD", gc)
				case 'y':
					dispatch("COPY-TO-CLIPBOARD", gc)
				case '~':
					dispatch("SAVE-FILE", gc)
				case 'i':
					dispatch("INPUT-MODE", gc)
				case '0':
					dispatch("START-OF-LINE", gc)
				case '^':
					dispatch("START-OF-TEXT-ON-LINE", gc)
				case '$':
					dispatch("END-OF-LINE", gc)
				case 'A':
					dispatch("END-OF-LINE", gc)
					dispatch("INPUT-MODE", gc)
				case 'a':
					gc.ActiveBuffer.Formatter.Cursor++
					dispatch("INPUT-MODE", gc)
				case 'k':
					gc.ActiveBuffer.Formatter.Cursor = scanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
				case 'j':
					gc.ActiveBuffer.Formatter.Cursor = scanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
				case 'l':
					dispatch("NEXT-CHARACTER", gc)
				case 'h':
					dispatch("PREVIOUS-CHARACTER", gc)
				case 'T':
					dispatch("TAIL", gc)
				case 'W':
					if gc.ActiveBuffer.Formatter.Outline {
						gc.ActiveBuffer.Formatter.Outline = false
					} else {
						gc.ActiveBuffer.Formatter.Outline = true
					}
				case 'S':
					dispatch("TOGGLE-VERTICAL-MODE", gc)
				case '+':
					dispatch("INCREASE-FONT", gc)
				case '-':
					dispatch("DECREASE-FONT", gc)
				case 'B':
					glim.ClearAllCaches()
					Log2Buff("Caches cleared")
					log.Println("Caches cleared")

				}
			}
		}

	}
	if gc.ActiveBuffer.Formatter.Cursor > gc.ActiveBuffer.Formatter.LastDrawnCharPos {
		log.Println("Advancing screen to cursor")
		//gc.ActiveBuffer.Formatter.FirstDrawnCharPos = scanToNextLine (gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.FirstDrawnCharPos)
		//gc.ActiveBuffer.Formatter.FirstDrawnCharPos = scanToPrevLine (gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	}

	if gc.ActiveBuffer.Formatter.Cursor < 0 {
		gc.ActiveBuffer.Formatter.Cursor = 0
	}
	if gc.ActiveBuffer.Formatter.Cursor < gc.ActiveBuffer.Formatter.FirstDrawnCharPos || gc.ActiveBuffer.Formatter.Cursor > gc.ActiveBuffer.Formatter.LastDrawnCharPos {
		scrollToCursor(gc.ActiveBuffer)
	}
}
*/
