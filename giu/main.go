package main

import (
	"fmt"
	"os"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/donomii/menu"
)

func onQuit() {
	os.Exit(0)
}

var name string

var first = false

func loop() {
	pred := menu.Predict([]byte(name))
	//label := g.Label("Hello world from giu")

	x := float32(0.0)
	y := float32(0.0)
	width := float32(400.0)
	height := float32(200.0)

	flags := imgui.WindowFlagsNoTitleBar |
		imgui.WindowFlagsNoCollapse |
		imgui.WindowFlagsNoScrollbar |
		imgui.WindowFlagsNoMove |
		imgui.WindowFlagsMenuBar |
		imgui.WindowFlagsNoResize
	if flags&imgui.WindowFlagsNoMove != 0 && flags&imgui.WindowFlagsNoResize != 0 {
		imgui.SetNextWindowPos(imgui.Vec2{X: x, Y: y})
		imgui.SetNextWindowSize(imgui.Vec2{X: width, Y: height})
	} else {
		imgui.SetNextWindowPosV(imgui.Vec2{X: x, Y: y}, imgui.ConditionFirstUseEver, imgui.Vec2{X: 0, Y: 0})
		imgui.SetNextWindowSizeV(imgui.Vec2{X: width, Y: height}, imgui.ConditionFirstUseEver)
	}
	thing := true
	imgui.BeginV("Fuck you", &thing, int(flags))

	if !first {
		g.SetKeyboardFocusHereV(0)
		first = true
	}
	//g.InputText("##name", 0, &name).Build()
	g.InputTextV("##name", 0.0, &name, 0, func(x imgui.InputTextCallbackData) int32 { fmt.Printf("data: %+v\n", x); return 0 }, func() { fmt.Println("Changed!") }).Build()

	for _, v := range pred {
		vv := v
		g.Button(v, func() { fmt.Println("Activating", vv); menu.Activate(vv); os.Exit(0) }).Build()
	}

	imgui.End()

}

func main() {
	wnd := g.NewMasterWindow("Menu", 400, 200, 0, nil)
	wnd.Main(loop)
}
