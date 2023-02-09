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
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	zenroom "github.com/dyne/Zenroom/bindings/golang/zenroom"
)

const GQL_PERSON_PUBKEY string = "query($id: ID!) {personPubkey(id: $id)}"

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

// Fills ZenroomData with the public key requested to zenflows (from the email)
func (data *ZenroomData) requestPublicKey(url string, id string) error {
	query, err := json.Marshal(map[string]interface{}{
		"query": GQL_PERSON_PUBKEY,
		"variables": map[string]string{
			"id": id,
		},
	})
	resp, err := http.Post(url, "application/json", bytes.NewReader(query))
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result map[string]map[string]string
	json.Unmarshal(body, &result)
	if result["data"]["personPubkey"] == "" {
		return errors.New(string(body))
	}
	data.EdDSAPublicKey = result["data"]["personPubkey"]
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
