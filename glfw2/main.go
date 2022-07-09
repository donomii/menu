package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"bufio"
	"encoding/json"
	"runtime"
	"time"

	"github.com/donomii/glim"
	"github.com/donomii/goof"
	"github.com/donomii/menu"

	"io/ioutil"
	"log"

	"github.com/agnivade/levenshtein"
	"github.com/schollz/closestmatch"

	//".."
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() { runtime.LockOSThread() }

var ed *GlobalConfig
var confFile string
var pic []uint8

var pred []string
var predAction []string
var input, status string
var selected int
var update bool = true
var form *glim.FormatParams
var edWidth = 700
var edHeight = 500
var mode = "searching"
var window *glfw.Window

var wantWindow = true
var createWin = true
var preserveWindow = true

func Seq(min, max int) []int {
	size := max - min + 1
	if size < 1 {
		return []int{}
	}
	a := make([]int, size)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func UpdateBuffer(ed *GlobalConfig, input string) {
	ClearActiveBuffer(ed)
	if selected > len(pred)-1 {
		selected = 0
	}
	if mode == "searching" {
		ActiveBufferInsert(ed, "\n?> ")
		ActiveBufferInsert(ed, input)
		ActiveBufferInsert(ed, "\n\n")
		pred, predAction = menu.Predict([]byte(input))

		log.Printf("predictions %#v, %#v\n", pred, predAction)
		if len(pred) > 0 {
			pred = append(pred, "Menu Settings")
			predAction = append(predAction, "Menu Settings") //FIXME make this a file:// url
			for _, v := range Seq(selected, len(pred)-1) {
				if v == selected {
					ActiveBufferInsert(ed, "\n\n")
					ActiveBufferInsert(ed, "        "+pred[v]+"\n\n")
				} else {
					ActiveBufferInsert(ed, "        "+pred[v]+"\n")
				}
			}

			for _, v := range Seq(0, selected-1) {
				if v == selected {
					ActiveBufferInsert(ed, "\n")
					ActiveBufferInsert(ed, pred[v]+"\n\n")
				} else {
					ActiveBufferInsert(ed, "        "+pred[v]+"\n")
				}
			}
		}
	} else {
		ActiveBufferInsert(ed, "Loading\n\n")
		ActiveBufferInsert(ed, status)
	}
}

//{"name": "select", "sent": "complete : number {value}", "matches": {"value": "one"}, "conf": 0.6667726998777387, "input": "complete: number one"}

type intentInput struct {
	Name    string
	Sent    []string
	Matches map[string]string
	Conf    float64
	Input   string
}

func monitorSTDIN() {
	//Build a map translating  words into numbers
	var word2num map[string]int
	word2num = make(map[string]int)
	word2num["one"] = 1
	word2num["two"] = 2
	word2num["three"] = 3
	word2num["four"] = 4
	word2num["five"] = 5
	word2num["six"] = 6
	word2num["seven"] = 7
	word2num["eight"] = 8
	word2num["nine"] = 9
	word2num["ten"] = 10
	word2num["eleven"] = 11
	word2num["twelve"] = 12
	word2num["thirteen"] = 13
	word2num["fourteen"] = 14
	word2num["fifteen"] = 15
	word2num["sixteen"] = 16
	word2num["seventeen"] = 17
	word2num["eighteen"] = 18
	word2num["nineteen"] = 19
	word2num["twenty"] = 20
	word2num["thirty"] = 30
	scanner := bufio.NewScanner(os.Stdin)
	for {
		//Read a line from stdin

		scanner.Scan()
		// Holds the string that scanned
		input := scanner.Text()

		fmt.Printf("Input: %s\n", input)

		if len(input) > 0 {
			//Unmarshall json from input line
			var in intentInput
			err := json.Unmarshal([]byte(input), &in)
			if err != nil {
				fmt.Println("Error unmarshalling json:", err)
				continue
			}
			log.Printf("Decoded json: %#v\n", in)

			if !wantWindow {
				if in.Name == "showmenu" {
					popWindow()
				}
			} else {
				//Case statement
				switch in.Name {
				case "select":
					num := word2num[in.Matches["value"]]
					menu.Activate(Menu.SubNodes[num].Command)
				case "showmenu":
					popWindow()
				case "hidemenu":
					hideWindow()
				default:
					var menu_options_text []string
					//Loop over menu subnodes
					for _, v := range Menu.SubNodes {
						menu_options_text = append(menu_options_text, v.Name)
					}

					input_string := in.Input
					//Remove "partial: " prefix from input_string
					input_string = strings.Replace(input_string, "partial: ", "", 1)
					//Remove "complete: " prefix from input_string
					input_string = strings.Replace(input_string, "complete: ", "", 1)

					wordsToTest := menu_options_text

					log.Printf("Words to test: %#v\n", wordsToTest)

					// Choose a set of bag sizes, more is more accurate but slower
					bagSizes := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

					// Create a closestmatch object
					cm := closestmatch.New(wordsToTest, bagSizes)
					log.Printf("Finding matches for: %#v\n", input_string)
					predictions := cm.ClosestN(input_string, 5)

					log.Printf("Predicted: %#v\n", predictions)
					if len(predictions) > 0 {
						levenshtein_distance := levenshtein.ComputeDistance(predictions[0], input_string)
						log.Printf("Levenshtein distance: %#v\n", levenshtein_distance)

						target_menu_item := predictions[0]
						//Find the index of the target_menu_item in the menu_options_text
						target_menu_item_index := -1
						for i, v := range menu_options_text {
							if v == target_menu_item {
								target_menu_item_index = i
								break
							}
						}
						if target_menu_item_index != -1 {
							log.Printf("Activating menu item: %#v\n", target_menu_item)
							menu.Activate(Menu.SubNodes[target_menu_item_index].Command)
						}

					}
				}
			}
		}
	}
}

func handleKeys(window *glfw.Window) {
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

		log.Printf("Got key %c,%v,%v,%v", key, key, mods, action)

		/*if key == 301 {
			hideWindow()
			return
		}
		*/
		if action > 0 {

			//ESC
			if key == 256 {
				//os.Exit(0)
				log.Println("Escape pressed")
				toggleWindow()
				return
			}

			if key == 265 {
				selected -= 1
				if selected < 0 {
					selected = 0
				}
			}

			if key == 264 {
				selected += 1
				if selected > len(pred)-1 {
					selected = len(pred) - 1
				}
			}

			if key == 257 {
				if len(pred) > 0 {
					status = "Loading " + pred[selected] + predAction[selected]
					mode = "loading"
					update = true
					go func(thread_selected int) {
						if pred[thread_selected] == "Menu Settings" {
							recallFile := menu.RecallFilePath()

							log.Println("Opening for edit: ", recallFile)

							//goof.QC([]string{"open", recallFile})
							go goof.Command("c:\\Windows\\System32\\cmd.exe", []string{"/c", "start", recallFile})
							go goof.Command("/usr/bin/open", []string{recallFile})
						}
						value := predAction[thread_selected]
						if strings.HasPrefix(value, "internal://") {
							cmd := strings.TrimPrefix(value, "internal://")

							switch cmd {
							case "exit":
								os.Exit(0)
							case "reload":
								menu.RecallCache = nil
							default:
								log.Println("unsupported command when trying to run internal://")
							}
						}
						menu.Activate(predAction[thread_selected])

						toggleWindow()

						return
						//os.Exit(0)
					}(selected)
					//FIXME some kind of transition here?
					mode = "searching"
					input = ""
					status = ""
					update = true
					selected = 0
				} else {
					toggleWindow()
				}
			}

			if key == 259 {
				if len(input) > 0 {
					input = input[0 : len(input)-1]
				}
			}

			UpdateBuffer(ed, input)
			update = true
		}

	})

	window.SetCharModsCallback(func(w *glfw.Window, char rune, mods glfw.ModifierKey) {

		text := fmt.Sprintf("%c", char)
		val, _ := strconv.ParseInt(text, 10, strconv.IntSize)
		fmt.Println("Activating menu option", val)
		item := Menu.SubNodes[val]

		fmt.Println("Activating menu option", item.Name)
		SwitchToCwd()

		menu.Activate(item.Command)
		hideWindow()
	})
}

