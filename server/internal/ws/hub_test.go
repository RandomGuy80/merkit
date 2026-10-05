package ws

import (
	"encoding/json"
	"testing"
)

func TestHubSendAndReceive(t *testing.T) {
	hub := NewHub()

	// register a fake client with nil conn (we only test channel delivery)
	c := &client{userID: "user1", send: make(chan []byte, 8)}
	hub.mu.Lock()
	hub.clients["user1"] = append(hub.clients["user1"], c)
	hub.mu.Unlock()

	msg := Message{Type: "new_order", Payload: map[string]string{"order_id": "abc"}}
	hub.Send("user1", msg)

	select {
	case data := <-c.send:
		var got Message
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Type != "new_order" {
			t.Fatalf("expected new_order, got %s", got.Type)
		}
	default:
		t.Fatal("expected message in channel")
	}

	// send to unknown user — should not panic
	hub.Send("unknown-user", msg)
}

func TestHubUnregister(t *testing.T) {
	hub := NewHub()
	c := &client{userID: "user2", send: make(chan []byte, 8)}
	hub.mu.Lock()
	hub.clients["user2"] = append(hub.clients["user2"], c)
	hub.mu.Unlock()

	hub.Unregister(c)

	hub.mu.RLock()
	count := len(hub.clients["user2"])
	hub.mu.RUnlock()

	if count != 0 {
		t.Fatalf("expected 0 clients after unregister, got %d", count)
	}
}
