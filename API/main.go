package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"path"
)

const scraperPath = "/home/ubuntu/skybox_scraper"

type APIError struct {
	Message string `json:"message"`
}

func startTCGScraper(w http.ResponseWriter, r *http.Request) {
	outputBytes, err := exec.Command("python3", path.Join(scraperPath, "TCG_Scraper_Ver2.0.py")).Output()
	if err != nil {
		respondWithError(w, r, 500, string(outputBytes))
		return
	}
	respondWithJSON(w, r, 200, string(outputBytes))
}

func startSportsScraper(w http.ResponseWriter, r *http.Request) {
	outputBytes, err := exec.Command("python3", path.Join(scraperPath, "TCG_Scraper_Ver2.0.py")).Output()
	if err != nil {
		respondWithError(w, r, 500, string(outputBytes))
		return
	}
	respondWithJSON(w, r, 200, string(outputBytes))
}

func startTestScraper(w http.ResponseWriter, r *http.Request) {
	outputBytes, err := exec.Command("python3", path.Join(scraperPath, "test.py")).Output()
	if err != nil {
		respondWithError(w, r, 500, string(outputBytes))
		return
	}
	respondWithJSON(w, r, 200, string(outputBytes))
}

func handleRequests() {
	http.HandleFunc("/scrapers/tcg/start", startTCGScraper)
	http.HandleFunc("/scrapers/sports/start", startSportsScraper)
	http.HandleFunc("/scrapers/test/start", startTestScraper)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Println(payload, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		log.Println(payload, err)
		return err
	}

	return nil
}

func respondWithError(w http.ResponseWriter, r *http.Request, code int, message string) {
	respondWithJSON(w, r, code, APIError{Message: message})
}

func main() {
	handleRequests()
}
