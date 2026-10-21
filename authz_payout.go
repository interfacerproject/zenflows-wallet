package main

import "github.com/tarantool/go-tarantool"

// AlreadyPaid reports whether a payout has already been recorded for eventId.
// Integration-tier: reads the PAYOUTS space (primary key = event id).
func (storage *TTStorage) AlreadyPaid(eventId string) (bool, error) {
	resp, err := storage.db.Select("PAYOUTS", "primary", 0, 1, tarantool.IterEq, []interface{}{eventId})
	if err != nil {
		return false, err
	}
	return len(resp.Data) > 0, nil
}

// MarkPaid records that eventId has paid out, so a later issuance for the same
// event is suppressed by AlreadyPaid. Idempotent on the event id.
func (storage *TTStorage) MarkPaid(eventId string) error {
	_, err := storage.db.Insert("PAYOUTS", []interface{}{eventId})
	return err
}
