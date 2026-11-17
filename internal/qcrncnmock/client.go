package qcrncnmock

import (
	"context"
	"encoding/json"

	"github.com/nx-a/mock/internal/mcp"
)

type Client struct {
	mcp *mcp.Client
}

func NewClient(mcpClient *mcp.Client) *Client {
	return &Client{mcp: mcpClient}
}

func (c *Client) Connect(ctx context.Context) error {
	return c.mcp.Connect(ctx)
}

func (c *Client) Tools(ctx context.Context) (json.RawMessage, error) {
	return c.mcp.ListTools(ctx)
}

func (c *Client) Call(ctx context.Context, tool string, arguments json.RawMessage) (json.RawMessage, error) {
	return c.mcp.CallTool(ctx, tool, arguments)
}
