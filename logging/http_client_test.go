package logging

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClient_Do(t *testing.T) {
	expectedBody := "Hello, world!"
	expectedHeaders := http.Header{
		"Content-Type": []string{"text/plain"},
	}

	r := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, body)
			return
		}
		if r.Header.Get("Content-Type") != expectedHeaders.Get("Content-Type") {
			t.Errorf("expected header %q, got %q", expectedHeaders, r.Header)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot compute"))
	}))
	defer r.Close()

	buffer := &bytes.Buffer{}
	logger := &HTTPClient{
		Wrapped: r.Client(),
		Out:     buffer,
	}

	req, _ := http.NewRequest(http.MethodPost, r.URL, bytes.NewBufferString(expectedBody))
	req.Header = expectedHeaders
	res, err := logger.Do(req)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, res.StatusCode)
		return
	}

	var resBody bytes.Buffer
	_, _ = io.Copy(&resBody, res.Body)
	if resBody.String() != "Cannot compute" {
		t.Errorf("expected body %q, got %q", "Cannot compute", resBody.String())
		return
	}
}
