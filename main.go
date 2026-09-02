package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
)

const (
	DefaultPort = "8080"
	Usage       = `StaticM
Serve static files based on wildcard paths for mocking external servers

Options:
  -port string
    	The port to listen on (default "8080")
  -root string
    	The directory to search for documents within. (default "./documents/")

Flags:
  -watch
    	Watch the directory for file changes.
    	This will parse changes on each request, its wasteful but adequate for a tool like this.
  -v
    	Print verbose output
`
)

var (
	root    string
	port    string
	watch   bool
	verbose bool
)

func main() {
	flag.StringVar(&port, "port", DefaultPort, "The port to listen on")
	flag.BoolVar(&watch, "watch", false, "Watch the directory for file changes.\nThis will parse changes on each request, its wasteful but adequate for a tool like this.")
	flag.BoolVar(&verbose, "v", false, "Print verbose output")
	flag.Usage = func() {
		fmt.Print(Usage)
	}

	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Please provide a directory to serve documents from")
		fmt.Fprint(os.Stderr, Usage)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	registry := NewRegistry(logger, flag.Arg(0), watch)
	if err := registry.Parse(true); err != nil {
		logger.Error("failed to load docs", "error", err)
		os.Exit(1)
	}

	logger.Info("server starting", "port", port)
	http.HandleFunc("/", buildHandler(logger, registry))
	http.ListenAndServe(":"+port, nil)
}

func buildHandler(logger *slog.Logger, registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := registry.Refresh(verbose); err != nil {
			logger.Error("failed to refresh registry", "error", err)
		}

		logger.Info("processing request", "method", r.Method, "url", r.URL.String())

		var doc *Document
		var args map[string]any
		score := uint(math.MaxUint)

		for _, subject := range registry.Docs {
			match, s, d := subject.Match(r.URL, r.Method)
			if !match {
				continue
			}

			if verbose {
				logger.Info("found match", "score", s, "pattern", subject.url.String())
			}

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

		if verbose {
			logger.Info("rendering output", "pattern", doc.url.String(), "args", args, "score", score)
		}

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
