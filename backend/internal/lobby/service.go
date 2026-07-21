package lobby

import (
	"time"

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
	Name     string
	HostID   string
	Public   bool
	Settings *LobbySettings // nil → use DefaultSettings()
}

func (s *Service) Create(req CreateLobbyRequest) (*Lobby, error) {
	if err := ValidateName(req.Name); err != nil {
		return nil, err
	}

	if err := ValidateID(req.HostID); err != nil {
		return nil, err
	}

	settings := DefaultSettings()

	if req.Settings != nil {
		if err := ValidateSettings(*req.Settings); err != nil {
			return nil, err
		}

		settings = *req.Settings
	}

	l := New(
		uuid.New().String(),
		req.Name,
		req.HostID,
		req.Public,
		settings,
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

// ListPublic returns only lobbies created with Public: true.
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

// SetReady moves a lobby from waiting to ready.
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

// UpdateSettings applies a partial settings patch to the lobby. Only
// allowed while nothing is in flight (waiting or between rounds).
func (s *Service) UpdateSettings(lobbyID string, update SettingsUpdate) (*Lobby, error) {
	l, err := s.manager.Get(lobbyID)

	if err != nil {
		return nil, err
	}

	if err := l.UpdateSettings(update); err != nil {
		return nil, err
	}

	s.manager.Touch(l)

	return l, nil
}

// StartCountdown begins the pre-buzz countdown ending at endsAt.
func (s *Service) StartCountdown(lobbyID string, endsAt time.Time) (*Lobby, error) {
	l, err := s.manager.Get(lobbyID)

	if err != nil {
		return nil, err
	}

	if err := l.StartCountdown(endsAt); err != nil {
		return nil, err
	}

	s.manager.Touch(l)

	return l, nil
}

// OpenBuzz opens the buzzer for a round.
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

// NextRound closes out the finished round and resets to StateReady.
func (s *Service) NextRound(lobbyID string) (*Lobby, error) {
	l, err := s.manager.Get(lobbyID)

	if err != nil {
		return nil, err
	}

	if _, err := l.NextRound(); err != nil {
		return nil, err
	}

	s.manager.Touch(l)

	return l, nil
}
