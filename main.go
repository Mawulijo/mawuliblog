package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

type SlugReader interface {
	Read(slug string) (string, error)
}

type PostDta struct {
	Title   string
	Content template.HTML
	Author  string
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
		mRenderer := goldmark.New(goldmark.WithExtensions(highlighting.NewHighlighting(highlighting.WithStyle("dracula"))))
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			// TODO: Handle different errors in the future
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		var buf bytes.Buffer
		err = mRenderer.Convert([]byte(postMarkdown), &buf)
		if err != nil {
			http.Error(w, "Error converting markdown", http.StatusInternalServerError)
			return
		}

		tpl, err := template.ParseFiles("post.html")
		if err != nil {
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}
		// TODO: Stop hardcoding post data. Parse from frontmatter.
		_ = tpl.Execute(w, PostDta{
			Title:   "My First Post",
			Content: template.HTML(buf.String()),
			Author:  "Joshua M.",
		})

		// io.Copy(w, &buf)
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
