package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "./data/bot.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM daily_completions")
	if err != nil {
		log.Fatal(err)
	}

	rows, _ := result.RowsAffected()
	fmt.Printf("Cleared %d stale completion records\n", rows)
}
