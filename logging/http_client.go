package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/kr/text"
)

// HTTPClient is a wrapper around http.Client that logs requests and responses.
type HTTPClient struct {
	Wrapped *http.Client
	Out     io.Writer
}

// Do performs an HTTP request and logs the request and response.
func (cli *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	interpretControlSequences := func(text string) string {
		text = strings.ReplaceAll(text, `\n`, "\n")
		text = strings.ReplaceAll(text, `\t`, "\t")
		return text
	}

	if req.Body != nil {
		reqWriter := text.NewIndentWriter(cli.Out, []byte("> "))
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("couldn't read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(data))

		dataToLog := maybeJSONFromBody(data)
		fmt.Fprintf(reqWriter, interpretControlSequences(string(dataToLog))+"\n")
	}

	resWriter := text.NewIndentWriter(cli.Out, []byte("< "))
	res, resErr := cli.Wrapped.Do(req)
	if resErr != nil {
		_, _ = fmt.Fprintf(resWriter, "error: %s\n", resErr)
		return res, resErr
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't read response body: %w", err)
	}
	res.Body = io.NopCloser(bytes.NewReader(data))

	dataToLog := maybeJSONFromBody(data)
	fmt.Fprintf(resWriter, interpretControlSequences(string(dataToLog))+"\n")
	fmt.Fprintf(cli.Out, "\n")

	return res, nil
}

func maybeJSONFromBody(data []byte) []byte {
	var value interface{}
	if err := json.Unmarshal(data, &value); err == nil {
		marshalled, err := json.MarshalIndent(value, "", "  ")
		if err == nil {
			return marshalled
		}
	}

	return data
}

// ConcurrentSafeWriter is a wrapper around an io.Writer that makes it safe to use concurrently.
type ConcurrentSafeWriter struct {
	Out   io.Writer
	mutex sync.Mutex
}

// Write writes to the underlying io.Writer, locking the mutex while doing so.
func (w *ConcurrentSafeWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.Out.Write(p)
}
