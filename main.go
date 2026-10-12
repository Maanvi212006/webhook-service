package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const secret = "mysecret"

type Event struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	User   string `json:"user"`
}

var seen = map[string]bool{}
var mu sync.Mutex

func isValid(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func postToDiscord(text string) error {
	url := os.Getenv("DISCORD_URL")
	payload, _ := json.Marshal(map[string]string{"content": text})

	var err error
	wait := 1 * time.Second
	for attempt := 1; attempt <= 3; attempt++ {
		var resp *http.Response
		resp, err = http.Post(url, "application/json", bytes.NewReader(payload))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 300 {
				return nil
			}
			err = fmt.Errorf("discord returned status %d", resp.StatusCode)
		}
		fmt.Println("Attempt", attempt, "failed:", err)
		if attempt < 3 {
			time.Sleep(wait)
			wait *= 2
		}
	}
	return err
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	signature := r.Header.Get("X-Hub-Signature-256")
	if !isValid(body, signature) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var event Event
	err := json.Unmarshal(body, &event)
	if err != nil {
		http.Error(w, "bad JSON", http.StatusBadRequest)
		return
	}

		mu.Lock()
	if seen[event.ID] {
		mu.Unlock()
		fmt.Println("Duplicate, skipping:", event.ID)
		w.Write([]byte("OK"))
		return
	}
	seen[event.ID] = true
	mu.Unlock()

	msg := event.User + " did: " + event.Action
		err = postToDiscord(msg)
	if err != nil {
		fmt.Println("Gave up:", err)
	} else {
		fmt.Println("Sent:", msg)
	}
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
