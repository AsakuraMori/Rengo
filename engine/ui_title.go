// engine/ui/title.go

package engine

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yuin/gopher-lua"
	"log"
	"math"
	"time"
)

type TitleUI struct {
	engine      *Engine
	luaState    *lua.LState
	images      map[string]*ebiten.Image
	uiElements  []UIElement
	buttons     []Button
	startTime   time.Time
	onStartGame func()
	winX, winY  int
	//titleBarHeight int
	isHandCursor bool
	screen       *ebiten.Image // 用于绘制的屏幕
}

type UIElement struct {
	Image    *ebiten.Image
	X, Y     float64
	Scale    float64
	Alpha    float64
	Rotation float64
}

type Button struct {
	Name              string
	fName             string
	X, Y              float64
	Width             float64
	Height            float64
	IsHovered         bool
	Image             *ebiten.Image
	Alpha             float64
	CurrentHoverIndex int
}

func NewTitleUI(engine *Engine, onStartGame func()) *TitleUI {
	ui := &TitleUI{
		engine:      engine,
		luaState:    lua.NewState(),
		images:      make(map[string]*ebiten.Image),
		uiElements:  make([]UIElement, 0),
		buttons:     make([]Button, 0),
		startTime:   time.Now(),
		onStartGame: onStartGame,
		//titleBarHeight: 70,
		isHandCursor: false,
	}

	ui.loadImages()
	ui.setupLuaFunctions()
	ui.loadLuaScript()

	return ui
}

func (ui *TitleUI) loadImages() {
	imageNames := []string{
		"background", "logo",
		"start_game_normal", "start_game_hover",
		"load_game_normal", "load_game_hover",
		"config_normal", "config_hover",
		"exit_normal", "exit_hover",
	}
	for _, name := range imageNames {
		img, _, err := ebitenutil.NewImageFromFile(fmt.Sprintf("./resource/sys/title/%s.png", name))
		if err != nil {
			log.Printf("Failed to load image %s: %v", name, err)
			continue
		}
		ui.images[name] = img
		log.Printf("Loaded image: %s", name)
	}
}

func (ui *TitleUI) setupLuaFunctions() {
	ui.luaState.SetGlobal("loadImage", ui.luaState.NewFunction(ui.loadImage))
	ui.luaState.SetGlobal("drawImageEx", ui.luaState.NewFunction(ui.drawImageEx))
	ui.luaState.SetGlobal("getImageSize", ui.luaState.NewFunction(ui.getImageSize))
	ui.luaState.SetGlobal("drawImageCentered", ui.luaState.NewFunction(ui.drawImageCentered))
	ui.luaState.SetGlobal("drawImage", ui.luaState.NewFunction(ui.drawImage))
	ui.luaState.SetGlobal("drawImageFull", ui.luaState.NewFunction(ui.drawImageFull))
	ui.luaState.SetGlobal("addUIElement", ui.luaState.NewFunction(ui.addUIElement))
	ui.luaState.SetGlobal("setUIElement", ui.luaState.NewFunction(ui.setUIElement))
	ui.luaState.SetGlobal("getImageDimensions", ui.luaState.NewFunction(ui.getImageDimensions))
	ui.luaState.SetGlobal("registerButton", ui.luaState.NewFunction(ui.registerButton))
	ui.luaState.SetGlobal("setButtonImage", ui.luaState.NewFunction(ui.setButtonImage))
	ui.luaState.SetGlobal("setButtonAlpha", ui.luaState.NewFunction(ui.setButtonAlpha))
	ui.luaState.SetGlobal("onStartGame", ui.luaState.NewFunction(ui.luaOnStartGame))
}

func (ui *TitleUI) luaOnStartGame(L *lua.LState) int {
	// 调用 Go 代码中的 onStartGame 函数
	ui.onStartGame()
	return 0
}

func (ui *TitleUI) drawImageFull(L *lua.LState) int {
	imageName := L.ToString(1)
	screenWidth := float64(L.ToNumber(2))
	screenHeight := float64(L.ToNumber(3))
	alpha := float64(L.ToNumber(4))

	img, ok := ui.images[imageName]
	if !ok {
		log.Printf("Image not found: %s", imageName)
		return 0
	}

	imgWidth, imgHeight := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())

	op := &ebiten.DrawImageOptions{}

	// 计算缩放比例，确保图片填满屏幕
	scaleX := screenWidth / imgWidth
	scaleY := screenHeight / imgHeight
	scale := math.Max(scaleX, scaleY)

	op.GeoM.Scale(scale, scale)

	// 居中图片
	x := (screenWidth - imgWidth*scale) / 2
	y := (screenHeight - imgHeight*scale) / 2
	op.GeoM.Translate(x, y)

	// 设置透明度
	op.ColorM.Scale(1, 1, 1, alpha)

	ui.screen.DrawImage(img, op)
	return 0
}

