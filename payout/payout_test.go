package payout

import "testing"

func TestShouldMintOnlyWhenNotAlreadyPaid(t *testing.T) {
	if !ShouldMint(false) {
		t.Fatal("an unpaid event must mint")
	}
	if ShouldMint(true) {
		t.Fatal("an already-paid event must not mint again (idempotency)")
	}
}
