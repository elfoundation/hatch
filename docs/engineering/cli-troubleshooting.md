# CLI Troubleshooting Guide

Comprehensive troubleshooting guide for Hatch CLI issues.

## Quick Diagnostics

### Health Check

First, verify the Hatch server is running and accessible:

```bash
# Check if server is responding
curl -s http://localhost:8080/healthz

# Expected output: ok

# If using custom URL
curl -s $HATCH_URL/healthz
```

### Version Check

Verify you're using the correct version:

```bash
hatch version
# Expected output: hatch v0.1.0

# Check if binary is working
hatch --help
```

### Connectivity Test

Test network connectivity to the server:

```bash
# Test with verbose output
curl -v http://localhost:8080/healthz

# Test with specific timeout
curl --connect-timeout 5 http://localhost:8080/healthz
```

## Common Error Messages

### Connection Issues

#### `Error: request failed: connection refused`

**Cause:** The Hatch server is not running or not accessible.

**Solutions:**

1. Start the Hatch server:
   ```bash
   hatch serve
   ```

2. Check if the server is listening on the expected port:
   ```bash
   # Linux/macOS
   netstat -tlnp | grep 8080
   # or
   lsof -i :8080
   
   # Windows
   netstat -ano | findstr :8080
   ```

3. Verify the HATCH_URL environment variable:
   ```bash
   echo $HATCH_URL
   # Should be http://localhost:8080 or your server URL
   ```

4. Check firewall settings:
   ```bash
   # Linux
   sudo iptables -L -n | grep 8080
   
   # macOS
   sudo pfctl -sr | grep 8080
   ```

#### `Error: request failed: dial tcp: lookup localhost: no such host`

**Cause:** DNS resolution failure.

**Solutions:**

1. Use IP address instead of hostname:
   ```bash
   export HATCH_URL=http://127.0.0.1:8080
   ```

2. Check `/etc/hosts` file:
   ```bash
   grep localhost /etc/hosts
   # Should contain: 127.0.0.1 localhost
   ```

#### `Error: request failed: context deadline exceeded`

**Cause:** Request timeout.

**Solutions:**

1. Check server load:
   ```bash
   top -bn1 | grep hatch
   ```

2. Increase timeout (if using curl):
   ```bash
   curl --max-time 30 http://localhost:8080/healthz
   ```

3. Check network latency:
   ```bash
   ping localhost
   ```

### Authentication Errors

#### `Error (HTTP 401): unauthorized`

**Cause:** Authentication required but not provided.

**Solutions:**

1. Check if authentication is enabled:
   ```bash
   # Check server configuration
   docker exec hatch-container env | grep AUTH
   ```

2. Provide authentication token:
   ```bash
   # For API requests
   curl -H "Authorization: Bearer $HATCH_AUTH_TOKEN" http://localhost:8080/v1/endpoints
   ```

3. For development, disable authentication:
   ```bash
   # Start server without auth
   HATCH_AUTH_ENABLED=false hatch serve
   ```

#### `Error (HTTP 403): forbidden`

**Cause:** Insufficient permissions.

**Solutions:**

1. Check user roles and permissions
2. Verify API token has required scopes
3. Contact administrator for access

### Request/Response Errors

#### `Error (HTTP 400): bad request`

**Cause:** Invalid request format.

**Solutions:**

1. Validate JSON format:
   ```bash
   echo '{"invalid":json}' | jq .
   # Should show parse error
   ```

2. Check Content-Type header:
   ```bash
   curl -H "Content-Type: application/json" -d '{"valid":"json"}' ...
   ```

3. Verify request body size:
   ```bash
   # Check body size
   echo -n '{"data":"..."}' | wc -c
   ```

#### `Error (HTTP 404): endpoint not found`

**Cause:** Requested resource doesn't exist.

**Solutions:**

1. List available endpoints:
   ```bash
   curl http://localhost:8080/v1/endpoints
   ```

2. Check endpoint ID spelling:
   ```bash
   # Case-sensitive
   hatch inspect MyEndpoint  # Wrong
   hatch inspect myendpoint  # Correct
   ```

3. Verify endpoint has captured requests:
   ```bash
   hatch inspect myendpoint -limit 1
   ```

#### `Error (HTTP 413): payload too large`

**Cause:** Request body exceeds size limit.

**Solutions:**

