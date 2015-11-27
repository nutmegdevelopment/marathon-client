package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventBus(t *testing.T) {

	var raw RawEvent
	raw.Name = "api_post_event"
	raw.Data = []byte(re.ReplaceAllString(event_tests["api_post_event"], ""))

	in := make(chan RawEvent)
	out := make(chan Event)

	go eventBus(in, out)

	select {

	case parsed, ok := <-out:
		if ok {
			assert.Equal(t, "api_post_event", parsed.ApiPostEvent.EventType)
			assert.Equal(t, "0:0:0:0:0:0:0:1", parsed.ApiPostEvent.ClientIp)
			assert.Equal(t, "2014-03-01 23:29:30.158 +0000 UTC", parsed.ApiPostEvent.Timestamp.String())
		}

	default:
		in <- raw
		close(in)

	}

}
