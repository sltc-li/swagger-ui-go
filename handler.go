package swagger_ui

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	swaggerFiles "github.com/swaggo/files"
)

const (
	uiAssignment = "      window.ui = ui"
	appendScript = `
      // hide download url
      document.querySelector(".download-url-wrapper").style.display = "none"
      //const timerId = setInterval(function () {
      //  const a = document.querySelector("hgroup.main > a")
      //  if (!!a) {
      //    a.style.display = "none"
      //    clearInterval(timerId)
      //  }
      //}, 1)`
)

var (
	urlPattern = regexp.MustCompile(`(?m)^\s+url: ".+",$`)
)

func buildIndexHTML(schemaPath string) ([]byte, error) {
	indexHTML := swaggerFiles.FileIndexHTML

	results := urlPattern.FindAll(indexHTML, -1)
	if len(results) != 1 {
		return nil, fmt.Errorf("unable to find url pattern")
	}
	indexHTML = urlPattern.ReplaceAll(indexHTML, []byte(`        url: "`+schemaPath+`",`))

	if bytes.Count(indexHTML, []byte(uiAssignment)) != 1 {
		return nil, fmt.Errorf("unable to find ui assignment")
	}
	return bytes.Replace(indexHTML, []byte(uiAssignment), []byte(uiAssignment+"\n"+appendScript), 1), nil
}

func HandlerWithPath(schemaPath string, pathPrefix string) (http.Handler, error) {
	dir, base := filepath.Dir(schemaPath), filepath.Base(schemaPath)

	indexHTML, err := buildIndexHTML(pathPrefix + "/" + base)
	if err != nil {
		return nil, fmt.Errorf("build index html: %w", err)
	}

	// Disable cache
	fs := noCache(http.FileServer(http.Dir(dir)))

	return handlerWithIndexHTML(indexHTML, pathPrefix, func(path string) (http.Handler, bool) {
		if _, err := os.Stat(filepath.Join(dir, path)); err == nil {
			return fs, true
		}

		return nil, false
	}), nil
}

func HandlerWithURL(schemaURL string, pathPrefix string) (http.Handler, error) {
	indexHTML, err := buildIndexHTML(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("build index html: %w", err)
	}

	return handlerWithIndexHTML(indexHTML, pathPrefix, func(path string) (http.Handler, bool) {
		return nil, false
	}), nil
}

func handlerWithIndexHTML(indexHTML []byte, pathPrefix string, getHandler func(path string) (http.Handler, bool)) http.Handler {
	filesHandler := swaggerFiles.Handler
	filesHandler.Prefix = pathPrefix

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		path := r.URL.Path
		if pathPrefix != "" {
			path = strings.TrimPrefix(path, pathPrefix)
			if len(path) == len(r.URL.Path) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}

		if path == "/" || path == "/index.html" {
			w.Write(indexHTML)
			return
		}

		if handler, ok := getHandler(path); ok {
			r.URL.Path = path
			handler.ServeHTTP(w, r)
			return
		}

		filesHandler.ServeHTTP(w, r)
	})
}

var noCacheHeaders = map[string]string{
	"Expires":         time.Unix(0, 0).Format(time.RFC1123),
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delete any ETag headers that may have been set
		for _, v := range etagHeaders {
			if r.Header.Get(v) != "" {
				r.Header.Del(v)
			}
		}

		// Set our NoCache headers
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	})
}
