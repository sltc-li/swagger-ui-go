package swagger_ui

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	swaggerFiles "github.com/swaggo/files"
)

const (
	schemaPath   = "/schema.json"
	uiAssignment = "      window.ui = ui"
	appendScript = `
      // hide download url
      document.querySelector(".download-url-wrapper").style.display = "none"
      const timerId = setInterval(function () {
        const a = document.querySelector("hgroup.main > a")
        if (!!a) {
          a.style.display = "none"
          clearInterval(timerId)
        }
      }, 1)`
)

var (
	urlPattern = regexp.MustCompile(`(?m)^\s+url: ".+",$`)
)

func buildIndexHTML(pathPrefix string) ([]byte, error) {
	indexHTML := swaggerFiles.FileIndexHTML

	results := urlPattern.FindAll(indexHTML, -1)
	if len(results) != 1 {
		return nil, fmt.Errorf("unable to find url pattern")
	}
	indexHTML = urlPattern.ReplaceAll(indexHTML, []byte(`        url: "`+pathPrefix+schemaPath+`",`))

	if bytes.Count(indexHTML, []byte(uiAssignment)) != 1 {
		return nil, fmt.Errorf("unable to find ui assignment")
	}
	return bytes.Replace(indexHTML, []byte(uiAssignment), []byte(uiAssignment+"\n"+appendScript), 1), nil
}

func New(schema []byte, pathPrefix string) (http.Handler, error) {
	indexHTML, err := buildIndexHTML(pathPrefix)
	if err != nil {
		return nil, fmt.Errorf("build index html: %w", err)
	}

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

		switch path {
		case "/", "/index.html":
			w.Write(indexHTML)
		case schemaPath:
			w.Write(schema)
		default:
			filesHandler.ServeHTTP(w, r)
		}
	}), nil
}
