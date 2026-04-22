package entity

// DailyQuestion represents the LeetCode daily coding challenge.
type DailyQuestion struct {
	Date       string   // e.g. "2026-04-22"
	Title      string   // e.g. "Two Sum"
	TitleSlug  string   // e.g. "two-sum"
	QuestionID string   // e.g. "1"
	Difficulty string   // "Easy", "Medium", "Hard"
	Link       string   // Full URL
	TopicTags  []string // e.g. ["Array", "Hash Table"]
}
