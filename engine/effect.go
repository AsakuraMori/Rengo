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
	target     *ebiten.Image // 目标图像
	progress   float64       // 动画进度 (0.0 - 1.0)
	speed      float64       // 动画速度
	isReversed bool          // 是否反向
	maskShader *ebiten.Shader
}

// EffectSystem manages multiple effects
type EffectSystem struct {
	effects    []Effect
	maskShader *ebiten.Shader
	engine     *Engine
}

// NewEffectSystem creates a new EffectSystem
func NewEffectSystem(engine *Engine) *EffectSystem {

	maskShaderRead, err := os.ReadFile("./resource/kage/mask_shader.kage")
	if err != nil {
		panic(err)
	}
	maskShader, err := ebiten.NewShader(maskShaderRead)
	if err != nil {
		panic(err)
	}
	e := &EffectSystem{
		effects:    make([]Effect, 0),
		maskShader: maskShader,
		engine:     engine,
	}
	return e
}

// AddMaskEffect adds a mask effect
func (es *EffectSystem) AddMaskEffect(target *ebiten.Image, maskPath string, speed float64, isReversed bool) error {
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
		target:     target,
		progress:   0,
		speed:      speed,
		isReversed: isReversed,
		maskShader: es.maskShader,
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
	if m.target == nil || m.mask == nil {
		panic("target or mask is nil")
	}

	// 创建绘制选项
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = m.target // 目标图像（对应 imageSrc0）
	op.Images[1] = m.mask   // 遮罩图像（对应 imageSrc1）
	op.Uniforms = map[string]interface{}{
		"Progress": float32(m.progress), // 传递进度值
	}
	op.Blend = ebiten.BlendSourceOver
	// 获取目标图像的尺寸
	targetW, targetH := m.target.Size()

	// 使用着色器绘制
	screen.DrawRectShader(targetW, targetH, m.maskShader, op)
}
