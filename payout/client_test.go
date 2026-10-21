package payout

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecodeDecision(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/authz/decide" {
			t.Errorf("path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Response{Decision: "allow", Payouts: []Payout{{Recipient: "O", Token: "idea", Amount: 10000}}})
	}))
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	resp, err := c.Decide(context.Background(), BuildIssueRequest("r", "pk", Proof{}, EventRef{Type: "accept", ID: "1"}))
	if err != nil || resp.Decision != "allow" || len(resp.Payouts) != 1 {
		t.Fatalf("decode: resp=%+v err=%v", resp, err)
	}
}

func TestDecideEmptyBaseURLErrors(t *testing.T) {
	c := &Client{BaseURL: "", HTTP: http.DefaultClient}
	if _, err := c.Decide(context.Background(), Request{}); err == nil {
		t.Fatal("empty base URL must error")
	}
}

func TestDecideNon2xxErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	if _, err := c.Decide(context.Background(), Request{}); err == nil {
		t.Fatal("non-2xx must error")
	}
}
