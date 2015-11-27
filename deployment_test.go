package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"
)

var deploymentId = "867ed450-f6a8-4d33-9b0e-e11c5513990b"

var targetLogOutput = `Tracking deployment ID: 867ed450-f6a8-4d33-9b0e-e11c5513990b
Healthcheck added for /my-app
/my-app ScaleApplication Succeeded
/my-app running on host slave-1234.acme.org
Healthcheck status for /my-app changed to true
Tracking deployment ID: 867ed450-f6a8-4d33-9b0e-e11c5513990b
Healthcheck added for /my-app
/my-app ScaleApplication Failed
/my-app running on host slave-1234.acme.org
Healthcheck failed for /my-app
`

func TestTrackDeployment(t *testing.T) {

	// Test that we log the correct output

	debug = true
	var w bytes.Buffer
	log.SetOutput(&w)
	log.SetFlags(0)

	ch := make(chan Event, 64)

	// Test a successful deployment
	go func() {

		eventList := []string{
			"api_post_event",
			"group_change_success",
			"deployment_info",
			"add_health_check_event",
			"deployment_step_success",
			"status_update_event",
			"health_status_changed_event",
			"deployment_success"}

		for i := range eventList {
			e, err := runEvent(eventList[i])
			if err != nil {
				close(ch)
				t.Fatal(err)
			}

			ch <- e
		}

	}()

	_, err := TrackDeployment(deploymentId, ch)
	if err != nil {
		t.Error(err)
	}

	// Test a failed deployment
	go func() {

		eventList := []string{
			"api_post_event",
			"group_change_success",
			"deployment_info",
			"add_health_check_event",
			"deployment_step_failure",
			"status_update_event",
			"failed_health_check_event",
			"deployment_failed"}

		for i := range eventList {
			e, err := runEvent(eventList[i])
			if err != nil {
				t.Fatal(err)
			}

			ch <- e
		}

	}()

	_, err = TrackDeployment(deploymentId, ch)
	if err == nil {
		t.Error("No error on failed deployment")
	}

	debug = false

	output, err := ioutil.ReadAll(&w)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, targetLogOutput, string(output))

}
