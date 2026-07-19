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

func (s *Service) Create(req CreateLobbyRequest) (*Lobby, error) {
	if err := ValidateName(req.Name); err != nil {
		return nil, err
	}

	if err := ValidateID(req.HostID); err != nil {
		return nil, err
	}

	l := New(
		uuid.New().String(),
		req.Name,
		req.HostID,
		req.Public,
	)

	s.manager.Add(l)

	return l, nil
}

func (s *Service) Get(id string) (*Lobby, error) {
	return s.manager.Get(id)
}

// List returns every known lobby. Prefer ListPublic for anything shown on
// a public homepage.
func (s *Service) List() []*Lobby {
	return s.manager.List()
}

// ListPublic returns only lobbies created with Public: true, which is
// what an unauthenticated homepage should ever display.
func (s *Service) ListPublic() []*Lobby {
	all := s.manager.List()
	public := make([]*Lobby, 0, len(all))

	for _, l := range all {
		if l.Snapshot().Public {
			public = append(public, l)
		}
	}

	return public
}

func (s *Service) Join(lobbyID string, player *Player) error {
	if err := ValidateID(player.ID); err != nil {
		return err
	}

	if err := ValidateName(player.Name); err != nil {
		return err
	}

	l, err := s.manager.Get(lobbyID)

	if err != nil {
		return err
	}

	if err := l.AddPlayer(player); err != nil {
		return err
	}

	s.manager.Touch(l)

	return nil
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

	s.manager.Touch(l)

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

	s.manager.Touch(l)

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

	s.manager.Touch(l)

	return l, nil
}
