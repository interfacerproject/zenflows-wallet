package payout

// Proof carries the exact signed bytes (base64) and their signature, so the
// PDP re-verifies the user's own signature.
type Proof struct {
	Signature  string `json:"signature"`
	PayloadB64 string `json:"payload_b64"`
}

// EventRef names the ValueFlows/social event that justifies an issuance. The
// authority (PDP) resolves it to recipients, amount and paid-state.
type EventRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Request is the wire request the wallet sends to the PDP. It carries NO
// owner/amount — those are derived server-side from the event.
type Request struct {
	RequestID string `json:"request_id"`
	Action    string `json:"action"`
	Principal struct {
		Pk string `json:"pk"`
	} `json:"principal"`
	Proof Proof    `json:"proof"`
	Event EventRef `json:"event"`
}

// Payout is one mint the PDP authorized.
type Payout struct {
	Recipient string `json:"recipient"`
	Token     string `json:"token"`
	Amount    int64  `json:"amount"`
}

// Response is the PDP decision.
type Response struct {
	Decision   string   `json:"decision"`
	ReasonCode string   `json:"reason_code"`
	Payouts    []Payout `json:"payouts"`
}

// Outcome is what the handler enforces.
type Outcome struct {
	Mint       bool
	Status     int
	ReasonCode string
	Payouts    []Payout
}

// BuildIssueRequest assembles the wire request. The action is fixed here, not
// taken from the client.
func BuildIssueRequest(requestID, pk string, proof Proof, ev EventRef) Request {
	var r Request
	r.RequestID = requestID
	r.Action = "wallet.points.issue"
	r.Principal.Pk = pk
	r.Proof = proof
	r.Event = ev
	return r
}

// InterpretPayouts maps a PDP response (or transport error) to an outcome.
// Fail closed: only an "allow" carrying at least one payout mints.
func InterpretPayouts(resp Response, transportErr error) Outcome {
	if transportErr != nil {
		return Outcome{Mint: false, Status: 403, ReasonCode: "pdp_unreachable"}
	}
	if resp.Decision == "allow" {
		if len(resp.Payouts) == 0 {
			return Outcome{Mint: false, Status: 403, ReasonCode: "no_payouts"}
		}
		return Outcome{Mint: true, Status: 200, ReasonCode: resp.ReasonCode, Payouts: resp.Payouts}
	}
	reason := resp.ReasonCode
	if reason == "" {
		reason = "denied"
	}
	return Outcome{Mint: false, Status: 403, ReasonCode: reason}
}
