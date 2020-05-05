package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 13
)

// HashPassword input string output Hash string and error
func HashPassword(pwd string) (string, error) {
	hpwdm, err := bcrypt.GenerateFromPassword([]byte(pwd), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hpwdm), nil
}

// ComparePassword input 2 string Check Hash Compare output true & false
func ComparePassword(hpwd, pwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hpwd), []byte(pwd))
	return err == nil
}

// RandStr random string input size
func RandStr(strSize int) string {

	var dictionary string

	dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

//GetNewID Create New ID (input Table)
func GetNewID(q Queryer, table string) (int64, error) {

	var id int64
	err := q.QueryRow("Select id From " + table + " Order BY id DESC").Scan(&id)
	if err == sql.ErrNoRows {
		return 1000000000000000, nil
	}

	if err != nil {
		return 0, err
	}

	return id + 1, err
}

//convJSON convert interface ti json
func convJSON(is interface{}) []byte {
	json, _ := json.Marshal(is)
	return json
}
