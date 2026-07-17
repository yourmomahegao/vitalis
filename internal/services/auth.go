package services

import (
	"encoding/hex"
	"log"
	"math/rand"
	"time"
	"vitalis/internal/database"
	"vitalis/internal/enviroment"
)

func CheckSessionKeyUnique(sessionKey string) (bool, error) {
	rows, err := database.Database.Query(`select * 
							from auth_session_keys 
							where session_key = $1`, sessionKey)

	if err != nil {
		log.Printf("Error occured in CheckSessionKeyUnique(): %v", err)
		return false, err
	}
	defer rows.Close()

	countRows := 0

	for rows.Next() {
		countRows++
	}

	if countRows == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func CheckSessionKey(sessionKey string) (bool, error) {
	currentTimestamp, err := database.GetDatabaseTime()
	if err != nil {
		log.Printf("Error while getting current timestamp in CheckSessionKey(): %v", err)
		return false, err
	}

	rows, err := database.Database.Query(`select id from auth_session_keys where session_key = $1 and valid_until > $2`, sessionKey, currentTimestamp)

	if err != nil {
		log.Printf("Error while getting session keys in CheckSessionKey(): %v", err)
		return false, err
	}

	var count int

	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			log.Printf("Error while scan in CheckSessionKey(): %v", err)
			return false, err
		}
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func SaveSessionKey(sessionKey string) error {
	currentTimestamp, err := database.GetDatabaseTime()
	if err != nil {
		log.Printf("Error occured in SaveSessionKey(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`delete from auth_session_keys 
										where valid_until < $1 or id in (select 
																			id from (select row_number() over (order by id desc) as id_rn, 
																			id 
																		from auth_session_keys) 
																			where id_rn > $2)`, currentTimestamp, enviroment.ENV.MAX_SESSION_KEYS_AMOUNT-1)

	if err != nil {
		log.Printf("Error while deleting old session keys in SaveSessionKey(): %v", err)
		return err
	}

	validUntil := currentTimestamp.Add(time.Duration(enviroment.ENV.ACCESS_TOKEN_LIFETIME_MINUTES) * time.Minute)

	_, err = database.Database.Exec(`insert into auth_session_keys 
										(session_key, valid_until)
										values ($1, $2)`, sessionKey, validUntil)

	if err != nil {
		log.Printf("Error while saving new session key in SaveSessionKey(): %v", err)
		return err
	}

	return nil
}

func GenerateSessionKey() (string, error) {
	sessionKey := GenerateSecureToken(16)

	uniqueStatus, err := CheckSessionKeyUnique(sessionKey)

	if err != nil {
		log.Printf("Error occured in GenerateSessionKey(): %v", err)
		return "", err
	}

	if !uniqueStatus {
		return GenerateSessionKey()
	}

	err = SaveSessionKey(sessionKey)
	if err != nil {
		log.Printf("Error occured in GenerateSessionKey(): %v", err)
		return "", err
	}

	return sessionKey, nil
}
