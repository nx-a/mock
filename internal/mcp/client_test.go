package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSessionToolsAndSSECall(t *testing.T) {
	var sessionSeen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json, text/event-stream" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		var request rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		switch request.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "session-1")
			writeRPC(t, w, request.ID, map[string]any{"protocolVersion": ProtocolVersion})
		case "notifications/initialized":
			if r.Header.Get("Mcp-Session-Id") != "session-1" {
				t.Errorf("missing session on initialized notification")
			}
			w.WriteHeader(http.StatusAccepted)
		case "tools/list":
			if r.Header.Get("Mcp-Session-Id") != "session-1" {
				t.Errorf("missing session on tools/list")
			}
			writeRPC(t, w, request.ID, map[string]any{"tools": []map[string]any{{"name": "list_kafka_servers"}}})
		case "tools/call":
			sessionSeen = true
			w.Header().Set("Content-Type", "text/event-stream")
			id, _ := json.Marshal(request.ID)
			_, _ = fmt.Fprintf(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":{\"content\":[{\"type\":\"text\",\"text\":\"ok\"}]}}\n\n", id)
		default:
			t.Fatalf("unexpected method %q", request.Method)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	if err := client.Connect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListTools(context.Background()); err != nil {
		t.Fatal(err)
	}
	result, err := client.CallTool(context.Background(), "list_kafka_servers", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if !sessionSeen || string(result) == "" {
		t.Fatalf("sessionSeen = %v, result = %s", sessionSeen, result)
	}
}

func TestClientHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client := NewClient(server.URL, server.Client())
	if err := client.Connect(context.Background()); err == nil {
		t.Fatal("expected HTTP error")
	}
}

func writeRPC(t *testing.T, w http.ResponseWriter, id any, result any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": id, "result": result}); err != nil {
		t.Fatal(err)
	}
}
