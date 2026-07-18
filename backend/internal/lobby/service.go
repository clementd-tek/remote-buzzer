package lobby

import (
	"github.com/google/uuid"
)

type Service struct {
	manager *Manager
}

func NewService(manager *Manager) *Service {
	return &Service{
		manager: manager,
	}
}

type CreateLobbyRequest struct {
	Name   string
	HostID string
	Public bool
}

func (s *Service) Create(req CreateLobbyRequest) *Lobby {
	l := New(
		uuid.New().String(),
		req.Name,
		req.HostID,
		req.Public,
	)

	s.manager.Add(l)

	return l
}

func (s *Service) Get(id string) (*Lobby, error) {
	return s.manager.Get(id)
}

func (s *Service) List() []*Lobby {
	return s.manager.List()
}

func (s *Service) Join(lobbyID string, player *Player) error {
	l, err := s.manager.Get(lobbyID)

	if err != nil {
		return err
	}

	return l.AddPlayer(player)
}
