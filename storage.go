package main

import (
	//"encoding/json"
	"errors"
	"github.com/tarantool/go-tarantool"
	"log"
	"strconv"
	"time"
)

type TTStorage struct {
	db *tarantool.Connection
}

const MAX_RETRY int = 10

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
	timestamp := uint64(now.UnixNano())
	_, err := storage.db.Insert("TXS", []interface{}{nil, owner, token, timestamp, strconv.FormatInt(int64(amount), 10)})
	if err != nil {
		return err
	}
	return nil

}

func (storage *TTStorage) Read(owner, token string) (int64, error) {
	// SELECT SUM(amount) ... doesn't work???? Problem with numbers/integer
	// Now numbers are stored as strings....
	//resp, err := storage.db.Execute("SELECT amount FROM txs WHERE owner = ? and token = ?", []interface{}{owner, token})

	var amount int64 = 0
	const limit uint32 = 5
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
			amount = amount + n
		}
		offset = offset + limit
	}
	return amount, nil
}
