# mock

A small Go client for the qcrncnmock Model Context Protocol (MCP) server. It connects to a Streamable HTTP endpoint, performs the MCP initialization handshake, lists tools, and calls any tool with a JSON object.

The client uses only the Go standard library and builds for Linux and Windows without CGO.

## Build

```sh
make build       # native binary: dist/mock
make linux       # dist/mock-linux-amd64
make windows     # dist/mock-windows-amd64.exe
make release     # clean build of both release binaries
```

## Configure the server

Pass the full MCP endpoint URL with `--server` or set `QCRNCNMOCK_URL`. The default is `http://localhost:8080/mcp`.

```sh
export QCRNCNMOCK_URL="https://mock.example.com/mcp"
```

For a test server with a self-signed certificate, add `--insecure`. Do not use that flag for untrusted endpoints. Request timeout defaults to 60 seconds and can be changed with `--timeout`.

## List tools

```sh
./dist/mock --server "https://mock.example.com/mcp" tools
```

## Call a tool

```sh
./dist/mock call list_kafka_servers '{}'

./dist/mock call create_generator '{
  "sysname": "REQUEST_ID",
  "kind": "uuid"
}'

./dist/mock call create_mock '{
  "type": "rest",
  "url": "/api/example",
  "header": "Content-Type: application/json",
  "bodies": [
    {"filter_cond": "", "body": "{\"status\":\"ok\"}"}
  ]
}'

./dist/mock call create_scenario '{
  "sys_name": "example_scenario",
  "steps": [
    {"position": 1, "type": "delay", "config": "{\"duration_ms\":1000}"}
  ]
}'

./dist/mock call create_topic_copy '{
  "topic": "events",
  "source_server_id": 0,
  "target_server_id": 2
}'
```

Use `--server` before the command when the environment variable is not set:

```sh
./dist/mock --server "http://localhost:8080/mcp" tools
./dist/mock --server "http://localhost:8080/mcp" call list_kafka_servers '{}'
```

Output is indented JSON on stdout. Usage errors return exit code 2; connection and protocol errors return exit code 1.

## Architecture

- `cmd/mock` — process entry point and signal handling.
- `internal/app` — CLI parsing, validation, configuration, and output.
- `internal/qcrncnmock` — qcrncnmock-facing client operations.
- `internal/mcp` — MCP JSON-RPC transport, initialization, session headers, JSON and SSE responses.

## Test

```sh
make check
```

Tests cover the initialization handshake, session propagation, JSON and SSE responses, CLI calls, and validation. They use local HTTP test servers and do not require a real qcrncnmock instance.

## Security

Run the client only against trusted MCP servers. Tool calls can create or change server-side mocks, generators, scenarios, and Kafka forwarding rules. See [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE)
