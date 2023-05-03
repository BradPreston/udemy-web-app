package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"web-app/pkg/repository"
	"web-app/pkg/repository/dbrepo"
)

const port = 8090

type application struct{
	DSN string
	DB repository.DatabaseRepo
	Domain string
	JWTSecret string
}

func main() {
	app := application{}
	flag.StringVar(&app.Domain, "domain", "example.com", "domain for application eg: company.com")
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "postgres connection")
	flag.StringVar(&app.JWTSecret, "jwt-secret", "verysecret", "signing secret")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	log.Printf("starting api on port %d\n", port)
	
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}