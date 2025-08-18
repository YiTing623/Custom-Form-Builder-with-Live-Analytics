package ws

import (
	"sync"
)

type Subscriber chan []byte

type Hub struct {
	mu    sync.RWMutex
	subs  map[string]map[Subscriber]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[string]map[Subscriber]struct{}),
	}
}

func (h *Hub) Subscribe(formID string) Subscriber {
	ch := make(Subscriber, 8)
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subs[formID]; !ok {
		h.subs[formID] = make(map[Subscriber]struct{})
	}
	h.subs[formID][ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(formID string, ch Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.subs[formID]; ok {
		if _, exist := set[ch]; exist {
			delete(set, ch)
			close(ch)
			if len(set) == 0 {
				delete(h.subs, formID)
			}
		}
	}
}

func (h *Hub) Broadcast(formID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if set, ok := h.subs[formID]; ok {
		for ch := range set {
			select {
			case ch <- payload:
			default:
			}
		}
	}
}