func (ui *TitleUI) drawImageEx(L *lua.LState) int {
	imageName := L.ToString(1)
	x := float64(L.ToNumber(2))
	y := float64(L.ToNumber(3))
	scale := float64(L.ToNumber(4))
	alpha := float64(L.ToNumber(5))

	img, ok := ui.images[imageName]
	if !ok {
		log.Printf("Image not found: %s", imageName)
		return 0
	}

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(width)/2, -float64(height)/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorM.Scale(1, 1, 1, alpha)

	ui.screen.DrawImage(img, op)
	return 0
}

func (ui *TitleUI) drawImageCentered(L *lua.LState) int {
	imageName := L.ToString(1)
	centerX := float64(L.ToNumber(2))
	centerY := float64(L.ToNumber(3))

	img, ok := ui.images[imageName]
	if !ok {
		log.Printf("Image not found: %s", imageName)
		return 0
	}

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	x := centerX - float64(width)/2
	y := centerY - float64(height)/2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)

	ui.screen.DrawImage(img, op)

	// 添加调试信息
	//log.Printf("Drawing %s at (%.2f, %.2f) with size %dx%d", imageName, x, y, width, height)

	return 0
}

func (ui *TitleUI) getImageSize(L *lua.LState) int {
	imageName := L.ToString(1)

	img, ok := ui.images[imageName]
	if !ok {
		log.Printf("Image not found: %s", imageName)
		L.Push(lua.LNumber(0))
		L.Push(lua.LNumber(0))
		return 2
	}

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	L.Push(lua.LNumber(width))
	L.Push(lua.LNumber(height))
	return 2
}
func (ui *TitleUI) loadImage(L *lua.LState) int {
	imageName := L.ToString(1)
	img, _, err := ebitenutil.NewImageFromFile(fmt.Sprintf("./resource/sys/title/%s.png", imageName))
	if err != nil {
		log.Printf("Failed to load image %s: %v", imageName, err)
		return 0
	}
	ui.images[imageName] = img
	log.Printf("Loaded image: %s", imageName)
	return 0
}

