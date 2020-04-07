package jogly

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// Jog - jogly payload to send info into slack
type Jog struct {
	endpoint string
	Payload  interface{}
}

// New returns a new jogly request object
func New(endpoint string) *Jog {
	return &Jog{
		endpoint: endpoint,
	}
}

// Serialise will set a new payload
func (j *Jog) Serialise(i interface{}) *Jog {
	j.Payload = i
	return j
}

// Post will serialise the payload and send it
func (j *Jog) Post() error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(j)
	if err != nil {
		return err
	}
	if j.endpoint == "" {
		j.log()
		return nil
	}
	resp, err := http.Post(j.endpoint, "application/json", buf)
	if err != nil {
		log.Println("Couldn't serialise data", j.Payload)
		return err
	}
	defer resp.Body.Close()
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Println(string(data), err)
	}

	return err
}

func (j *Jog) log() {
	log.Println("---------------")
	log.Println("Jogly: No endpoint provided, logging message instead")
	log.Println(j.Payload)
	log.Println("---------------")
}
