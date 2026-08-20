package main

import (
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

const DocumentDir = "./documents/"

var documents map[string]Document

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	documents = make(map[string]Document)
	err := parseDocuments(DocumentDir)
	if err != nil {
		logger.Error("failed to load docs", "error", err)
		os.Exit(1)
	}

	http.HandleFunc("/", buildHandler(logger))
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

		d, err := NewDocument(path, contents)
		if err != nil {
			return err
		}

		documents[path] = *d
		return nil
	})
}

func buildHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("processing request", "method", r.Method, "url", r.URL.String())

		var doc *Document
		var args map[string]string
		score := uint(math.MaxUint)

		for _, subject := range documents {
			match, s, d := subject.Match(r.URL, r.Method)
			if !match {
				continue
			}

			logger.Info("found match", "score", s, "pattern", subject.url.String())
			if s >= score {
				continue
			}

			score = s
			args = d
			doc = &subject
		}

		if doc == nil {
			logger.Error("no match found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		data, err := doc.Render(args)
		if err != nil {
			logger.Error("render failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if doc.Mime != "" {
			w.Header().Add("Content-Type", doc.Mime)
		}

		logger.Info("rendering output", "pattern", doc.url.String(), "args", args, "score", score)
		if doc.Response != nil {
			for k, v := range doc.Response.Headers {
				w.Header().Add(k, v)
			}

			if doc.Response.Status != 0 {
				w.WriteHeader(doc.Response.Status)
			}
		}

		w.Write(data)
	}
}
