package app

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nx-a/mock/internal/mcp"
	"github.com/nx-a/mock/internal/qcrncnmock"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("mock", flag.ContinueOnError)
	flags.SetOutput(stderr)
	defaultURL := os.Getenv("QCRNCNMOCK_URL")
	if defaultURL == "" {
		defaultURL = "http://localhost:8080/mcp"
	}
	serverURL := flags.String("server", defaultURL, "qcrncnmock MCP endpoint URL")
	insecure := flags.Bool("insecure", false, "skip TLS certificate verification")
	timeout := flags.Duration("timeout", 60*time.Second, "request timeout")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *timeout <= 0 {
		fmt.Fprintln(stderr, "mock: timeout must be positive")
		return 2
	}
	remaining := flags.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(stderr, "mock: command required: tools or call")
		return 2
	}
	if remaining[0] == "call" {
		if len(remaining) < 2 || len(remaining) > 3 {
			fmt.Fprintln(stderr, "mock: usage: mock [flags] call <tool> [arguments-json]")
			return 2
		}
		if len(remaining) == 3 && (!json.Valid(json.RawMessage(remaining[2])) || !strings.HasPrefix(strings.TrimSpace(remaining[2]), "{")) {
			fmt.Fprintln(stderr, "mock: arguments must be a JSON object")
			return 2
		}
	}
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if *insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicitly requested by the user
	}
	httpClient := &http.Client{Transport: transport, Timeout: *timeout}
	client := qcrncnmock.NewClient(mcp.NewClient(*serverURL, httpClient))
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintln(stderr, "mock:", err)
		return 1
	}
	var result json.RawMessage
	var err error
	switch remaining[0] {
	case "tools":
		if len(remaining) != 1 {
			fmt.Fprintln(stderr, "mock: tools takes no arguments")
			return 2
		}
		result, err = client.Tools(ctx)
	case "call":
		if len(remaining) < 2 || len(remaining) > 3 {
			fmt.Fprintln(stderr, "mock: usage: mock [flags] call <tool> [arguments-json]")
			return 2
		}
		rawArguments := "{}"
		if len(remaining) == 3 {
			rawArguments = remaining[2]
		}
		arguments := json.RawMessage(rawArguments)
		if !json.Valid(arguments) || !strings.HasPrefix(strings.TrimSpace(rawArguments), "{") {
			fmt.Fprintln(stderr, "mock: arguments must be a JSON object")
			return 2
		}
		result, err = client.Call(ctx, remaining[1], arguments)
	default:
		fmt.Fprintf(stderr, "mock: unknown command %q; use tools or call\n", remaining[0])
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, "mock:", err)
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintln(stderr, "mock:", err)
		return 1
	}
	return 0
}
