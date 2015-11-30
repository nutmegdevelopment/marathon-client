package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
)

//
// HTTP sending and recieving
//
const (
	eventEndPoint = "/v2/events"
	groupEndPoint = "/v2/groups"
	appEndPoint   = "/v2/apps"
)

var (
	nameRexp, dataRexp *regexp.Regexp
)

func init() {
	nameRexp = regexp.MustCompile(`^event: ([[:graph:]]+)$`)
	dataRexp = regexp.MustCompile(`^data: ([[:graph:]]+)$`)
}

func EventListener(url string, ch chan<- RawEvent) (err error) {

	client := new(http.Client)

	req, err := http.NewRequest("GET", url+eventEndPoint, nil)
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

		defer close(ch)

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

// We are only interested in the deploymentId of the response
type Response struct {
	DeploymentId string
}

func DeployApplication(url string, job Job) (deploymentId string, err error) {

	var jobUrl string

	if job.IsGroup() {
		jobUrl = url + groupEndPoint
	} else {
		jobUrl = url + appEndPoint
	}

	client := new(http.Client)

	// Check if we should do a POST or PUT
	req, err := http.NewRequest("GET", jobUrl+job.Id(), nil)
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
		method = "PUT"
		jobUrl += job.Id()

	// New job
	case 404:
		method = "POST"

	// Error, abort
	default:
		err = errors.New(resp.Status)
		return
	}

	data, err := job.Data()
	if err != nil {
		return
	}

	req, err = http.NewRequest(method, jobUrl, bytes.NewReader(data))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	if authenticate {
		req.SetBasicAuth(user, pass)
	}

	resp, err = client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		err = errors.New(resp.Status)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var r Response

	err = json.Unmarshal(body, &r)

	deploymentId = r.DeploymentId
	return

}
