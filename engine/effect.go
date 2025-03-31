package engine

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	_ "image/png"
	"os"
)

// Effect is an interface for all visual effects
type Effect interface {
	Update() bool
	Draw(screen *ebiten.Image)
}

// MaskEffect represents a mask effect
type MaskEffect struct {
	mask       *ebiten.Image // 遮罩图像
	source     *ebiten.Image // 源图像(渐变开始)
	target     *ebiten.Image // 目标图像(渐变结束)
	progress   float64       // 动画进度 (0.0 - 1.0)
	speed      float64       // 动画速度
	maskShader *ebiten.Shader
	layerIndex int // 应用到的图层索引
	engine     *Engine
}

// EffectSystem manages multiple effects
type EffectSystem struct {
	effects          []Effect
	transitionShader *ebiten.Shader
	engine           *Engine
}

// NewEffectSystem creates a new EffectSystem
func NewEffectSystem(engine *Engine) *EffectSystem {

	transitionShaderSrc, err := os.ReadFile("./resource/kage/mask_shader.kage")
	if err != nil {
		panic(err)
	}
	transitionShader, err := ebiten.NewShader(transitionShaderSrc)
	if err != nil {
		panic(err)
	}
	e := &EffectSystem{
		effects:          make([]Effect, 0),
		transitionShader: transitionShader,
		engine:           engine,
	}
	return e
}

// AddMaskEffect adds a mask effect
func (es *EffectSystem) AddTransitionEffect(layerIndex int, source, target *ebiten.Image, maskPath string, speed float64) error {
	maskFile, err := os.Open(maskPath)
	if err != nil {
		return fmt.Errorf("failed to open mask file: %v", err)
	}
	defer maskFile.Close()

	maskImg, _, err := image.Decode(maskFile)
	if err != nil {
		return fmt.Errorf("failed to decode mask image: %v", err)
	}
	mask := ebiten.NewImageFromImage(maskImg)

	effect := &MaskEffect{
		mask:       mask,
		source:     source,
		target:     target,
		progress:   0,
		speed:      speed,
		maskShader: es.transitionShader,
		layerIndex: layerIndex,
		engine:     es.engine,
	}
	es.effects = append(es.effects, effect)
	return nil
}

// Update updates all effects and removes finished ones
func (es *EffectSystem) Update() {
	for i := 0; i < len(es.effects); i++ {
		if es.effects[i].Update() {
			// Effect is finished, remove it
			es.engine.ScriptEngine.waitingForInput = true
			es.engine.TextDisplay.SetText(es.engine.ScriptEngine.currentText)
			es.effects = append(es.effects[:i], es.effects[i+1:]...)
			i--
		}
	}
}

// Draw draws all active effects
func (es *EffectSystem) Draw(screen *ebiten.Image) {
	for _, effect := range es.effects {
		effect.Draw(screen)
	}
}

// HasActiveEffects checks if there are any active effects
func (es *EffectSystem) HasActiveEffects() bool {
	return len(es.effects) > 0
}

// Update updates the mask effect progress
func (m *MaskEffect) Update() bool {
	m.progress += m.speed
	if m.progress >= 1 {
		m.progress = 1
		return true // 特效完成
	}
	return false // 特效未完成
}

// Draw implements the mask effect
func (m *MaskEffect) Draw(screen *ebiten.Image) {
	if m.source == nil || m.target == nil || m.mask == nil {
		return
	}

	// 获取图层
	layer := m.engine.Layers[m.layerIndex]

	// 创建新的图像作为绘制目标
	w, h := m.source.Size()
	if layer.ImageDisplay.current == nil || layer.ImageDisplay.current.Bounds().Dx() != w || layer.ImageDisplay.current.Bounds().Dy() != h {
		layer.ImageDisplay.current = ebiten.NewImage(w, h)
	} else {
		layer.ImageDisplay.current.Clear()
	}

	// 创建绘制选项
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = m.source
	op.Images[1] = m.target
	op.Images[2] = m.mask
	op.Uniforms = map[string]interface{}{
		"Progress": float32(m.progress),
	}

	// 绘制到图层图像
	layer.ImageDisplay.current.DrawRectShader(w, h, m.maskShader, op)

	// 将图层绘制到屏幕
	screen.DrawImage(layer.ImageDisplay.current, nil)
}
