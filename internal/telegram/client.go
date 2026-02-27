package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

type Update struct {
	UpdateID int64           `json:"update_id"`
	Raw      json.RawMessage `json:"-"`
}

type apiResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

type WebhookInfo struct {
	URL                  string `json:"url"`
	HasCustomCertificate bool   `json:"has_custom_certificate"`
	PendingUpdateCount   int    `json:"pending_update_count"`
	LastErrorDate        int64  `json:"last_error_date"`
	LastErrorMessage     string `json:"last_error_message"`
	MaxConnections       int    `json:"max_connections"`
	IPAddr               string `json:"ip_address"`
}

func NewClient(apiBase, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(apiBase, "/"),
		token:   token,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) GetMe(ctx context.Context) (json.RawMessage, error) {
	return c.call(ctx, "getMe", nil)
}

func (c *Client) SendMessage(ctx context.Context, chatID, text string) (json.RawMessage, error) {
	params := map[string]any{"chat_id": chatID, "text": text}
	return c.call(ctx, "sendMessage", params)
}

func (c *Client) SetWebhook(ctx context.Context, webhookURL string) error {
	_, err := c.call(ctx, "setWebhook", map[string]any{"url": webhookURL})
	return err
}

func (c *Client) GetWebhookInfo(ctx context.Context) (*WebhookInfo, error) {
	res, err := c.call(ctx, "getWebhookInfo", nil)
	if err != nil {
		return nil, err
	}
	var info WebhookInfo
	if err := json.Unmarshal(res, &info); err != nil {
		return nil, fmt.Errorf("decode getWebhookInfo result: %w", err)
	}
	return &info, nil
}

func (c *Client) DeleteWebhook(ctx context.Context) error {
	_, err := c.call(ctx, "deleteWebhook", map[string]any{})
	return err
}

func (c *Client) GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]Update, error) {
	endpoint, err := c.buildURL("getUpdates")
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if offset > 0 {
		q.Set("offset", strconv.FormatInt(offset, 10))
	}
	q.Set("timeout", strconv.Itoa(timeoutSec))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out struct {
		OK          bool              `json:"ok"`
		Description string            `json:"description"`
		Result      []json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode getUpdates: %w", err)
	}
	if !out.OK {
		if out.Description == "" {
			out.Description = "unknown telegram api error"
		}
		return nil, fmt.Errorf("getUpdates: %s", out.Description)
	}

	updates := make([]Update, 0, len(out.Result))
	for _, raw := range out.Result {
		var idOnly struct {
			UpdateID int64 `json:"update_id"`
		}
		if err := json.Unmarshal(raw, &idOnly); err != nil {
			return nil, fmt.Errorf("decode update id: %w", err)
		}
		updates = append(updates, Update{UpdateID: idOnly.UpdateID, Raw: raw})
	}

	return updates, nil
}

func (c *Client) call(ctx context.Context, method string, params map[string]any) (json.RawMessage, error) {
	endpoint, err := c.buildURL(method)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if params != nil {
		payload, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params for %s: %w", method, err)
		}
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out apiResponse
	if err := json.Unmarshal(resBody, &out); err != nil {
		return nil, fmt.Errorf("decode %s: %w", method, err)
	}
	if !out.OK {
		if out.Description == "" {
			out.Description = "unknown telegram api error"
		}
		return nil, fmt.Errorf("%s: %s", method, out.Description)
	}
	return out.Result, nil
}

func (c *Client) buildURL(method string) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse api base: %w", err)
	}
	u.Path = path.Join(u.Path, "bot"+c.token, method)
	return u.String(), nil
}
