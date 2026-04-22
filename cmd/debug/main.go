package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/tien/go-discord-bot/internal/adapter/outbound/leetcode"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := leetcode.NewClient(logger)
	ctx := context.Background()

	// 1. Fetch daily question
	fmt.Println("=== DAILY QUESTION ===")
	daily, err := client.GetDailyQuestion(ctx)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("Date: %s\n", daily.Date)
	fmt.Printf("Title: %s\n", daily.Title)
	fmt.Printf("TitleSlug: %q\n", daily.TitleSlug)
	fmt.Printf("Link: %s\n", daily.Link)
	fmt.Printf("Difficulty: %s\n", daily.Difficulty)

	// 2. Fetch recent submissions for vantien0105
	fmt.Println("\n=== RECENT SUBMISSIONS (vantien0105) ===")
	subs, err := client.GetRecentAcceptedSubmissions(ctx, "vantien0105", 20)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("Total recent submissions: %d\n", len(subs))
	for i, s := range subs {
		match := ""
		if s.TitleSlug == daily.TitleSlug {
			match = " <<<< MATCH DAILY!"
		}
		fmt.Printf("  [%d] %q (slug: %q)%s\n", i, s.Title, s.TitleSlug, match)
	}

	// 3. Check match
	fmt.Println("\n=== RESULT ===")
	found := false
	for _, s := range subs {
		if s.TitleSlug == daily.TitleSlug {
			found = true
			break
		}
	}
	if found {
		fmt.Println("✅ User HAS solved today's daily question!")
	} else {
		fmt.Println("❌ User has NOT solved today's daily question.")
		fmt.Printf("Looking for slug: %q\n", daily.TitleSlug)
	}

	// 4. Test non-existent user
	fmt.Println("\n=== TEST NON-EXISTENT USER ===")
	_, err = client.GetUserProfile(ctx, "thisuserdoesnotexist12345xyz")
	if err != nil {
		fmt.Printf("✅ Correctly returned error: %v\n", err)
	} else {
		fmt.Println("❌ BUG: Should have returned error for non-existent user!")
	}
}
