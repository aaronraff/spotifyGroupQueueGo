package main

import (
	"log"
	"encoding/json"
	"golang.org/x/oauth2"
	"database/sql"
	_ "github.com/lib/pq"
)

func GetTokenFromCode(db *sql.DB, roomCode string) *oauth2.Token {	
	row := db.QueryRow("SELECT oauth_token from rooms WHERE room_code = $1", roomCode)

	tok := &oauth2.Token{}
	res := new([]byte)
	if err := row.Scan(res); err != nil {
		if err != sql.ErrNoRows {
			log.Println(err)
		}

		return nil
	}

	json.Unmarshal(*res, tok)

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

	if err != nil {
		log.Println(err)
		return
	}

	_, err = db.Exec("INSERT INTO rooms VALUES ($1, $2, $3)", roomCode, userID, tokJ)

	if err != nil {
		log.Println(err)
	}
}

