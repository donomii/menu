// +build !VR

package main

import (
	"io"
	"net"
	"os"
	"unicode"

	"github.com/donomii/glim"
	"golang.org/x/crypto/ssh/agent"

	//"github.com/mitchellh/go-homedir"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/atotto/clipboard"
	"golang.org/x/crypto/ssh"
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
	buf := &Buffer{}
	buf.Data = &BufferData{}
	buf.Formatter = glim.NewFormatter()
	buf.Data.Text = ""
	buf.Data.FileName = ""
	return buf
}

func NewEditor() *GlobalConfig {
	var gc GlobalConfig
	gc.ActiveBuffer = NewBuffer()
	gc.ActiveBuffer.Formatter = glim.NewFormatter()
	gc.ActiveBuffer.Data.Text = `
 Welcome to the shonky editor`
	return &gc

}

func Log2Buff(gc *GlobalConfig, s string) {
	gc.StatusBuffer.Data.Text = s
}
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

func LoadFile(gc *GlobalConfig, fileName string) {
	data, _ := ioutil.ReadFile(fileName)
	gc.ActiveBuffer.Data.Text = ""
	ActiveBufferAppend(gc, string(data))
	gc.ActiveBuffer.Formatter.Cursor = len(gc.ActiveBuffer.Data.Text)

}

func BuffAppend(gc *GlobalConfig, buffId int, txt string) {
	gc.BufferList[1].Data.Text = strings.Join([]string{gc.BufferList[1].Data.Text, txt}, "")
}

func ActiveBufferAppend(gc *GlobalConfig, txt string) {
	gc.ActiveBuffer.Data.Text = strings.Join([]string{gc.ActiveBuffer.Data.Text, txt}, "")
}

func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

/*
func StartSshConn(buffId int, user, password, serverAndPort string) {
	activeBufferAppend("Starting ssh connection\n")
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			SSHAgent(),
		},
	}

	// Dial your ssh server.
	activeBufferAppend(fmt.Sprintf("Connecting to ssh server as user %v: ", user))
	activeBufferAppend(serverAndPort)
	conn, err := ssh.Dial("tcp", serverAndPort, config)
	if err != nil {
		log.Fatal("unable to connect: ", err)
	}

	session, err := conn.NewSession()
	if err != nil {
		//return fmt.Errorf("Failed to create session: %s", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		//return fmt.Errorf("Unable to setup stdin for session: %v", err)
	}
	go io.Copy(stdin, os.Stdin)

	stdout, err := session.StdoutPipe()
	if err != nil {
		//return fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	//go io.Copy(os.Stdout, stdout)
	go processPort(stdout)
	dispatch("TAIL", gc)

	stderr, err := session.StderrPipe()
	if err != nil {
		//return fmt.Errorf("Unable to setup stderr for session: %v", err)
	}
	go io.Copy(os.Stderr, stderr)

	err = session.Run("dude tail-all-logs")
	defer conn.Close()
}
*/

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
			log.Println("Clipping from ", buf.Formatter.SelectStart, " to ", buf.Formatter.SelectEnd)
			buf.Data.Text = fmt.Sprintf("%s%s",
				buf.Data.Text[:buf.Formatter.SelectStart],
				buf.Data.Text[buf.Formatter.SelectEnd+1:])
			buf.Formatter.SelectStart = 0
			buf.Formatter.SelectEnd = 0
		}
	}
}

func ReduceFont(buf *Buffer) {
	buf.Formatter.FontSize -= 1
	glim.ClearAllCaches()

}

func IncreaseFont(buf *Buffer) {
	buf.Formatter.FontSize += 1
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
	log.Printf("Next buffer: %v", gc.ActiveBufferId)
}

func ToggleVerticalMode(gc *GlobalConfig) {
	if gc.ActiveBuffer.Formatter.Vertical {
		dispatch("HORIZONTAL-MODE", gc)
	} else {
		dispatch("VERTICAL-MODE", gc)
	}
}

func PasteFromClipBoard(gc *GlobalConfig, buf *Buffer) {
	text, _ := clipboard.ReadAll()
	dispatch("EXCISE-SELECTION", gc)

	if gc.ActiveBuffer.Formatter.Cursor < 0 {
		gc.ActiveBuffer.Formatter.Cursor = 0
	}

	gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", gc.ActiveBuffer.Data.Text[:gc.ActiveBuffer.Formatter.Cursor], text, gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.Cursor:])
}

//This function carries out commands.  It is the interface between your scripting, and the actual engine operation
func dispatch(command string, gc *GlobalConfig) {
	switch command {
	case "WHEEL-UP":
		gc.ActiveBuffer.Formatter.Cursor = ScanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "WHEEL-DOWN":
		gc.ActiveBuffer.Formatter.Cursor = ScanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
	case "EXCISE-SELECTION": //Cut
		ExciseSelection(gc.ActiveBuffer)
	case "REDUCE-FONT":
		ReduceFont(gc.ActiveBuffer)
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
		//	case "PASTE-FROM-CLIPBOARD":
		//		PasteFromClipBoard(gc.ActiveBuffer)
	case "COPY-TO-CLIPBOARD":
		clipboard.WriteAll(gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.SelectStart : gc.ActiveBuffer.Formatter.SelectEnd+1])
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
					dispatch("REDUCE-FONT", gc)
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