func popWindow() {
	log.Println("Popping window")
	Menu = LoadUserMenu()
	SwitchToCwd()
	update = true
	window.Restore()
	window.Show()
	wantWindow = true

}

func hideWindow() {
	log.Println("Hiding window")
	window.Iconify()
	window.Hide()
	wantWindow = false
	if !preserveWindow {
		log.Println("Exiting by request")
		os.Exit(0)
	}
}
func toggleWindow() {
	Menu = LoadUserMenu()
	log.Println("Toggling window")
	wantWindow = !wantWindow
	if wantWindow {

		if preserveWindow {
			popWindow()
		} else {
			update = true
			createWin = true
		}
	} else {
		if preserveWindow {
			hideWindow()
		} else {
			createWin = false
		}
	}
}

func SwitchToCwd() {
	cwdb, err := os.ReadFile(goof.HomePath(".umh/cwd"))
	cwds := strings.TrimSpace(string(cwdb))
	if err != nil {
		fmt.Println("Error changing dir:", err)
		panic(err)
	}
	fmt.Println("Switching to dir", string(cwds))
	os.Chdir(string(cwds))
	fmt.Println("Dir after switch:", goof.Cwd())
}

func main() {
	//if runtime.GOOS == "darwin" {
	//	preserveWindow = false
	//}
	Menu = LoadUserMenu()
	SwitchToCwd()

	var doLogs bool
	flag.BoolVar(&doLogs, "debug", false, "Display logging information")
	flag.Parse()

	go WatchKeys()

	if !doLogs {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	go monitorSTDIN()
	for {
		if createWin {
			createWindow()
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func createWindow() {

	log.Println("Init glfw")
	if err := glfw.Init(); err != nil {
		panic("failed to initialize glfw: " + err.Error())
	}
	defer glfw.Terminate()

	log.Println("Setup window")
	monitor := glfw.GetPrimaryMonitor()
	mode := monitor.GetVideoMode()
	edWidth = mode.Width - int(float64(mode.Width)*0.2)
	edHeight = int(float64(mode.Height) * 0.8)

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	//glfw.WindowHint(glfw.GLFW_TRANSPARENT_FRAMEBUFFER, GLFW_TRUE)
	log.Println("Make window", edWidth, "x", edHeight)
	var err error
	window, err = glfw.CreateWindow(edWidth, edHeight, "Menu", nil, nil)

	if err != nil {
		panic(err)
	}
	window.SetPos(mode.Width/10.0, mode.Height/10.0)
	popWindow()
	log.Println("Make glfw window context current")
	window.MakeContextCurrent()
	log.Println("Allocate memory")
	pic = make([]uint8, 3000*3000*4)
	ed = NewEditor()
	//Create a text formatter.  This controls the appearance of the text, e.g. colour, size, layout
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 16)
	log.Println("Set up key handlers")
	handleKeys(window)

	//This should be SetFramebufferSizeCallback, but that doesn't work, so...
	window.SetSizeCallback(func(w *glfw.Window, width int, height int) {

		edWidth = width
		edHeight = height
		//renderEd(edWidth, edHeight, ed.ActiveBuffer.Data.Text)
		renderEd(edWidth, edHeight, ed.ActiveBuffer.Data.Text)
		blit(pic, edWidth, edHeight)
		window.SwapBuffers()
		update = true
	})

	log.Println("Init gl")
	if err := gl.Init(); err != nil {
		panic(err)
	}
	/*
		go func() {
			lastTime := glfw.GetTime()

			for {
				nowTime := glfw.GetTime()
				if nowTime-lastTime < 10000.0 {

					update = true
					fmt.Println("Forece refresh")
				} else {
					return
				}
			}
		}()
	*/

	lastTime := glfw.GetTime()
	frames := 0

	UpdateBuffer(ed, input)
	log.Println("Start rendering")
	for !window.ShouldClose() && createWin {
		time.Sleep(35 * time.Millisecond)
		frames++
		nowTime := glfw.GetTime()
		if nowTime-lastTime >= 1.0 {
			//status = fmt.Sprintf("%.3f ms/f  %.0ffps\n", 1000.0/float32(frames), float32(frames))
			frames = 0
			lastTime += 1.0
		}

		if update {

			renderEd(edWidth, edHeight, renderMenuText(Menu))
			blit(pic, edWidth, edHeight)
			window.SwapBuffers()
			update = false
		}
		glfw.PollEvents()
	}
	log.Println("Normal glfw shutdown")
}

func blit(pix []uint8, w, h int) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Viewport(0, 0, int32(w)*screenScale(), int32(h)*screenScale())
	gl.Ortho(0, 1, 1, 0, 0, -1)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D, 0,
		gl.RGBA,
		int32(w), int32(h), 0,
		gl.RGBA,
		gl.UNSIGNED_BYTE, gl.Ptr(pix),
	)

	gl.Enable(gl.TEXTURE_2D)
	{
		gl.Begin(gl.QUADS)
		{
			gl.TexCoord2i(0, 0)
			gl.Vertex2i(0, 0)

			gl.TexCoord2i(1, 0)
			gl.Vertex2i(1, 0)

			gl.TexCoord2i(1, 1)
			gl.Vertex2i(1, 1)

			gl.TexCoord2i(0, 1)
			gl.Vertex2i(0, 1)
		}
		gl.End()
	}
	gl.Disable(gl.TEXTURE_2D)

	gl.Flush()

	gl.DeleteTextures(1, &texture)
}
