package mqtt

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"

	mqttv1 "github.com/morvencao/event-controller/pkg/apis/v1"
)

type EventClient struct {
	// Buffered channel of outbound messages.
	Channel chan event.GenericEvent
}

func NewEventClient() *EventClient {
	return &EventClient{
		Channel: make(chan event.GenericEvent, 1),
	}
}

type EventHub struct {
	// lock for registered clients
	mu sync.RWMutex

	// Registered clients.
	clients map[*EventClient]struct{}

	// Inbound messages from the clients.
	broadcast chan event.GenericEvent
}

func NewEventHub() *EventHub {
	return &EventHub{
		broadcast: make(chan event.GenericEvent),
		clients:   make(map[*EventClient]struct{}),
	}
}

func (h *EventHub) Register(c *EventClient) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *EventHub) Unregister(c *EventClient) {
	h.mu.Lock()
	delete(h.clients, c)
	close(c.Channel)
	h.mu.Unlock()
}

func (h *EventHub) Broadcast(e event.GenericEvent) {
	h.broadcast <- e
}

func (h *EventHub) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			for client := range h.clients {
				close(client.Channel)
			}
			return nil
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.Channel <- message:
				default:
					close(client.Channel)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Decoder Sink
func (h *EventHub) Delete(uid types.UID) error {
	// noop
	return nil
}

func (h *EventHub) Update(msg *mqttv1.ResourceMessage) error {
	h.Broadcast(event.GenericEvent{
		Object: msg.Content,
	})
	return nil
}
