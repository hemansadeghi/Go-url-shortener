package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func generateShortLink(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func ensureJSONExists(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		f, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Error creating file: %v", err)
		}
		f.Write([]byte("{}"))
		f.Close()
	}
}

func loadURLs() map[string]string {
	data, err := os.ReadFile("urls.json")
	if err != nil {
		return make(map[string]string)
	}
	var urls map[string]string
	json.Unmarshal(data, &urls)
	return urls
}

func saveURLs(urls map[string]string) {
	data, _ := json.MarshalIndent(urls, "", "  ")
	os.WriteFile("urls.json", data, 0644)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	type Request struct {
		URL string `json:"url"`
	}

	type Response struct {
		ShortURL string `json:"short_url"`
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortCode := generateShortLink(6)

	urls := loadURLs()
	urls[shortCode] = req.URL
	saveURLs(urls)

	resp := Response{
		ShortURL: fmt.Sprintf("https://%s/%s", r.Host, shortCode),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	if code == "" {
		http.NotFound(w, r)
		return
	}

	urls := loadURLs()
	if original, ok := urls[code]; ok {
		http.Redirect(w, r, original, http.StatusFound)
		return
	}

	http.NotFound(w, r)
}

func main() {
	ensureJSONExists("urls.json")

	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler) // این مسیر مهمه

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running at http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
