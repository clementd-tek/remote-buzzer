package lobby

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("lobby not found")

type Manager struct {
	lobbies sync.Map
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Add(lobby *Lobby) {
	m.lobbies.Store(
		lobby.ID,
		lobby,
	)
}

func (m *Manager) Get(id string) (*Lobby, error) {
	value, ok := m.lobbies.Load(id)

	if !ok {
		return nil, ErrNotFound
	}

	return value.(*Lobby), nil
}

func (m *Manager) Delete(id string) {
	m.lobbies.Delete(id)
}

func (m *Manager) List() []*Lobby {
	result := make([]*Lobby, 0)

	m.lobbies.Range(
		func(key, value any) bool {
			result = append(
				result,
				value.(*Lobby),
			)

			return true
		},
	)

	return result
}
