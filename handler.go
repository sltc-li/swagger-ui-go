package swagger_ui

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"

	swaggerFiles "github.com/swaggo/files"
)

const (
	schemaPath     = "/schema.json"
	urlReplacement = `        url: "` + schemaPath + `",`
	uiAssignment   = "      window.ui = ui"
	appendScript   = `
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

func buildIndexHTML() ([]byte, error) {
	indexHTML := swaggerFiles.FileIndexHTML

	results := urlPattern.FindAll(indexHTML, -1)
	if len(results) != 1 {
		return nil, fmt.Errorf("unable to find url pattern")
	}
	indexHTML = urlPattern.ReplaceAll(indexHTML, []byte(urlReplacement))

	if bytes.Count(indexHTML, []byte(uiAssignment)) != 1 {
		return nil, fmt.Errorf("unable to find ui assignment")
	}
	return bytes.Replace(indexHTML, []byte(uiAssignment), []byte(uiAssignment+"\n"+appendScript), 1), nil
}

func New(schema []byte) (http.Handler, error) {
	indexHTML, err := buildIndexHTML()
	if err != nil {
		return nil, fmt.Errorf("build index html: %w", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch r.URL.Path {
		case "/", "/index.html":
			w.Write(indexHTML)
		case schemaPath:
			w.Write(schema)
		default:
			swaggerFiles.Handler.ServeHTTP(w, r)
		}
	}), nil
}
