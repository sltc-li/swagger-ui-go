package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/li-go/swagger-ui-go"
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

	handler, err := swagger_ui.New(func() ([]byte, error) {
		return readSchemaAsJSON(flag.Arg(0))
	}, "")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening :%d", *port)
	log.Printf("open http://localhost:%d/ in browser", *port)
	if err := http.ListenAndServe(":"+strconv.Itoa(*port), handler); err != nil {
		log.Fatal(err)
	}
}

func readSchemaAsJSON(path string) ([]byte, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("get remote file: %w", err)
		}
		defer resp.Body.Close()
		return readJSON(resp.Body)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return readJSON(file)
}

func readJSON(reader io.Reader) ([]byte, error) {
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}

	var v interface{}
	// valid json
	if err := json.Unmarshal(buf, &v); err == nil {
		return buf, nil
	}

	// valid yaml
	if buf, err := yaml.YAMLToJSON(buf); err == nil {
		return buf, nil
	}

	return nil, fmt.Errorf("unsupported media type")
}
