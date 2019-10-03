// gui.go
package main

import (
	"bytes"
	"strings"
	"time"

	//	"time"

	"github.com/atotto/clipboard"
	"github.com/schollz/closestmatch"

	//"time"
	"runtime"
	//"unsafe"
	"io/ioutil"
	"strconv"

	"github.com/donomii/glim"
	"github.com/donomii/nuklear-templates"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"

	"github.com/mattn/go-shellwords"

	//"text/scanner"

	"fmt"

	"log"
	"os"

	//"os/exec"

	//"github.com/donomii/glim"
	"github.com/donomii/goof"
)

const ( // iota is reset to 0
	mode_search     = iota
	mode_editor     = iota
	mode_buttonmenu = iota
)

var mapTex *nktemplates.Texture
var lastEnterDown bool
var lastBackspaceDown bool

func defaultMenu(ctx *nk.Context) {
	col := nk.NewColor()
	col.SetRGBA(nk.Byte(255), nk.Byte(255), nk.Byte(255), nk.Byte(255))
	nk.SetBackgroundColor(ctx, *col)
	if 0 < nk.NkButtonLabel(ctx, "---------") {
	}

	if 0 < nk.NkButtonLabel(ctx, "Edit Config") {
		LoadFileIfNotLoaded(ed, confFile)
		getCurrentNode().Name = "File Manager"
	}

	if 0 < nk.NkButtonLabel(ctx, "Run command") {
		cmd := strings.Join(NodesToStringArray(currentThing[1:]), " ")
		result = goof.Command("cmd", []string{"/c", cmd})
		result = result + goof.Command("/bin/sh", []string{"-c", cmd})
	}

	if 0 < nk.NkButtonLabel(ctx, "Run command interactively") {
		goof.QCI(NodesToStringArray(currentThing[1:]))

	}
	if 0 < nk.NkButtonLabel(ctx, "Change directory") {
		path := strings.Join(NodesToStringArray(currentThing[1:]), "/")
		os.Chdir(path)
		updateCurrentNode(makeStartNode())
		currentThing = []*Node{getCurrentNode()}
	}

	if len(currentThing) > 1 {

		lastMenu := currentThing[len(currentThing)-2]

		if 0 < nk.NkButtonLabel(ctx, "Go back to "+lastMenu.Name) {
			if len(currentThing) > 1 {
				updateCurrentNode(currentThing[len(currentThing)-2])
				currentThing = currentThing[:len(currentThing)-1]
			}
		}
	}
	if 0 < nk.NkButtonLabel(ctx, "Exit") {

		fmt.Println(strings.Join(NodesToStringArray(currentThing), " ") + "\n")
		if ui {
			app.Stop()
		}
		os.Exit(0)
	}
}

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
		//static size_t prog = 40;
		//static int slider = 10;
		check := int32(1)
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
		//if (nk.NkMenuItemLabel(ctx, "About", NK_TEXT_LEFT))
		//    show_app_about = nk_true;
		//			nk.NkProgress(ctx, &prog, 100, nk.Modifiable)
		//			nk.NkSliderInt(ctx, 0, &slider, 16, 1)
		nk.NkCheckboxLabel(ctx, "check", &check)
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
			log.Println("Starting", cmd)
			switch runtime.GOOS {
			case "linux":
				log.Println("Starting ", cmd)
				result = goof.Command("/bin/sh", []string{"-c", cmd})
				result = result + goof.Command("cmd", []string{"/c", cmd})
				return true
			case "windows":
				go goof.Command("cmd", []string{"/c", cmd})
				time.Sleep(100000000 * time.Nanosecond) //FIXME use cmd.Exec or w/e to start program then exit
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
	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyEnter) > 0 {
		fmt.Printf("Enter: %+v\n", ctx.Input().GetKeyboard())
		if lastEnterDown == false {
			ed.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", ed.ActiveBuffer.Data.Text[:ed.ActiveBuffer.Formatter.Cursor], "\n", ed.ActiveBuffer.Data.Text[ed.ActiveBuffer.Formatter.Cursor:])
			ed.ActiveBuffer.Formatter.Cursor++

			//If menu launches too quickly, the user's finger will still be on Enter, but there will be nothing to select
			log.Println("Last elem selected:", len(comboCallback(userbytes, lastUserbytes)), ",", lastElemSelectedIndex)
			if len(comboCallback(userbytes, lastUserbytes)) > lastElemSelectedIndex {
				lastEnterDown = true
				if activate(-1, comboCallback(userbytes, lastUserbytes)[lastElemSelectedIndex]) {
					os.Remove(pidPath())
					os.Exit(0)
				}
			} else {
				lastEnterDown = false
			}
			lastEnterDown = true //FIXME  If I take this out it crashes

		}
	} else {
		lastEnterDown = false
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyBackspace) > 0 {
		//fmt.Printf("Back: %+v\n", ctx.Input().GetKeyboard())
		if lastBackspaceDown == false {
			dispatch("DELETE-LEFT", ed)
		}
		lastBackspaceDown = true
	} else {
		lastBackspaceDown = false
	}
}

