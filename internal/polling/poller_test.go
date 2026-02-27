package polling

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/example/tgbot-cli/internal/telegram"
)

type fakeAPI struct {
	deleteCalled bool
	updates      [][]telegram.Update
	idx          int
}

func (f *fakeAPI) DeleteWebhook(_ context.Context) error {
	f.deleteCalled = true
	return nil
}

func (f *fakeAPI) GetUpdates(_ context.Context, _ int64, _ int) ([]telegram.Update, error) {
	if f.idx >= len(f.updates) {
		return nil, context.Canceled
	}
	u := f.updates[f.idx]
	f.idx++
	return u, nil
}

func TestFormatUpdatePretty(t *testing.T) {
	out, err := formatUpdate([]byte(`{"update_id":1,"message":{"text":"hi"}}`), "pretty")
	if err != nil {
		t.Fatalf("formatUpdate returned error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "\n  \"update_id\": 1") {
		t.Fatalf("expected indented pretty json, got: %s", s)
	}
}

func TestFormatUpdateJSONL(t *testing.T) {
	out, err := formatUpdate([]byte(`{"update_id":1}`), "jsonl")
	if err != nil {
		t.Fatalf("formatUpdate returned error: %v", err)
	}
	if string(out) != "{\"update_id\":1}\n" {
		t.Fatalf("unexpected jsonl output: %q", string(out))
	}
}

func TestPollerRunWritesFormattedOutput(t *testing.T) {
	api := &fakeAPI{updates: [][]telegram.Update{{{UpdateID: 100, Raw: []byte(`{"update_id":100,"message":{"text":"hello"}}`)}}}}
	p := New(api, Options{Once: true, OutputFormat: "pretty"})
	var out, errOut strings.Builder

	if err := p.Run(context.Background(), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(out.String(), "\"text\": \"hello\"") {
		t.Fatalf("expected pretty output to contain message text, got %s", out.String())
	}
}

func TestPollerRunDeleteWebhook(t *testing.T) {
	api := &fakeAPI{updates: [][]telegram.Update{{}}}
	p := New(api, Options{Once: true, DeleteWebhook: true, OutputFormat: "jsonl"})
	if err := p.Run(context.Background(), &strings.Builder{}, &strings.Builder{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !api.deleteCalled {
		t.Fatalf("expected DeleteWebhook to be called")
	}
}

func TestPollerRunErrorOnInvalidJSON(t *testing.T) {
	api := &fakeAPI{updates: [][]telegram.Update{{{UpdateID: 1, Raw: []byte(`{not-json}`)}}}}
	p := New(api, Options{Once: true, OutputFormat: "pretty"})
	err := p.Run(context.Background(), &strings.Builder{}, &strings.Builder{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "decode update json") {
		t.Fatalf("unexpected error: %v", err)
	}
	if errors.Is(err, context.Canceled) {
		t.Fatalf("should not be context canceled")
	}
}
