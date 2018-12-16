// gui.go
package main

import (
	"strings"

	//"unsafe"

	"io/ioutil"

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
		//QuickFileEditor(ctx)
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
					var err error
					EditBytes, err = ioutil.ReadFile(vv)
					log.Println(err)
				}

			}
		}
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)

		//nk.NkLayoutRowStatic(ctx, 100, 100, 3)
		nk.NkLayoutRowDynamic(ctx, float32(winHeight), 1)
		{
			if EditBytes != nil {
				//var lenStr = int32(len(EditBytes))
				//nk.NkEditString(ctx, nk.EditMultiline|nk.EditAlwaysInsertMode, EditBytes, &lenStr, 512, nk.NkFilterAscii) FIXME
				nk.NkLabelWrap(ctx, string(EditBytes))
			}
		}
		nk.NkGroupEnd(ctx)
	}

}
