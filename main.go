package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Request body this API expects on POST /request:
//
//	{
//	  "method":  "POST",                              // required
//	  "url":     "https://api.example.com/v1/things",  // required
//	  "headers": {"Content-Type": "application/json"}, // optional
//	  "body":    "{\"name\":\"test\"}"                 // optional, raw string
//	}
//
// The HTTP response you get back from calling POST /request IS the upstream
// response: same status code, same headers, same body. Nothing is wrapped.
type bridgeRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

var client = &http.Client{Timeout: 60 * time.Second}

// headers we never forward as-is (recalculated or not meaningful upstream)
var skipHeaders = map[string]bool{
	"content-length": true,
	"host":           true,
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		if r.Header.Get("X-Api-Key") != apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed reading request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req bridgeRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Method == "" || req.URL == "" {
		http.Error(w, `"method" and "url" are required`, http.StatusBadRequest)
		return
	}

	upstreamReq, err := http.NewRequest(
		strings.ToUpper(req.Method),
		req.URL,
		bytes.NewReader([]byte(req.Body)),
	)
	if err != nil {
		http.Error(w, "failed building upstream request: "+err.Error(), http.StatusBadRequest)
		return
	}
	for k, v := range req.Headers {
		if skipHeaders[strings.ToLower(k)] {
			continue
		}
		upstreamReq.Header.Set(k, v)
	}

	resp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// mirror the upstream response back exactly: status, headers, body
	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/request", requestHandler)

	log.Printf("bridge API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
