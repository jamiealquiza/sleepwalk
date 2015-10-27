package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"encoding/json"
	"os"
	"time"
	"strings"
)


type Setting struct {
	StartHH, StartMM string
	Value   *strings.Reader
}

func getSettings() (string, error) {
	resp, err := http.Get("http://localhost:9200/_cluster/settings")
	if err != nil {
		return "", err
	}

	contents, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return string(contents), nil
}

func putSettings(template *strings.Reader) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("PUT", "http://localhost:9200/_cluster/settings", template)
	if err != nil {
		return "", err
	}

	r, err := client.Do(req)
	if err != nil {
		return "", err
	}

	resp, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	return string(resp), nil
}

func parseTemplate() ([]Setting, error) {
	s := Setting{}
	settings := []Setting{}

	f, err := os.Open("template")
	if err != nil {
		return settings, err
	}

	lines := []string{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Get value (setting) from the template.
	s.Value = strings.NewReader(lines[1])
	// Get start time.
	hhmm := strings.Split(lines[0], ":")
	s.StartHH, s.StartMM = hhmm[0], hhmm[1]

	settings = append(settings, s)
	return settings, nil
}

func main() {
	settings, _ := parseTemplate()

	now := time.Now()
	tz, _ := time.Now().Zone()

	for i := range settings {
		ts := fmt.Sprintf("%d-%d-%d %s:%s %s",
			now.Year(), now.Month(), now.Day(), settings[i].StartHH, settings[i].StartMM, tz)

		target, err := time.Parse("2006-01-02 15:04 MST", ts)
		if err != nil {
			log.Println(err)
		}

		if now.After(target) {
			resp, _ := putSettings(settings[i].Value)
			log.Printf("Pushing settings: %s", resp)
			cSettings, _ := getSettings()
			log.Printf("Current settings: %s", cSettings)
		} else {
			log.Println("No settings to push")
		}
	}
}