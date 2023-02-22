package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/lib/pq"
)

var connection *sql.DB

func Connect() {
	var e error
	connection, e = sql.Open("postgres",
		`host=10.14.206.28
		port=5432 user=student password=1234
		dbname=test05 sslmode=disable
		`)
	if e != nil {
		panic(e.Error())
	}

	e = connection.Ping()
	if e != nil {
		panic(e.Error())
	}
}

func IsExistsName(name string) bool {
	row := connection.QueryRow(`SELECT COUNT(*) <> 0 FROM "Player" WHERE "Login"=$1`, name)

	var isExists bool
	e := row.Scan(&isExists)
	if e != nil {
		fmt.Println(e.Error())
		return true
	}

	return isExists
}

func InsertPlayerName(player *PlayerType) {
	encryptedPassword, _ := EncryptPassword(player.Password)
	_, e := connection.Exec(
		`INSERT INTO "Player"("Login", "Password", "Nickname", "RegistrationDate") VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`,
		player.Login, encryptedPassword, player.Nickname)
	if e != nil {
		fmt.Println(e.Error())
	}
}

func SingIn(player *PlayerType) bool {
	encryptedPassword, _ := EncryptPassword(player.Password)

	row := connection.QueryRow(
		`SELECT "Nickname" FROM "Player" 
                  WHERE "Login"=$1 AND "Password"=$2`,
		player.Login, encryptedPassword)

	e := row.Scan(&player.Nickname)
	fmt.Println(e)
	return e == nil
}

func EncryptPassword(value string) (string, error) {
	hash := sha256.New()
	_, e := hash.Write([]byte(value + "SomeSecretKey"))
	if e != nil {
		return "", e
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func InsertMessage(message *Message) {
	row := connection.QueryRow(`INSERT INTO "Message"("Player", "Text", "Date") VALUES ($1, $2, CURRENT_TIMESTAMP) returning "Date"`, message.Name, message.Message)
	e := row.Scan(&message.Time)
	if e != nil {
		fmt.Println(e.Error())
	}
}

func SelectUsers() []string {
	rows, e := connection.Query(`SELECT "Nickname" FROM "Player" ORDER BY "Nickname"`)
	if e != nil {
		fmt.Println(e)
		return nil
	}

	defer rows.Close()

	users := make([]string, 0)
	var name string
	for rows.Next() {
		e = rows.Scan(&name)
		if e != nil {
			fmt.Println(e)
			return nil
		}

		users = append(users, name)
	}

	return users
}
