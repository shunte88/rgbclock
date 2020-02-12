package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

//SSE name constants
const (
	idTag    = "id"
	eventTag = "event"
	dataTag  = "data"
)

type (

	// SSEvent represents a Server-Sent Event
	SSEvent struct {
		Active bool
		URI    string
		Type   string
		Name   string
		ID     string
		Data   io.Reader
	}

	// Channel data - visualization
	Channel struct {
		Name        string  `json:"name"`
		Accumulated int32   `json:"accumulated,omitempty"`
		DBfs        int32   `json:"dBfs,omitempty"`
		DB          int32   `json:"dB,omitempty"`
		Linear      int32   `json:"linear,omitempty"`
		Scaled      int32   `json:"scaled,omitempty"`
		NumFFT      int32   `json:"numFFT,omitempty"`
		FFT         []int32 `json:"FFT,omitempty"`
	}
	// Meter implementation, VU, Spectrum, RMS etc ...
	Meter struct {
		Type     string    `json:"type,omitempty"`
		Channels []Channel `json:"channel,omitempty"`
	}
)

var (
	//ErrNilChan will be returned by Notify if it is passed a nil channel
	ErrNilChan = fmt.Errorf("nil channel given")
)

// sseClient is the default client used for requests.
var sseClient = &http.Client{}

func hasPrefix(s []byte, prefix string) bool {
	return bytes.HasPrefix(s, []byte(prefix))
}

func liveReq(verb, uri string, body io.Reader) (*http.Request, error) {
	req, err := getReq(verb, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/event-stream")

	return req, nil
}

var getReq = func(verb, uri string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(verb, uri, body)
}

func ssenotify(uri string, evCh chan<- *SSEvent) error {

	if evCh == nil {
		fmt.Println(`ssenotify`, ErrNilChan)
		return ErrNilChan
	}

	var thisEvent *SSEvent

	// prime for exception status
	thisEvent = &SSEvent{URI: uri, Active: false}

	req, err := liveReq("GET", uri, nil)
	if err != nil {
		ef := fmt.Errorf("error getting sse request: %v", err)
		fmt.Printf("%v\n", ef)
		evCh <- thisEvent // deactivate consumer and channel
		return ef
	}

	res, err := sseClient.Do(req)
	if err != nil {
		ef := fmt.Errorf("error performing request for %s: %v", uri, err)
		fmt.Printf("%v\n", ef)
		evCh <- thisEvent // deactivate consumer and channel
		return ef
	}

	br := bufio.NewReader(res.Body)
	defer res.Body.Close()

	delim := []byte{':', ' '}

	for {
		bs, err := br.ReadBytes('\n')

		if err != nil && err != io.EOF {
			return err
		}

		if len(bs) < 2 {
			continue
		}

		spl := bytes.Split(bs, delim)

		if len(spl) < 2 {
			continue
		}

		thisEvent = &SSEvent{URI: uri, Active: true}
		switch string(spl[0]) {
		case eventTag:
			thisEvent.Type = string(bytes.TrimSpace(spl[1]))
		case dataTag:
			thisEvent.Data = bytes.NewBuffer(bytes.TrimSpace(spl[1]))
			evCh <- thisEvent
		}
		if err == io.EOF {
			break
		}
	}

	return nil
}
