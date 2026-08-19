package main

import (
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

const DocumentDir = "./documents/"

var documents map[string]Document

func main() {
	documents = make(map[string]Document)
	err := parseDocuments(DocumentDir)
	if err != nil {
		log.Fatal("failed to load docs: %s", err)
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func parseDocuments(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if _, found := documents[path]; found {
			return nil
		}

		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		d, err := NewDocument(contents)
		if err != nil {
			return err
		}

		documents[path] = *d
		return nil
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s", r.Method, r.URL.String())

	var doc *Document
	var args map[string]string
	score := uint(math.MaxUint)

	for _, subject := range documents {
		match, s, d := subject.Match(r.URL, r.Method)
		if !match || s >= score {
			continue
		}

		score = s
		args = d
		doc = &subject
	}

	if doc == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := doc.Render(args)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("FOUND: [%s] %s", doc.Method, doc.url.String())
	w.Write(data)
}
