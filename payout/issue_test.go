package payout

import (
	"errors"
	"testing"
)

func TestBuildIssueRequestCarriesEventAndProof(t *testing.T) {
	r := BuildIssueRequest("rid", "pkOwner", Proof{Signature: "SIG", PayloadB64: "e30="}, EventRef{Type: "accept", ID: "01EVT"})
	if r.Action != "wallet.points.issue" || r.Principal.Pk != "pkOwner" {
		t.Fatalf("action/pk: %+v", r)
	}
	if r.Event.Type != "accept" || r.Event.ID != "01EVT" {
		t.Fatalf("event not carried: %+v", r.Event)
	}
	if r.Proof.Signature != "SIG" || r.Proof.PayloadB64 != "e30=" {
		t.Fatalf("proof not carried: %+v", r.Proof)
	}
}

func TestInterpretAllowWithPayoutsMints(t *testing.T) {
	out := InterpretPayouts(Response{Decision: "allow", Payouts: []Payout{{Recipient: "O", Token: "idea", Amount: 10000}}}, nil)
	if !out.Mint || out.Status != 200 || len(out.Payouts) != 1 {
		t.Fatalf("allow+payouts must mint: %+v", out)
	}
}

func TestInterpretAllowEmptyPayoutsFailsClosed(t *testing.T) {
	out := InterpretPayouts(Response{Decision: "allow", Payouts: nil}, nil)
	if out.Mint || out.Status != 403 {
		t.Fatalf("allow with no payouts must fail closed: %+v", out)
	}
}

func TestInterpretDenyFailsClosed(t *testing.T) {
	out := InterpretPayouts(Response{Decision: "deny", ReasonCode: "already_paid"}, nil)
	if out.Mint || out.Status != 403 || out.ReasonCode != "already_paid" {
		t.Fatalf("deny must fail closed: %+v", out)
	}
}

func TestInterpretTransportErrorFailsClosed(t *testing.T) {
	out := InterpretPayouts(Response{}, errors.New("refused"))
	if out.Mint || out.Status != 403 || out.ReasonCode != "pdp_unreachable" {
		t.Fatalf("transport error must fail closed: %+v", out)
	}
}
