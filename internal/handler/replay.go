package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/elfoundation/hatch/internal/store"
)

const maxReplayBodyBytes = 64 * 1024

type replayRequest struct {
	TargetURL string `json:"target_url"`
}

type replayResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Error   string            `json:"error,omitempty"`
}

func HandleReplay(repo store.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.PathValue("requestID")
		if requestID == "" {
			writeError(w, http.StatusBadRequest, "missing request_id")
			return
		}

		var replay replayRequest
		if err := json.NewDecoder(r.Body).Decode(&replay); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
			return
		}
		if replay.TargetURL == "" {
			writeError(w, http.StatusBadRequest, "target_url is required")
			return
		}

		target, err := url.Parse(replay.TargetURL)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid target_url: "+err.Error())
			return
		}
		if target.Scheme != "http" && target.Scheme != "https" {
			writeError(w, http.StatusBadRequest, "target_url scheme must be http or https")
			return
		}

		if !allowPrivateReplay() && isPrivateAddr(target.Host) {
			writeError(w, http.StatusForbidden, "replay to private/loopback addresses is denied. Set HATCH_ALLOW_PRIVATE_REPLAY=true to allow")
			return
		}

		capReq, err := repo.GetRequest(r.Context(), requestID)
		if err != nil {
			writeError(w, http.StatusNotFound, "request not found: "+err.Error())
			return
		}

		replayResult, err := doReplay(capReq, target)
		if err != nil {
			writeError(w, http.StatusBadGateway, "replay failed: "+err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(replayResult)
	}
}

func doReplay(capReq *store.Request, targetURL *url.URL) (*replayResponse, error) {
	outURL := *targetURL
	outURL.Path = joinPath(targetURL.Path, capReq.Path)
	if capReq.Query != "" {
		outURL.RawQuery = capReq.Query
	}

	var bodyReader io.Reader
	if capReq.Body != nil && len(capReq.Body) > 0 {
		bodyReader = strings.NewReader(string(capReq.Body))
	}

	req, err := http.NewRequest(capReq.Method, outURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	var headers map[string]string
	if capReq.Headers != "" {
		if err := json.Unmarshal([]byte(capReq.Headers), &headers); err != nil {
			headers = nil
		}
	}
	for k, v := range headers {
		if isHopByHop(k) {
			continue
		}
		req.Header.Set(k, v)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !allowPrivateReplay() && isPrivateAddr(req.URL.Host) {
				return fmt.Errorf("redirect to private address %s denied", req.URL.Host)
			}
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxReplayBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	respHeaders := make(map[string]string, len(resp.Header))
	for k, vs := range resp.Header {
		respHeaders[k] = strings.Join(vs, ", ")
	}

	return &replayResponse{
		Status:  resp.StatusCode,
		Headers: respHeaders,
		Body:    string(body),
	}, nil
}

// isPrivateAddr reports whether addr falls within a private, loopback, or reserved range.
func isPrivateAddr(hostport string) bool {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = hostport
	}
	if host == "" {
		return false
	}

	// String-based checks first (catch names and the unspecified address 0.0.0.0).
	if host == "localhost" || host == "0.0.0.0" || host == "::1" || host == "[::1]" {
		return true
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
	}
	return false
}

func allowPrivateReplay() bool {
	return strings.EqualFold(os.Getenv("HATCH_ALLOW_PRIVATE_REPLAY"), "true")
}

var hopByHopHeaders = map[string]bool{
	"connection":          true,
	"keep-alive":          true,
	"proxy-authenticate":  true,
	"proxy-authorization": true,
	"te":                  true,
	"trailer":             true,
	"transfer-encoding":   true,
	"upgrade":             true,
}

func isHopByHop(header string) bool {
	return hopByHopHeaders[strings.ToLower(header)]
}

func joinPath(base, extra string) string {
	base = strings.TrimRight(base, "/")
	extra = strings.TrimLeft(extra, "/")
	if base == "" {
		return "/" + extra
	}
	if extra == "" {
		return base
	}
	return base + "/" + extra
}
