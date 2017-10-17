package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

//
// HTTP sending and recieving
//
const (
	eventPath = "/v2/events"
	groupPath = "/v2/groups"
	appPath   = "/v2/apps"
)

var (
	nameRexp, dataRexp *regexp.Regexp
)

func init() {
	nameRexp = regexp.MustCompile(`^event: ([[:graph:]]+)$`)
	dataRexp = regexp.MustCompile(`^data: ([[:graph:]]+)$`)
}

func EventListener(rawurl string, ch chan<- RawEvent) (err error) {

	client := new(http.Client)

	eventUrl, err := url.Parse(rawurl)
	if err != nil {
		return
	}
	eventUrl.Path = eventPath

	req, err := http.NewRequest("GET", eventUrl.String(), nil)
	if err != nil {
		close(ch)
		return
	}
	req.Header.Add("Accept", "text/event-stream")

	if authenticate {
		req.SetBasicAuth(user, pass)
	}

	resp, err := client.Do(req)
	if err != nil {
		close(ch)
		return
	}

	if resp.StatusCode != 200 {
		close(ch)
		return errors.New("Error, got response " + resp.Status)
	}

	go func() {

		reader := bufio.NewReader(resp.Body)

		var ev RawEvent

		for {

			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}

			switch {

			// name of event
			case bytes.HasPrefix(line, []byte("event:")):
				ev.Name = string(line[7:])

			// event data
			case bytes.HasPrefix(line, []byte("data:")):
				ev.Data = line[6:]
				ch <- ev

			}
		}

	}()

	return nil

}

// We are only interested in the deploymentId of the response.
// Unfortunately, the marathon API is very inconsistent, and the
// response we get varies.
type Response struct {
	DeploymentId string
	Deployments  []struct {
		Id string
	}
}

func DeployApplication(rawurl string, job Job) (deploymentId string, err error) {

	// var jobUrl string

	jobUrl, err := url.Parse(rawurl)
	if err != nil {
		return
	}

	if job.IsGroup() {
		jobUrl.Path = groupPath
	} else {
		jobUrl.Path = appPath
	}

	client := new(http.Client)

	// Check if we should do a POST or PUT
	req, err := http.NewRequest("GET", jobUrl.String()+job.Id(), nil)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	if authenticate {
		req.SetBasicAuth(user, pass)
	}

	var method string

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	switch resp.StatusCode {
	// Existing job found, update it
	case 200:
		if debug {
			log.Println("Existing job found, updating")
		}
		method = "PUT"
		jobUrl.Path += job.Id()

		if force {
			jobUrl.RawQuery = "force=true"
		}

	// New job
	case 404:
		if debug {
			log.Println("Creating new job")
		}
		method = "POST"

	// Error, abort
	default:
		err = fmt.Errorf("Unexpected response code. HTTP status code: %s", resp.Status)
		return
	}

	data, err := job.Data()
	if err != nil {
		return
	}

	req, err = http.NewRequest(method, jobUrl.String(), bytes.NewReader(data))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	if authenticate {
		req.SetBasicAuth(user, pass)
	}

Loop:
	for {
		resp, err = client.Do(req)
		if err != nil {
			return
		}

		switch resp.StatusCode {

		case 200:
			break Loop

		case 201:
			break Loop

		case 409:
			if !force {
				time.Sleep(30 * time.Second)
				continue
			}

		default:
			break Loop

		}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode > 399 {
		err = fmt.Errorf("ERROR - marathon returned an error response. HTTP status: %s, message: %s", resp.Status, string(body))
		return
	}

	var r Response

	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}

	switch {
	case r.DeploymentId != "":
		deploymentId = r.DeploymentId
	case len(r.Deployments) >= 1:
		deploymentId = r.Deployments[0].Id
	default:
		err = errors.New("No deployment ID detected in response")
	}

	return

}