func gfxMain(win *glfw.Window, ctx *nk.Context, state *State) {

	maxVertexBuffer := 512 * 1024
	maxElementBuffer := 128 * 1024

	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(50, 50, 230, 250)
	update := nk.NkBegin(ctx, "Menu", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)
	nk.NkWindowSetPosition(ctx, "Menu", nk.NkVec2(0, 0))
	nk.NkWindowSetSize(ctx, "Menu", nk.NkVec2(float32(winWidth), float32(winHeight)))
	handleKeys(ctx)

	if update > 0 {
		nk.NkStyleSetFont(ctx, fontSmall.Handle())
		if AppMode == mode_editor {
			drawmenu(ctx, state)
		}

		if AppMode == mode_buttonmenu {
			nk.NkLayoutRowDynamic(ctx, 20, 3)
			{
				nk.NkLabel(ctx, strings.Join(NodesToStringArray(currentThing), " > "), nk.TextLeft)
				if 0 < nk.NkButtonLabel(ctx, "Undo") {
					if len(currentThing) > 1 {
						updateCurrentNode(currentThing[len(currentThing)-2])
						currentThing = currentThing[:len(currentThing)-1]
					}
				}
				if 0 < nk.NkButtonLabel(ctx, "Go Back") {
					if len(currentThing) > 1 {
						updateCurrentNode(currentThing[len(currentThing)-2])
					}
				}
			}
			ButtonBox(ctx)

			nk.NkLayoutRowDynamic(ctx, 20, 3)
			{
				nk.NkLabel(ctx, strings.Join(NodesToStringArray(currentThing), " "), nk.TextLeft)
				if 0 < nk.NkButtonLabel(ctx, "Run") {
					cmd := strings.Join(NodesToStringArray(currentThing[1:]), " ")
					result = goof.Command("cmd", []string{"/c", cmd})
					result = result + goof.Command("/bin/sh", []string{"-c", cmd})
				}

				if 0 < nk.NkButtonLabel(ctx, "Run interactive") {
					goof.QCI(NodesToStringArray(currentThing[1:]))

				}
			}
			nk.NkLayoutRowDynamic(ctx, 25, 1)
			{

				//pf := nk.NewPluginFilterRef(unsafe.Pointer(&nk.NkFilterDefault))

				size := nk.NkVec2(nk.NkWidgetWidth(ctx), 400)
				if nk.NkComboBeginColor(ctx, state.bgColor, size) > 0 {
					nk.NkLayoutRowDynamic(ctx, 120, 1)
					//Linux doesn't like this?
					//state.bgColor = nk.NkColorPicker(ctx, state.bgColor, nk.ColorFormatRGBA)
					nk.NkLayoutRowDynamic(ctx, 25, 1)
					r, g, b, a := state.bgColor.RGBAi()
					r = nk.NkPropertyi(ctx, "#R:", 0, r, 255, 1, 1)
					g = nk.NkPropertyi(ctx, "#G:", 0, g, 255, 1, 1)
					b = nk.NkPropertyi(ctx, "#B:", 0, b, 255, 1, 1)
					a = nk.NkPropertyi(ctx, "#A:", 0, a, 255, 1, 1)
					state.bgColor.SetRGBAi(r, g, b, a)
					nk.NkComboEnd(ctx)
				}
			}
		}
		if AppMode == mode_editor {
			QuickFileEditor(ctx)

		}
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
		//fmt.Println("Width:", width)
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
	if len(optionsList) > 3 {
		optionsList = optionsList[:3]
	}

	//pf := nk.NewPluginFilterRef(unsafe.Pointer(&nk.NkFilterDefault))

	for i, v := range optionsList {
		if i == 0 {
			lastElemSelectedIndex = 0
			lastElemSelected = v
		}

		nk.NkLayoutRowDynamic(ctx, 50, 1)
		clicked := nk.NkButtonLabel(ctx, v)
		if clicked > 0 {
			if activate(-1, v) {
				os.Remove(pidPath())
				os.Exit(0)
			}
		}
	}
}

func ButtonBox(ctx *nk.Context) {
	nk.NkLayoutRowDynamic(ctx, 400, 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			for _, vv := range getCurrentNode().SubNodes {
				//node := vv.SubNodes[i]
				name := vv.Name
				command := vv.Command
				v := vv
				if nk.NkButtonLabel(ctx, name) > 0 {
					fmt.Println("Data:", vv.Data)
					result = vv.Data
					if !strings.HasPrefix(command, "!") && !strings.HasPrefix(command, "&") {

						currentThing = append(currentThing, v)
						updateCurrentNode(v)
					} else {

						log.Println("Running command", command)
						if strings.HasPrefix(command, "!!") {
							args, _ := shellwords.Parse(command[2:])
							fmt.Println("Running", args)
							goof.QCI(args)
						}
						if strings.HasPrefix(command, "!") {

							//It's a shell command

							cmd := command[1:]
							result = goof.Command("/bin/sh", []string{"-c", cmd})
							result = result + goof.Command("cmd", []string{"/c", cmd})
						}

						if strings.HasPrefix(name, "&") {

							//It's an internal command

							cmd := command[1:]
							if cmd == "lslR" {
								result = strings.Join(goof.LslR("."), "\n")
							}
							if cmd == "ls" {
								result = strings.Join(goof.Ls("."), "\n")
							}
						}

						if result != "" {
							log.Println("Ran command, got result", result)
							execNode := makeNodeShort("Exec", []*Node{})
							addTextNodesFromString(execNode, result)
							updateCurrentNode(execNode)
						}

					}

					//list.Clear()
					//populateList(list)
					//app.Stop()
				}
			}

			defaultMenu(ctx)
		}
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 10, 1)
		{
			//Control the display
			nk.NkLayoutRowDynamic(ctx, 20, 3)
			{

				if 0 < nk.NkButtonLabel(ctx, "None") {
					displaySplit = "None"
				}

				if 0 < nk.NkButtonLabel(ctx, "Spaces") {
					displaySplit = "Spaces"
				}

				if 0 < nk.NkButtonLabel(ctx, "Tabs") {
					displaySplit = "Tabs"
				}

				if displaySplit == "None" {
					nk.NkLayoutRowDynamic(ctx, 10, 1)
					{
						results := strings.Split(result, "\n")
						for _, v := range results {
							//nk.NkLabel(ctx, v, nk.WindowBorder)
							if nk.NkButtonLabel(ctx, v) > 0 {
								n := makeNodeShort(v, []*Node{})
								currentThing = append(currentThing, n)

							}
						}
					}
				}

				if displaySplit == "Spaces" {
					nk.NkLayoutRowDynamic(ctx, 10, 5)
					{
						results := strings.Split(result, "\n")
						for _, line := range results {
							bits := strings.Split(line, " ")
							for i := 0; i < 5; i++ {
								label := ""
								if i < len(bits) {
									label = bits[i]
								} else {
									label = ""
								}
								//nk.NkLabel(ctx, v, nk.WindowBorder)
								if nk.NkButtonLabel(ctx, label) > 0 {
									n := makeNodeShort(label, []*Node{})
									currentThing = append(currentThing, n)

								}
							}
						}
					}
				}

			}

		}
		nk.NkGroupEnd(ctx)
	}

}