1. Reduce request body size
2. Compress data before sending
3. Split into smaller chunks

#### `Error (HTTP 429): too many requests`

**Cause:** Rate limiting.

**Solutions:**

1. Implement backoff:
   ```bash
   # Exponential backoff script
   for i in {1..5}; do
     sleep $((2**i))
     hatch capture /api -body '{"retry":'$i'}' && break
   done
   ```

2. Check rate limit headers:
   ```bash
   curl -I http://localhost:8080/v1/endpoints
   # Look for X-RateLimit-* headers
   ```

### Data Errors

#### `Error: invalid character '}' looking for beginning of value`

**Cause:** Malformed JSON in request body.

**Solutions:**

1. Validate JSON:
   ```bash
   echo '{"key":"value"}' | jq .
   ```

2. Use proper escaping:
   ```bash
   # Bash
   hatch capture /api -body '{"key":"value with \"quotes\""}'
   
   # Use file input
   echo '{"key":"value"}' > request.json
   hatch capture /api -body @request.json
   ```

3. Check for invisible characters:
   ```bash
   cat -A request.json
   ```

#### `Error: unexpected end of JSON input`

**Cause:** Incomplete JSON response.

**Solutions:**

1. Check server logs for errors
2. Verify network connection wasn't interrupted
3. Try with smaller payload

### Storage Errors

#### `Error: database is locked`

**Cause:** SQLite database contention.

**Solutions:**

1. Reduce concurrent operations
2. Check for long-running transactions:
   ```bash
   # Monitor database locks
   sqlite3 hatch.db "PRAGMA journal_mode=WAL;"
   ```

3. Consider using PostgreSQL for production

#### `Error: no space left on device`

**Cause:** Disk space exhausted.

**Solutions:**

1. Check disk space:
   ```bash
   df -h
   du -sh /var/lib/hatch
   ```

2. Clean old data:
   ```bash
   # Remove requests older than 7 days
   curl -X DELETE "http://localhost:8080/v1/endpoints/myendpoint/requests?older_than=7d"
   ```

3. Configure data retention:
   ```bash
   # Set retention policy
   HATCH_RETENTION_DAYS=30 hatch serve
   ```

## Platform-Specific Issues

### Linux

#### Permission Denied

```bash
# Fix binary permissions
chmod +x /usr/local/bin/hatch

# Or install to user directory
mkdir -p ~/.local/bin
mv hatch ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
```

#### SELinux Issues

```bash
# Check SELinux status
getenforce

# If enabled, add context
sudo semanage fcontext -a -t bin_t /usr/local/bin/hatch
sudo restorecon -v /usr/local/bin/hatch
```

### macOS

#### Gatekeeper Blocked

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/hatch

# Or allow in System Preferences > Security & Privacy
```

#### Homebrew Path Issues

```bash
# Ensure Homebrew bin is in PATH
export PATH="/usr/local/bin:$PATH"

# Or create symlink
ln -s /usr/local/bin/hatch /opt/homebrew/bin/hatch
```

### Windows

#### Windows Defender Blocked

1. Open Windows Security
2. Go to Virus & threat protection
3. Allow app through firewall

#### PowerShell Execution Policy

```bash
# Run as Administrator
Set-ExecutionPolicy RemoteSigned

# Or bypass for current session
powershell -ExecutionPolicy Bypass -File script.ps1
```

#### PATH Not Updated

```bash
# Add to PATH permanently
$env:PATH += ";C:\path\to\hatch"

# Or use System Properties > Environment Variables
```

## Performance Issues

### Slow Response Times

#### Diagnose

```bash
# Measure response time
time curl -s http://localhost:8080/v1/endpoints > /dev/null

# Check server resources
top -bn1 | grep hatch

# Monitor network
iftop -i eth0
```

#### Solutions

1. **Increase server resources:**
   ```bash
   # Docker
   docker update --cpus="2.0" --memory="2g" hatch-container
   ```

2. **Optimize database:**
   ```bash
   # Vacuum SQLite database
   sqlite3 hatch.db "VACUUM;"
   ```

3. **Enable caching:**
   ```bash
   HATCH_CACHE_TTL=300 hatch serve
   ```

### High Memory Usage

#### Diagnose

```bash
# Check memory usage
ps aux | grep hatch
pmap -x $(pgrep hatch) | tail -1

