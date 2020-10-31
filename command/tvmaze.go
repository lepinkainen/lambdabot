package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/lepinkainen/lambdabot/lambda"

	log "github.com/sirupsen/logrus"
)

// TVMazeResponse asd
type TVMazeResponse struct {
	ID           int         `json:"id"`
	URL          string      `json:"url"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Language     string      `json:"language"`
	Genres       []string    `json:"genres"`
	Status       string      `json:"status"`
	Runtime      interface{} `json:"runtime"`
	Premiered    string      `json:"premiered"`
	OfficialSite string      `json:"officialSite"`
	Schedule     Schedule    `json:"schedule"`
	Weight       int         `json:"weight"`
	Network      WebChannel  `json:"network"`
	WebChannel   WebChannel  `json:"webChannel"`
	Summary      string      `json:"summary"`
	Updated      int         `json:"updated"`
	Embedded     Embedded    `json:"_embedded"`
}

// Schedule - when does the show air?
type Schedule struct {
	Time string   `json:"time"`
	Days []string `json:"days"`
}

// WebChannel - the network the show is on, pretty much
type WebChannel struct {
	ID      int         `json:"id"`
	Name    string      `json:"name"`
	Country interface{} `json:"country"`
}

// Episodes - all episodes for the show
type Episodes struct {
	ID       int       `json:"id"`
	URL      string    `json:"url"`
	Name     string    `json:"name"`
	Season   int       `json:"season"`
	Number   int       `json:"number"`
	Type     string    `json:"type"`
	Airdate  string    `json:"airdate"`
	Airtime  string    `json:"airtime"`
	Airstamp time.Time `json:"airstamp"`
	Runtime  int       `json:"runtime"`
	Summary  string    `json:"summary"`
}

// Nextepisode - info about the upcoming episode
type Nextepisode struct {
	ID       int         `json:"id"`
	URL      string      `json:"url"`
	Name     string      `json:"name"`
	Season   int         `json:"season"`
	Number   int         `json:"number"`
	Type     string      `json:"type"`
	Airdate  string      `json:"airdate"`
	Airtime  string      `json:"airtime"`
	Airstamp time.Time   `json:"airstamp"`
	Runtime  interface{} `json:"runtime"`
	Summary  interface{} `json:"summary"`
}

// Embedded extras, episodes and next episode data
type Embedded struct {
	Episodes    []Episodes  `json:"episodes"`
	Nextepisode Nextepisode `json:"nextepisode"`
}

func parseResponse(bytes []byte) (TVMazeResponse, error) {
	data := TVMazeResponse{}

	err := json.Unmarshal(bytes, &data)
	if err != nil {
		log.Errorf("Unable to unmarshal result JSON: %v", err)
		return data, err
	}

	return data, nil
}

func nextEpResponse(data TVMazeResponse) string {
	// Next episode of The Mandalorian 2x02 'Chapter 10: The Confrontation' airs 2020-11-06 (5 days) on Disney+
	seriesname := data.Name

	sxep := fmt.Sprintf("%dx%02d", data.Embedded.Nextepisode.Season, data.Embedded.Nextepisode.Number)
	epname := data.Embedded.Nextepisode.Name
	airdate := data.Embedded.Nextepisode.Airdate

	network := ""
	if data.WebChannel == (WebChannel{}) {
		network = data.Network.Name
	} else {
		network = data.WebChannel.Name
	}

	return fmt.Sprintf("Next episode of %s %s '%s' airs %s on %s", seriesname, sxep, epname, airdate, network)

}

func latestEpResponse(data TVMazeResponse) string {
	lastEp := data.Embedded.Episodes[len(data.Embedded.Episodes)-1]

	// Next episode of The Mandalorian 2x02 'Chapter 10: The Confrontation' airs 2020-11-06 (5 days) on Disney+
	seriesname := data.Name

	//fmt.Printf("%v\n", lastEp)

	sxep := fmt.Sprintf("%dx%02d", lastEp.Season, lastEp.Number)
	epname := lastEp.Name
	airdate := lastEp.Airdate
	network := ""
	if data.WebChannel == (WebChannel{}) {
		network = data.Network.Name
	} else {
		network = data.WebChannel.Name
	}

	// Show has ended for some reason
	status := ""
	if data.Status == "Ended" {
		status = " [Ended]"
	}

	// No "next episode" data, but there is a latest episode in a non-eded show ->
	// We know there will be more, we just don't know _when_
	if airdate == "" {
		airdate = "[UNKNOWN]"
	}

	return fmt.Sprintf("Latest episode of %s %s '%s' airs %s on %s%s", seriesname, sxep, epname, airdate, network, status)
}

// TVMaze search for tvmaze and list next episode in series
func TVMaze(args string) (string, error) {
	query := url.QueryEscape(args)
	apiurl := fmt.Sprintf("http://api.tvmaze.com/singlesearch/shows?q=%s&embed[]=episodes&embed[]=nextepisode", query)

	res, err := http.Get(apiurl)
	if err != nil {
		log.Errorf("Unable to get API response from TVMaze: %v", err)
		return "", errors.Wrap(err, "Unable to get API response")
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Unable to read response from TVMaze: %v", err)
		return "", errors.Wrap(err, "Unable to read response")
	}

	response, err := parseResponse(bytes)
	if err != nil {
		log.Errorf("Could not parse TVMaze response")
		return "", err
	}

	// Show has known next episode, hasn't ended
	if !(response.Embedded.Nextepisode == (Nextepisode{})) {
		return nextEpResponse(response), nil
	}

	return latestEpResponse(response), nil
}

func init() {
	lambda.RegisterHandler("ep", TVMaze)
}
