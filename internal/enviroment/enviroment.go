package enviroment

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type EnviromentStruct struct {
	RunAddress string

	DatabaseAddress  string
	DatabasePort     string
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string

	SecretKey string

	AccessTokenLifetimeMinutes int
	MaxInfoRowsAmount          int
	MaxSessionKeysAmount       int
}

var Env EnviromentStruct

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func Preload() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// ========== DATABASE_ADDRESS ==========
	Env.DatabaseAddress = os.Getenv("DATABASE_ADDRESS")
	if Env.DatabaseAddress == "" {
		Env.DatabaseAddress = "localhost"
		log.Printf("DATABASE_ADDRESS not defined, using localhost:5432")
	}

	// ========== DATABASE_PORT ==========
	Env.DatabasePort = os.Getenv("DATABASE_PORT")
	if Env.DatabasePort == "" {
		Env.DatabasePort = "5432"
		log.Printf("DATABASE_PORT not defined, using 5432")
	}

	// ========== DATABASE_NAME ==========
	Env.DatabaseName = os.Getenv("DATABASE_NAME")
	if Env.DatabaseName == "" {
		Env.DatabaseName = "vitalis"
		log.Printf("DATABASE_NAME not defined, using vitalis")
	}

	// ========== DATABASE_USER ==========
	Env.DatabaseUser = os.Getenv("DATABASE_USER")
	if Env.DatabaseUser == "" {
		Env.DatabaseUser = "root"
		log.Printf("DATABASE_USER not defined, using root")
	}

	// ========== DATABASE_PASSWORD ==========
	Env.DatabasePassword = os.Getenv("DATABASE_PASSWORD")
	if Env.DatabasePassword == "" {
		Env.DatabasePassword = "1"
		log.Printf("DATABASE_PASSWORD not defined, using 1")
	}

	// ========== RUN_ADDRESS ==========

	Env.RunAddress = os.Getenv("RUN_ADDRESS")
	if Env.RunAddress == "" {
		Env.RunAddress = "0.0.0.0:8080"
		log.Printf("RUN_ADDRESS not defined, using 0.0.0.0:8080")
	}

	// ========== MAX_INFO_ROWS_AMOUNT ==========

	maxInfoRowsAmount := 10000
	maxInfoRowsAmountRaw := os.Getenv("MAX_INFO_ROWS_AMOUNT")

	if maxInfoRowsAmountRaw == "" {
		log.Printf("MAX_INFO_ROWS_AMOUNT not defined, using default (10000).")
	} else {
		var err error
		maxInfoRowsAmount, err = strconv.Atoi(maxInfoRowsAmountRaw)

		if err != nil {
			log.Printf("MAX_INFO_ROWS_AMOUNT is invalid, using default (10000).")
		}
	}

	Env.MaxInfoRowsAmount = maxInfoRowsAmount

	// ========== MAX_SESSION_KEYS_AMOUNT ==========

	maxSessionKeysAmount := 100
	maxSessionKeysAmountRaw := os.Getenv("MAX_SESSION_KEYS_AMOUNT")

	if maxSessionKeysAmountRaw == "" {
		log.Printf("MAX_SESSION_KEYS_AMOUNT not defined, using default (100).")
	} else {
		var err error
		maxSessionKeysAmount, err = strconv.Atoi(maxSessionKeysAmountRaw)

		if err != nil {
			log.Printf("MAX_SESSION_KEYS_AMOUNT is invalid, using default (100).")
		}
	}

	Env.MaxSessionKeysAmount = maxSessionKeysAmount

	// ========== SECRET_KEY ==========

	Env.SecretKey = os.Getenv("SECRET_KEY")

	if Env.SecretKey == "" {
		log.Printf("SECRET_KEY not defined, generating new one...")

		log.Printf("=================================")
		log.Printf("Save and use this secret key: %s", GenerateSecureToken(32))
		log.Printf("=================================")

		return
	}

	// ========== ACCESS_TOKEN_LIFETIME_MINUTES ==========

	accessTokenLifetimeMinutes := 60
	accessTokenLifetimeMinutesRaw := os.Getenv("ACCESS_TOKEN_LIFETIME_MINUTES")

	if accessTokenLifetimeMinutesRaw == "" {
		log.Printf("ACCESS_TOKEN_LIFETIME_MINUTES not defined, using default (60 minutes).")
	} else {
		var err error
		accessTokenLifetimeMinutes, err = strconv.Atoi(accessTokenLifetimeMinutesRaw)

		if err != nil {
			log.Printf("ACCESS_TOKEN_LIFETIME_MINUTES is invalid, using default (60 minutes).")
		}
	}

	Env.AccessTokenLifetimeMinutes = accessTokenLifetimeMinutes
}