# Monitor over time
while true; do
  echo "$(date): $(ps -o rss= -p $(pgrep hatch))KB"
  sleep 60
done
```

#### Solutions

1. **Limit request history:**
   ```bash
   HATCH_MAX_REQUESTS=10000 hatch serve
   ```

2. **Enable pagination:**
   ```bash
   hatch inspect myendpoint -limit 100
   ```

3. **Restart periodically:**
   ```bash
   # Cron job
   0 0 * * * systemctl restart hatch
   ```

### High CPU Usage

#### Diagnose

```bash
# Check CPU usage
top -bn1 | grep hatch

# Profile if needed
go tool pprof http://localhost:6060/debug/pprof/profile
```

#### Solutions

1. **Reduce logging:**
   ```bash
   HATCH_LOG_LEVEL=warn hatch serve
   ```

2. **Limit concurrent connections:**
   ```bash
   HATCH_MAX_CONNECTIONS=100 hatch serve
   ```

## Network Issues

### Proxy Configuration

```bash
# Set proxy environment variables
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1

# Or configure in hatch
HATCH_PROXY=$HTTP_PROXY hatch serve
```

### SSL/TLS Issues

```bash
# Skip SSL verification (development only)
curl -k https://hatch.example.com/healthz

# Or add certificate
curl --cacert /path/to/ca.crt https://hatch.example.com/healthz

# For self-signed certificates
export CURL_CA_BUNDLE=/path/to/cert.pem
```

### DNS Resolution

```bash
# Test DNS resolution
nslookup localhost
dig localhost

# Use IP address
export HATCH_URL=http://127.0.0.1:8080
```

## Docker Issues

### Container Won't Start

```bash
# Check container logs
docker logs hatch-container

# Check container status
docker ps -a | grep hatch

# Inspect container
docker inspect hatch-container
```

### Port Conflicts

```bash
# Find what's using port 8080
lsof -i :8080
netstat -tlnp | grep 8080

# Use different port
docker run -p 9090:8080 hatch:latest
```

### Volume Mount Issues

```bash
# Check volume permissions
ls -la /path/to/volume

# Fix permissions
sudo chown -R 1000:1000 /path/to/volume

# Test mount
docker run -v /path/to/volume:/data busybox ls -la /data
```

## Debugging Techniques

### Enable Debug Logging

```bash
# Set debug level
DEBUG=true hatch serve

# Or use environment variable
HATCH_LOG_LEVEL=debug hatch serve
```

### Verbose Output

```bash
# Use curl verbose mode
curl -v http://localhost:8080/v1/endpoints

# Capture full request/response
curl -v -w '\n' http://localhost:8080/v1/endpoints > debug.txt 2>&1
```

### Network Tracing

```bash
# Linux
strace -e trace=network hatch serve

# macOS
dtruss -n hatch

# Windows
netsh trace start capture=yes tracefile=trace.etl
```

### Core Dumps

```bash
# Enable core dumps
ulimit -c unlimited

# Run with core dump
hatch serve &
echo $! > /tmp/hatch.pid

# If crash occurs, analyze
gdb /usr/local/bin/hatch core
```

## Getting Help

### Documentation

- [CLI Reference](cli.md) - Command documentation
- [Examples](../../examples/) - Usage examples
- [Architecture](hatch-architecture.md) - System design

### Community

- [GitHub Issues](https://github.com/elfoundation/hatch/issues) - Bug reports
- [Discussions](https://github.com/elfoundation/hatch/discussions) - Questions

### Logs

```bash
# Check application logs
docker logs -f hatch-container

# Check system logs
journalctl -u hatch -f

# Check access logs
tail -f /var/log/hatch/access.log
```

## Reporting Issues

When reporting issues, include:

1. **Environment details:**
   ```bash
   hatch version
   go version
   uname -a  # Linux/macOS
   systeminfo  # Windows
   ```

2. **Configuration:**
   ```bash
   env | grep HATCH
   ```

3. **Reproduction steps:**
   - Exact commands run
   - Expected behavior
   - Actual behavior

4. **Logs:**
   ```bash
   # Attach relevant logs
   docker logs hatch-container > hatch.log 2>&1
   ```

5. **Network info:**
   ```bash
   curl -v http://localhost:8080/healthz > network-debug.txt 2>&1
   ```