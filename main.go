package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// const ChunkSize = 64 * 1024 // 64 KB

func serveHTML(w http.ResponseWriter, r *http.Request) {
	htmlFile, err := os.ReadFile("./index.html")
	if err != nil {
		http.Error(w, "couldn't serve html file.", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(htmlFile); err != nil {
		http.Error(w, "couldn't serve html file.", http.StatusInternalServerError)
		return
	}
}

func recieveChunk(w http.ResponseWriter, r *http.Request) {
	chunk, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "couldn't read chunk data.", http.StatusInternalServerError)
		log.Println("couldn't read chunk data.", err)
		r.Body.Close()
		return
	}
	defer r.Body.Close()

	fileName, _ := url.QueryUnescape(r.Header.Get("file-name"))
	fd, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "couldn't write chunk data.", http.StatusInternalServerError)
		log.Println("couldn't write chunk data.", err)
		return
	}
	if _, err := fd.Write(chunk); err != nil {
		http.Error(w, "couldn't write chunk data.", http.StatusInternalServerError)
		log.Println("couldn't write chunk data.", err)
		return
	}
}

func main() {
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/upload", recieveChunk)

	fmt.Println("server started at http://localhost:8080...")
	http.ListenAndServe(":8080", nil)
}
