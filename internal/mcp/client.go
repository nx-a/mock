package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
)

const ProtocolVersion = "2024-11-05"

type Client struct {
	url     string
	http    *http.Client
	session string
	nextID  atomic.Int64
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(url string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{url: url, http: httpClient}
}

func (c *Client) Connect(ctx context.Context) error {
	params := map[string]any{
		"protocolVersion": ProtocolVersion,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "mock", "version": "1.0.0"},
	}
	if _, err := c.request(ctx, "initialize", params, true); err != nil {
		return fmt.Errorf("initialize MCP session: %w", err)
	}
	if _, err := c.request(ctx, "notifications/initialized", map[string]any{}, false); err != nil {
		return fmt.Errorf("confirm MCP initialization: %w", err)
	}
	return nil
}

func (c *Client) ListTools(ctx context.Context) (json.RawMessage, error) {
	return c.request(ctx, "tools/list", map[string]any{}, true)
}

func (c *Client) CallTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error) {
	params := map[string]any{"name": name, "arguments": arguments}
	return c.request(ctx, "tools/call", params, true)
}

func (c *Client) request(ctx context.Context, method string, params any, expectResult bool) (json.RawMessage, error) {
	reqBody := rpcRequest{JSONRPC: "2.0", Method: method, Params: params}
	if expectResult {
		reqBody.ID = c.nextID.Add(1)
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", ProtocolVersion)
	if c.session != "" {
		req.Header.Set("Mcp-Session-Id", c.session)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if session := resp.Header.Get("Mcp-Session-Id"); session != "" {
		c.session = session
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if !expectResult {
		return nil, nil
	}
	var rpcResp rpcResponse
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		rpcResp, err = decodeSSE(resp.Body)
	} else {
		err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	}
	if err != nil {
		return nil, fmt.Errorf("decode MCP response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	if len(rpcResp.Result) == 0 {
		return nil, fmt.Errorf("empty MCP result")
	}
	return rpcResp.Result, nil
}

func decodeSSE(reader io.Reader) (rpcResponse, error) {
	var result rpcResponse
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64<<10), 2<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return result, err
		}
		return result, nil
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}
	return result, io.EOF
}
