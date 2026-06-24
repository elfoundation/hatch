# Webhook Integration Examples

Real-world examples of using Hatch for webhook capture, testing, and debugging.

## Example 1: Stripe Webhook Testing

Capture and test Stripe webhook events in development.

### Setup

```bash
# Start Hatch server
hatch serve &

# Capture Stripe webhook events
hatch capture /webhooks/stripe \
  -method POST \
  -header 'Stripe-Signature:whsec_test123' \
  -body '{"id":"evt_123","type":"payment_intent.succeeded","data":{"object":{"id":"pi_456","amount":2000,"currency":"usd"}}}'
```

### Development Workflow

```bash
# 1. Capture incoming webhook
hatch capture /webhooks/stripe -body '{"type":"checkout.session.completed","data":{"object":{"id":"cs_789"}}}'

# 2. Inspect captured events
hatch inspect webhooks/stripe

# 3. Search for specific event types
hatch search webhooks/stripe -query 'payment_intent.succeeded'

# 4. Replay to test handler
hatch replay <event-id> \
  -endpoint webhooks/stripe \
  -target http://localhost:3000/stripe/webhook

# 5. Configure mock for frontend testing
hatch mock set webhooks/stripe \
  -status 200 \
  -body '{"received":true}'
```

### Load Testing

```bash
# Capture multiple events
for i in {1..10}; do
  hatch capture /webhooks/stripe \
    -body "{\"id\":\"evt_$i\",\"type\":\"payment_intent.succeeded\"}"
done

# Replay all to load test handler
hatch inspect webhooks/stripe -limit 10 | \
  jq -r '.[] | "hatch replay \(.id) -endpoint webhooks/stripe -target http://localhost:3000/stripe/webhook"' | \
  bash
```

## Example 2: GitHub Webhook Debugging

Debug GitHub webhook delivery issues.

### Capture GitHub Events

```bash
# Capture push events
hatch capture /webhooks/github \
  -method POST \
  -header 'X-GitHub-Event:push' \
  -header 'X-Hub-Signature:sha1=abc123' \
  -body '{"ref":"refs/heads/main","commits":[{"id":"def456","message":"Fix bug"}]}'

# Capture pull request events
hatch capture /webhooks/github \
  -method POST \
  -header 'X-GitHub-Event:pull_request' \
  -body '{"action":"opened","pull_request":{"number":42,"title":"New feature"}}'
```

### Debug Failed Deliveries

```bash
# Find failed webhook deliveries
hatch search webhooks/github -query 'status:500'

# Get details of failed request
hatch inspect webhooks/github -limit 5

# Replay to local debugging server
hatch replay <request-id> \
  -endpoint webhooks/github \
  -target http://localhost:8080/debug
```

### Mock GitHub API Responses

```bash
# Mock GitHub API for local development
hatch mock set api/github \
  -status 200 \
  -header 'Content-Type:application/json' \
  -body '{"login":"test-user","id":12345}'
```

## Example 3: Slack Integration Testing

Test Slack app integrations without hitting real Slack APIs.

### Capture Slack Events

```bash
# Capture Slack events
hatch capture /slack/events \
  -method POST \
  -header 'X-Slack-Signature:v0=abc123' \
  -body '{"type":"event_callback","event":{"type":"message","text":"Hello world"}}'

# Capture Slack interactive payloads
hatch capture /slack/actions \
  -method POST \
  -header 'X-Slack-Signature:v0=def456' \
  -body '{"type":"interactive_message","actions":[{"type":"button","value":"approve"}]}'
```

### Development Setup

```bash
# Mock Slack API responses
hatch mock set slack/api \
  -status 200 \
  -header 'Content-Type:application/json' \
  -body '{"ok":true}'

# Test message sending
hatch replay <message-event-id> \
  -endpoint slack/events \
  -target http://localhost:3000/slack/events
```

## Example 4: Payment Provider Integration

Test payment provider webhooks across different environments.

### Multi-Environment Setup

```bash
# Development environment
hatch mock set payments/webhook \
  -status 200 \
  -body '{"status":"received"}'

# Staging environment
hatch capture payments/webhook \
  -body '{"type":"payment.completed","amount":1000,"currency":"USD"}'

# Replay to different environments
hatch replay <payment-id> \
  -endpoint payments/webhook \
  -target http://localhost:8080/payments/webhook  # Local
hatch replay <payment-id> \
  -endpoint payments/webhook \
  -target https://staging.example.com/payments/webhook  # Staging
```

### Error Simulation

```bash
# Simulate payment failure
hatch mock set payments/webhook \
  -status 200 \
  -body '{"status":"failed","error":"insufficient_funds"}'

# Simulate network timeout (use with caution)
hatch mock set payments/webhook \
  -status 504 \
  -body '{"error":"gateway_timeout"}'
```

## Example 5: API Documentation Generation

Generate OpenAPI documentation from real traffic patterns.

### Capture API Traffic

