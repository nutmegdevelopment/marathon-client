package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

var api_post_event = `{
  "eventType": "api_post_event",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "clientIp": "0:0:0:0:0:0:0:1",
  "uri": "/v2/apps/my-app",
  "appDefinition": {
    "args": [],
    "backoffFactor": 1.15, 
    "backoffSeconds": 1, 
    "cmd": "sleep 30", 
    "constraints": [], 
    "container": null, 
    "cpus": 0.2, 
    "dependencies": [], 
    "disk": 0.0, 
    "env": {}, 
    "executor": "", 
    "healthChecks": [], 
    "id": "/my-app", 
    "instances": 2, 
    "mem": 32.0, 
    "ports": [10001], 
    "requirePorts": false, 
    "storeUrls": [], 
    "upgradeStrategy": {
        "minimumHealthCapacity": 1.0
    }, 
    "uris": [], 
    "user": null, 
    "version": "2014-09-09T05:57:50.866Z"
  }
}`

var status_update_event = `{
  "eventType": "status_update_event",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "slaveId": "20140909-054127-177048842-5050-1494-0",
  "taskId": "my-app_0-1396592784349",
  "taskStatus": "TASK_RUNNING",
  "appId": "/my-app",
  "host": "slave-1234.acme.org",
  "ports": [31372],
  "version": "2014-04-04T06:26:23.051Z"
}`

var add_health_check_event = `{
  "eventType": "add_health_check_event",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "appId": "/my-app",
  "healthCheck": {
    "protocol": "HTTP",
    "path": "/health",
    "portIndex": 0,
    "gracePeriodSeconds": 5,
    "intervalSeconds": 10,
    "timeoutSeconds": 10,
    "maxConsecutiveFailures": 3
  }
}`

var remove_health_check_event = `{
  "eventType": "remove_health_check_event",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "appId": "/my-app",
  "healthCheck": {
    "protocol": "HTTP",
    "path": "/health",
    "portIndex": 0,
    "gracePeriodSeconds": 5,
    "intervalSeconds": 10,
    "timeoutSeconds": 10,
    "maxConsecutiveFailures": 3
  }
}`

var failed_health_check_event = `{
  "eventType": "failed_health_check_event",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "appId": "/my-app",
  "taskId": "my-app_0-1396592784349",
  "healthCheck": {
    "protocol": "HTTP",
    "path": "/health",
    "portIndex": 0,
    "gracePeriodSeconds": 5,
    "intervalSeconds": 10,
    "timeoutSeconds": 10,
    "maxConsecutiveFailures": 3
  }
}`

var health_status_changed_event = `{
  "eventType": "health_status_changed_event",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "appId": "/my-app",
  "taskId": "my-app_0-1396592784349",
  "version": "2014-04-04T06:26:23.051Z",
  "alive": true
}`

var group_change_success = `{
  "eventType": "group_change_success",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "groupId": "/product-a/backend",
  "version": "2014-04-04T06:26:23.051Z"
}`

var group_change_failed = `{
  "eventType": "group_change_failed",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "groupId": "/product-a/backend",
  "version": "2014-04-04T06:26:23.051Z",
  "reason": ""
}`

var deployment_success = `{
  "eventType": "deployment_success",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "id": "867ed450-f6a8-4d33-9b0e-e11c5513990b"
}`

var deployment_failed = `{
  "eventType": "deployment_failed",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "id": "867ed450-f6a8-4d33-9b0e-e11c5513990b"
}`

