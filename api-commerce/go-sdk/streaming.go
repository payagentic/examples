package payagentic

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Event represents a server-sent event from the PayAgentic streaming API.
type Event struct {
	// Type is the event type (e.g. "transaction.confirmed", "payment.settled").
	Type string `json:"type"`
	// Data is the raw JSON payload of the event.
	Data json.RawMessage `json:"data"`
}

// EventStream provides an iterator over server-sent events (SSE) from the API.
type EventStream struct {
	client  *Client
	resp    *http.Response
	scanner *bufio.Scanner
	err     error
}

// Subscribe opens an SSE connection to the given event stream path.
// The returned EventStream must be closed when no longer needed.
func (c *Client) Subscribe(ctx context.Context, path string) (*EventStream, error) {
	url := c.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("User-Agent", "payagentic-go/"+Version)
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	if c.config.AgentID != "" {
		req.Header.Set("X-Agent-ID", c.config.AgentID)
	}

	// Use the underlying http.Client directly; no retry for streaming connections.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connecting to event stream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("event stream returned status %d", resp.StatusCode)
	}

	return &EventStream{
		client:  c,
		resp:    resp,
		scanner: bufio.NewScanner(resp.Body),
	}, nil
}

// Next reads the next event from the stream. It blocks until an event is available,
// an error occurs, or the stream is closed. Returns nil, nil when the stream ends.
func (es *EventStream) Next() (*Event, error) {
	if es.err != nil {
		return nil, es.err
	}

	var eventType string
	var dataLines []string

	for es.scanner.Scan() {
		line := es.scanner.Text()

		if line == "" {
			// Empty line signals end of an event.
			if eventType != "" || len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				return &Event{
					Type: eventType,
					Data: json.RawMessage(data),
				}, nil
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
		// Ignore comments (lines starting with ":") and other fields.
	}

	if err := es.scanner.Err(); err != nil {
		es.err = err
		return nil, fmt.Errorf("reading event stream: %w", err)
	}

	// Stream ended cleanly.
	return nil, nil
}

// Close closes the event stream and releases resources.
func (es *EventStream) Close() error {
	if es.resp != nil && es.resp.Body != nil {
		return es.resp.Body.Close()
	}
	return nil
}
