package events

import (
	"io"
	"testing"
)

func TestWriteError_RoundTrip(t *testing.T) {
	pr, pw := io.Pipe()

	go func() {
		WriteError(pw, "cname %s does not point to %s", "a.com", "b.jprq.live")
		pw.Close()
	}()

	var result Event[TunnelOpened]
	if err := result.Read(pr); err != nil {
		t.Fatalf("error reading: %v", err)
	}

	want := "cname a.com does not point to b.jprq.live"
	if result.Data.ErrorMessage != want {
		t.Errorf("expected %q, got %q", want, result.Data.ErrorMessage)
	}
}

func TestWriteError_Args(t *testing.T) {
	tests := []struct {
		name    string
		message string
		args    []string
		want    string
	}{
		{"no args", "subdomain is busy", nil, "subdomain is busy"},
		{"single arg", "subdomain is busy: %s", []string{"foo"}, "subdomain is busy: foo"},
		{"multi arg", "cname %s does not point to %s", []string{"a.com", "b.jprq.live"}, "cname a.com does not point to b.jprq.live"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteError(io.Discard, tt.message, tt.args...)
			if err.Error() != tt.want {
				t.Errorf("expected %q, got %q", tt.want, err.Error())
			}
		})
	}
}

func TestEvent_EncodeDecode(t *testing.T) {
	event := Event[ConnectionReceived]{
		Data: &ConnectionReceived{
			ClientIP:    []byte{127, 0, 0, 1},
			ClientPort:  8000,
			RateLimited: false,
		},
	}

	data, err := event.encode()
	if err != nil {
		t.Errorf("error encoding %v", err)
	}

	var result Event[ConnectionReceived]
	err = result.decode(data)
	if err != nil {
		t.Errorf("error decoding %v", err)
	}

	if result.Data.ClientIP.String() != event.Data.ClientIP.String() {
		t.Logf("expected %s, got %s", event.Data.ClientIP, result.Data.ClientIP)
		t.Fail()
	}

	if result.Data.ClientPort != event.Data.ClientPort {
		t.Logf("expected %d, got %d", event.Data.ClientPort, result.Data.ClientPort)
		t.Fail()
	}
}

func TestEvent_ReadWrite(t *testing.T) {
	pr, pw := io.Pipe()
	event := Event[ConnectionReceived]{
		Data: &ConnectionReceived{
			ClientIP:    []byte{127, 0, 0, 1},
			ClientPort:  8000,
			RateLimited: false,
		},
	}

	go func() {
		if err := event.Write(pw); err != nil {
			t.Errorf("error writing: %v", err)
		}
	}()

	var result Event[ConnectionReceived]
	if err := result.Read(pr); err != nil {
		t.Errorf("error reading: %v", err)
	}

	if result.Data.ClientIP.String() != event.Data.ClientIP.String() {
		t.Logf("expected %s, got %s", event.Data.ClientIP, result.Data.ClientIP)
		t.Fail()
	}

	if result.Data.ClientPort != event.Data.ClientPort {
		t.Logf("expected %d, got %d", event.Data.ClientPort, result.Data.ClientPort)
		t.Fail()
	}
}
