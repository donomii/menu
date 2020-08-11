// gui.go
package main

import (
	"github.com/donomii/glim"
	"github.com/donomii/nuklear-templates"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"

	"fmt"

	"log"
)

var foreColour, backColour *glim.RGBA
var mapTex *nktemplates.Texture
var mapTex1 *nktemplates.Texture
var lastEnterDown bool
var lastBackspaceDown bool

func gfxMain(win *glfw.Window, ctx *nk.Context, state *State) {

	log.Println("Starting gfx")
	width, height := win.GetSize()
	log.Printf("glfw: window %vx%v", width, height)
	gl.Viewport(0, 0, int32(width-1), int32(height-1))
	appName := "Menu"

	maxVertexBuffer := 512 * 1024
	maxElementBuffer := 128 * 1024

	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(50, 50, 230, 250)
	nk.NkBegin(ctx, appName, bounds, nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)

	col := nk.NewColor()
	col.SetRGBA(nk.Byte(255), nk.Byte(255), nk.Byte(255), nk.Byte(255))
	wbd := ctx.Style().Window().GetFixedBackground().GetData()
	wbd[0] = 255
	wbd[1] = 255
	wbd[2] = 255
	wbd[3] = 255
	wbg := ctx.Style().GetButton().GetTextBackground()
	wbg.SetRGBAi(255, 255, 255, 255)

	nk.NkWindowSetPosition(ctx, appName, nk.NkVec2(0, 0))
	nk.NkWindowSetSize(ctx, appName, nk.NkVec2(float32(winWidth), float32(winHeight)))

	//if update || nkupdate > 0 {
	log.Println("Draw menu")
	//drawmenu(ctx, state)
	log.Println("Draw editor")
	QuickFileEditor(ctx)

	//}
	nk.NkEnd(ctx)
	log.Println("update complete")
	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	width, height = win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))

	//gl.Clear(gl.COLOR_BUFFER_BIT)
	//gl.ClearColor(0.0, 0.0, 0.0, 0.0) // Everything crashes if you move htis
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
	log.Println("Finished gfx")
}

func QuickFileEditor(ctx *nk.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in QuickFileEditor", r)
		}
	}()

	nk.NkLayoutRowDynamic(ctx, float32(0), 2)
	{

		//log.Println("Check mouse")
		/*
			butts := ctx.Input().Mouse().GetButtons()

			mouseX, mouseY := int32(-1000), int32(-1000)

			for _, v := range butts {
				if *v.GetClicked() > 0 {
					mouseX, mouseY = ctx.Input().Mouse().Pos()

					//log.Println("Click at ", mouseX, mouseY)
				}
			}
		*/
		bounds := nk.NkWidgetBounds(ctx)
		left := int(*bounds.GetX())
		top := int(*bounds.GetY())
		nuHeight := 400
		log.Println("Starting row draw")
		nk.NkLayoutRowDynamic(ctx, float32(0), 1)
		{

			if ed != nil {
				log.Println("Starting editor draw")
				width := int(nk.NkWidgetWidth(ctx))

				size := width * nuHeight * 4
				log.Println("Clearing", size, "bytes(", width, "x", nuHeight, ")")
				for i := 0; i < size; i = i + 1 {
					pic[i] = ((*backColour)[0])
				}

				form = ed.ActiveBuffer.Formatter
				form.Colour = foreColour
				form.Colour = backColour
				form.Outline = true

				mouseX := 10
				mouseY := 10
				displayText := ed.ActiveBuffer.Data.Text

				log.Println("Render paragraph", string(displayText))

				ed.ActiveBuffer.Formatter.FontSize = 16
				glim.RenderPara(ed.ActiveBuffer.Formatter,
					0, 0, 0, 0,
					width, nuHeight, width, nuHeight,
					int(mouseX)-left, int(mouseY)-top, pic, displayText,
					false, true, true)
				log.Println("Finished render paragraph")
				log.Println("Render image (", len(pic), " ", width, " ", nuHeight)
				doImage(ctx, pic, width, nuHeight)
			}
		}
	}

	log.Println("Finish terminal display")
}

func doImage(ctx *nk.Context, pic []uint8, width, nuHeight int) {
	log.Printf("Rendering image, %vx%v", width, nuHeight)
	nk.NkLayoutRowDynamic(ctx, float32(nuHeight), 1)
	{
		var err error = nil
		log.Println("calling RawTexture")

		mapTex1, err = nktemplates.RawTexture(glim.Uint8ToBytes(pic, picBytes), int32(width), int32(nuHeight), mapTex1)
		log.Println("rawtex comlplete, starting nkimageid")
		if err == nil {
			testim := nk.NkImageId(int32(mapTex1.Handle))
			log.Println("nk image")
			nk.NkImage(ctx, testim)
		} else {
			log.Println(err)
		}
	}
}
