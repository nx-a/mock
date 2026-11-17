package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunCallWithoutArguments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		method := request["method"]
		result := map[string]any{}
		if method == "tools/call" {
			params := request["params"].(map[string]any)
			if params["name"] != "list_kafka_servers" {
				t.Fatalf("params = %+v", params)
			}
			if arguments, ok := params["arguments"].(map[string]any); !ok || len(arguments) != 0 {
				t.Fatalf("arguments = %#v", params["arguments"])
			}
			result = map[string]any{"content": []map[string]any{{"type": "text", "text": "ok"}}}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request["id"], "result": result})
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	status := Run(context.Background(), []string{"--server", server.URL, "call", "list_kafka_servers"}, &stdout, &stderr)
	if status != 0 || stderr.Len() != 0 {
		t.Fatalf("status = %d, stderr = %s", status, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"text": "ok"`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := Run(context.Background(), []string{"call", "create_mock", "[]"}, &stdout, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), "JSON object") {
		t.Fatalf("status = %d, stderr = %s", status, stderr.String())
	}
}
