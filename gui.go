// gui.go
package main

import (
	"bytes"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/schollz/closestmatch"

	"runtime"

	"github.com/donomii/nuklear-templates"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"

	"fmt"

	"log"
	"os"

	"github.com/donomii/goof"
)

var mapTex *nktemplates.Texture
var lastEnterDown bool
var lastBackspaceDown bool

func drawmenu(ctx *nk.Context, state *State) {
	nk.NkMenubarBegin(ctx)

	/* menu #1 */
	nk.NkLayoutRowBegin(ctx, nk.Static, 25, 5)
	nk.NkLayoutRowPush(ctx, 45)

	if nk.NkMenuBeginLabel(ctx, "File", nk.TextLeft, nk.NkVec2(120, 200)) > 0 {
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "Save", nk.TextLeft) > 0 {
			dispatch("SAVE-FILE", ed)
		}
		if nk.NkMenuItemLabel(ctx, "Exit", nk.TextLeft) > 0 {
			os.Exit(0)
		}
		nk.NkMenuEnd(ctx)
	}

	if nk.NkMenuBeginLabel(ctx, "Fonts", nk.TextLeft, nk.NkVec2(120, 200)) > 0 {
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "Text direction", nk.TextLeft) > 0 {
			dispatch("TOGGLE-VERTICAL-MODE", ed)
		}
		if nk.NkMenuItemLabel(ctx, "Increase font", nk.TextLeft) > 0 {
			dispatch("INCREASE-FONT", ed)
		}
		if nk.NkMenuItemLabel(ctx, "Decrease font", nk.TextLeft) > 0 {
			dispatch("DECREASE-FONT", ed)
		}
		if nk.NkMenuItemLabel(ctx, "8 point", nk.TextLeft) > 0 {
			SetFont(ed.ActiveBuffer, 8)
		}
		if nk.NkMenuItemLabel(ctx, "12 point", nk.TextLeft) > 0 {
			SetFont(ed.ActiveBuffer, 12)
		}
		if nk.NkMenuItemLabel(ctx, "20 point", nk.TextLeft) > 0 {
			SetFont(ed.ActiveBuffer, 20)
		}
		nk.NkMenuEnd(ctx)
	}

	if nk.NkMenuBeginLabel(ctx, "Buffers", nk.TextLeft, nk.NkVec2(120, 200)) > 0 {
		//static size_t prog = 40;
		//static int slider = 10;
		check := int32(1)
		nk.NkLayoutRowDynamic(ctx, 25, 1)

		if nk.NkMenuItemLabel(ctx, "Next Buffer", nk.TextLeft) > 0 {
			dispatch("NEXT-BUFFER", ed)
			fmt.Println("NExt buffer")
		}
		if nk.NkMenuItemLabel(ctx, "Previous Buffer", nk.TextLeft) > 0 {
			dispatch("PREVIOUS-BUFFER", ed)
		}

		if nk.NkMenuItemLabel(ctx, "---------------", nk.TextLeft) > 0 {
		}

		for i, v := range ed.BufferList {
			if nk.NkMenuItemLabel(ctx, fmt.Sprintf("%v) %v", i, v.Data.FileName), nk.TextLeft) > 0 {
				ed.ActiveBuffer = ed.BufferList[i]
			}
		}

		//if (nk.NkMenuItemLabel(ctx, "About", NK_TEXT_LEFT))
		//    show_app_about = nk_true;
		//			nk.NkProgress(ctx, &prog, 100, nk.Modifiable)
		//			nk.NkSliderInt(ctx, 0, &slider, 16, 1)
		nk.NkCheckboxLabel(ctx, "check", &check)
		nk.NkMenuEnd(ctx)
	}

	nk.NkMenubarEnd(ctx)
}

var recallCache [][]string

func comboCallback(newString, oldString []byte) []string {
	news := string(newString)
	//log.Println("Processing ", news)
	if appCache == nil {
		appCache = Apps()
	}
	if recallCache == nil {
		recallCache = Recall()
	}
	var names []string
	for _, details := range appCache {
		names = append(names, details[0])
	}
	for _, details := range recallCache {
		names = append(names, details[0])
	}
	wordsToTest := names

	//log.Println("Words to test: ", wordsToTest)

	// Choose a set of bag sizes, more is more accurate but slower
	bagSizes := []int{2}

	// Create a closestmatch object
	cm := closestmatch.New(wordsToTest, bagSizes)
	return cm.ClosestN(news, 5)
}

