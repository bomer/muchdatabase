package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

var schemas = [...]string{`
CREATE TABLE person (
    firstname text,
    lastname text,
    email text
);`,
	`
CREATE TABLE place (
    country text,
    city text NULL,
    telcode integer
);
`}
var (
	db *sqlx.DB
)

type Person struct {
	Firstname string
	Lastname  string
	Email     string
}

type Place struct {
	Country string
	City    sql.NullString
	TelCode int
}

func setupDb() {
	db.MustExec("drop table IF EXISTS person")
	db.MustExec("drop table IF EXISTS place")

	db.MustExec(schemas[0])
	db.MustExec(schemas[1])

	for i := 0; i < 1000; i++ {
		tx := db.MustBegin()
		tx.MustExec("INSERT INTO person (firstname, lastname, email) VALUES (?, ?, ?)", "Jason", "Moiron", "jmoiron@jmoiron.net")
		tx.MustExec("INSERT INTO person (firstname, lastname, email) VALUES (?, ?, ?)", "John", "Doe", "johndoeDNE@gmail.net")
		tx.MustExec("INSERT INTO place (country, city, telcode) VALUES (?, ?, ?)", "United States", "New York", "1")
		tx.MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Hong Kong", "852")
		tx.MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Singapore", "65")
		// Named queries can use structs, so if you have an existing struct (i.e. person := &Person{}) that you have populated, you can pass it in as &person
		tx.NamedExec("INSERT INTO person (firstname, lastname, email) VALUES (:firstname, :lastname, :email)", &Person{"Jane", "Citizen", "jane.citzen@example.com"})
		tx.Commit()
	}

}

func main() {
	// this Pings the database trying to connect, panics on error
	// use sqlx.Open() for sql.Open() semantics
	var err error
	db, err = sqlx.Connect("mysql", "root:password@/go_muchdatabase")
	if err != nil {
		log.Fatalln(err)
	}
	db.Ping()
	fmt.Printf("Connected I think\n")

	setupDb()

	// Query the database, storing results in a []Person (wrapped in []interface{})
	people := []Person{}
	db.Select(&people, "SELECT * FROM person ORDER BY firstname ASC")
	jason, john := people[0], people[1]

	fmt.Printf("%#v\n%#v", jason, john)

	http.HandleFunc("/", showAllPeople)
	http.HandleFunc("/person", showPerson)
	http.ListenAndServe(":3000", nil)

}

func showAllPeople(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	people := []Person{}
	db.Select(&people, "SELECT * FROM person")
	json.NewEncoder(w).Encode(people)
}

func showPerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	people := []Person{}
	db.Select(&people, "SELECT * FROM person where firstname LIKE ? ORDER BY firstname ASC limit 1 ", name)
	json.NewEncoder(w).Encode(people)
}
