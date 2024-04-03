package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/yuin/goldmark"
)

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct{}

func (fs FileReader) Read(slug string) (string, error) {
	f, err := os.Open(slug + ".md")
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func PostHandler(sl SlugReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			// TODO: Handle different errors in the future
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		var buf bytes.Buffer
		err = goldmark.Convert([]byte(postMarkdown), &buf)
		if err != nil {
			http.Error(w, "Error converting markdown", http.StatusInternalServerError)
			return
		}

		io.Copy(w, &buf)
	}
}

func main() {
	port := ":3000"
	mux := http.NewServeMux()
	mux.HandleFunc("GET /post/{slug}", PostHandler(FileReader{}))

	fmt.Printf("Server running on %s", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
