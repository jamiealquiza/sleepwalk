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
	StartHH, StartMM, EndHH, EndMM string
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

func parseTsRange(tsrange string) (string, string, string, string) {
	// Get 01:00 - 02:00 timestamp range.
	r := strings.Split(tsrange, "-")
	// Get start elements.
	start := strings.Split(r[0], ":")
	// Get end elements.
	end := strings.Split(r[1], ":")

	return start[0], start[1], end[0], end[1]
}

func parseTemplate() (Setting, error) {
	s := Setting{}

	f, err := os.Open("template")
	if err != nil {
		return s, err
	}

	lines := []string{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Get value (setting) from the template.
	s.Value = strings.NewReader(lines[1])
	s.StartHH, s.StartMM, s.EndHH, s.EndMM = parseTsRange(lines[0])

	return s, nil
}

func main() {
	f, _ := parseTemplate()

	now := time.Now()
	tz, _ := time.Now().Zone()

	ts := fmt.Sprintf("%d-%d-%d %s:%s %s",
		now.Year(), now.Month(), now.Day(), f.StartHH, f.StartMM, tz)

	target, err := time.Parse("2006-01-02 15:04 MST", ts)
	if err != nil {
		log.Println(err)
	}

	if now.After(target) {
		resp, _ := putSettings(f.Value)
		log.Printf("Pushing settings: %s", resp)
		cSettings, _ := getSettings()
		log.Printf("Current settings: %s", cSettings)
	} else {
		log.Println("No settings to push")
	}

}
