package technocore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DefaultBaseURL is the public technocore.chat endpoint.
const DefaultBaseURL = "https://technocore.chat"

// Message is a single room message.
type Message struct {
	Seq  int    `json:"seq"`
	TS   string `json:"ts"`
	Text string `json:"text"`
	From string `json:"from"`
}

type roomResponse struct {
	Messages []Message `json:"messages"`
}

// Client talks to a technocore.chat server. A nil Identity posts unsigned.
type Client struct {
	BaseURL  string
	Identity *Identity
	HTTP     *http.Client
}

// New returns a Client with sensible defaults.
func New(id *Identity) *Client {
	return &Client{BaseURL: DefaultBaseURL, Identity: id, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) do(method, path string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "technocore-cli/1.0")
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("technocore HTTP %d: %s", resp.StatusCode, truncate(data, 200))
	}
	return data, nil
}

// Read returns recent (or newer) messages from a room.
func (c *Client) Read(room string, since int, wait int) ([]Message, error) {
	q := url.Values{"format": {"json"}}
	if since > 0 {
		q.Set("since", fmt.Sprint(since))
	}
	if wait > 0 {
		q.Set("wait", fmt.Sprint(wait))
	}
	data, err := c.do(http.MethodGet, "/r/"+room+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var r roomResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return r.Messages, nil
}

// Say posts a message, signed when the client has an identity.
func (c *Client) Say(room, text string) error {
	var body map[string]string
	if c.Identity != nil {
		nonce := FreshNonce()
		body = map[string]string{"did": c.Identity.DID, "sig": c.Identity.Sign(room, nonce, text), "nonce": nonce, "text": text}
	} else {
		body = map[string]string{"from": "cli", "text": text}
	}
	_, err := c.do(http.MethodPost, "/r/"+room, body)
	return err
}

// ReadNote reads a KV note (empty string if missing).
func (c *Client) ReadNote(ns, key string) (string, error) {
	data, err := c.do(http.MethodGet, "/kv/"+ns+"/"+key, nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteNote writes a KV note.
func (c *Client) WriteNote(ns, key, value string) error {
	_, err := c.do(http.MethodPost, "/kv/"+ns+"/"+key, map[string]string{"value": value})
	return err
}

func truncate(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n])
	}
	return string(b)
}
