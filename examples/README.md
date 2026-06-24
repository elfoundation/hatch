# Hatch Examples

Real-world examples and integration guides for using Hatch.

## Examples Index

| Example | Description | Use Case |
|---------|-------------|----------|
| [Webhook Integration](webhook-integration.md) | Comprehensive webhook capture, testing, and debugging | Payment processors, GitHub, Slack, API integrations |

## Quick Start Examples

### Capture Your First Request

```bash
# Start Hatch server
hatch serve &

# Capture a request
curl -X POST http://localhost:8080/my-endpoint \
  -H "Content-Type: application/json" \
  -d '{"event":"test","data":{"id":123}}'

# View captured requests
hatch inspect my-endpoint
```

### Mock API for Development

```bash
# Configure mock response
hatch mock set api/users \
  -status 200 \
  -header "Content-Type:application/json" \
  -body '[{"id":1,"name":"John Doe"}]'

# Test your frontend
curl http://localhost:8080/api/users
```

### Debug Failed Webhooks

```bash
# Find errors
hatch search webhooks -query "status:500"

# Replay to debug
hatch replay <request-id> \
  -endpoint webhooks \
  -target http://localhost:3000/debug
```

## Integration Guides

### Stripe Webhooks

See [Webhook Integration - Stripe Example](webhook-integration.md#example-1-stripe-webhook-testing)

### GitHub Webhooks

See [Webhook Integration - GitHub Example](webhook-integration.md#example-2-github-webhook-debugging)

### Slack Integration

See [Webhook Integration - Slack Example](webhook-integration.md#example-3-slack-integration-testing)

## Use Cases

### 1. Development Environment

Use Hatch as a local mock server for frontend development.

```bash
# Configure mocks for your API
hatch mock set api/users -status 200 -body '[]'
hatch mock set api/auth -status 200 -body '{"token":"mock-token"}'

# Frontend code uses http://localhost:8080 as API base
```

### 2. Testing & QA

Capture production traffic patterns and replay them in testing.

```bash
# Capture traffic
hatch inspect api/critical-path -limit 1000 > traffic.json

# Replay in test environment
cat traffic.json | jq -r '.[].id' | while read id; do
  hatch replay "$id" -endpoint api/critical-path -target http://staging:8080
done
```

### 3. API Documentation

Generate OpenAPI specs from real traffic.

```bash
# Capture representative traffic
hatch capture /api/users -method GET
hatch capture /api/users -method POST -body '{"name":"test"}'

# Generate documentation
hatch doc generate api/users > users-api.json
```

### 4. Load Testing

Replay captured traffic for load testing.

```bash
# Capture production patterns
hatch inspect api/critical-path -limit 1000 > load-test-traffic.json

# Create load test script
cat > load-test.sh << 'EOF'
#!/bin/bash
for i in {1..100}; do
  hatch inspect api/critical-path -limit 10 | \
    jq -r '.[] | "hatch replay \(.id) -endpoint api/critical-path -target http://loadtest:8080"' | \
    bash &
done
wait
EOF
chmod +x load-test.sh
```

## Best Practices

1. **Use descriptive endpoint names**
   ```bash
   # Good
   hatch capture /webhooks/stripe-payments
   hatch capture /api/v2/users
   
   # Bad
   hatch capture /a
   hatch capture /test
   ```

2. **Include relevant headers**
   ```bash
   hatch capture /webhooks/github \
     -header 'X-GitHub-Event:push' \
     -header 'X-Hub-Signature:sha1=abc123'
   ```

3. **Organize by environment**
   ```bash
   # Development
   hatch mock set dev/api/users -status 200 -body '[]'
   
   # Staging
   hatch mock set staging/api/users -status 200 -body '[{"id":1}]'
   ```

4. **Use search for debugging**
   ```bash
   # Find all errors
   hatch search api/webhooks -query 'status:500'
   
   # Find specific events
   hatch search webhooks/stripe -query 'payment_intent'
   ```

## Troubleshooting

See [CLI Troubleshooting Guide](../docs/engineering/cli-troubleshooting.md) for common issues and solutions.

## Contributing Examples

To add a new example:

1. Create a markdown file in this directory
2. Follow the naming convention: `<use-case>.md`
3. Include:
   - Clear use case description
   - Step-by-step instructions
   - Complete code examples
   - Expected output
   - Common pitfalls
4. Update this README.md to include your example in the index

## Resources

- [CLI Reference](../docs/engineering/cli.md) - Complete command documentation
- [Quick Reference](../docs/engineering/cli-quick-reference.md) - Command cheat sheet
- [Troubleshooting](../docs/engineering/cli-troubleshooting.md) - Common issues and solutions