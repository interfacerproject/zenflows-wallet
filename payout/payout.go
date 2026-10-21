// Package payout holds the wallet's pure issuance logic, kept free of the
// zenroom/cgo imports in package main so it is unit-testable on any host.
package payout

// ShouldMint is the idempotency guard: a payout for an event mints only if
// that event has not already paid out. Economic points mint once per event;
// a replay of an already-paid event mints nothing.
func ShouldMint(alreadyPaid bool) bool {
	return !alreadyPaid
}
