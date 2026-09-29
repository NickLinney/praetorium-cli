package coordinator

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExecutePostsToCoordinatorExecutePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/execute" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"task":"test"}` {
			t.Fatalf("unexpected request body: %s", body)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"succeeded"}`))
	}))
	defer server.Close()

	body, status, err := (Client{Endpoint: server.URL, Timeout: time.Second}).Execute(context.Background(), []byte(`{"task":"test"}`))
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusAccepted || !strings.Contains(string(body), "succeeded") {
		t.Fatalf("unexpected response: %d %s", status, body)
	}
}

func TestExecuteRejectsInvalidEndpoint(t *testing.T) {
	if _, _, err := (Client{Endpoint: "ftp://example.test", Timeout: time.Second}).Execute(context.Background(), nil); err == nil {
		t.Fatal("expected invalid endpoint to fail")
	}
}
