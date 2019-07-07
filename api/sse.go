package api

// From github.com/andrewstuart/go-sse

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Event represents a Server-Sent Event
type Event struct {
	Name string
	ID   string
	Data map[string]interface{}
}

// OpenURL opens a connection to a stream of server sent events
func OpenURL(user *User, url string) (events chan Event, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cookie", fmt.Sprintf("r_session=%s", user.session))

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("got response status code %d\n", resp.StatusCode)
	}

	events = make(chan Event)
	var buf bytes.Buffer

	go func() {
		ev := Event{}

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					log.Printf("%s sse disconnected\n", user.UserId)
				} else {
					log.Printf("%s sse error during read: %s\n", user.UserId, err)
				}
				close(events)
				break
			}

			switch {
			// OK line
			case bytes.HasPrefix(line, []byte(":ok")):
				// Do nothing

			// id of event
			case bytes.HasPrefix(line, []byte("id: ")):
				ev.ID = string(line[4:])
			case bytes.HasPrefix(line, []byte("id:")):
				ev.ID = string(line[3:])

			// name of event
			case bytes.HasPrefix(line, []byte("event: ")):
				ev.Name = string(line[7 : len(line)-1])
			case bytes.HasPrefix(line, []byte("event:")):
				ev.Name = string(line[6 : len(line)-1])

			// event data
			case bytes.HasPrefix(line, []byte("data: ")):
				buf.Write(line[6:])
			case bytes.HasPrefix(line, []byte("data:")):
				buf.Write(line[5:])

			// end of event
			case bytes.Equal(line, []byte("\n")):
				b := buf.Bytes()

				if bytes.HasPrefix(b, []byte("{")) {
					var data map[string]interface{}
					err := json.Unmarshal(b, &data)

					if err == nil {
						ev.Data = data
						buf.Reset()
						events <- ev
						ev = Event{}
					}
				}
			}
		}
	}()

	return events, nil
}
