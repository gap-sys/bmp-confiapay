package helpers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// CurlString contains exec.Command compatible slice + helpers
type CurlString []string

// append appends a string to the CurlCommand
func (c *CurlString) append(newSlice ...string) {
	*c = append(*c, newSlice...)
}

// String returns a ready to copy/paste command
func (c *CurlString) String() string {
	return strings.Join(*c, " ")
}

func bashEscape(str string) string {
	return `'` + strings.ReplaceAll(str, `'`, `'\''`) + `'`
}

// GetCurlCommand returns a CurlCommand corresponding to an http.Request
func GetCurlCommand(req *http.Request) (*CurlString, error) {
	if req.URL == nil {
		return nil, fmt.Errorf("getCurlCommand: invalid request, req.URL is nil")
	}

	command := CurlString{}

	command.append("curl")

	schema := req.URL.Scheme
	requestURL := req.URL.String()
	if schema == "" {
		schema = "http"
		if req.TLS != nil {
			schema = "https"
		}
		requestURL = schema + "://" + req.Host + req.URL.Path
	}

	if schema == "https" {
		command.append("-k")
	}

	command.append("-X", bashEscape(req.Method))

	if req.Body != nil {
		var buff bytes.Buffer
		_, err := buff.ReadFrom(req.Body)
		if err != nil {
			return nil, fmt.Errorf("getCurlCommand: buffer read from body error: %w", err)
		}
		// reset body for potential re-reads
		req.Body = io.NopCloser(bytes.NewBuffer(buff.Bytes()))
		bodyContent := buff.String()
		if len(bodyContent) > 0 {
			bodyEscaped := bashEscape(bodyContent)
			//bodyEscaped = strings.ReplaceAll(bodyEscaped, `\"`, `"`)
			command.append("-d", bodyEscaped)
		}
	}

	var keys []string

	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		command.append("-H", bashEscape(fmt.Sprintf("%s: %s", k, strings.Join(req.Header[k], " "))))
	}

	command.append(bashEscape(requestURL))

	command.append("--compressed")

	return &command, nil
}
