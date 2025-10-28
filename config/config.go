package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env         string `mapstructure:"ENV"`
	LoggerLevel string `mapstructure:"LOGGER_LEVEL"`

	// Database config
	Dsn string `mapstructure:"DSN"`

	// JWT config
	JwtSecret      string        `mapstructure:"JWT_SECRET"`
	JwtExpDiration time.Duration `mapstructure:"JWT_EXP_DURATION"`
	JwtIssuer      string        `mapstructure:"JWT_ISSUER"`

	// Server config
	ServerPort             string `mapstructure:"SERVER_PORT"`
	ServerReadTimeout      int    `mapstructure:"SERVER_READ_TIMEOUT"`
	ServerWriteTimeout     int    `mapstructure:"SERVER_WRITE_TIMEOUT"`
	ServerShutdownDeadline int    `mapstructure:"SERVER_SHUTDOWN_DEADLINE"`

	// CORS config
	CorsAllowedOrigins   []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	CorsAllowedMethods   []string `mapstructure:"CORS_ALLOWED_METHODS"`
	CorsAllowedHeaders   []string `mapstructure:"CORS_ALLOWED_HEADERS"`
	CorsAllowCredentials bool     `mapstructure:"CORS_ALLOW_CREDENTIALS"`
	CorsDebug            bool     `mapstructure:"CORS_DEBUG"`
	CorsExposedHeaders   string   `mapstructure:"CORS_EXPOSED_HEADERS"`
}

func GetConfig() (conf Config, err error) {

	viper.AddConfigPath(".")
	viper.SetConfigName(".env.local") // default env file
	viper.SetConfigType("env")

	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		viper.SetDefault("ENV", "local")
		viper.SetDefault("LOGGER_LEVEL", "local")
		viper.SetDefault("DSN", "postgres://enginecareuser:mypassword@postgres:5432/enginecaredb?sslmode=disable")
		viper.SetDefault("JWT_SECRET", "our_secret_key")
		viper.SetDefault("JWT_EXP_DURATION", 24)
		viper.SetDefault("JWT_ISSUER", "enginecare-api")
		viper.SetDefault("SERVER_PORT", "4000")
		viper.SetDefault("SERVER_READ_TIMEOUT", 15)
		viper.SetDefault("SERVER_WRITE_TIMEOUT", 15)
		viper.SetDefault("SERVER_SHUTDOWN_DEADLINE", 15)

		// CORS config defaults
		viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")
		viper.SetDefault("CORS_ALLOWED_METHODS", "DELETE,GET,OPTIONS,POST,PUT")
		viper.SetDefault("CORS_ALLOWED_HEADERS", "*")
		viper.SetDefault("CORS_ALLOW_CREDENTIALS", "true")
		viper.SetDefault("CORS_DEBUG", "false")
		viper.SetDefault("CORS_EXPOSED_HEADERS", "*")
	}

	err = viper.Unmarshal(&conf)
	if err != nil {
		return Config{}, err
	}

	return conf, nil
}
