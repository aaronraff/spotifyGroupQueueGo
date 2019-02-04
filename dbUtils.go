package main

import (
	"log"
	"math/rand"
	"time"
	"os"
	"encoding/json"
	"golang.org/x/oauth2"
	"database/sql"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/nacl/secretbox"
)

var encryptionKey [32]byte 

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	var key [32]byte
	orig := []byte(os.Getenv("ENCRYPTION_KEY"))
	copy(key[:], orig)
	encryptionKey = key
}

func GetAllRoomCodes(db *sql.DB) []string {
	rows, err := db.Query("SELECT room_code FROM rooms")
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		} else {
			log.Println(err)
		}
	}

	res := make([]string, 0)
	
	for rows.Next() {
		val := ""
		if err := rows.Scan(&val); err != nil {
			log.Println(err)
			continue
		}

		res = append(res, val)
	}

	return res
}

func GetTokenFromCode(db *sql.DB, roomCode string) *oauth2.Token {	
	row := db.QueryRow("SELECT oauth_token FROM rooms WHERE room_code = $1", roomCode)

	var tokString []byte
	tok := &oauth2.Token{}
	if err := row.Scan(&tokString); err != nil {
		if err != sql.ErrNoRows {
			log.Println(err)
		}

		return nil
	}

	// Nonce is stored in first 24 chars
	var nonce [24]byte
	copy(nonce[:], tokString[:24])

	res, ok := secretbox.Open(nil, tokString[24:], &nonce, &encryptionKey)

	if !ok {
		log.Println("The message could not be decrypted")
	}

	json.Unmarshal(res, tok)
	return tok
}

func DoesRoomExist(db *sql.DB, roomCode string) bool {	
	row := db.QueryRow("SELECT user_id from rooms WHERE room_code = $1", roomCode)

	res := new(string)
	if err := row.Scan(res); err != nil {
		return false
	}

	return true
}

func GetRoomCode(db *sql.DB, userID string) string {
	row := db.QueryRow("SELECT room_code from rooms WHERE user_id = $1", userID)

	roomCode := new(string)
	if err := row.Scan(roomCode); err != nil {
		if err != sql.ErrNoRows {
			log.Println(err)
		}

		return "The room is not active."
	}

	return *roomCode
}

func DeleteRoom(db *sql.DB, userID string) {
	_, err := db.Exec("DELETE FROM rooms WHERE user_id = $1", userID)

	if err != nil {
		log.Println(err)
	}
}

func InsertRoom(db *sql.DB, roomCode string, userID string, tok *oauth2.Token) {
	tokJ, err := json.Marshal(tok)	
	tokS := string(tokJ)

	if err != nil {
		log.Println(err)
		return
	}
	
	nonce := make([]byte, 24)
	rand.Read(nonce)
	var sizedNonce [24]byte
	copy(sizedNonce[:], nonce)

	// Store nonce at the beginning of the message
	out := make([]byte, 24)
	copy(out, nonce)

	res := secretbox.Seal(out, []byte(tokS), &sizedNonce, &encryptionKey)
	_, err = db.Exec("INSERT INTO rooms VALUES ($1, $2, $3)", roomCode, userID, res)

	if err != nil {
		log.Println(err)
	}
}

