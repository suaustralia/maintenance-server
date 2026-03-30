package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var port = ":80"

type Message struct {
	Content string
}

func main() {
	// env overriding port
	if os.Getenv("SERVER_PORT") != "" {
		port = os.Getenv("SERVER_PORT")
	}

	// ensure we have an index.html file
	if _, err := os.Stat("index.html"); os.IsNotExist(err) {
		log.Fatal(err)
	}

	// load index.html it into buffer
	index, err := ioutil.ReadFile("index.html")

	if err != nil {
		log.Fatal(err)
	}

	// This handler has been reduced to almost nothing because there is very basic issue in the std lib: Go serves files
	// with the correct status code and implies it knows whats best for the developer. You can do whatever you want when
	// the file does not exist but as soon as http.ServeFile(s) or http.FileServer resolve a file and you want to rely
	// on the best proven way to serve them, you get an http status code 200.

	// This makes it impossible to rely on those functions and all there is deliver bare bone strings through a direct
	// write to the http.ResponseWriter. You could try to use w.WriteHeader before using http.ServeFile(s) or
	// http.FileServer but as documented it starts streaming to the client and might break or corrupt the headers,
	// as the std lib also handles modified times and such.

	// So you have the option to override the index.html but nothing else. It will be buffered into memory and served.
	// If you need to use images, base64 encode and embed them into the index.html.

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// filter dotted
		if strings.HasPrefix(filepath.Base(r.URL.Path), ".") {
			return
		}

		// Override status to 503 and serve index
		if isJsonRequest(r){
			w.Header().Set("Content-Type", "application/json")
			content := Message{ Content: "We are currently under maintenance. Please try again after some moments"}
			json.NewEncoder(w).Encode(content)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, err := w.Write(index)

			if err != nil {
				log.Print("Failed to serve index.html")
			}
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	log.Print(fmt.Sprintf("Listening on %s...", port))

	err = http.ListenAndServe(port, nil)
	log.Fatal(err)
}

func isJsonRequest(r *http.Request) bool  {
	return r.Header.Get("Content-Type") == "application/json" ||
		r.Header.Get("Accept") == "application/json" ||
		strings.HasSuffix(r.URL.Path, ".json")
}
