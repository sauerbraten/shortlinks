package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"

	"github.com/sauerbraten/shortlinks"
)

func main() {
	i, err := shortlinks.NewSQLiteIndex("file:shortlinks.sqlite?_journal_mode=wal") // would come from command-line argument or env var
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s := shortlinks.NewServer(i, nil)

	r := chi.NewRouter()

	s.SetupRoutes(r)

	log.Println("server running on port 5656")

	err = http.ListenAndServe(":5656", r)
	if err != nil {
		fmt.Println(err)
	}
}
