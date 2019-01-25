// gui.go
package main

import (
	"strings"

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

	//"github.com/donomii/glim"
	"github.com/donomii/goof"
)

var mapTex *nktemplates.Texture

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

	if update > 0 {
		nk.NkMenubarBegin(ctx)

		/* menu #1 */
		nk.NkLayoutRowBegin(ctx, nk.Static, 25, 5)
		nk.NkLayoutRowPush(ctx, 45)
		if nk.NkMenuBeginLabel(ctx, "MENU", nk.TextLeft, nk.NkVec2(120, 200)) > 0 {
			//static size_t prog = 40;
			//static int slider = 10;
			check := int32(1)
			nk.NkLayoutRowDynamic(ctx, 25, 1)
			//if (nk.NkMenuItemLabel(ctx, "Hide", NK_TEXT_LEFT))
			//    show_menu = nk_false;
			//if (nk.NkMenuItemLabel(ctx, "About", NK_TEXT_LEFT))
			//    show_app_about = nk_true;
			//			nk.NkProgress(ctx, &prog, 100, nk.Modifiable)
			//			nk.NkSliderInt(ctx, 0, &slider, 16, 1)
			nk.NkCheckboxLabel(ctx, "check", &check)
			nk.NkMenuEnd(ctx)
		}

		nk.NkMenubarEnd(ctx)

		nk.NkLayoutRowDynamic(ctx, 20, 3)
		{
			nk.NkLabel(ctx, strings.Join(NodesToStringArray(currentThing), " "), nk.TextLeft)
			if 0 < nk.NkButtonLabel(ctx, "Undo") {
				if len(currentThing) > 1 {
					currentNode = currentThing[len(currentThing)-2]
					currentThing = currentThing[:len(currentThing)-1]
				}
			}
			if 0 < nk.NkButtonLabel(ctx, "Go Back") {
				if len(currentThing) > 1 {
					currentNode = currentThing[len(currentThing)-2]
				}
			}
		}
		QuickFileEditor(ctx)
		//ButtonBox(ctx)
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
				state.bgColor = nk.NkColorPicker(ctx, state.bgColor, nk.ColorFormatRGBA)
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
	nk.NkEnd(ctx)

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func ButtonBox(ctx *nk.Context) {
	nk.NkLayoutRowDynamic(ctx, 400, 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			for _, vv := range currentNode.SubNodes {
				//node := vv.SubNodes[i]
				name := vv.Name
				command := vv.Command
				v := vv
				if nk.NkButtonLabel(ctx, name) > 0 {
					fmt.Println("Data:", vv.Data)
					result = vv.Data
					if !strings.HasPrefix(command, "!") && !strings.HasPrefix(command, "&") {
						currentThing = append(currentThing, v)
						currentNode = v
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
							currentNode = execNode
						}

					}

					//list.Clear()
					//populateList(list)
					//app.Stop()
				}
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
				currentNode = makeStartNode()
				currentThing = []*Node{currentNode}
			}

			if 0 < nk.NkButtonLabel(ctx, "Go back") {
				if len(currentThing) > 1 {
					currentNode = currentThing[len(currentThing)-2]
					currentThing = currentThing[:len(currentThing)-1]
				}
			}
			if 0 < nk.NkButtonLabel(ctx, "Exit") {

				fmt.Println(strings.Join(NodesToStringArray(currentThing), " ") + "\n")
				app.Stop()
				os.Exit(0)
			}
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
					LoadFile(ed, vv)
					var err error
					EditBytes, err = ioutil.ReadFile(vv)
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
				newCursor, _, _ := glim.RenderPara(form, 0, 0, 0, 0, width, height, width, height, int(mouseX)-left, int(mouseY)-top, pic, ed.ActiveBuffer.Data.Text, true, true, true)
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
					nk.NkButtonImage(ctx, testim)
					//}
				} else {
					log.Println(err)
				}
			}
		}
		nk.NkGroupEnd(ctx)
	}

}
