package main

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

var eventMatch *regexp.Regexp

func init() {
	eventMatch = regexp.MustCompile(`(event|data): ([[:graph:]]+)`)
}

type PathId string

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
	AppId       PathId
	Version     Timestamp
	HealthCheck map[string]interface{}
	EventCommon
}

type FailedHealthCheck struct {
	AppId       PathId
	TaskId      string
	HealthCheck map[string]interface{}
	EventCommon
}

type HealthStatusChanged struct {
	AppId   PathId
	TaskId  string
	Version Timestamp
	Alive   bool
	EventCommon
}

// Group events
type GroupChangeSuccess struct {
	GroupId PathId
	Version string
	EventCommon
}

type GroupChangeFailed struct {
	GroupId PathId
	Version string
	Reason  string
	EventCommon
}

// Deployment events
type DeploymentStep struct {
	Actions []struct {
		Type string
		App  PathId
	}
}

type DeploymentPlan struct {
	Id       string
	Original map[string]interface{}
	Target   map[string]interface{}
	Steps    []DeploymentStep
	Version  Timestamp
}

type DeploymentStatus struct {
	Id   string
	Plan DeploymentPlan
	EventCommon
}

// Mesos events
type MesosStatusUpdateEvent struct {
	SlaveId    string
	TaskId     string
	TaskStatus string
	Message    string
	AppId      PathId
	Host       string
	Ports      []int
	Version    string
	EventCommon
}

func (e *Event) Unmarshal(in string) (err error) {

	data := eventMatch.FindAllStringSubmatch(in, -1)
	if len(data) != 2 || len(data[0]) < 3 || len(data[1]) < 3 {
		return errors.New("Bad event object")
	}

	e.Name = data[0][2]
	payload := []byte(data[1][2])

	switch e.Name {

	case "api_post_event":
		err = json.Unmarshal(payload, &e.ApiPostEvent)

	case "add_health_check_event":
		err = json.Unmarshal(payload, &e.AddHealthCheck)

	case "failed_health_check_event":
		err = json.Unmarshal(payload, &e.FailedHealthCheck)

	case "health_status_changed_event":
		err = json.Unmarshal(payload, &e.HealthStatusChanged)

	case "group_change_success":
		err = json.Unmarshal(payload, &e.GroupChangeSuccess)

	case "group_change_failed":
		err = json.Unmarshal(payload, &e.GroupChangeFailed)

	case "deployment_success":
		err = json.Unmarshal(payload, &e.DeploymentStatus)

	case "deployment_failed":
		err = json.Unmarshal(payload, &e.DeploymentStatus)

	case "deployment_info":
		err = json.Unmarshal(payload, &e.DeploymentStatus)

	case "deployment_step_success":
		err = json.Unmarshal(payload, &e.DeploymentStatus)

	case "deployment_step_failure":
		err = json.Unmarshal(payload, &e.DeploymentStatus)

	case "status_update_event":
		err = json.Unmarshal(payload, &e.MesosStatusUpdateEvent)

	default:
		err = errors.New("Unhandled event")

	}

	return
}
