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

	fmt.Println("=== USERS ===")
	rows, err := db.Query("SELECT id, discord_id, guild_id, leetcode_username FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var discordID, guildID, lcUser string
		rows.Scan(&id, &discordID, &guildID, &lcUser)
		fmt.Printf("  id=%d discord=%s guild=%s lc=%s\n", id, discordID, guildID, lcUser)
	}

	fmt.Println("\n=== DAILY COMPLETIONS ===")
	rows2, err := db.Query("SELECT id, user_id, date, question_slug FROM daily_completions")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var id, userID int64
		var date, slug string
		rows2.Scan(&id, &userID, &date, &slug)
		fmt.Printf("  id=%d user_id=%d date=%s slug=%s\n", id, userID, date, slug)
	}
}
