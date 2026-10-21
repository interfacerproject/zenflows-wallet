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
	"context"
	"crypto/rand"
	_ "embed"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/interfacerproject/zenflows-wallet/payout"
)

type Config struct {
	Port       int
	Host       string
	DidUrl     string
	DidContext string
	TTHost     string
	TTUser     string
	TTPass     string
}

type Transaction struct {
	Id        uint64 `json:"id"`
	Timestamp uint64 `json:"timestamp"`
	Amount    int64  `json:"amount"`
}

type Storage interface {
	AddDiff(string, string, int64) error
	Read(string, string, uint64) (int64, error)
	ReadTxs(string, string, int) ([]Transaction, error)
	AlreadyPaid(string) (bool, error)
	MarkPaid(string) error
}

type Wallet struct {
	Storage Storage
	Config  *Config
	Pdp     *payout.Client
}

// issueRequest is the event-backed body: the caller names the event that
// justifies the issuance. No owner/token/amount — those are derived by the PDP.
type issueRequest struct {
	Event payout.EventRef `json:"event"`
}

func newRequestID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, did-pk, did-sign")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (wallet *Wallet) addTokensHandler(c *gin.Context) {
	// Setup json response
	result := map[string]interface{}{
		"success": false,
	}
	defer c.JSON(http.StatusOK, result)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		result["error"] = err.Error()
		return
	}

	// Local authn floor: the body must be signed by the caller's DID key.
	zenroomData := ZenroomData{
		Gql:            b64.StdEncoding.EncodeToString(body),
		EdDSASignature: c.Request.Header.Get("did-sign"),
		EdDSAPublicKey: c.Request.Header.Get("did-pk"),
	}

	var req issueRequest
	if err := json.Unmarshal(body, &req); err != nil {
		result["error"] = err.Error()
		return
	}
	if req.Event.ID == "" || req.Event.Type == "" {
		result["error"] = "missing event reference"
		return
	}

	if err := zenroomData.verifyDid(
		wallet.Config.DidContext,
		wallet.Config.DidUrl); err != nil {
		result["error"] = err.Error()
		return
	}
	if err := zenroomData.isAuth(); err != nil {
		result["error"] = err.Error()
		return
	}

	// Idempotency: never pay the same event twice. This check→mint→MarkPaid is
	// not atomic: two concurrent issuances for the same event id (e.g. a
	// double-submitted accept — the accept mutation itself is idempotent, so
	// the id is identical) could both pass this guard and double-credit. It
	// fails toward a rare over-credit on a single-user action, never a security
	// hole or an unbacked mint. Accepted for this slice; it hardens when the
	// authoritative payout ledger moves Zenflows-side (design §5).
	paid, err := wallet.Storage.AlreadyPaid(req.Event.ID)
	if err != nil {
		result["error"] = err.Error()
		return
	}
	if !payout.ShouldMint(paid) {
		result["success"] = true
		result["already_paid"] = true
		return
	}

	// Authorize + derive payouts from the owning authority (the Zenflows PDP).
	// owner/token/amount are NOT taken from the body; they come from the
	// attestation of the referenced event.
	proof := payout.Proof{
		Signature:  c.Request.Header.Get("did-sign"),
		PayloadB64: b64.StdEncoding.EncodeToString(body),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	resp, derr := wallet.Pdp.Decide(ctx,
		payout.BuildIssueRequest(newRequestID(), c.Request.Header.Get("did-pk"), proof, req.Event))
	out := payout.InterpretPayouts(resp, derr)
	if !out.Mint {
		result["error"] = "not authorized for this issuance"
		result["reason"] = out.ReasonCode
		return
	}

	// Mint exactly the attested payouts, then mark the event paid.
	for _, p := range out.Payouts {
		if err := wallet.Storage.AddDiff(p.Recipient, p.Token, p.Amount); err != nil {
			result["error"] = err.Error()
			return
		}
	}
	if err := wallet.Storage.MarkPaid(req.Event.ID); err != nil {
		result["error"] = err.Error()
		return
	}

	log.Printf("issued points for event %s (%s): %d payouts", req.Event.ID, req.Event.Type, len(out.Payouts))
	result["success"] = true
	return
}

func (wallet *Wallet) getTokenHandler(c *gin.Context) {
	// Setup json response
	result := map[string]interface{}{
		"success": false,
	}
	defer c.JSON(http.StatusOK, result)

	token := c.Param("token")
	owner := c.Param("owner")

	until, err := strconv.ParseUint(c.DefaultQuery("until", "0"), 10, 64)
	if err != nil {
		result["error"] = fmt.Sprintln("Not a number ", until)
		return
	}

	if val, err := wallet.Storage.Read(owner, token, until); err != nil {
		result["error"] = err.Error()
	} else {
		result["success"] = true
		result["amount"] = strconv.FormatInt(val, 10)
	}

	return
}

func (wallet *Wallet) getTxsHandler(c *gin.Context) {
	// Setup json response
	result := map[string]interface{}{
		"success": false,
	}
	defer c.JSON(http.StatusOK, result)

	token := c.Param("token")
	owner := c.Param("owner")

	n, err := strconv.Atoi(c.Param("n"))
	if err != nil || n < 0 {
		result["error"] = fmt.Sprintln("Not a number ", n)
		return
	}

	if val, err := wallet.Storage.ReadTxs(owner, token, n); err != nil {
		result["error"] = err.Error()
	} else {
		result["success"] = true
		result["txs"] = val
	}

	return
}

// TODO: improve parsing and validation of env vars
func loadEnvConfig() Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	return Config{
		Host:       os.Getenv("HOST"),
		Port:       port,
		DidUrl:     os.Getenv("DID_URL"),
		DidContext: os.Getenv("DID_CONTEXT"),
		TTHost:     os.Getenv("TT_HOST"),
		TTUser:     os.Getenv("TT_USER"),
		TTPass:     os.Getenv("TT_PASS"),
	}
}

func main() {
	config := loadEnvConfig()
	log.Printf("Using DID %s\n", config.DidUrl)

	storage := &TTStorage{}
	err := storage.Init(config.TTHost, config.TTUser, config.TTPass)
	if err != nil {
		log.Fatal(err.Error())
	}
	wallet := Wallet{
		Storage: storage,
		Config:  &config,
		Pdp:     payout.NewClientFromEnv(),
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(CORS())
	r.POST("/token", wallet.addTokensHandler)
	r.GET("/token/:token/:owner", wallet.getTokenHandler)
	r.GET("/token/:token/:owner/last/:n", wallet.getTxsHandler)
	r.Run()
}
