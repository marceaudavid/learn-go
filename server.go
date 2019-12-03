package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var key = securecookie.GenerateRandomKey(32)
var store = sessions.NewCookieStore([]byte(key))

// Subscriber ...
type Subscriber struct {
	Username string
	Password string
	Email    string
}

// User ...
type User struct {
	Username string
	Password string
}

func main() {
	DSN := os.Getenv("DSN")
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// db.Exec("DROP TABLE IF EXISTS users")
	// db.Exec("CREATE TABLE IF NOT EXISTS users (username VARCHAR(255), password VARCHAR(255), email VARCHAR(255))")

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%f", rand.Float64())
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var user User
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&user)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				fmt.Fprintf(w, "%s", err.Error())
			} else {
				var login string
				err := db.QueryRow("SELECT username FROM users WHERE username = ? AND password = ?", user.Username, user.Password).Scan(&login)
				if err != nil {
					if err.Error() == "sql: no rows in result set" {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
					}
				} else {
					session, _ := store.Get(r, "auth")
					session.Values["authenticated"] = true
					session.Save(r, w)
					fmt.Fprintf(w, "OK")
				}
			}
		}
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var subscriber Subscriber
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&subscriber)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				fmt.Fprintf(w, "%s", err.Error())
			} else {
				var username string
				err := db.QueryRow("SELECT username FROM users WHERE username = ?", subscriber.Username).Scan(&username)
				if err != nil {
					if err.Error() == "sql: no rows in result set" {
						res, err := db.Exec("INSERT INTO users (username, password, email) VALUES (?, ?, ?)", subscriber.Username, subscriber.Password, subscriber.Email)
						_ = res
						if err != nil {
							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						}
						fmt.Fprintf(w, "OK")
					}
				} else {
					http.Error(w, "Username is already taken", http.StatusUnauthorized)
				}
			}
		}
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		auth, ok := session.Values["authenticated"].(bool)
		if !auth || !ok {
			fmt.Fprintf(w, "Visiteur")
		} else {
			fmt.Fprintf(w, "Authentifié")
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		session.Values["authenticated"] = false
		session.Save(r, w)
		fmt.Fprintf(w, "Vous avez été deconnecté")
	})

	http.ListenAndServe(":1337", nil)
}