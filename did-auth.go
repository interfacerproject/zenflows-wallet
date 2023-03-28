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
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	zenroom "github.com/dyne/Zenroom/bindings/golang/zenroom"
)

//go:embed zenflows-crypto/src/verify_graphql.zen
var VERIFY string

// Input and output of sign_graphql.zen
type ZenroomData struct {
	Gql            string `json:"gql"`
	EdDSASignature string `json:"eddsa_signature"`
	EdDSAPublicKey string `json:"eddsa_public_key"`
}

type ZenroomResult struct {
	Output []string `json:"output"`
}

func (data *ZenroomData) verifyDid(context string, baseUrl string) error {
	// TODO: improve URL management and concat
	url := fmt.Sprintf("%s%s:%s", baseUrl, context, data.EdDSAPublicKey)
	log.Printf("Fetching %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Problem fetching DID, status: %d", resp.StatusCode)
	}
	return nil
}

// Used to verify the signature with `zenflows-crypto`
func (data *ZenroomData) isAuth() error {
	var err error

	jsonData, _ := json.Marshal(data)

	// Verify the signature
	result, success := zenroom.ZencodeExec(VERIFY, "", string(jsonData), "")
	if !success {
		return errors.New(result.Logs)
	}
	var zenroomResult ZenroomResult
	err = json.Unmarshal([]byte(result.Output), &zenroomResult)
	if err != nil {
		return err
	}
	if zenroomResult.Output[0] != "1" {
		return errors.New("Signature is not authentic")
	}
	return nil
}
