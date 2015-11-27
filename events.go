package main

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

//
// Marathon API support
//

// Raw event from the SSE stream
type RawEvent struct {
	Name string
	Data []byte
}

type Timestamp string

func (t Timestamp) String() string {
	parsed, err := time.Parse(time.RFC3339, string(t))
	if err != nil {
		return ""
	}
	return parsed.String()
}

func (t Timestamp) Time() time.Time {
	parsed, err := time.Parse(time.RFC3339, string(t))
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// Event is the main event container
type Event struct {
	Name                   string
	ApiPostEvent           ApiPostEvent
	AddHealthCheck         AddHealthCheck
	FailedHealthCheck      FailedHealthCheck
	HealthStatusChanged    HealthStatusChanged
	GroupChangeSuccess     GroupChangeSuccess
	GroupChangeFailed      GroupChangeFailed
	DeploymentStatus       DeploymentStatus
	MesosStatusUpdateEvent MesosStatusUpdateEvent
}

type EventCommon struct {
	EventType string
	Timestamp Timestamp
}

func (e EventCommon) Type() string {
	return e.EventType
}

func (e EventCommon) Date() time.Time {
	return e.Timestamp.Time()
}

// POST event
type ApiPostEvent struct {
	ClientIp      string
	Uri           string
	AppDefinition map[string]interface{}
	EventCommon
}

// Health check events
type AddHealthCheck struct {
	AppId       string
	Version     Timestamp
	HealthCheck map[string]interface{}
	EventCommon
}

type FailedHealthCheck struct {
	AppId       string
	TaskId      string
	HealthCheck map[string]interface{}
	EventCommon
}

type HealthStatusChanged struct {
	AppId   string
	TaskId  string
	Version Timestamp
	Alive   bool
	EventCommon
}

// Group events
type GroupChangeSuccess struct {
	GroupId string
	Version string
	EventCommon
}

type GroupChangeFailed struct {
	GroupId string
	Version string
	Reason  string
	EventCommon
}

type Action struct {
	Action string
	App    string
}

type DeploymentPlan struct {
	Id       string
	Original map[string]interface{}
	Target   map[string]interface{}
	Steps    []Action
	Version  Timestamp
}

type DeploymentStatus struct {
	Id          string
	Plan        DeploymentPlan
	CurrentStep struct {
		Actions []struct {
			Type string
			App  string
		}
	}
	EventCommon
}

// Mesos events
type MesosStatusUpdateEvent struct {
	SlaveId    string
	TaskId     string
	TaskStatus string
	Message    string
	AppId      string
	Host       string
	Ports      []int
	Version    string
	EventCommon
}

func (e *Event) Unmarshal(in RawEvent) (err error) {

	if in.Name == "" || len(in.Data) == 0 {
		return errors.New("Bad event object")
	}

	e.Name = strings.TrimRight(in.Name, "\r\n")

	switch e.Name {

	case "api_post_event":
		err = json.Unmarshal(in.Data, &e.ApiPostEvent)

	case "add_health_check_event":
		err = json.Unmarshal(in.Data, &e.AddHealthCheck)

	case "failed_health_check_event":
		err = json.Unmarshal(in.Data, &e.FailedHealthCheck)

	case "health_status_changed_event":
		err = json.Unmarshal(in.Data, &e.HealthStatusChanged)

	case "group_change_success":
		err = json.Unmarshal(in.Data, &e.GroupChangeSuccess)

	case "group_change_failed":
		err = json.Unmarshal(in.Data, &e.GroupChangeFailed)

	case "deployment_success":
		err = json.Unmarshal(in.Data, &e.DeploymentStatus)

	case "deployment_failed":
		err = json.Unmarshal(in.Data, &e.DeploymentStatus)

	case "deployment_info":
		err = json.Unmarshal(in.Data, &e.DeploymentStatus)

	case "deployment_step_success":
		err = json.Unmarshal(in.Data, &e.DeploymentStatus)

	case "deployment_step_failure":
		err = json.Unmarshal(in.Data, &e.DeploymentStatus)

	case "status_update_event":
		err = json.Unmarshal(in.Data, &e.MesosStatusUpdateEvent)

	default:
		err = errors.New("Unhandled event")

	}

	return
}
