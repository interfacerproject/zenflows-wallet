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
	Port        int
	Host        string
	ZenflowsUrl string
	RedisUrl    string
	TTHost      string
	TTUser      string
	TTPass      string
}

type Storage interface {
	AddDiff(string, string, int64) error
	Read(string, string) (int64, error)
}

type Wallet struct {
	Storage Storage
	Config  *Config
}

type AddTokens struct {
	Owner  string `json:"owner"`
	Amount int64  `json:"amount"`
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
	fmt.Println(body)

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
	if err := zenroomData.requestPublicKey(wallet.Config.ZenflowsUrl, c.Request.Header.Get("zenflows-id")); err != nil {
		result["error"] = err.Error()
		return
	}
	fmt.Println(c.Request.Header)

	if err := zenroomData.isAuth(); err != nil {
		result["error"] = err.Error()
		return
	}

	if err := wallet.Storage.AddDiff(addTokens.Owner, addTokens.Token, addTokens.Amount); err != nil {
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

	if val, err := wallet.Storage.Read(owner, token); err != nil {
		result["error"] = err.Error()
	} else {
		result["success"] = true
		result["amount"] = val
	}

	return
}
func loadEnvConfig() Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	return Config{
		Host:        os.Getenv("HOST"),
		Port:        port,
		ZenflowsUrl: fmt.Sprintf("%s/api", os.Getenv("ZENFLOWS_URL")),
		TTHost:      os.Getenv("TT_HOST"),
		TTUser:      os.Getenv("TT_USER"),
		TTPass:      os.Getenv("TT_PASS"),
	}
}

func main() {
	config := loadEnvConfig()
	log.Printf("Using backend %s\n", config.ZenflowsUrl)

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
	r.Run()
}
