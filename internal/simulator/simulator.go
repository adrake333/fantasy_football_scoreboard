package simulator




import (
	"math/rand"
	"sync"
	"time"

	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
)




type Simulator struct {
	mu			sync.RWMutex
	matchups	[]fantasy.Matchup
	subscribers	map[chan []fantasy.Matchup]bool
	stopChan	chan struct{}
} 

func NewSimulator(initialMatchups []fantasy.Matchup) *Simulator {
	return &Simulator{
		matchups:		initialMatchups,
		subscribers:	make(map[chan []fantasy.Matchup]bool),
		stopChan:		make(chan struct{}),
	}
}

func (s *Simulator) Start(tickInterval time.Duration) {
	go func() {
		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				s.tick()
			}
		}
	}()
}

func (s *Simulator) tick() {
	s.mu.Lock()
	if len(s.matchups) == 0 {
		s.mu.Unlock()
		return
	}

	idx := rand.Intn(len(s.matchups))

	pointGain := float64(rand.Intn(55)+5) / 10.0

	if rand.Float32() < 0.5 {
		s.matchups[idx].UserScore += pointGain
	} else {
		s.matchups[idx].OpponentScore += pointGain
	}

	updatedCopy := make([]fantasy.Matchup, len(s.matchups))
	copy(updatedCopy, s.matchups)
	s.mu.Unlock()

	s.broadcast(updatedCopy)
}

func (s *Simulator) Subscribe() chan []fantasy.Matchup {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan []fantasy.Matchup, 1)
	s.subscribers[ch] = true
	return ch
}

func (s *Simulator) Unsubscribe(ch chan []fantasy.Matchup) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subscribers, ch)
	close(ch)
}

func (s *Simulator) broadcast(data []fantasy.Matchup) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.subscribers {
		select {
		case ch <- data:
		default:
		}
	}
}