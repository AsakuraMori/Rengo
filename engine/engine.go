package engine

import (
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"log"
	"os"
	"sort"
	"sync"
)

type Layer struct {
	ImageDisplay *ImageDisplay
	CharDisplay  *CharacterDisplay
	Visible      bool
	ZIndex       int
}

type Engine struct {
	Layers            []*Layer
	titleUI           *TitleUI
	CurrentImageLayer int
	CurrentCharLayer  int
	currentChoices    []Choice
	AffectionSystem   *AffectionSystem
	ChoiceSystem      *ChoiceManager
	EffectSystem      *EffectSystem
	TextDisplay       *TextDisplay
	Width, Height     int
	ScriptEngine      *ScriptEngine
	mutex             sync.RWMutex
	FontFace          *font.Face
	state             string
}

const (
	defaultFontSize = 24
	defaultFontDPI  = 72
)

var defaultFont font.Face

func loadDefaultFont() font.Face {
	if defaultFont != nil {
		return defaultFont
	}

	fontBytes, err := os.ReadFile("./resource/font/font.ttf")
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load font file: %v", err))
	}

	parsedFont, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse font: %v", err))
	}

	defaultFont = truetype.NewFace(parsedFont, &truetype.Options{
		Size:    defaultFontSize,
		DPI:     defaultFontDPI,
		Hinting: font.HintingFull,
	})

	return defaultFont
}

func NewEngine(width, height, layerCount int) *Engine {
	loadDefaultFont()
	e := &Engine{
		Layers:            make([]*Layer, layerCount),
		CurrentImageLayer: -1,
		CurrentCharLayer:  -1,
		TextDisplay:       NewTextDisplay(200, float64(height-100)),
		ChoiceSystem:      NewChoiceManager(defaultFont),
		AffectionSystem:   NewAffectionSystem(),
		Width:             width,
		Height:            height,
		state:             "title",
	}

	for i := range e.Layers {
		e.Layers[i] = &Layer{
			ImageDisplay: NewImageDisplay(),
			CharDisplay:  NewCharacterDisplay(),
			Visible:      true,
			ZIndex:       i,
		}
	}
	e.ScriptEngine = NewScriptEngine(e)
	e.EffectSystem = NewEffectSystem(e)
	e.TextDisplay.SetFont(defaultFont)

	e.titleUI = NewTitleUI(e, func() {
		e.state = "game"
		err := e.ScriptEngine.LoadScript("./resource/script/first.rgo")
		if err != nil {
			log.Fatal(err)
		}
		e.ScriptEngine.ExecuteStep()
	})
	return e
}

func (e *Engine) ShowChoices(choices []Choice) {
	e.currentChoices = choices
	// 在游戏界面上显示选项
	for i, choice := range choices {
		fmt.Printf("%d. %s\n", i+1, choice.Text)
	}
}

func (e *Engine) GetUserChoice() int {
	// 在这里实现获取用户输入的逻辑
	// 返回用户选择的索引（从0开始）
	// 如果用户还没有选择，返回-1
	var input int
	_, err := fmt.Scanf("%d", &input)
	if err != nil || input < 1 || input > len(e.currentChoices) {
		return -1
	}
	return input - 1
}

func (e *Engine) SetLayerImage(layerIndex int, imageName, imagePath string) error {
	if layerIndex >= 0 && layerIndex < len(e.Layers) {
		// 清除当前活动的图片图层
		if e.CurrentImageLayer >= 0 && e.CurrentImageLayer < len(e.Layers) {
			e.Layers[e.CurrentImageLayer].ImageDisplay.Clear()
		}

		layer := e.Layers[layerIndex]
		if err := layer.ImageDisplay.LoadImage(imageName, imagePath); err != nil {
			return err
		}
		layer.ImageDisplay.SetImage(imageName)
		layer.Visible = true
		e.CurrentImageLayer = layerIndex
		return nil
	}
	return fmt.Errorf("layer index out of range")
}

func (e *Engine) SetLayerCharacter(layerIndex int, characterName, characterPath string) error {
	if layerIndex >= 0 && layerIndex < len(e.Layers) {
		// 清除当前活动的立绘图层
		if e.CurrentCharLayer >= 0 && e.CurrentCharLayer < len(e.Layers) {
			e.Layers[e.CurrentCharLayer].CharDisplay.Clear()
		}

		layer := e.Layers[layerIndex]
		if err := layer.CharDisplay.LoadCharacter(characterName, characterPath); err != nil {
			return err
		}
		layer.CharDisplay.SetCharacter(characterName)
		layer.Visible = true
		e.CurrentCharLayer = layerIndex
		return nil
	}
	return fmt.Errorf("layer index out of range")
}
func (e *Engine) ClearLayer(layerIndex int) error {
	if layerIndex >= 0 && layerIndex < len(e.Layers) {
		layer := e.Layers[layerIndex]
		layer.ImageDisplay.Clear()
		layer.CharDisplay.Clear()
		return nil
	}
	return fmt.Errorf("layer index out of range")
}

var isMouseButtonPressed bool // 用于记录鼠标按钮状态

func (e *Engine) Update() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.state == "title" {
		return e.titleUI.Update()
	}
	// 更新文字显示进度
	e.TextDisplay.Update()

	// 检测鼠标左键点击
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !isMouseButtonPressed { // 只在按下时触发一次
			isMouseButtonPressed = true

			// 如果当前正在等待输入（显示文字）
			if e.ScriptEngine.waitingForInput {
				if !e.TextDisplay.IsReady {
					// 如果文字没有完全显示，则立即显示完整文字
					e.TextDisplay.CompleteText()
				} else {
					// 如果文字已经完全显示，则继续执行下一步
					e.ScriptEngine.waitingForInput = false
				}
			}

			// 如果当前正在等待选择（显示选项）
			if e.ScriptEngine.waitingForChoice {
				// 可以在这里添加处理选择的逻辑
			}
		}
	} else {
		isMouseButtonPressed = false // 重置状态
	}

	// 执行脚本步骤
	if e.ScriptEngine.waitingForChoice {
		selected, jumpTo := e.ChoiceSystem.HandleInput()
		if selected {
			e.ScriptEngine.jumpToLabel(jumpTo)
			e.ScriptEngine.waitingForChoice = false
			e.ChoiceSystem.IsActive = false
		}
	} else if !e.ScriptEngine.waitingForInput {
		e.ScriptEngine.ExecuteStep()
	}

	// 更新特效系统
	e.EffectSystem.Update()

	// 更新选择系统
	if e.ScriptEngine.waitingForChoice {
		e.ChoiceSystem.SetChoices(e.ScriptEngine.choicesToShow, e.Width)
	}

	return nil
}

func (e *Engine) Draw(screen *ebiten.Image) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	if e.state == "title" {
		e.titleUI.Draw(screen)
		return
	}
	sort.Slice(e.Layers, func(i, j int) bool {
		return e.Layers[i].ZIndex < e.Layers[j].ZIndex
	})

	for _, layer := range e.Layers {
		if layer.Visible {
			layer.ImageDisplay.Draw(screen)
			layer.CharDisplay.Draw(screen)
		}
	}

	e.TextDisplay.Draw(screen)
	e.ChoiceSystem.Draw(screen)

	e.EffectSystem.Draw(screen)

}

func (e *Engine) Run() error {
	ebiten.SetWindowSize(e.Width, e.Height)
	ebiten.SetWindowTitle("Visual Novel Engine")
	if err := ebiten.RunGame(e); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (e *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
	return e.Width, e.Height
}
