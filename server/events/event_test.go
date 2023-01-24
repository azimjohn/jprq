package events

import (
	"testing"
)

func TestEvent_EncodeDecode(t *testing.T) {
	event := Event[ConnectionReceived]{
		Data: &ConnectionReceived{
			ClientIP:    "127.0.0.1",
			ClientPort:  8000,
			RateLimited: false,
		},
	}

	data, err := event.Encode()
	if err != nil {
		t.Errorf("error encoding %v", err)
	}

	var result Event[ConnectionReceived]
	err = result.Decode(data)
	if err != nil {
		t.Errorf("error decoding %v", err)
	}

	if result.Data.ClientIP != event.Data.ClientIP {
		t.Logf("expected %s, got %s", event.Data.ClientIP, result.Data.ClientIP)
		t.Fail()
	}

	if result.Data.ClientPort != event.Data.ClientPort {
		t.Logf("expected %d, got %d", event.Data.ClientPort, result.Data.ClientPort)
		t.Fail()
	}
}