func QuickFileEditor(ctx *nk.Context) {

	nk.NkLayoutRowDynamic(ctx, float32(winHeight), 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{

			files := goof.Ls(".")
			for _, vv := range files {
				if nk.NkButtonLabel(ctx, vv) > 0 {
					LoadFileIfNotLoaded(ed, vv)
					var err error
					EditBytes, err = ioutil.ReadFile(vv)
					//AddActiveBuffer(ed, string(EditBytes), vv)
					log.Println(err)
				}

			}
		}
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)

		//nk.NkLayoutRowStatic(ctx, 100, 100, 3)
		//nk.NkLayoutRowDynamic(ctx, float32(winHeight), 1)
		height := 1000
		butts := ctx.Input().Mouse().GetButtons()
		keys := ctx.Input().Keyboard()

		text := keys.GetText()
		var l *int32
		l = keys.GetTextLen()
		ll := *l
		if ll > 0 {
			fmt.Printf("input: %+v\n", ctx.Input())
			s := fmt.Sprintf("\"%vu%04x\"", `\`, int(text[0]))
			s2, _ := strconv.Unquote(s)
			/*log.Println(err)
			log.Printf("Text: %v, %v\n", s, s2)
			newBytes := append(EditBytes[:form.Cursor], []byte(s2)...)
			newBytes = append(newBytes, EditBytes[form.Cursor:]...)
			form.Cursor++
			*/
			if ed.ActiveBuffer.Formatter.Cursor < 0 {
				ed.ActiveBuffer.Formatter.Cursor = 0
			}

			fmt.Printf("Inserting at %v, length %v\n", ed.ActiveBuffer.Formatter.Cursor, len(ed.ActiveBuffer.Data.Text))
			ed.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", ed.ActiveBuffer.Data.Text[:ed.ActiveBuffer.Formatter.Cursor], fmt.Sprintf("%v", s2), ed.ActiveBuffer.Data.Text[ed.ActiveBuffer.Formatter.Cursor:])
			ed.ActiveBuffer.Formatter.Cursor++
		}
		mouseX, mouseY := int32(-1000), int32(-1000)

		for _, v := range butts {
			if *v.GetClicked() > 0 {
				//fmt.Printf("%+v\n", ctx.Input())
				mouseX, mouseY = ctx.Input().Mouse().Pos()

				log.Println("Click at ", mouseX, mouseY)
			}
		}
		nk.NkLayoutRowDynamic(ctx, 1000, 1)
		{
			//if EditBytes != nil {
			if ed != nil {
				nkwidth := nk.NkWidgetWidth(ctx)
				width := int(nkwidth)

				//fmt.Println("Width:", width)
				//var lenStr = int32(len(EditBytes))
				//nk.NkEditString(ctx, nk.EditMultiline|nk.EditAlwaysInsertMode, EditBytes, &lenStr, 512, nk.NkFilterAscii) FIXME
				//nk.NkLabelWrap(ctx, string(EditBytes))
				pic := make([]uint8, width*height*4)

				form.Colour = &glim.RGBA{255, 255, 255, 255}
				//form.Cursor = 20
				form.FontSize = 12
				bounds := nk.NkWidgetBounds(ctx)
				left := int(*bounds.GetX())
				top := int(*bounds.GetY())
				newCursor, _, _ := glim.RenderPara(ed.ActiveBuffer.Formatter, 0, 0, 0, 0, width, height, width, height, int(mouseX)-left, int(mouseY)-top, pic, ed.ActiveBuffer.Data.Text, true, true, true)
				for _, v := range butts {
					if *v.GetClicked() > 0 {
						form.Cursor = newCursor
						ed.ActiveBuffer.Formatter.Cursor = newCursor
					}
				}

				//pic, width, height := glim.GFormatToImage(im, nil, width, height)
				//gl.DeleteTextures(testim)
				//t, err := nktemplates.LoadImageFile(fmt.Sprintf("%v/progress%05v.png", output, fnum), width, height)
				//t := nktemplates.LoadImageData(globalPic, width, height)
				mapTex, _ = nktemplates.RawTexture(glim.Uint8ToBytes(pic), int32(width), int32(height), mapTex)
				var err error = nil
				if err == nil {
					testim := nk.NkImageId(int32(mapTex.Handle))
					//nk.NkLayoutRowStatic(ctx, 400, 400, 1)
					//{
					//log.Println("Drawing image")

					/*
						if nk.NkButtonImage(ctx, testim) > 0 {
							ed.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", ed.ActiveBuffer.Data.Text[:ed.ActiveBuffer.Formatter.Cursor], "\n", ed.ActiveBuffer.Data.Text[ed.ActiveBuffer.Formatter.Cursor:])
							ed.ActiveBuffer.Formatter.Cursor++
						}
					*/
					nk.NkImage(ctx, testim)
					//}
				} else {
					log.Println(err)
				}
			}
		}
		nk.NkGroupEnd(ctx)
	}

}
