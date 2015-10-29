package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var SleepwalkSettings struct {
	address   string
	interval  int
	templates string
}

var (
	// Template naming scheme.
	templateFileName  *regexp.Regexp = regexp.MustCompile(".conf$")
	templateDateRange *regexp.Regexp = regexp.MustCompile("[0-9]{2}")
)

// An ElasticSearch cluster setting and timestamp describing
// a start time for the setting to go into effect.
type Setting struct {
	StartHH, StartMM, EndHH, EndMM string
	Value                          *strings.Reader
}

func init() {
	flag.StringVar(&SleepwalkSettings.address, "address", "http://localhost:9200", "ElasticSearch Address")
	flag.IntVar(&SleepwalkSettings.interval, "interval", 300, "Update interval in seconds")
	flag.StringVar(&SleepwalkSettings.templates, "templates", "./templates", "Template path")
	flag.Parse()
}

// getSettings fetches the current ElasticSearch cluster settings.
func getSettings() (string, error) {
	resp, err := http.Get(SleepwalkSettings.address + "/_cluster/settings")
	if err != nil {
		return "", fmt.Errorf("Error getting settings: %s", err)
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
		return "", fmt.Errorf("Request error: %s", err)
	}

	r, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error pushing settings: %s", err)
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

func validateSetting(setting Setting, i int) (int, bool) {
	// Validate start/end HH/MM.
	// Needs to do something smarter than just a /[0-9]{2}/ match.
	switch {
	case !templateDateRange.MatchString(setting.StartHH):
		return i + 1, false
	case !templateDateRange.MatchString(setting.StartMM):
		return i + 1, false
	case !templateDateRange.MatchString(setting.EndHH):
		return i + 1, false
	case !templateDateRange.MatchString(setting.EndMM):
		return i + 1, false
	}

	// Validate setting.Value as at least being valid JSON.
	settingValue, _ := ioutil.ReadAll(setting.Value)
	js := make(map[string]interface{})
	if err := json.Unmarshal(settingValue, &js); err != nil {
		return i + 2, false
	}

	return i, true
}

// parseTemplate reads a Sleepwalk settings template and returns an array of
// Setting structs.
func parseTemplate(template string) ([]Setting, error) {
	settings := []Setting{}

	f, err := os.Open(SleepwalkSettings.templates + "/" + template)
	if err != nil {
		return settings, fmt.Errorf("Template error: %s", err)
	}

	lines := []string{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// No safeties yet. Assumes that the template is a perfectly formatted
	// time range and associated setting in alternating lines.
	for i := 0; i < len(lines); i = i + 2 {
		s := Setting{}
		// Get value (the setting) from the template.
		s.Value = strings.NewReader(lines[i+1])
		// Get time range.
		s.StartHH, s.StartMM, s.EndHH, s.EndMM = parseTsRange(lines[i])

		// Validate setting. We have to pass the index i we are on to
		// determine the line the failed validation.
		if line, valid := validateSetting(s, i); valid {
			settings = append(settings, s)
		} else {
			return settings, fmt.Errorf("Template parsing error from %s:%d",
				template, line)
		}
	}

	return settings, nil
}

// getTs takes HH:MM pairs and a reference timestamp (for current date-time and zone)
// and returns a formatted time.Time stamp.
func getTs(hh, mm string, ref time.Time) (time.Time, error) {
	tz, _ := time.Now().Zone()
	tsString := fmt.Sprintf("%d-%d-%d %s:%s %s",
		ref.Year(), ref.Month(), ref.Day(), hh, mm, tz)

	ts, err := time.Parse("2006-01-02 15:04 MST", tsString)
	if err != nil {
		return ts, err
	}

	return ts, nil
}

// applyTemplate parses a template file and applies each setting.
func applyTemplate(template string) {
	log.Printf("Reading template: %s\n", template)

	settings, err := parseTemplate(template)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	now := time.Now()
	// Count of how many settings
	// were applied from this template.
	applied := 0

	// We have a bunch of Setting structs from our template.
	for i := range settings {
		start, _ := getTs(settings[i].StartHH, settings[i].StartMM, now)
		end, _ := getTs(settings[i].EndHH, settings[i].EndMM, now)

		if now.After(start) && now.Before(end) {
			cSettings, err := getSettings()
			if err != nil {
				log.Println(err)
			}

			log.Printf("Pushing setting from template: %s\n", template)

			_, err = putSettings(settings[i].Value)
			if err != nil {
				log.Println(err)
			}
			applied++

			nSettings, err := getSettings()
			if err != nil {
				log.Println(err)
			}

			if cSettings != nSettings {
				log.Printf("Settings changed from %s to %s\n", cSettings, nSettings)
			} else {
				log.Printf("No settings changed")
			}
		}
	}

	if applied < 1 {
		log.Printf("No settings to apply from: %s\n", template)
	}
}

// getTemplates returns a list of template files
// from the template path.
func getTemplates(path string) []string {
	templates := []string{}
	fs, _ := ioutil.ReadDir(path)

	for _, f := range fs {
		if templateFileName.MatchString(f.Name()) {
			templates = append(templates, f.Name())
		}
	}

	return templates
}

func main() {
	log.Println("Sleepwalk Running")

	templates := getTemplates("./templates")
	for _, t := range templates {
		applyTemplate(t)
	}

	run := time.Tick(time.Duration(SleepwalkSettings.interval) * time.Second)
	for _ = range run {
		for _, t := range templates {
			applyTemplate(t)
		}
	}
}
