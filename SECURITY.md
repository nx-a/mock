# Security

Use the client only with trusted qcrncnmock MCP servers. Tool calls can modify server-side configuration and Kafka forwarding.

TLS certificate verification is enabled by default. Use `--insecure` only for a trusted test endpoint with a self-signed certificate.

Do not publish server addresses, access tokens, request logs, or configuration containing credentials in GitHub issues. Report suspected vulnerabilities through a private channel appropriate for your deployment.
