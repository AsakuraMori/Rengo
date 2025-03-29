package engine

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"log"
	"sync"
)

type CharacterDisplay struct {
	characters           map[string]*Character
	current              *Character
	positionX, positionY float64
	mutex                sync.RWMutex
}

type Character struct {
	Name  string
	Image *ebiten.Image
}

func NewCharacterDisplay() *CharacterDisplay {
	return &CharacterDisplay{
		characters: make(map[string]*Character),
	}
}

func (cd *CharacterDisplay) LoadCharacter(characterName, imagePath string) error {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	img, _, err := ebitenutil.NewImageFromFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to load character image %s: %v", imagePath, err)
	}

	cd.characters[characterName] = &Character{
		Name:  characterName,
		Image: img,
	}
	return nil
}

func (cd *CharacterDisplay) SetCurrentCharacter(name string) {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	if char, ok := cd.characters[name]; ok {
		cd.current = char
	}
}

func (cd *CharacterDisplay) SetPosition(x, y float64) {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	cd.positionX = x
	cd.positionY = y
}

func (cd *CharacterDisplay) Update() {
	// Add any character animation or movement logic here
}

func (cd *CharacterDisplay) Draw(screen *ebiten.Image) {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	if cd.current != nil && cd.current.Image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(cd.positionX, cd.positionY)
		screen.DrawImage(cd.current.Image, op)
	}
}
func (cd *CharacterDisplay) SetCharacter(characterName string) {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	if char, exists := cd.characters[characterName]; exists {
		cd.current = char
	} else {
		log.Printf("Warning: Character %s not found", characterName)
	}
}

func (cd *CharacterDisplay) Clear() {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	cd.current = nil
}
func (cd *CharacterDisplay) Clone() *CharacterDisplay {
	newCD := &CharacterDisplay{
		characters: make(map[string]*Character),
		positionX:  cd.positionX,
		positionY:  cd.positionY,
	}
	for k, v := range cd.characters {
		newCD.characters[k] = v.Clone()
	}
	if cd.current != nil {
		newCD.current = cd.current.Clone()
	}
	return newCD
}
func (c *Character) Clone() *Character {
	return &Character{
		Name:  c.Name,
		Image: c.Image,
	}
}
func (cd *CharacterDisplay) IsReady() bool {
	// 根据你的需求实现这个方法
	// 例如，可以检查是否所有效果都已完成
	return true
}
