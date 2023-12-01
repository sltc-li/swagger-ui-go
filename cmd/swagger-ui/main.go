package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	swagger_ui "github.com/sltc-li/swagger-ui-go"
)

const usage = `Usage:
swagger-ui ./schema.{json,yaml}
swagger-ui https://example.com/schema.{json,yaml}
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}

	port := flag.Int("port", 8080, "port to host")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	handler, err := getHandler(flag.Arg(0), "/swagger-ui")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/swagger-ui/", handler)

	log.Printf("listening :%d", *port)
	log.Printf("open http://localhost:%d/swagger-ui/ in browser", *port)
	if err := http.ListenAndServe(":"+strconv.Itoa(*port), mux); err != nil {
		log.Fatal(err)
	}
}

func getHandler(path string, prefix string) (http.Handler, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return swagger_ui.HandlerWithURL(path, prefix)
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to get user's home directory: %w", err)
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return swagger_ui.HandlerWithPath(path, prefix)
}
