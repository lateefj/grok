package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func serve() {
	// Makrdown suffix suffixes
	mdSuffixes := []string{".markdown", ".mdown", ".mkdn", ".md", ".mkd", ".mdwn", ".mdtxt", ".mdtext", ".text"}

	var port int
	flag.IntVar(&port, "port", 7070, "Port number for http server")
	flag.Parse()
	var defaultFile = flag.Arg(0)
	if defaultFile == "" {
		defaultFile = "README.md"
	}

	r := mux.NewRouter()

	r.HandleFunc("/_grok/img/{imageId}.png", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fmt.Printf("ImageId: %v\n", vars["imageId"])
		imgId, err := strconv.Atoi(vars["imageId"])
		if err != nil {
			log.Printf("ERROR: %s\n", err)
			http.Error(w, "Error with image id", http.StatusBadRequest)
			return
		}

		if imgId > len(imageCache)-1 {
			log.Printf("ERROR imgId is %d but cache is only %d\n", imgId, len(imageCache))
			http.Error(w, "Error trying to find image", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(imageCache[imgId])
	})
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path[1:]
		var path string
		_, err := os.Stat(urlPath)
		if os.IsNotExist(err) {
			_, errDefault := os.Stat(defaultFile)
			if os.IsNotExist(errDefault) {
				_, errIndex := os.Stat("index.html")
				if os.IsNotExist(errIndex) {
					http.Error(w, fmt.Sprintf("Failed to find file to read tried README.md and index.html .... %s", path), http.StatusInternalServerError)
					return
				} else {
					path = "index.html"
				}
			} else {
				path = defaultFile
			}
		} else {
			path = urlPath
		}

		md := false
		for _, s := range mdSuffixes {
			if strings.HasSuffix(path, s) {
				md = true
				break
			}
		}
		if md {
			input, err := os.Open(path) // For read access.
			if err != nil {
				log.Fatal(err)
			}
			mdoc, err := parseMarkdown(input)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to processed %s", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(mdoc)
			//output := blackfriday.Run(b)
			//w.Write(output)
			return
		} else if strings.HasSuffix(r.URL.Path, ".json") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		} else if strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.ServeFile(w, r, path)
	})
	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%d", port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