```bash
# Capture various endpoints
hatch capture /api/users -method GET
hatch capture /api/users -method POST -body '{"name":"John Doe"}'
hatch capture /api/users/123 -method GET
hatch capture /api/users/123 -method PUT -body '{"name":"Jane Doe"}'
hatch capture /api/posts -method GET
hatch capture /api/posts -method POST -body '{"title":"Hello World"}'
```

### Generate Documentation

```bash
# Generate OpenAPI specs for each endpoint
hatch doc generate api/users > users-api.json
hatch doc generate api/posts > posts-api.json

# Combine into single spec
jq -s '.[0] * .[1]' users-api.json posts-api.json > combined-api.json

# Serve with Swagger UI
docker run -p 8081:8080 \
  -e SWAGGER_JSON=/openapi.json \
  -v $(pwd):/usr/share/nginx/html/swagger \
  swaggerapi/swagger-ui
```

## Example 6: Load Testing with Real Traffic

Capture and replay production traffic for load testing.

### Capture Production Patterns

```bash
# Capture high-traffic endpoint
hatch capture /api/critical-path -limit 1000

# Export traffic patterns
hatch inspect api/critical-path -limit 1000 > traffic-patterns.json
```

### Analyze Traffic

```bash
# Analyze request distribution
jq -r '.[] | "\(.method) \(.path)"' traffic-patterns.json | \
  sort | uniq -c | sort -rn

# Find slow requests (if timing data available)
jq '.[] | select(.duration > 1000)' traffic-patterns.json

# Identify error patterns
jq '.[] | select(.status >= 400)' traffic-patterns.json
```

### Replay for Load Testing

```bash
# Create load test script
cat > load-test.sh << 'EOF'
#!/bin/bash
ENDPOINT="api/critical-path"
TARGET="https://loadtest.example.com"
REQUESTS=$(hatch inspect $ENDPOINT -limit 100)

echo "$REQUESTS" | jq -r '.[] | .id' | while read -r id; do
  echo "Replaying request $id"
  hatch replay "$id" -endpoint "$ENDPOINT" -target "$TARGET"
  sleep 0.1  # Rate limiting
done
EOF

chmod +x load-test.sh
./load-test.sh
```

## Example 7: Mock Server for Frontend Development

Set up a comprehensive mock server for frontend development.

### Create Mock Endpoints

```bash
# User API mock
hatch mock set api/users \
  -status 200 \
  -header 'Content-Type:application/json' \
  -body '[
    {"id":1,"name":"John Doe","email":"john@example.com"},
    {"id":2,"name":"Jane Smith","email":"jane@example.com"}
  ]'

# Authentication mock
hatch mock set api/auth \
  -status 200 \
  -header 'Content-Type:application/json' \
  -body '{"token":"mock-jwt-token-123","expires_in":3600}'

# Error responses
hatch mock set api/errors \
  -status 401 \
  -header 'Content-Type:application/json' \
  -body '{"error":"unauthorized","message":"Invalid credentials"}'
```

### Frontend Integration

```javascript
// Frontend code using mock API
const API_BASE = 'http://localhost:8080';

// Fetch users
const response = await fetch(`${API_BASE}/api/users`);
const users = await response.json();

// Login
const loginResponse = await fetch(`${API_BASE}/api/auth`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email: 'user@example.com', password: 'pass' })
});
```

## Best Practices

### 1. Use Descriptive Endpoint Names

```bash
# Good
hatch capture /webhooks/stripe-payments
hatch capture /api/v2/users

# Bad
hatch capture /a
hatch capture /test
```

### 2. Include Relevant Headers

```bash
# Capture with all relevant headers
hatch capture /webhooks/github \
  -header 'X-GitHub-Event:push' \
  -header 'X-Hub-Signature:sha1=abc123' \
  -header 'Content-Type:application/json'
```

### 3. Organize by Environment

```bash
# Development
hatch mock set dev/api/users -status 200 -body '[]'

# Staging
hatch mock set staging/api/users -status 200 -body '[{"id":1}]'
```

### 4. Use Search for Debugging

```bash
# Find all errors
hatch search api/webhooks -query 'status:500'

# Find specific event types
hatch search webhooks/stripe -query 'payment_intent'
```

### 5. Document Your Mocks

```bash
# Add comments to mock configuration
hatch mock set api/users \
  -status 200 \
  -header 'X-Mock-Description:Returns list of test users'
```

## Troubleshooting

### Common Issues

1. **Connection refused**: Ensure Hatch server is running (`hatch serve`)
2. **Endpoint not found**: Check endpoint ID with `curl http://localhost:8080/v1/endpoints`
3. **Request not captured**: Verify request format and headers
4. **Mock not working**: Check if mock is configured for the correct endpoint

### Debug Mode

```bash
# Enable debug logging
DEBUG=true hatch serve

# Check server logs
docker logs -f hatch-container
```

### Performance Issues

```bash
# Limit concurrent connections
hatch inspect api/heavy-endpoint -limit 10

# Use pagination for large datasets
hatch inspect api/large-dataset -limit 50
```