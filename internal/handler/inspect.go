package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/elfoundation/hatch/internal/store"
)

// inspectPageData is passed to the inspect template.
type inspectPageData struct {
	EndpointID string
	Requests   []*store.Request
}

// inspectTemplate is the server-rendered HTML for the inspect page.
var inspectTemplate = template.Must(template.New("inspect").Funcs(template.FuncMap{
	"prettyJSON": prettyJSON,
	"formatTime": formatTime,
}).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Hatch — {{.EndpointID}}</title>
<style>
  :root { color-scheme: light dark; }
  body { font-family: system-ui, sans-serif; max-width: 960px; margin: 0 auto; padding: 1rem; }
  h1 { font-size: 1.25rem; margin: 0 0 0.25rem; }
  .endpoint-url { color: #666; font-family: monospace; font-size: 0.85rem; margin-bottom: 1.5rem; word-break: break-all; }
  .request { border: 1px solid #e0e0e0; border-radius: 8px; margin-bottom: 1rem; overflow: hidden; }
  .request-header { display: flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1rem; background: #f5f5f5; cursor: pointer; user-select: none; }
  .request-header:hover { background: #eee; }
  @media (prefers-color-scheme: dark) {
    .request { border-color: #333; }
    .request-header { background: #1a1a1a; }
    .request-header:hover { background: #222; }
    .endpoint-url { color: #aaa; }
  }
  .method { font-weight: 700; font-size: 0.8rem; padding: 2px 8px; border-radius: 4px; }
  .method.GET { background: #d4edda; color: #155724; }
  .method.POST { background: #cce5ff; color: #004085; }
  .method.PUT { background: #fff3cd; color: #856404; }
  .method.PATCH { background: #fff3cd; color: #856404; }
  .method.DELETE { background: #f8d7da; color: #721c24; }
  .path { font-family: monospace; font-size: 0.9rem; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .time { color: #999; font-size: 0.8rem; white-space: nowrap; }
  .request-body { padding: 1rem; display: none; }
  .request-body.open { display: block; }
  .section { margin-bottom: 0.75rem; }
  .section-label { font-size: 0.75rem; font-weight: 600; text-transform: uppercase; color: #888; margin-bottom: 0.25rem; }
  pre { background: #f8f8f8; border: 1px solid #e0e0e0; border-radius: 4px; padding: 0.75rem; font-size: 0.8rem; overflow-x: auto; margin: 0; white-space: pre-wrap; word-break: break-all; }
  @media (prefers-color-scheme: dark) { pre { background: #111; border-color: #333; } }
  .empty { text-align: center; color: #999; padding: 3rem 0; }
  .empty h2 { font-size: 1rem; margin-bottom: 0.5rem; }
  .empty code { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; }
  @media (prefers-color-scheme: dark) { .empty code { background: #222; } }

  .replay-btn { margin-left: auto; padding: 4px 12px; font-size: 0.8rem; border: 1px solid #007bff; background: #007bff; color: #fff; border-radius: 4px; cursor: pointer; }
  .replay-btn:hover { background: #0056b3; }
  .replay-btn:disabled { opacity: 0.6; cursor: not-allowed; }
  .replay-panel { display: none; padding: 1rem; border-top: 1px solid #e0e0e0; }
  @media (prefers-color-scheme: dark) { .replay-panel { border-color: #333; } }
  .replay-panel.open { display: block; }
  .replay-url-input { width: 100%; padding: 0.5rem; border: 1px solid #ccc; border-radius: 4px; font-family: monospace; font-size: 0.85rem; box-sizing: border-box; margin-bottom: 0.5rem; }
  .replay-url-input:focus { outline: none; border-color: #007bff; box-shadow: 0 0 0 2px rgba(0,123,255,0.25); }
  .replay-actions { display: flex; gap: 0.5rem; align-items: center; }
  .replay-send-btn { padding: 6px 16px; border: none; background: #28a745; color: #fff; border-radius: 4px; cursor: pointer; font-size: 0.85rem; }
  .replay-send-btn:hover { background: #218838; }
  .replay-send-btn:disabled { opacity: 0.6; cursor: not-allowed; }
  .replay-cancel-btn { padding: 6px 12px; border: 1px solid #ccc; background: transparent; border-radius: 4px; cursor: pointer; font-size: 0.85rem; }
  .replay-result { margin-top: 0.75rem; }
  .replay-error { color: #dc3545; font-size: 0.85rem; }
  .status-badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.8rem; font-weight: 600; margin-right: 0.5rem; }
  .status-2xx { background: #d4edda; color: #155724; }
  .status-3xx { background: #fff3cd; color: #856404; }
  .status-4xx { background: #f8d7da; color: #721c24; }
  .status-5xx { background: #f8d7da; color: #721c24; }
</style>
</head>
<body>
<h1>Hatch</h1>
<div class="endpoint-url">Endpoint: <strong>{{.EndpointID}}</strong></div>

{{if .Requests}}
  {{range .Requests}}
  <div class="request" id="req-{{.ID}}">
    <div class="request-header" onclick="toggleRequest('{{.ID}}')">
      <span class="method {{.Method}}">{{.Method}}</span>
      <span class="path">{{.Path}}</span>
      <span class="time" title="{{.CreatedAt}}">{{formatTime .CreatedAt}}</span>
      <button class="replay-btn" onclick="event.stopPropagation(); openReplay('{{.ID}}')" title="Replay this request">&#x21BA; Replay</button>
    </div>
    <div class="request-body" id="body-{{.ID}}">
      {{if .Headers}}
      <div class="section">
        <div class="section-label">Headers</div>
        <pre>{{prettyJSON .Headers}}</pre>
      </div>
      {{end}}
      {{if .Query}}
      <div class="section">
        <div class="section-label">Query</div>
        <pre>{{.Query}}</pre>
      </div>
      {{end}}
      {{if .Body}}
      <div class="section">
        <div class="section-label">Body</div>
        <pre>{{printf "%s" .Body}}</pre>
      </div>
      {{end}}
    </div>
    <div class="replay-panel" id="replay-{{.ID}}">
      <input type="text" class="replay-url-input" id="replay-url-{{.ID}}" placeholder="https://myapp.local/webhook" value="">
      <div class="replay-actions">
        <button class="replay-send-btn" onclick="doReplay('{{.ID}}')">Send</button>
        <button class="replay-cancel-btn" onclick="closeReplay('{{.ID}}')">Cancel</button>
        <span class="replay-error" id="replay-error-{{.ID}}"></span>
      </div>
      <div class="replay-result" id="replay-result-{{.ID}}"></div>
    </div>
  </div>
  {{end}}
{{else}}
  <div class="empty">
    <h2>Waiting for requests...</h2>
    <p>Send a request to <code>/{{.EndpointID}}</code> to see it appear here.</p>
  </div>
{{end}}

<script>
function toggleRequest(id) {
  document.getElementById('body-' + id).classList.toggle('open');
}

function openReplay(id) {
  document.querySelectorAll('.replay-panel.open').forEach(function(p) { p.classList.remove('open'); });
  document.getElementById('replay-' + id).classList.add('open');
  var input = document.getElementById('replay-url-' + id);
  input.focus();
  if (!input.value) input.value = 'http://localhost:8080';
  document.getElementById('replay-result-' + id).innerHTML = '';
  document.getElementById('replay-error-' + id).textContent = '';
}

function closeReplay(id) {
  document.getElementById('replay-' + id).classList.remove('open');
}

function doReplay(id) {
  var targetUrl = document.getElementById('replay-url-' + id).value.trim();
  if (!targetUrl) {
    document.getElementById('replay-error-' + id).textContent = 'Please enter a target URL.';
    return;
  }
  var sendBtn = document.querySelector('#replay-' + id + ' .replay-send-btn');
  var errorEl = document.getElementById('replay-error-' + id);
  var resultEl = document.getElementById('replay-result-' + id);
  var endpointId = window.location.pathname.replace(/^\/e\//, '');

  sendBtn.disabled = true;
  errorEl.textContent = '';
  resultEl.innerHTML = '<em>Sending...</em>';

  fetch('/e/' + endpointId + '/requests/' + id + '/replay', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ target_url: targetUrl })
  }).then(function(resp) {
    return resp.json().then(function(data) { return { status: resp.status, data: data }; });
  }).then(function(result) {
    if (result.data.error) {
      errorEl.textContent = result.data.error;
      resultEl.innerHTML = '';
    } else {
      renderReplayResult(resultEl, result.data);
    }
  }).catch(function(err) {
    errorEl.textContent = 'Network error: ' + err.message;
    resultEl.innerHTML = '';
  }).finally(function() { sendBtn.disabled = false; });
}

function renderReplayResult(el, data) {
  var sc = 'status-2xx';
  if (data.status >= 300 && data.status < 400) sc = 'status-3xx';
  else if (data.status >= 400 && data.status < 500) sc = 'status-4xx';
  else if (data.status >= 500) sc = 'status-5xx';
  var html = '<div class="section"><div class="section-label">Response</div>';
  html += '<span class="status-badge ' + sc + '">' + data.status + '</span></div>';
  if (data.headers && Object.keys(data.headers).length > 0) {
    html += '<div class="section"><div class="section-label">Response Headers</div><pre>';
    for (var h in data.headers) { html += escapeHtml(h) + ': ' + escapeHtml(data.headers[h]) + '\n'; }
    html += '</pre></div>';
  }
  if (data.body) {
    html += '<div class="section"><div class="section-label">Response Body</div><pre>' + escapeHtml(data.body) + '</pre></div>';
  }
  el.innerHTML = html;
}

function formatTimeStr(ts) {
  try { var d = new Date(ts); if (isNaN(d.getTime())) return ts; } catch(_) { return ts; }
  var now = new Date();
  var diff = Math.floor((now - d) / 1000);
  if (diff < 5) return 'just now';
  if (diff < 60) return diff + 's ago';
  if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
  if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
  if (diff < 604800) return Math.floor(diff / 86400) + 'd ago';
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + ', ' + d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
}

function buildRequestBody(req) {
  var html = '';
  if (req.headers) {
    var h = req.headers;
    if (typeof h === 'string') { try { h = JSON.parse(h); } catch(_) { h = null; } }
    if (h && Object.keys(h).length > 0) {
      html += '<div class="section"><div class="section-label">Headers</div><pre>' + escapeHtml(JSON.stringify(h, null, 2)) + '</pre></div>';
    }
  }
  if (req.query) {
    html += '<div class="section"><div class="section-label">Query</div><pre>' + escapeHtml(req.query) + '</pre></div>';
  }
  if (req.body) {
    var b = typeof req.body === 'string' ? req.body : '';
    if (b) {
      html += '<div class="section"><div class="section-label">Body</div><pre>' + escapeHtml(b) + '</pre></div>';
    }
  }
  return html;
}

function escapeHtml(str) {
  var div = document.createElement('div');
  div.appendChild(document.createTextNode(str));
  return div.innerHTML;
}

// Live updates via EventSource
(function() {
  var endpointId = window.location.pathname.replace(/^\/e\//, '');
  if (!endpointId) return;
  var es = new EventSource('/e/' + endpointId + '/events');
  es.onmessage = function(e) {
    try {
      var req = JSON.parse(e.data);
      var empty = document.querySelector('.empty');
      if (empty) empty.remove();
      var div = document.createElement('div');
      div.className = 'request';
      div.id = 'req-' + req.id;
      div.innerHTML = '<div class="request-header" onclick="toggleRequest(\'' + req.id + '\')">' +
        '<span class="method ' + req.method + '">' + req.method + '</span>' +
        '<span class="path">' + req.path + '</span>' +
        '<span class="time" title="' + req.created_at + '">' + formatTimeStr(req.created_at) + '</span>' +
        '<button class="replay-btn" onclick="event.stopPropagation(); openReplay(\'' + req.id + '\')">&#x21BA; Replay</button>' +
        '</div>' +
        '<div class="request-body" id="body-' + req.id + '">' + buildRequestBody(req) + '</div>' +
        '<div class="replay-panel" id="replay-' + req.id + '">' +
        '<input type="text" class="replay-url-input" id="replay-url-' + req.id + '" placeholder="https://myapp.local/webhook" value="">' +
        '<div class="replay-actions">' +
        '<button class="replay-send-btn" onclick="doReplay(\'' + req.id + '\')">Send</button>' +
        '<button class="replay-cancel-btn" onclick="closeReplay(\'' + req.id + '\')">Cancel</button>' +
        '<span class="replay-error" id="replay-error-' + req.id + '"></span></div>' +
        '<div class="replay-result" id="replay-result-' + req.id + '"></div></div>';
      var ref = document.querySelector('.endpoint-url');
      ref.insertAdjacentElement('afterend', div);
    } catch(_) {}
  };
})();
</script>
</body>
</html>
`))

// HandleInspect serves the inspect page at GET /e/{endpointID}.
func HandleInspect(repo store.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpointID := r.PathValue("endpointID")
		if endpointID == "" {
			http.Error(w, "missing endpoint ID", http.StatusBadRequest)
			return
		}
		if _, err := repo.GetEndpoint(r.Context(), endpointID); err != nil {
			repo.CreateEndpoint(r.Context(), endpointID)
		}
		requests, err := repo.ListRequests(r.Context(), endpointID, 100)
		if err != nil {
			http.Error(w, "failed to list requests", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		inspectTemplate.Execute(w, inspectPageData{
			EndpointID: endpointID,
			Requests:   requests,
		})
	}
}

// prettyJSON formats a JSON string with indentation.
func prettyJSON(raw string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return raw
	}
	return string(b)
}

// formatTime displays an ISO 8601 timestamp as a short relative or absolute string.
func formatTime(raw string) string {
	t, err := time.Parse("2006-01-02T15:04:05.000Z07:00", raw)
	if err != nil {
		// Try a few common variants.
		t, err = time.Parse("2006-01-02T15:04:05Z07:00", raw)
	}
	if err != nil {
		return raw
	}
	d := time.Since(t)
	switch {
	case d < 5*time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2, 15:04")
	}
}
