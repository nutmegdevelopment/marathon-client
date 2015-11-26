package main

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"regexp"
)

//
// HTTP sending and recieving
//

var (
	nameRexp, dataRexp *regexp.Regexp
)

func init() {
	nameRexp = regexp.MustCompile(`^event: ([[:graph:]]+)$`)
	dataRexp = regexp.MustCompile(`^data: ([[:graph:]]+)$`)
}

func EventListener(url string, ch chan<- RawEvent) (err error) {

	client := new(http.Client)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		close(ch)
		return
	}
	req.Header.Add("Accept", "text/event-stream")

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
