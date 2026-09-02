package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type Registry struct {
	Docs   map[string]Document
	mux    sync.Mutex
	watch  bool
	root   string
	logger *slog.Logger
}

func NewRegistry(logger *slog.Logger, root string, watch bool) *Registry {
	return &Registry{
		Docs:   make(map[string]Document),
		watch:  watch,
		root:   root,
		logger: logger,
	}
}

func (r *Registry) Parse(verbose bool) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	parsed := make(map[string]struct{})

	err := filepath.WalkDir(r.root, func(path string, entry os.DirEntry, err error) error {
		info, err := entry.Info()
		if err != nil {
			return err
		}

		if doc, found := r.Docs[path]; found {
			if os.SameFile(doc.info, info) {
				return nil
			}
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

		d, err := NewDocument(path, contents, info)
		if err != nil {
			return err
		}

		if verbose {
			r.logger.Info("Document registered", "path", d.Path, "method", d.Method)
		}

		r.Docs[path] = *d
		parsed[path] = struct{}{}

		return nil
	})

	if err != nil {
		return err
	}

	for k, _ := range r.Docs {
		if _, found := parsed[k]; !found {
			delete(r.Docs, k)
			if verbose {
				r.logger.Info("Document removed", "path", k)
			}
		}
	}

	return nil
}

func (r *Registry) Refresh(verbose bool) error {
	if !r.watch {
		return nil
	}

	return r.Parse(verbose)
}
