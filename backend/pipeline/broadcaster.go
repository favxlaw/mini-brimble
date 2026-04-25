package pipeline

import "sync"

type LogBroadcaster struct {
	mu          sync.RWMutex
	subscribers map[string][]chan string
}

func NewLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{
		subscribers: make(map[string][]chan string),
	}
}

func (b *LogBroadcaster) Subscribe(deploymentID string) (<-chan string, func()) {
	ch := make(chan string, 100)

	b.mu.Lock()
	b.subscribers[deploymentID] = append(b.subscribers[deploymentID], ch)
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		subs := b.subscribers[deploymentID]
		for i, sub := range subs {
			if sub == ch {
				subs[i] = subs[len(subs)-1]
				b.subscribers[deploymentID] = subs[:len(subs)-1]
				break
			}
		}
		close(ch)
	}

	return ch, cancel
}

func (b *LogBroadcaster) Publish(deploymentID, line string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers[deploymentID] {
		select {
		case ch <- line:
		default:
		}
	}
}
