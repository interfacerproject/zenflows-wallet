//go:build ignore

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

	"github.com/go-redis/redis/v8"
	"context"
)

type Config struct {
	Port        int
	Host        string
	ZenflowsUrl string
	RedisUrl    string
}

type Wallet struct {
	Ctx context.Context
	RDB *redis.Client
	ZenflowsUrl string
}

type AddTokens struct {
	Owner  string `json:"owner"`
	Amount int64    `json:"amount"`
	Token  string `json:"token"`
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
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
		EdDSASignature: c.Request.Header.Get("zenflows-sign"),
	}

	var addTokens AddTokens
	
	if err := json.Unmarshal(body, &addTokens); err != nil {
		result["error"] = err.Error()
		return
	}
	if err := zenroomData.requestPublicKey(wallet.ZenflowsUrl, addTokens.Owner); err != nil {
		result["error"] = err.Error()
		return
	}
	
	if err := zenroomData.isAuth(); err != nil {
		result["error"] = err.Error()
		return
	}

	key := fmt.Sprintf("%s:%s", addTokens.Owner, addTokens.Token)

	if err := wallet.RDB.IncrBy(wallet.Ctx, key, addTokens.Amount).Err(); err != nil {
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

	key := fmt.Sprintf("%s:%s", owner, token) 

	if val, err := wallet.RDB.Get(wallet.Ctx, key).Result(); err != nil {
		result["error"] = err.Error()
	} else if amount, err := strconv.ParseInt(val, 10, 64); err != nil {
		result["error"] = err.Error()
	} else {
		result["success"] = true
		result["amount"] = amount
	}

	return
}
func loadEnvConfig() Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	return Config{
		Host:        os.Getenv("HOST"),
		Port:        port,
		ZenflowsUrl: fmt.Sprintf("%s/api", os.Getenv("ZENFLOWS_URL")),
		RedisUrl:   os.Getenv("REDIS_URL"),
	}
}

func main() {
	config := loadEnvConfig()
	log.Printf("Using backend %s\n", config.ZenflowsUrl)

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisUrl,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	wallet := Wallet{ 
		RDB: rdb,
		Ctx: ctx,
		ZenflowsUrl: config.ZenflowsUrl,
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(CORS())
	r.POST("/token", wallet.addTokensHandler)
	r.GET("/token/:token/:owner", wallet.getTokenHandler)
	r.Run()
}
