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

// SetReady moves a lobby from waiting to ready, so the host can then open
// the buzzer for a round.
func (s *Service) SetReady(lobbyID string) (*Lobby, error) {
	l, err := s.manager.Get(lobbyID)

	if err != nil {
		return nil, err
	}

	if err := l.SetReady(); err != nil {
		return nil, err
	}

	return l, nil
}

// OpenBuzz opens the buzzer for a round, allowing players to buzz in.
func (s *Service) OpenBuzz(lobbyID string) (*Lobby, error) {
	l, err := s.manager.Get(lobbyID)

	if err != nil {
		return nil, err
	}

	if err := l.OpenBuzz(); err != nil {
		return nil, err
	}

	return l, nil
}

// Buzz registers a player's buzz for the current round.
func (s *Service) Buzz(lobbyID string, playerID string) (*Lobby, error) {
	l, err := s.manager.Get(lobbyID)

	if err != nil {
		return nil, err
	}

	if _, err := l.Buzz(playerID); err != nil {
		return nil, err
	}

	return l, nil
}
