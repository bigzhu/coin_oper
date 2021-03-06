package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	bitcoind "github.com/lomocoin/go-bitcoind"
	"github.com/name5566/leaf/log"
	"github.com/spf13/viper"
)

var DB = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	viper.SetConfigName("conf")
	viper.AddConfigPath(".")
	viper.SetConfigType("json")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	Host := viper.GetString("Coins.BTC.Host")
	RPCPort := viper.GetInt("Coins.BTC.RPCPort")
	RPCUser := viper.GetString("Coins.BTC.RPCUser")
	RPCPassword := viper.GetString("Coins.BTC.RPCPassword")

	BC, err := bitcoind.New(Host, RPCPort, RPCUser, RPCPassword, false)
	if err != nil {
		log.Fatal(err.Error())
		panic(fmt.Errorf("Fatal create bitcoind: %s", err))
	}
	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.GET("/listunspent", func(c *gin.Context) {

		transactions, err := BC.ListUnspent(0, 999999)
		if err != nil {
			log.Error(err.Error())
		}
		// c.String(200, "pong")
		c.JSON(200, gin.H{"err": err, "data": transactions})
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := DB[user]
		if ok {
			c.JSON(200, gin.H{"user": user, "value": value})
		} else {
			c.JSON(200, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			DB[user] = json.Value
			c.JSON(200, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":3001")
}
