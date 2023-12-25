package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	_defaultLogLevel = "info"
	_defaultENV      = "local"
)

func initConfig() {
	viper.SetDefault("ENV", _defaultENV)
	viper.SetDefault("LOG_LEVEL", _defaultLogLevel)
	viper.SetDefault("HTTP_SERVER_LISTEN_ADDR", "8082")
	viper.SetDefault("SECRET_KEY", "VERY_SECRET_KEY")
	viper.SetDefault("ACCESS_TOKEN_EXPIRATION", "15m")
	viper.SetDefault("REFRESH_TOKEN_EXPIRATION", "24h")
	viper.SetDefault("HTTP_SERVER_REQUEST_TIMEOUT", "5s")
	viper.SetDefault("MONGO_URI", "mongodb://localhost:27017")

	viper.AutomaticEnv()
}

type Config struct {
	ENV                    string
	LogLevel               string
	Mongo                  *Mongo
	SecretKey              string
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
	HTTPServer             *HTTPServer
}

func NewConfig() *Config {
	initConfig()

	return &Config{
		ENV:                    viper.GetString("ENV"),
		LogLevel:               viper.GetString("LOG_LEVEL"),
		SecretKey:              viper.GetString("SECRET_KEY"),
		AccessTokenExpiration:  viper.GetDuration("ACCESS_TOKEN_EXPIRATION"),
		RefreshTokenExpiration: viper.GetDuration("REFRESH_TOKEN_EXPIRATION"),
		Mongo:                  NewMongo(),
		HTTPServer:             NewHTTPServer(),
	}
}

type Mongo struct {
	URI string
}

func NewMongo() *Mongo {
	return &Mongo{
		URI: viper.GetString("MONGO_URI"),
	}
}

type HTTPServer struct {
	ListenAddr     string
	RequestTimeout time.Duration
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		ListenAddr:     viper.GetString("HTTP_SERVER_LISTEN_ADDR"),
		RequestTimeout: viper.GetDuration("HTTP_SERVER_REQUEST_TIMEOUT"),
	}
}
