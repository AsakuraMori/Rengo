package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"image/color"
	"strings"
	"unicode/utf8"
	//"strings"
)

type TextDisplay struct {
	CurrentText     string
	IsReady         bool
	X, Y            float64
	Font            font.Face
	Color           color.Color
	MaxWidth        int
	CharIndex       int
	CharDelay       int
	FrameCount      int
	WaitingForInput bool
}

func NewTextDisplay(x, y float64) *TextDisplay {
	return &TextDisplay{
		Color:     color.White,
		MaxWidth:  600,
		X:         x,
		Y:         y,
		CharDelay: 2, // 每个字符之间的帧数延迟
	}
}

func (td *TextDisplay) CompleteText() {
	td.CharIndex = utf8.RuneCountInString(td.CurrentText)
	td.IsReady = true
	td.WaitingForInput = true
}

func (td *TextDisplay) SetText(s string) {
	td.CurrentText = s
	td.IsReady = false
	td.CharIndex = 0
	td.FrameCount = 0
	td.WaitingForInput = false
}

func (td *TextDisplay) Update() {
	if !td.IsReady && td.CharIndex < utf8.RuneCountInString(td.CurrentText) {
		td.FrameCount++
		if td.FrameCount >= td.CharDelay {
			td.CharIndex++
			td.FrameCount = 0
		}
	} else if td.CharIndex >= utf8.RuneCountInString(td.CurrentText) {
		td.IsReady = true
		td.WaitingForInput = true
	}
}

func (td *TextDisplay) Draw(screen *ebiten.Image) {
	if td.Font == nil {
		return
	}

	displayText := string([]rune(td.CurrentText)[:td.CharIndex])
	lines := td.wrapText(displayText)

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
	for i, line := range lines {
		text.Draw(
			screen,
			line,
			td.Font,
			int(td.X+shadowOffsetX),
			int(td.Y+shadowOffsetY)+i*td.Font.Metrics().Height.Round(),
			color.Black, // 阴影颜色
		)
	}

	// 绘制黑色边框
	for _, offset := range borderOffsets {
		for i, line := range lines {
			text.Draw(
				screen,
				line,
				td.Font,
				int(td.X+offset.x),
				int(td.Y+offset.y)+i*td.Font.Metrics().Height.Round(),
				color.Black, // 边框颜色
			)
		}
	}

	// 绘制原始文字
	for i, line := range lines {
		text.Draw(
			screen,
			line,
			td.Font,
			int(td.X),
			int(td.Y)+i*td.Font.Metrics().Height.Round(),
			td.Color, // 文字颜色
		)
	}
}

func (td *TextDisplay) wrapText(s string) []string {
	var lines []string
	words := strings.Split(s, " ")
	if len(words) == 0 {
		return lines
	}

	var line string
	for i, word := range words {
		if i == 0 && strings.HasSuffix(word, ":") {
			// 这是人名，直接添加到第一行
			lines = append(lines, word)
			line = ""
			continue
		}
		if text.BoundString(td.Font, line+" "+word).Dx() <= td.MaxWidth {
			if line == "" {
				line = word
			} else {
				line += " " + word
			}
		} else {
			if line != "" {
				lines = append(lines, line)
			}
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func (td *TextDisplay) SetFont(f font.Face) {
	td.Font = f
}

func (td *TextDisplay) SetColor(c color.Color) {
	td.Color = c
}

func (td *TextDisplay) SetPosition(x, y float64) {
	td.X = x
	td.Y = y
}

func (td *TextDisplay) IsTextComplete() bool {
	return td.IsReady && td.WaitingForInput
}

func (td *TextDisplay) ClearText() {
	td.CurrentText = ""
	td.IsReady = false
	td.CharIndex = 0
	td.FrameCount = 0
	td.WaitingForInput = false
}
