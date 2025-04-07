package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
	"io/ioutil"
)

type URLMap map[string]string

const storageFile = "urls.json"

func loadURLMap() URLMap {
	data, err := ioutil.ReadFile(storageFile)
	if err != nil {
		return make(URLMap)
	}
	var urls URLMap
	json.Unmarshal(data, &urls)
	return urls
}

func saveURLMap(urls URLMap) {
	data, _ := json.MarshalIndent(urls, "", "  ")
	ioutil.WriteFile(storageFile, data, 0644)
}

func generateShortCode(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		URL string `json:"url"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.URL == "" {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	urls := loadURLMap()
	code := generateShortCode(6)
	urls[code] = body.URL
	saveURLMap(urls)

	resp := map[string]string{"short_url": fmt.Sprintf("https://%s/%s", r.Host, code)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	if code == "" {
		http.ServeFile(w, r, "./frontend/index.html")
		return
	}
	urls := loadURLMap()
	longURL, ok := urls[code]
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

func main() {
	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/shorten", shortenHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