func (ui *TitleUI) drawImage(L *lua.LState) int {
	imageName := L.ToString(1)
	x := float64(L.ToNumber(2))
	y := float64(L.ToNumber(3))
	width := float64(L.ToNumber(4))
	height := float64(L.ToNumber(5))

	img, ok := ui.images[imageName]
	if !ok {
		log.Printf("Image not found: %s", imageName)
		return 0
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(width/float64(img.Bounds().Dx()), height/float64(img.Bounds().Dy()))
	op.GeoM.Translate(x, y)

	ui.screen.DrawImage(img, op)
	return 0
}

func (ui *TitleUI) setButtonImage(L *lua.LState) int {
	buttonName := L.ToString(1)
	imageName := L.ToString(2)

	// 查找对应的按钮
	var targetButton *Button
	for i := range ui.buttons {
		if ui.buttons[i].Name == buttonName {
			targetButton = &ui.buttons[i]
			break
		}
	}

	if targetButton == nil {
		log.Printf("Button not found: %s", buttonName)
		return 0
	}

	// 查找对应的图片
	image, ok := ui.images[imageName]
	if !ok {
		log.Printf("Image not found: %s", imageName)
		return 0
	}

	// 设置按钮图片
	targetButton.Image = image
	log.Printf("Set image '%s' for button '%s'", imageName, buttonName)

	return 0
}

func (ui *TitleUI) setButtonAlpha(L *lua.LState) int {
	buttonName := L.ToString(1)
	alpha := float64(L.ToNumber(2))

	// 确保 alpha 值在 0 到 1 之间
	alpha = math.Max(0, math.Min(1, alpha))

	// 查找对应的按钮
	var targetButton *Button
	for i := range ui.buttons {
		if ui.buttons[i].Name == buttonName {
			targetButton = &ui.buttons[i]
			break
		}
	}

	if targetButton == nil {
		log.Printf("Button not found: %s", buttonName)
		return 0
	}

	// 设置按钮透明度
	targetButton.Alpha = alpha
	log.Printf("Set alpha %.2f for button '%s'", alpha, buttonName)

	return 0
}

func (ui *TitleUI) getImageDimensions(L *lua.LState) int {
	imageName := L.ToString(1)
	if img, ok := ui.images[imageName]; ok {
		width, height := img.Size()
		L.Push(lua.LNumber(width))
		L.Push(lua.LNumber(height))
		return 2
	}
	return 0
}

func (ui *TitleUI) loadLuaScript() {
	if err := ui.luaState.DoFile("./resource/sys/title.lua"); err != nil {
		log.Fatal(err)
	}
}

func (ui *TitleUI) addUIElement(L *lua.LState) int {
	imageName := L.ToString(1)
	x := float64(L.ToNumber(2))
	y := float64(L.ToNumber(3))
	scale := float64(L.ToNumber(4))
	alpha := float64(L.ToNumber(5))
	rotation := float64(L.ToNumber(6))

	ui.uiElements = append(ui.uiElements, UIElement{
		Image:    ui.images[imageName],
		X:        x,
		Y:        y,
		Scale:    scale,
		Alpha:    alpha,
		Rotation: rotation,
	})
	return 0
}

func (ui *TitleUI) setUIElement(L *lua.LState) int {
	index := L.ToInt(1) - 1 // Lua索引从1开始
	x := float64(L.ToNumber(2))
	y := float64(L.ToNumber(3))
	scale := float64(L.ToNumber(4))
	alpha := float64(L.ToNumber(5))
	rotation := float64(L.ToNumber(6))

	if index >= 0 && index < len(ui.uiElements) {
		ui.uiElements[index].X = x
		ui.uiElements[index].Y = y
		ui.uiElements[index].Scale = scale
		ui.uiElements[index].Alpha = alpha
		ui.uiElements[index].Rotation = rotation
	}
	return 0
}
func (ui *TitleUI) registerButton(L *lua.LState) int {
	name := L.ToString(1)
	x := float64(L.ToNumber(2))
	y := float64(L.ToNumber(3))
	width := float64(L.ToNumber(4))
	height := float64(L.ToNumber(5))

	normalImage := ui.images[name]
	if normalImage == nil {
		log.Printf("Normal image for button %s not found", name)
		return 0
	}

	ui.buttons = append(ui.buttons, Button{
		Name:              name,
		X:                 x,
		Y:                 y,
		Width:             width,
		Height:            height,
		IsHovered:         false,
		Image:             normalImage,
		CurrentHoverIndex: 0,
	})

	fmt.Printf("Registered button: %s at (%.0f, %.0f) with size %.0fx%.0f\n", name, x, y, width, height)

	return 0
}

// 在 engine/ui/title.go 文件中
func (ui *TitleUI) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()

	dt := 1.0 / 60.0 // 假设 60 FPS，可以根据实际情况调整

	// 调用 Lua 的 update 函数
	if err := ui.luaState.CallByParam(lua.P{
		Fn:      ui.luaState.GetGlobal("update"),
		NRet:    0,
		Protect: true,
	}, lua.LNumber(dt)); err != nil {
		return fmt.Errorf("error calling Lua update function: %v", err)
	}

	if err := ui.luaState.CallByParam(lua.P{
		Fn:      ui.luaState.GetGlobal("isGameStarted"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		return fmt.Errorf("error calling Lua isGameStarted function: %v", err)
	}
	gameStarted := ui.luaState.ToBool(-1)
	ui.luaState.Pop(1)

	if !gameStarted {
		// 调用 Lua 的 onMouseHover 函数
		if err := ui.luaState.CallByParam(lua.P{
			Fn:      ui.luaState.GetGlobal("onMouseHover"),
			NRet:    1,
			Protect: true,
		}, lua.LNumber(mouseX), lua.LNumber(mouseY)); err != nil {
			return fmt.Errorf("error calling Lua onMouseHover function: %v", err)
		}
		hoveredButton := ui.luaState.Get(-1).String()
		ui.luaState.Pop(1)

		// 添加调试信息
		//log.Printf("Mouse position: (%d, %d), Hovered button: %s", mouseX, mouseY, hoveredButton)
		shouldBeHandCursor := hoveredButton != "none"
		if shouldBeHandCursor != ui.isHandCursor {
			if shouldBeHandCursor {
				ebiten.SetCursorShape(ebiten.CursorShapePointer)
			} else {
				ebiten.SetCursorShape(ebiten.CursorShapeDefault)
			}
			ui.isHandCursor = shouldBeHandCursor
		}

		// 处理鼠标点击
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if err := ui.luaState.CallByParam(lua.P{
				Fn:      ui.luaState.GetGlobal("onMouseClick"),
				NRet:    1,
				Protect: true,
			}, lua.LNumber(mouseX), lua.LNumber(mouseY)); err != nil {
				return fmt.Errorf("error calling Lua onMouseClick function: %v", err)
			}
			clickedButton := ui.luaState.Get(-1).String()
			ui.luaState.Pop(1)

			//log.Printf("Clicked button: %s", clickedButton)

			// 处理按钮点击
			switch clickedButton {
			case "start_game":
				log.Println("Start game clicked")
				//ui.startGame()
				// 开始游戏逻辑
				ebiten.SetCursorShape(ebiten.CursorShapeDefault) // 重置鼠标光标形状
				ui.isHandCursor = false
			case "load_game":
				log.Println("Load game clicked")
				// 加载游戏逻辑
			case "config":
				log.Println("Config clicked")
				// 配置逻辑
			case "exit":
				log.Println("Exit clicked")
				// 退出逻辑
			}
		}
	}

	return nil
}

func (ui *TitleUI) Draw(screen *ebiten.Image) {
	ui.screen = screen // 保存屏幕引用

	// 调用 Lua 的 draw 函数
	if err := ui.luaState.CallByParam(lua.P{
		Fn:      ui.luaState.GetGlobal("draw"),
		NRet:    0,
		Protect: true,
	}); err != nil {
		log.Printf("Error calling Lua draw function: %v", err)
	}

}

func (ui *TitleUI) startGame() {
	if ui.onStartGame != nil {
		ui.onStartGame()
	}
}

func (ui *TitleUI) Close() {
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
}
