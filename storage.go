// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2023 Dyne.org foundation <foundation@dyne.org>.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	//"encoding/json"
	"errors"
	"github.com/tarantool/go-tarantool"
	"log"
	"sort"
	"strconv"
	"time"
)

type TTStorage struct {
	db *tarantool.Connection
}

const MAX_RETRY int = 10

type byTimestamp []Transaction

func (s byTimestamp) Len() int {
	return len(s)
}
func (s byTimestamp) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byTimestamp) Less(i, j int) bool {
	return s[i].Timestamp < s[j].Timestamp
}

// TODO: use big integer for numbers (save them as strings)
func (storage *TTStorage) Init(host, user, pass string) error {
	var err error
	for done, retry := false, 0; !done; retry++ {
		storage.db, err = tarantool.Connect(host, tarantool.Opts{
			User: user,
			Pass: pass,
		})
		done = retry == MAX_RETRY || err == nil
		if !done {
			log.Println("Could not connect to tarantool, retrying...")
			time.Sleep(3 * time.Second)
		} else {
			log.Println("Connected to tarantool")
		}
	}
	return err
}

func (storage *TTStorage) AddDiff(owner, token string, amount int64) error {
	now := time.Now()
	timestamp := uint64(now.UnixMilli())
	_, err := storage.db.Insert("TXS", []interface{}{nil, owner, token, timestamp, strconv.FormatInt(int64(amount), 10)})
	if err != nil {
		return err
	}
	return nil

}

func (storage *TTStorage) Read(owner, token string, untilTimestamp uint64) (int64, error) {
	// SELECT SUM(amount) ... doesn't work???? Problem with numbers/integer
	// Now numbers are stored as strings....
	//resp, err := storage.db.Execute("SELECT amount FROM txs WHERE owner = ? and token = ?", []interface{}{owner, token})

	var amount int64 = 0
	const limit uint32 = 100
	var offset uint32 = 0
	for {
		resp, err := storage.db.Select("TXS", "owner_token", offset, limit, tarantool.IterEq, []interface{}{owner, token})
		if err != nil {
			return 0, err
		}
		if resp.Error != "" {
			return 0, errors.New(resp.Error)
		}
		if len(resp.Data) == 0 {
			break
		}
		for i := 0; i < len(resp.Data); i = i + 1 {
			n, err := strconv.ParseInt(resp.Data[i].([]interface{})[4].(string), 10, 64)
			if err != nil {
				return 0, err
			}
			if untilTimestamp <= 0 || resp.Data[i].([]interface{})[3].(uint64) <= untilTimestamp {
				amount = amount + n
			}
		}
		offset = offset + limit
	}
	return amount, nil
}
func (storage *TTStorage) ReadTxs(owner, token string, n int) ([]Transaction, error) {
	const limit uint32 = 100
	var offset uint32 = 0
	txs := []Transaction{}

	for {
		resp, err := storage.db.Select("TXS", "owner_token", offset, limit, tarantool.IterEq, []interface{}{owner, token})
		if err != nil {
			return nil, err
		}
		if resp.Error != "" {
			return nil, errors.New(resp.Error)
		}
		if len(resp.Data) == 0 {
			break
		}
		for i := 0; i < len(resp.Data); i = i + 1 {
			n, err := strconv.ParseInt(resp.Data[i].([]interface{})[4].(string), 10, 64)
			if err != nil {
				return nil, err
			}
			txs = append(txs, Transaction{
				Id:        resp.Data[i].([]interface{})[0].(uint64),
				Timestamp: resp.Data[i].([]interface{})[3].(uint64),
				Amount:    n,
			})
		}
		offset = offset + limit
	}
	sort.Sort(byTimestamp(txs))
	if len(txs) > n {
		n = len(txs) - n
	} else {
		n = 0
	}
	return txs[n:len(txs)], nil
}
