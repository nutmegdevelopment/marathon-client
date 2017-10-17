package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

//
// Track deployment lifecycle
//

type appFailures struct {
	apps    []string
	actions []string
}

func (a *appFailures) print() string {
	str := "Failure reason(s):"
	for i := range a.apps {
		str = fmt.Sprintf("%s\nApplication: %s\nAction: %s\n",
			str, a.apps[i], a.actions[i])
	}
	return str
}

func (a *appFailures) add(app, action string) {
	apps := append(a.apps, app)
	a.apps = apps

	actions := append(a.actions, action)
	a.actions = actions
}

// lookupApp looks for an AppId in a list of Actions
func lookupApp(list []Action, appId string) bool {
	for i := range list {
		if list[i].App == appId {
			return true
		}
	}
	return false
}

func TrackDeployment(id string, events <-chan Event) (duration time.Duration, err error) {

	if debug {
		log.Println("Tracking deployment ID:", id)
	}

	// Actions for this deployment
	actions := make([]Action, 0)

	var failures appFailures

	var start, end time.Time

	for e := range events {

		switch {

		// Build list of actions, and set start time
		case e.Name == "deployment_info" &&
			e.DeploymentStatus.Plan.Id == id:

			start = e.DeploymentStatus.Timestamp.Time()
			actions = e.DeploymentStatus.Plan.Steps

		case e.Name == "deployment_step_success" &&
			e.DeploymentStatus.Plan.Id == id:

			if debug {
				log.Println(
					e.DeploymentStatus.CurrentStep.Actions[0].App,
					e.DeploymentStatus.CurrentStep.Actions[0].Type,
					"Succeeded")
			}

		case e.Name == "deployment_step_failure" &&
			e.DeploymentStatus.Plan.Id == id:

			failures.add(
				e.DeploymentStatus.CurrentStep.Actions[0].App,
				e.DeploymentStatus.CurrentStep.Actions[0].Type)

			if debug {
				log.Println(
					e.DeploymentStatus.CurrentStep.Actions[0].App,
					e.DeploymentStatus.CurrentStep.Actions[0].Type,
					"Failed")
			}

		case e.Name == "add_health_check_event" &&
			lookupApp(actions, e.AddHealthCheck.AppId):

			if debug {
				log.Println("Healthcheck added for", e.AddHealthCheck.AppId)
			}

		case e.Name == "failed_health_check_event" &&
			lookupApp(actions, e.FailedHealthCheck.AppId):

			failures.add(e.FailedHealthCheck.AppId, "HealthCheck")

			if debug {
				log.Println("Healthcheck failed for", e.FailedHealthCheck.AppId)
			}

		case e.Name == "health_status_changed_event" &&
			lookupApp(actions, e.HealthStatusChanged.AppId):

			if debug {
				log.Println(
					"Healthcheck status for",
					e.HealthStatusChanged.AppId,
					"changed to",
					e.HealthStatusChanged.Alive)

			}

		case e.Name == "status_update_event" &&
			lookupApp(actions, e.MesosStatusUpdateEvent.AppId):

			if lookupApp(actions, e.MesosStatusUpdateEvent.AppId) && debug {
				log.Println(e.MesosStatusUpdateEvent.AppId,
					"running on host", e.MesosStatusUpdateEvent.Host)
			}

		case e.Name == "deployment_success" &&
			e.DeploymentStatus.Id == id:

			end = e.DeploymentStatus.Timestamp.Time()
			if start.Year() == 1 {
				start = end
			}
			return end.Sub(start), nil

		case e.Name == "deployment_failed" &&
			e.DeploymentStatus.Id == id:

			end = e.DeploymentStatus.Timestamp.Time()
			if start.Year() == 1 {
				start = end
			}
			err = fmt.Errorf("%s:\n%s", "Deployment failed", failures.print())
			return end.Sub(start), err

		}
	}

	return 0, errors.New("Failed to track deployment")

}
