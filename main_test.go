package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testJson = `
{
	"id": "/product/service/myApp",
	"cmd": "env && sleep 300",
	"args": ["/bin/sh", "-c", "env && sleep 300"]
}
`

var testBadJson = `
{
	"cmd": "env && sleep 300",
	"args": ["/bin/sh", "-c", "env && sleep 300"]
}
`

var testEmptyIDJson = `
{
	"id": "",
	"cmd": "env && sleep 300",
	"args": ["/bin/sh", "-c", "env && sleep 300"]
}
`

var testGroupJson = `
{
	"id": "/product",
	"apps": [ 
		{
			"id":"/product/service",
			"cmd": "env && sleep 300"
		}
	]
}
`

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

func TestNewJob(t *testing.T) {

	j, err := NewJob([]byte(testJson))
	if err != nil {
		t.Error(err)
	}

	if j.IsGroup() {
		t.Error("Should not be a group")
	}

	j, err = NewJob([]byte(testBadJson))
	if err == nil {
		t.Error("No error with missing ID")
	}

	j, err = NewJob([]byte(testEmptyIDJson))
	if err == nil {
		t.Error("No error with empty ID")
	}

	j, err = NewJob([]byte(testGroupJson))
	if err != nil {
		t.Error(err)
	}

	if !j.IsGroup() {
		t.Error("Should be a group", j)
	}

}
