package main

import (
	"bufio"
	"fmt"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	//"encoding/json"
	"os"
	"time"
	"strings"
)

var SleepwalkSettings struct {
	address string
}

// An ElasticSearch cluster setting and timestamp describing
// a start time for the setting to go into effect.
type Setting struct {
	StartHH, StartMM, EndHH, EndMM string
	Value   *strings.Reader
}

func init() {
	flag.StringVar(&SleepwalkSettings.address, "address", "http://localhost:9200", "ElasticSearch Address")
	flag.Parse()
}

// getSettings fetches the current ElasticSearch cluster settings.
func getSettings() (string, error) {
	resp, err := http.Get(SleepwalkSettings.address+"/_cluster/settings")
	if err != nil {
		return "", err
	}

	contents, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return string(contents), nil
}

// putSettings pushes a cluster setting to ElasticSearch.
func putSettings(setting *strings.Reader) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("PUT", SleepwalkSettings.address+"/_cluster/settings", setting)
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

// parseTsRange takes a 09:30-15:30 format start / end time range
// and reteurns the start HH, start MM, end HH, end MM elements.
func parseTsRange(tsrange string) (string, string, string, string) {
	// Break start / stop times.
	r := strings.Split(tsrange, "-")
	// Get start elements.
	start := strings.Split(r[0], ":")
	// Get end elements.
	end := strings.Split(r[1], ":")

	return start[0], start[1], end[0], end[1]
}

// parseTemplate reads a Sleepwalk settings template and returns an array of
// Setting structs.
func parseTemplate(template string) ([]Setting, error) {
	settings := []Setting{}

	f, err := os.Open(template)
	if err != nil {
		return settings, err
	}

	lines := []string{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// No safeties yet. Assumes that the template is a perfectly formatted
	// time range and associated setting in alternating lines.
	for i := 0; i < len(lines); i = i+2 {
		s := Setting{}
		// Get value (the setting) from the template.
		s.Value = strings.NewReader(lines[i+1])
		// Get time range.
		s.StartHH, s.StartMM, s.EndHH, s.EndMM = parseTsRange(lines[i])

		settings = append(settings, s)
	}

	return settings, nil
}

// getTs takes HH:MM pairs and a reference timestamp (for current date-time and zone)
// and returns a formatted time.Time stamp.
func getTs(hh, mm string, ref time.Time) time.Time {
	tz, _ := time.Now().Zone()
	tsString := fmt.Sprintf("%d-%d-%d %s:%s %s",
		ref.Year(), ref.Month(), ref.Day(), hh, mm, tz)

	ts, err := time.Parse("2006-01-02 15:04 MST", tsString)
	if err != nil {
		log.Println(err)
	}

	return ts
}

// applyTemplate parses a template file and applies each setting.
func applyTemplate() {
	settings, _ := parseTemplate("template")
	now := time.Now()

	for i := range settings {
		start := getTs(settings[i].StartHH, settings[i].StartMM, now)
		end := getTs(settings[i].EndHH, settings[i].EndMM, now)

		if now.After(start) &&  now.Before (end) {
			cSettings, _ := getSettings()
			log.Printf("Pushing settings. Current settings: %s\n", cSettings)
			_, _ = putSettings(settings[i].Value)
			nSettings, _ := getSettings()
			if cSettings != nSettings {
				log.Printf("New settings: %s", nSettings)
			}
		}
	}
}

func main() {
	log.Println("Sleepwalk Running")
	applyTemplate()
	run := time.Tick(15 * time.Second)
	for _ = range run {
		applyTemplate()
	}
}