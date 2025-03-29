package engine

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"log"
	"sync"
)

type ImageDisplay struct {
	images  map[string]*ebiten.Image
	current *ebiten.Image
	mask    *ebiten.Image
	effects map[string]*ImageEffect
	mutex   sync.RWMutex
}

type ImageEffect struct {
	Duration int
	Current  int
	Type     string
}

func NewImageDisplay() *ImageDisplay {
	return &ImageDisplay{
		images:  make(map[string]*ebiten.Image),
		effects: make(map[string]*ImageEffect),
	}
}

func (id *ImageDisplay) LoadImage(imageName, imagePath string) error {
	id.mutex.Lock()
	defer id.mutex.Unlock()

	img, _, err := ebitenutil.NewImageFromFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to load image %s: %v", imagePath, err)
	}

	id.images[imageName] = img
	return nil
}

func (id *ImageDisplay) SetImage(imageName string) {
	id.mutex.Lock()
	defer id.mutex.Unlock()

	if img, exists := id.images[imageName]; exists {
		id.current = img
	} else {
		log.Printf("Warning: Image %s not found", imageName)
	}
}
func (id *ImageDisplay) AddEffect(name string, effect *ImageEffect) {
	id.mutex.Lock()
	defer id.mutex.Unlock()

	id.effects[name] = effect
}

func (id *ImageDisplay) Update() {
	id.mutex.Lock()
	defer id.mutex.Unlock()

	for name, effect := range id.effects {
		effect.Current++
		if effect.Current >= effect.Duration {
			delete(id.effects, name)
		}
	}
}

func (id *ImageDisplay) Draw(screen *ebiten.Image) {
	id.mutex.RLock()
	defer id.mutex.RUnlock()

	if id.current != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(id.current, op)
	}

	// 绘制效果（如果有的话）
	/*	for _, effect := range id.effects {
		effect.Draw(screen)
	}*/
}

func (id *ImageDisplay) Clear() {
	id.mutex.Lock()
	defer id.mutex.Unlock()

	id.current = nil
	for k := range id.effects {
		delete(id.effects, k)
	}
}

func (id *ImageDisplay) IsReady() bool {
	// 根据你的需求实现这个方法
	// 例如，可以检查是否所有效果都已完成
	return len(id.effects) == 0
}