var deployment_info = `{
  "eventType": "deployment_info",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "plan": {
    "id": "867ed450-f6a8-4d33-9b0e-e11c5513990b",
    "original": {
      "apps": [], 
      "dependencies": [], 
      "groups": [], 
      "id": "/", 
      "version": "2014-09-09T06:30:49.667Z"
    },
    "target": {
      "apps": [
        {
          "args": [],
          "backoffFactor": 1.15, 
          "backoffSeconds": 1, 
          "cmd": "sleep 30", 
          "constraints": [], 
          "container": null, 
          "cpus": 0.2, 
          "dependencies": [], 
          "disk": 0.0, 
          "env": {}, 
          "executor": "", 
          "healthChecks": [], 
          "id": "/my-app", 
          "instances": 2, 
          "mem": 32.0, 
          "ports": [10001], 
          "requirePorts": false, 
          "storeUrls": [], 
          "upgradeStrategy": {
              "minimumHealthCapacity": 1.0
          }, 
          "uris": [], 
          "user": null, 
          "version": "2014-09-09T05:57:50.866Z"
        }
      ], 
      "dependencies": [], 
      "groups": [], 
      "id": "/", 
      "version": "2014-09-09T05:57:50.866Z"
    },
    "steps": [
      {
        "action": "ScaleApplication",
        "app": "/my-app"
      }
    ],
    "version": "2014-03-01T23:24:14.846Z"
  },
  "currentStep": {
    "actions": [
      {
        "type": "ScaleApplication",
        "app": "/my-app"
      }
    ]
  }
}`

var deployment_step_success = `{
  "eventType": "deployment_step_success",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "plan": {
    "id": "867ed450-f6a8-4d33-9b0e-e11c5513990b",
    "original": {
      "apps": [], 
      "dependencies": [], 
      "groups": [], 
      "id": "/", 
      "version": "2014-09-09T06:30:49.667Z"
    },
    "target": {
      "apps": [
        {
          "args": [],
          "backoffFactor": 1.15, 
          "backoffSeconds": 1, 
          "cmd": "sleep 30", 
          "constraints": [], 
          "container": null, 
          "cpus": 0.2, 
          "dependencies": [], 
          "disk": 0.0, 
          "env": {}, 
          "executor": "", 
          "healthChecks": [], 
          "id": "/my-app", 
          "instances": 2, 
          "mem": 32.0, 
          "ports": [10001], 
          "requirePorts": false, 
          "storeUrls": [], 
          "upgradeStrategy": {
              "minimumHealthCapacity": 1.0
          }, 
          "uris": [], 
          "user": null, 
          "version": "2014-09-09T05:57:50.866Z"
        }
      ], 
      "dependencies": [], 
      "groups": [], 
      "id": "/", 
      "version": "2014-09-09T05:57:50.866Z"
    },
    "steps": [
      {
        "action": "ScaleApplication",
        "app": "/my-app"
      }
    ],
    "version": "2014-03-01T23:24:14.846Z"
  },
  "currentStep": {
    "actions": [
      {
        "type": "ScaleApplication",
        "app": "/my-app"
      }
    ]
  }
}`

var deployment_step_failure = `{
  "eventType": "deployment_step_failure",
  "timestamp": "2014-03-01T23:29:30.158Z",
  "plan": {
    "id": "867ed450-f6a8-4d33-9b0e-e11c5513990b",
    "original": {
      "apps": [], 
      "dependencies": [], 
      "groups": [], 
      "id": "/", 
      "version": "2014-09-09T06:30:49.667Z"
    },
    "target": {
      "apps": [
        {
          "args": [],
          "backoffFactor": 1.15, 
          "backoffSeconds": 1, 
          "cmd": "sleep 30", 
          "constraints": [], 
          "container": null, 
          "cpus": 0.2, 
          "dependencies": [], 
          "disk": 0.0, 
          "env": {}, 
          "executor": "", 
          "healthChecks": [], 
          "id": "/my-app", 
          "instances": 2, 
          "mem": 32.0, 
          "ports": [10001], 
          "requirePorts": false, 
          "storeUrls": [], 
          "upgradeStrategy": {
              "minimumHealthCapacity": 1.0
          }, 
          "uris": [], 
          "user": null, 
          "version": "2014-09-09T05:57:50.866Z"
        }
      ], 
      "dependencies": [], 
      "groups": [], 
      "id": "/", 
      "version": "2014-09-09T05:57:50.866Z"
    },
    "steps": [
      {
        "action": "ScaleApplication",
        "app": "/my-app"
      }
    ],
    "version": "2014-03-01T23:24:14.846Z"
  },
  "currentStep": {
    "actions": [
      {
        "type": "ScaleApplication",
        "app": "/my-app"
      }
    ]
  }
}`

