package main // import "github.com/nutmegdevelopment/marathon-client"

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var (
	rawurl, file string
	user, pass   string
	debug        bool
	authenticate bool
	force        bool
	delete       bool
)

func init() {
	flag.StringVar(&rawurl, "m", "", "Marathon URL")
	flag.StringVar(&file, "f", "", "Job file")
	flag.StringVar(&user, "u", "", "Username for basic auth")
	flag.StringVar(&pass, "p", "", "Password for basic auth")
	flag.BoolVar(&debug, "d", false, "Debug output")
	flag.BoolVar(&force, "force", false, "Force deploy over any existing deployments")
	flag.BoolVar(&delete, "delete", false, "Delete an existing application")
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
	if j["id"] == "" {
		err = errors.New("ID is empty")
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
	if rawurl == "" {
		log.Fatal("Marathon URL (-m) is required")
	}

	if rawurl[0:4] != "http" {
		// default to http
		rawurl = "http://" + rawurl
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

	err = EventListener(rawurl, rawEvents)
	if err != nil {
		log.Fatal(err)
	}

	// Run the event bus
	go eventBus(rawEvents, events)

	// Start listening for events
	err = EventListener(rawurl, rawEvents)
	if err != nil {
		log.Fatal(err)
	}

	// Create the deployment job
	id, err := DeployApplication(rawurl, job)
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
