package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEventListener(t *testing.T) {
	data := re.ReplaceAllString(event_tests["api_post_event"], "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Ensure that the correct header is received
		if r.Header.Get("Accept") != "text/event-stream" {
			http.Error(w, "Error 405", 405)
		}

		// Simulate a bunch of blank lines along with the payload
		fmt.Fprintf(w, "\r\n\r\nevent: %s\r\ndata: %s\r\n\r\n", "api_post_event", data)
	}))
	defer ts.Close()

	ch := make(chan RawEvent)

	err := EventListener(ts.URL, ch)
	if err != nil {
		t.Error(err)
	}

	res := <-ch

	// Check that we recieved the event
	assert.Equal(t, "api_post_event\r\n", res.Name)
	assert.Equal(t, data+"\r\n", string(res.Data))

	// Test that we can unmarshal the event
	e := new(Event)
	err = e.Unmarshal(res)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "api_post_event", e.ApiPostEvent.EventType)
	assert.Equal(t, "0:0:0:0:0:0:0:1", e.ApiPostEvent.ClientIp)
	assert.Equal(t, "2014-03-01 23:29:30.158 +0000 UTC", e.ApiPostEvent.Timestamp.String())

}
