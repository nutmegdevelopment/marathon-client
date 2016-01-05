package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var (
	url, file    string
	user, pass   string
	debug        bool
	authenticate bool
	wait         bool
)

func init() {
	flag.StringVar(&url, "m", "", "Marathon URL")
	flag.StringVar(&file, "f", "", "Job file")
	flag.StringVar(&user, "u", "", "Username for basic auth")
	flag.StringVar(&pass, "p", "", "Password for basic auth")
	flag.BoolVar(&debug, "d", false, "Debug output")
	flag.BoolVar(&wait, "w", false, "Wait for existing deployments")

	flag.Parse()

	if user != "" && pass != "" {
		authenticate = true
	}
}

func eventBus(in <-chan RawEvent, out chan<- Event) {

	defer close(out)

	for {

		select {

		case in, ok := <-in:
			if !ok {
				break
			}

			e := new(Event)
			err := e.Unmarshal(in)

			switch {

			case err != nil && err.Error() == "Unhandled event":
				continue

			case err != nil:
				log.Println("Error parsing event:", err, in.Data)
				continue

			default:
				out <- *e

			}
		}
	}

	return
}

type Job map[string]interface{}

func NewJob(data []byte) (j Job, err error) {
	err = json.Unmarshal(data, &j)
	if err != nil {
		return
	}
	if _, ok := j["id"]; !ok {
		err = errors.New("Missing ID")
		return
	}
	switch j["id"].(type) {
	case string:
		return
	default:
		err = errors.New("Invalid JSON")
		return
	}
}

func (j Job) IsGroup() bool {
	if _, ok := j["groups"]; ok {
		return true
	}
	if _, ok := j["apps"]; ok {
		return true
	}
	return false
}

func (j Job) Id() string {
	id := j["id"]
	if id.(string)[0] == '/' {
		return id.(string)
	} else {
		return "/" + id.(string)
	}
}

func (j Job) Data() ([]byte, error) {
	return json.Marshal(&j)
}

func main() {
	if url == "" {
		log.Fatal("Marathon URL (-m) is required")
	}

	if url[0:4] != "http" {
		// default to http
		url = "http://" + url
	}

	if file == "" {
		log.Fatal("Marathon job (-f) is required")
	}

	var data []byte
	var err error

	if file == "-" {
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = ioutil.ReadFile(file)
	}
	if err != nil {
		log.Fatal(err)
	}

	job, err := NewJob(data)
	if err != nil {
		log.Fatal(err)
	}

	rawEvents := make(chan RawEvent, 64)
	events := make(chan Event, 64)

	// Exit cleanly
	defer close(rawEvents)

	err = EventListener(url, rawEvents)
	if err != nil {
		log.Fatal(err)
	}

	// Run the event bus
	go eventBus(rawEvents, events)

	// Start listening for events
	err = EventListener(url, rawEvents)
	if err != nil {
		log.Fatal(err)
	}

	// Create the deployment job
	id, err := DeployApplication(url, job)
	if err != nil {
		log.Fatal(err)
	}

	dur, err := TrackDeployment(id, events)
	if err != nil {
		log.Println("Deployment failed")
		log.Printf("%s: %6.2f %s\n", "Duration", dur.Seconds(), "seconds")
		log.Println("Reason:", err)
		os.Exit(1)
	} else {
		log.Println("Deployment succeeded")
		log.Printf("%s: %6.2f %s\n", "Duration", dur.Seconds(), "seconds")
	}
}
