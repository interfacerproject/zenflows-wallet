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
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
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
}

type Wallet struct {
	Storage Storage
	Config  *Config
}

type AddTokens struct {
	Owner  string `json:"owner"`
	Amount string  `json:"amount"`
	Token  string `json:"token"`
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, zenflows-sign, zenflows-id")
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

	// Verify signature request
	zenroomData := ZenroomData{
		Gql:            b64.StdEncoding.EncodeToString(body),
		EdDSASignature: c.Request.Header.Get("did-sign"),
		EdDSAPublicKey: c.Request.Header.Get("did-pk"),
	}

	var addTokens AddTokens

	if err := json.Unmarshal(body, &addTokens); err != nil {
		result["error"] = err.Error()
		return
	}

	amount, err := strconv.ParseInt(addTokens.Amount, 10, 64)
	if err != nil {
		result["error"] = fmt.Sprintln("Not a number ", amount)
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

	if err := wallet.Storage.AddDiff(addTokens.Owner, addTokens.Token, amount); err != nil {
		result["error"] = err.Error()
		return
	}

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
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(CORS())
	r.POST("/token", wallet.addTokensHandler)
	r.GET("/token/:token/:owner", wallet.getTokenHandler)
	r.GET("/token/:token/:owner/last/:n", wallet.getTxsHandler)
	r.Run()
}
