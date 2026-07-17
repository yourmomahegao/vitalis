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
	DATABASE_ADDRESS  string
	DATABASE_PORT     int
	DATABASE_NAME     string
	DATABASE_USER     string
	DATABASE_PASSWORD string

	RUN_ADDRESS string
	SECRET_KEY  string

	COLLECT_CPU_INFO_INTERVAL_SECONDS  int
	COLLECT_RAM_INFO_INTERVAL_SECONDS  int
	COLLECT_NET_INFO_INTERVAL_SECONDS  int
	COLLECT_FILE_INFO_INTERVAL_SECONDS int

	MAX_INFO_GROUPS_AMOUNT  int
	MAX_SESSION_KEYS_AMOUNT int

	ACCESS_TOKEN_LIFETIME_MINUTES int
}

var ENV EnviromentStruct

func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func getStringEnviromentVariable(name string, defaultValue string) string {
	variable := os.Getenv(name)

	if variable == "" {
		variable = defaultValue
		log.Printf("%s enviroment variable not defined, using %s", name, defaultValue)
	}

	return variable
}

func getIntEnviromentVariable(name string, defaultValue int) int {
	variableRaw := os.Getenv(name)
	variable := defaultValue

	if variableRaw == "" {
		log.Printf("%s enviroment variable not defined, using %d", name, defaultValue)
	} else {
		variableConverted, err := strconv.Atoi(variableRaw)

		if err != nil {
			log.Printf("%s enviroment variable is invalid, using %d", name, defaultValue)
		} else {
			variable = variableConverted
		}
	}

	return variable
}

func Preload() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ENV.DATABASE_ADDRESS = getStringEnviromentVariable("DATABASE_ADDRESS", "localhost")
	ENV.DATABASE_PORT = getIntEnviromentVariable("DATABASE_PORT", 5432)
	ENV.DATABASE_NAME = getStringEnviromentVariable("DATABASE_NAME", "vitalis")
	ENV.DATABASE_USER = getStringEnviromentVariable("DATABASE_USER", "root")
	ENV.DATABASE_PASSWORD = getStringEnviromentVariable("DATABASE_PASSWORD", "")

	ENV.RUN_ADDRESS = getStringEnviromentVariable("RUN_ADDRESS", "0.0.0.0:8080")
	ENV.SECRET_KEY = getStringEnviromentVariable("SECRET_KEY", "")

	ENV.COLLECT_CPU_INFO_INTERVAL_SECONDS = getIntEnviromentVariable("COLLECT_CPU_INFO_INTERVAL_SECONDS", 30)
	ENV.COLLECT_RAM_INFO_INTERVAL_SECONDS = getIntEnviromentVariable("COLLECT_RAM_INFO_INTERVAL_SECONDS", 30)
	ENV.COLLECT_NET_INFO_INTERVAL_SECONDS = getIntEnviromentVariable("COLLECT_NET_INFO_INTERVAL_SECONDS", 30)
	ENV.COLLECT_FILE_INFO_INTERVAL_SECONDS = getIntEnviromentVariable("COLLECT_FILE_INFO_INTERVAL_SECONDS", 30)

	ENV.MAX_INFO_GROUPS_AMOUNT = getIntEnviromentVariable("MAX_INFO_GROUPS_AMOUNT", 100)
	ENV.MAX_SESSION_KEYS_AMOUNT = getIntEnviromentVariable("MAX_SESSION_KEYS_AMOUNT", 60)

	ENV.ACCESS_TOKEN_LIFETIME_MINUTES = getIntEnviromentVariable("ACCESS_TOKEN_LIFETIME_MINUTES", 60)

	if ENV.SECRET_KEY == "" {
		log.Printf("SECRET_KEY not defined, generating new one...")

		log.Printf("=================================")
		log.Printf("Save and use this secret key: %s", generateSecureToken(32))
		log.Printf("=================================")

		return
	}
}
