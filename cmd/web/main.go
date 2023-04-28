package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"
	"web-app/pkg/data"
	"web-app/pkg/repository"
	"web-app/pkg/repository/dbrepo"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	DSN     string
	DB      repository.DatabaseRepo
	Session *scs.SessionManager
}

func main() {
	gob.Register(data.User{})
	// setup an app config
	app := application{}
	// get DSN
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "postgres connection")
	flag.Parse()
	// connect to database
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	// get a session manager
	app.Session = getSession()
	// print out a starting message
	log.Println("starting server on port 8080")
	// start the server
	err = http.ListenAndServe(":8080", app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