var event_tests = map[string]string{
	"api_post_event":              api_post_event,
	"status_update_event":         status_update_event,
	"add_health_check_event":      add_health_check_event,
	"remove_health_check_event":   remove_health_check_event,
	"failed_health_check_event":   failed_health_check_event,
	"health_status_changed_event": health_status_changed_event,
	"group_change_success":        group_change_success,
	"group_change_failed":         group_change_failed,
	"deployment_success":          deployment_success,
	"deployment_failed":           deployment_failed,
	"deployment_info":             deployment_info,
	"deployment_step_success":     deployment_step_success,
	"deployment_step_failure":     deployment_step_failure,
}

var re *regexp.Regexp

func init() {
	re = regexp.MustCompile(`[\n\t ]`)
}

func runEvent(name string) (Event, error) {
	var e Event
	err := e.Unmarshal(
		fmt.Sprintf(
			"event: %s\ndata: %s\n",
			name,
			re.ReplaceAllString(event_tests[name], "")))

	if err != nil {
		return e, err
	}
	if e.Name != name {
		return e, errors.New("Wrong event name for " + name)
	}
	return e, nil
}

func TestApiPostEvent(t *testing.T) {
	e, err := runEvent("api_post_event")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "api_post_event", e.ApiPostEvent.EventType)
	assert.Equal(t, "0:0:0:0:0:0:0:1", e.ApiPostEvent.ClientIp)
	assert.Equal(t, "2014-03-01 23:29:30.158 +0000 UTC", e.ApiPostEvent.Timestamp.String())
}

func TestStatusUpdateEvent(t *testing.T) {
	e, err := runEvent("status_update_event")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "status_update_event", e.MesosStatusUpdateEvent.EventType)
	assert.Equal(t, "20140909-054127-177048842-5050-1494-0", e.MesosStatusUpdateEvent.SlaveId)
	assert.Equal(t, 31372, e.MesosStatusUpdateEvent.Ports[0])
}

func TestAddHealthCheckEvent(t *testing.T) {
	e, err := runEvent("add_health_check_event")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "add_health_check_event", e.AddHealthCheck.EventType)
	assert.Equal(t, "/health", e.AddHealthCheck.HealthCheck["path"])
}

func TestRemoveHealthCheckEvent(t *testing.T) {
	_, err := runEvent("remove_health_check_event")
	assert.Equal(t, "Unhandled event", err.Error())
}

func TestFailedHealthCheckEvent(t *testing.T) {
	e, err := runEvent("failed_health_check_event")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "failed_health_check_event", e.FailedHealthCheck.EventType)
	assert.Equal(t, "/health", e.FailedHealthCheck.HealthCheck["path"])
}

func TestHealthCheckStatusChangedEvent(t *testing.T) {
	e, err := runEvent("health_status_changed_event")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "health_status_changed_event", e.HealthStatusChanged.EventType)
}

func TestGroupChangeSuccessEvent(t *testing.T) {
	e, err := runEvent("group_change_success")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "group_change_success", e.GroupChangeSuccess.EventType)
}

func TestGroupChangeFailedEvent(t *testing.T) {
	e, err := runEvent("group_change_failed")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "group_change_failed", e.GroupChangeFailed.EventType)
}

func TestDeploymentSuccessEvent(t *testing.T) {
	e, err := runEvent("deployment_success")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "deployment_success", e.DeploymentStatus.EventType)
}

func TestDeploymentFailedEvent(t *testing.T) {
	e, err := runEvent("deployment_failed")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "deployment_failed", e.DeploymentStatus.EventType)
}

func TestDeploymentInfoEvent(t *testing.T) {
	e, err := runEvent("deployment_info")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "deployment_info", e.DeploymentStatus.EventType)
}

func TestDeploymentStepSuccessEvent(t *testing.T) {
	e, err := runEvent("deployment_step_success")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "deployment_step_success", e.DeploymentStatus.EventType)
}

func TestDeploymentStepFailureEvent(t *testing.T) {
	e, err := runEvent("deployment_step_failure")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "deployment_step_failure", e.DeploymentStatus.EventType)
}
