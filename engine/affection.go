package engine

import (
	"sync"
)

type AffectionSystem struct {
	affections      map[string]int
	maxAffection    int
	minAffection    int
	defaultValue    int
	mutex           sync.RWMutex
	changeCallbacks []func(string, int)
}

func NewAffectionSystem() *AffectionSystem {
	return &AffectionSystem{
		affections:   make(map[string]int),
		maxAffection: 100,
		minAffection: 0,
		defaultValue: 50,
	}
}

func (as *AffectionSystem) SetAffection(character string, value int) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	as.affections[character] = value
	as.notifyCallbacks(character, value)
}

func (as *AffectionSystem) GetAffection(character string) int {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	if value, ok := as.affections[character]; ok {
		return value
	}
	return as.defaultValue
}

func (as *AffectionSystem) ChangeAffection(character string, delta int) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	current := as.affections[character]
	new := current + delta
	if new > as.maxAffection {
		new = as.maxAffection
	} else if new < as.minAffection {
		new = as.minAffection
	}
	as.affections[character] = new
	as.notifyCallbacks(character, new)
}

func (as *AffectionSystem) AddChangeCallback(callback func(string, int)) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	as.changeCallbacks = append(as.changeCallbacks, callback)
}

func (as *AffectionSystem) notifyCallbacks(character string, value int) {
	for _, callback := range as.changeCallbacks {
		callback(character, value)
	}
}
