package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"log"
	//"encoding/json"
	"os"
	"time"
)

var Config struct {

}

func getSettings() (string, error) {
	resp, err := http.Get("http://localhost:9200/_cluster/settings")
	if err != nil {
		return "" , err
	}

	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)

	return string(contents), nil
}

func putSettings(template *os.File) (string, error) {
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
	return string(resp), nil
}

func main() {
	f, _ := os.Open("template")

	now := time.Now()
	tz, _ := time.Now().Zone()

	ts := fmt.Sprintf("%d-%d-%d %02d:%02d %s", 
		now.Year(), now.Month(), now.Day(), 9, 30, tz)

	target, err := time.Parse("2006-01-02 15:04 MST", ts)
	if err != nil {log.Println(err)}
	
	if now.After(target) { 
			resp, _ := putSettings(f)
			log.Printf("Pushing settings: %s", resp)
		} else {
			log.Println("No settings to push")
	}

	cSettings, _ := getSettings()
	log.Printf("Current settings: %s", cSettings)
}