func activate(index int, value string) bool {

	log.Println("selected ", index, value)
	appCache := Apps()

	for i, v := range appCache {
		cmp := strings.Compare(value, v[0])

		if cmp == 0 {

			cmd := appCache[i][1][1:]

			switch runtime.GOOS {
			case "linux":
				log.Println("Starting ", cmd)
				result = goof.Command("/bin/sh", []string{"-c", cmd})
				result = result + goof.Command("cmd", []string{"/c", cmd})
				return true
			case "windows":
				cmdArray := []string{"/c", cmd}
				log.Println("Starting cmd", cmdArray)
				go goof.Command("c:\\Windows\\System32\\cmd.exe", cmdArray)
				time.Sleep(5 * time.Second) //FIXME use cmd.Exec or w/e to start program then exit
				return true
			case "darwin":
				result = result + goof.Command("/bin/sh", []string{"-c", cmd})
				return true
			default:
				log.Println("unsupported platform when trying to run application")
			}

		}
		for _, v := range Recall() {
			name := v[0]
			//log.Println("Searching for", value, name)
			cmp := strings.Compare(value, name)
			if cmp == 0 {
				//log.Println("Found", value, v[1])
				if v[1] == "recall" {

					//Copy to clipboard
					bits := strings.SplitN(name, " | ", 2)
					data := bits[1]
					if strings.HasPrefix(data, "http") {
						url := data
						log.Println("Opening ", data, "in browser")
						var err error
						switch runtime.GOOS {
						case "linux":
							goof.QC([]string{"xdg-open", url})
						case "windows":
							goof.QC([]string{"rundll32", "url.dll,FileProtocolHandler"})
						case "darwin":

							goof.QC([]string{"open", url})
						default:
							err = fmt.Errorf("unsupported platform")
						}
						if err != nil {
							log.Println(err)
						}
						return true
					}
					log.Println("Copying ", data, "to clipboard")
					if err := clipboard.WriteAll(data); err != nil {
						panic(err)
					}

					return true
				}

			}
		}
	}
	return false

}

func handleKeys(ctx *nk.Context) {
	if time.Now().Sub(lastKey).Seconds() < 0.1 {
		return
	}
	lastKey = time.Now()

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyBackspace) > 0 {
		if lastBackspaceDown == false {
			dispatch("DELETE-LEFT", ed)
		}
		lastBackspaceDown = true
	} else {
		lastBackspaceDown = false
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyUp) > 0 {

		activeSelection = activeSelection - 1
		if activeSelection < 0 {
			activeSelection = len(optionsList) - 1
		}
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyDown) > 0 {
		activeSelection = activeSelection + 1
		if activeSelection > len(optionsList) {
			activeSelection = 0
		}
	}
}

func gfxMain(win *glfw.Window, ctx *nk.Context, state *State) {

	maxVertexBuffer := 512 * 1024
	maxElementBuffer := 128 * 1024

	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(50, 50, 230, 250)
	update := nk.NkBegin(ctx, "Menu", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable)
	nk.NkWindowSetPosition(ctx, "Menu", nk.NkVec2(0, 0))
	nk.NkWindowSetSize(ctx, "Menu", nk.NkVec2(float32(winWidth), float32(winHeight)))
	handleKeys(ctx)

	if update > 0 {
		nk.NkStyleSetFont(ctx, fontSmall.Handle())
		SpeedSearch(ctx)

	}

	nk.NkEnd(ctx)

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOff, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func SpeedSearch(ctx *nk.Context) {

	nk.NkStyleSetFont(ctx, fontLarge.Handle())

	nk.NkLayoutRowDynamic(ctx, 50, 1)
	{

		var text string

		text = "Search"

		nk.NkButtonLabel(ctx, text)

		lastUserbytes = bytes.Map(func(r rune) rune { return r }, userbytes)
		var lenStr = int32(len(userbytes))
		nk.NkEditFocus(ctx, nk.EditAlwaysInsertMode)
		nk.NkEditString(ctx, nk.EditAlwaysInsertMode, userbytes, &lenStr, 512, nk.NkFilterAscii) //FIXME

		nk.NkLabelWrap(ctx, string(userbytes))

		ret := bytes.Compare(lastUserbytes, userbytes)

		if ret != 0 {
			//log.Println("Text string changed!", lastUserbytes, "|", userbytes)
			//Handle the change
			optionsList = comboCallback(userbytes, lastUserbytes)
		}
	}
	if len(optionsList) > 10 {
		optionsList = optionsList[:10]
	}

	for i, v := range optionsList {
		if i == 0 {
			lastElemSelectedIndex = 0
			lastElemSelected = v
		}

		nk.NkLayoutRowDynamic(ctx, 50, 1)
		var clicked int32
		if i == activeSelection {
			clicked = nk.NkOptionLabel(ctx, ""+v+"", 0)
		} else {
			clicked = nk.NkOptionLabel(ctx, v, 1)
		}
		if clicked > 0 {
			log.Printf("buttons: %+v", ctx.Input().GetMouse().GetButtons())
			butts := ctx.Input().GetMouse().GetButtons()

			if *butts[0].GetDown() > 0 {
				log.Println("clicked", clicked)
				if activate(-1, v) {
					os.Remove(pidPath())
					os.Exit(0)
				}
			}
		}

	}
	var text string
	text = "Add personal search results"

	clicked := nk.NkButtonLabel(ctx, text)

	if clicked > 0 {
		log.Printf("Opening config here")
		recallFile := goof.ConfigFilePath(".menu.recall.txt")
		loadEnsureRecallFile(recallFile)
		//goof.QC([]string{"open", recallFile})
		go goof.Command("c:\\Windows\\System32\\cmd.exe", []string{"/c", "start", recallFile})
		go goof.Command("/usr/bin/open", []string{recallFile})
		//butts := ctx.Input().GetMouse().GetButtons()
	}

}
