package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := flag.Int("port", 8080, "HTTP port")
	dir := flag.String("dir", "", "directory containing latest.html (default: ~/TechDigest)")
	flag.Parse()

	serveDir := *dir
	if serveDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("serve: home dir: %v", err)
		}
		serveDir = filepath.Join(home, "TechDigest")
	}

	serveDir, err := filepath.Abs(serveDir)
	if err != nil {
		log.Fatalf("serve: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		latest := filepath.Join(serveDir, "latest.html")
		if _, err := os.Stat(latest); err != nil {
			http.Error(w, "latest.html not found — run curator first", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, latest)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("serving %s at http://localhost%s/", serveDir, addr)
	if err := http.ListenAndServe(addr, logRequest(mux)); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/health") {
			log.Printf("%s %s", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}
