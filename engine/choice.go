package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"image/color"
)

type Choice struct {
	Text   string
	JumpTo string
	Rect   Rect // 用于检测鼠标位置
}

type Rect struct {
	X, Y, Width, Height int
}

type ChoiceManager struct {
	Choices      []Choice
	HoveredIndex int
	Font         font.Face
	IsActive     bool
}

func NewChoiceManager(font font.Face) *ChoiceManager {
	return &ChoiceManager{
		Font:         font,
		IsActive:     false,
		HoveredIndex: -1,
	}
}

func (cm *ChoiceManager) SetChoices(choices []Choice, screenWidth int) {
	cm.Choices = choices
	cm.IsActive = true
	cm.updateChoiceRects(screenWidth)
}

func (cm *ChoiceManager) updateChoiceRects(screenWidth int) {
	y := 100 // 起始 Y 坐标
	for i := range cm.Choices {
		bounds := text.BoundString(cm.Font, cm.Choices[i].Text)
		textWidth := bounds.Dx()
		textHeight := bounds.Dy()

		// 计算居中的 X 坐标
		x := (screenWidth - textWidth) / 2

		cm.Choices[i].Rect = Rect{
			X:      x - 10,          // 左边距
			Y:      y,               // 当前 Y 坐标
			Width:  textWidth + 20,  // 文本宽度 + 边距
			Height: textHeight + 10, // 文本高度 + 边距
		}
		y += textHeight + 20 // 根据文本高度动态调整间距
	}
}

func (cm *ChoiceManager) HandleInput() (selected bool, jumpTo string) {
	if !cm.IsActive {
		return false, ""
	}

	// 处理鼠标输入
	x, y := ebiten.CursorPosition()
	cm.HoveredIndex = -1
	for i, choice := range cm.Choices {
		rect := choice.Rect
		if x >= rect.X && x <= rect.X+rect.Width &&
			y >= rect.Y && y <= rect.Y+rect.Height {
			cm.HoveredIndex = i
			break
		}
	}

	// 处理鼠标点击
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && cm.HoveredIndex != -1 {
		selected = true
		jumpTo = cm.Choices[cm.HoveredIndex].JumpTo
		cm.IsActive = false // 关闭选项显示
	}

	return
}

func (cm *ChoiceManager) Draw(screen *ebiten.Image) {
	if !cm.IsActive {
		return
	}

	for i, choice := range cm.Choices {
		displayText := choice.Text
		textColor := color.RGBA{255, 255, 255, 255} // 默认白色

		// 悬停效果
		if i == cm.HoveredIndex {
			displayText = "> " + displayText
			textColor = color.RGBA{0, 255, 0, 255} // 绿色
		}

		// 点击效果
		if i == cm.HoveredIndex && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			textColor = color.RGBA{255, 0, 0, 255} // 红色
		}

		// 文本位置
		textX := choice.Rect.X + 10
		textY := choice.Rect.Y + (choice.Rect.Height / 2) + 10 // 垂直居中

		// 阴影偏移量
		shadowOffsetX := 2.0
		shadowOffsetY := 2.0

		// 边框偏移量
		borderOffsets := []struct {
			x, y float64
		}{
			{-1, -1}, {-1, 0}, {-1, 1},
			{0, -1}, {0, 1},
			{1, -1}, {1, 0}, {1, 1},
		}

		// 绘制黑色阴影
		text.Draw(
			screen,
			displayText,
			cm.Font,
			int(float64(textX))+int(shadowOffsetX),
			int(float64(textY))+int(shadowOffsetY),
			color.Black, // 阴影颜色
		)

		// 绘制黑色边框
		for _, offset := range borderOffsets {
			text.Draw(
				screen,
				displayText,
				cm.Font,
				int(float64(textX))+int(offset.x),
				int(float64(textY))+int(offset.y),
				color.Black, // 边框颜色
			)
		}

		// 绘制原始文字
		text.Draw(
			screen,
			displayText,
			cm.Font,
			textX,
			textY,
			textColor, // 文字颜色
		)
	}
}

func (cm *ChoiceManager) IsWaitingForChoice() bool {
	return cm.IsActive
}
