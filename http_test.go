package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testNewApp = `
{
	"id": "/new",
	"cmd": "env && sleep 300",
	"args": ["/bin/sh", "-c", "env && sleep 300"]
}
`

var testOldApp = `
{
	"id": "/old",
	"cmd": "env && sleep 300",
	"args": ["/bin/sh", "-c", "env && sleep 300"]
}
`

var testNewGroup = `
{
	"id": "/new",
	"apps": [{
		"id": "/new/app",
		"cmd": "env && sleep 300",
		"args": ["/bin/sh", "-c", "env && sleep 300"]
	}]
}
`

func TestEventListener(t *testing.T) {
	data := re.ReplaceAllString(event_tests["api_post_event"], "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Ensure that the correct header is received
		if r.Header.Get("Accept") != "text/event-stream" {
			http.Error(w, "Error 405", 405)
		}

		// Simulate a bunch of blank lines along with the payload
		w.WriteHeader(http.StatusOK)
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

func TestDeployApplication(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		putResponse := []byte(`{"deploymentId": "867ed450-f6a8-4d33-9b0e-e11c5513990b"}`)
		postResponse := []byte(`{"deployments":[{"id": "867ed450-f6a8-4d33-9b0e-e11c5513990b"}]}`)

		// Ensure that the correct header is received
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", 415)
		}

		switch r.Method {

		case "GET":

			// This isn't actually an error, but we want
			// to catch places where we GET the wrong url.
			if r.URL.Path == appPath {
				http.Error(w, "Not found", 500)
			}

			if r.URL.Path == appPath+"/new" {
				http.Error(w, "Not found", 404)
			}

			if r.URL.Path == groupPath+"/new" {
				http.Error(w, "Not found", 404)
			}

			if r.URL.Path == appPath+"/old" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "%s\r\n", `{"id":"/old"}`)
			}

		case "PUT":

			if r.URL.Path == appPath+"/new" {
				// Marathon doesn't actually error here, but we want
				// to conform to the strict API for future compat.
				http.Error(w, "Error", 500)
			}

			if r.URL.Path == groupPath+"/new" {
				http.Error(w, "Error", 500)
			}

			if r.URL.Path == appPath+"/old" {
				w.Header().Set("Content-Type", "application/json")
				w.Write(putResponse)
			}

		case "POST":

			if r.URL.Path == appPath {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				w.Write(postResponse)
			}

			if r.URL.Path == groupPath {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				w.Write(postResponse)
			}

		}

	}))
	defer ts.Close()

	j, err := NewJob([]byte(testNewApp))
	if err != nil {
		t.Fatal(err)
	}

	id, err := DeployApplication(ts.URL, j)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "867ed450-f6a8-4d33-9b0e-e11c5513990b", id)

	j, err = NewJob([]byte(testOldApp))
	if err != nil {
		t.Fatal(err)
	}

	id, err = DeployApplication(ts.URL, j)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "867ed450-f6a8-4d33-9b0e-e11c5513990b", id)

	j, err = NewJob([]byte(testNewGroup))
	if err != nil {
		t.Fatal(err)
	}

	id, err = DeployApplication(ts.URL, j)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "867ed450-f6a8-4d33-9b0e-e11c5513990b", id)

}
