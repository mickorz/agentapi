package httpapi

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/xerrors"
)

//go:embed chat/*
var chatStaticFiles embed.FS

// This must be kept in sync with the BASE_PATH in the Makefile.
const magicBasePath = "/magic-base-path-placeholder"

func createModifiedFS(baseFS fs.FS, oldBasePath string, newBasePath string) (*afero.HttpFs, error) {
	// Use in-memory filesystem to avoid Windows path issues
	memFS := afero.NewMemMapFs()

	// Use fs.WalkDir instead of afero.Walk to avoid path separator issues
	err := fs.WalkDir(baseFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return xerrors.Errorf("failed to walk: %w", err)
		}
		if d.IsDir() {
			return nil
		}

		// Read file content from embed.FS (always uses forward slash)
		content, err := fs.ReadFile(baseFS, path)
		if err != nil {
			return xerrors.Errorf("failed to read file %s: %w", path, err)
		}

		// Replace base path in text files
		if isTextFile(path) {
			contents := string(content)
			if newBasePath == "/" {
				contents = strings.ReplaceAll(contents, oldBasePath+"/", newBasePath)
			}
			contents = strings.ReplaceAll(contents, oldBasePath, newBasePath)
			content = []byte(contents)
		}

		// Write to memory filesystem with forward slash path
		if err := afero.WriteFile(memFS, path, content, 0o644); err != nil {
			return xerrors.Errorf("failed to write file %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return nil, xerrors.Errorf("fs.WalkDir: %w", err)
	}

	return afero.NewHttpFs(memFS), nil
}

func isTextFile(name string) bool {
	ext := strings.ToLower(name)
	textExts := []string{".html", ".css", ".js", ".json", ".txt", ".xml", ".svg", ".map"}
	for _, e := range textExts {
		if strings.HasSuffix(ext, e) {
			return true
		}
	}
	return false
}

// FileServerWithIndexFallback creates a file server that serves the given filesystem
// and falls back to index.html for any path that doesn't match a file
func FileServerWithIndexFallback(chatBasePath string) http.Handler {
	subFS, err := fs.Sub(chatStaticFiles, "chat")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, fmt.Sprintf("failed to get subfs: %s", err), http.StatusInternalServerError)
		})
	}
	chatFS, err := createModifiedFS(subFS, magicBasePath, chatBasePath)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, fmt.Sprintf("failed to create modified fs: %s", err), http.StatusInternalServerError)
		})
	}
	fileServer := http.FileServer(chatFS.Dir("."))
	isChatDirEmpty := false
	if _, err := chatFS.Open("index.html"); err != nil {
		isChatDirEmpty = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isChatDirEmpty {
			http.Error(w,
				"Looks like you're running an agentapi build without the chat UI. To rebuild the binary with the UI files embedded, run `make build`.",
				http.StatusNotFound)
			return
		}
		path := r.URL.Path
		trimmedPath := strings.TrimPrefix(path, "/")
		if trimmedPath == "" {
			trimmedPath = "index.html"
		}

		// Try to serve the file directly
		f, err := chatFS.Open(trimmedPath)
		if err == nil {
			defer func() {
				_ = f.Close()
			}()
			fileServer.ServeHTTP(w, r)
			return
		}

		// If file doesn't exist, serve 404.html for any path
		if os.IsNotExist(err) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL.Path = "/404.html"
			fileServer.ServeHTTP(w, r2)
			return
		}

		// For other errors, return the error as is
		http.Error(w, fmt.Sprintf("failed to serve file: %s", err), http.StatusInternalServerError)
	})
}
