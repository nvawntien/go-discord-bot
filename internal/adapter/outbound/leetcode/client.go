package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/tien/go-discord-bot/internal/domain/entity"
)

const (
	leetcodeGraphQLURL = "https://leetcode.com/graphql"
	leetcodeBaseURL    = "https://leetcode.com"
	maxRetries         = 3
	initialBackoff     = 1 * time.Second
)

// Client implements the LeetCode GraphQL API client with rate limiting and retry logic.
type Client struct {
	httpClient *http.Client
	logger     *slog.Logger
	mu         sync.Mutex
	lastReq    time.Time
	minDelay   time.Duration
}

// NewClient creates a new LeetCode API client.
func NewClient(logger *slog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger:   logger,
		minDelay: 1 * time.Second, // Rate limit: max 1 req/s
	}
}

// GetDailyQuestion fetches today's daily coding challenge from LeetCode.
func (c *Client) GetDailyQuestion(ctx context.Context) (*entity.DailyQuestion, error) {
	reqBody := graphqlRequest{
		Query: queryDailyQuestion,
	}

	var resp dailyQuestionResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, fmt.Errorf("fetch daily question: %w", err)
	}

	q := resp.Data.ActiveDailyCodingChallengeQuestion
	tags := make([]string, 0, len(q.Question.TopicTags))
	for _, t := range q.Question.TopicTags {
		tags = append(tags, t.Name)
	}

	return &entity.DailyQuestion{
		Date:       q.Date,
		Title:      q.Question.Title,
		TitleSlug:  q.Question.TitleSlug,
		QuestionID: q.Question.FrontendQuestionID,
		Difficulty: q.Question.Difficulty,
		Link:       leetcodeBaseURL + q.Link,
		TopicTags:  tags,
	}, nil
}

// GetUserProfile fetches a user's LeetCode profile statistics.
func (c *Client) GetUserProfile(ctx context.Context, username string) (*entity.UserStats, error) {
	reqBody := graphqlRequest{
		Query: queryUserProfile,
		Variables: map[string]any{
			"username": username,
		},
	}

	var resp userProfileResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, fmt.Errorf("fetch user profile: %w", err)
	}

	// Check for GraphQL errors (e.g. user not found)
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("user %q not found on LeetCode", username)
	}

	// MatchedUser is nil when the username doesn't exist
	if resp.Data.MatchedUser == nil {
		return nil, fmt.Errorf("user %q not found on LeetCode", username)
	}

	user := resp.Data.MatchedUser
	stats := &entity.UserStats{
		Username: user.Username,
		RealName: user.Profile.RealName,
		Ranking:  user.Profile.Ranking,
	}

	for _, s := range user.SubmitStatsGlobal.ACSubmissionNum {
		switch s.Difficulty {
		case "All":
			stats.TotalSolved = s.Count
		case "Easy":
			stats.EasySolved = s.Count
		case "Medium":
			stats.MediumSolved = s.Count
		case "Hard":
			stats.HardSolved = s.Count
		}
	}

	return stats, nil
}

// GetRecentAcceptedSubmissions fetches a user's recent accepted submissions.
func (c *Client) GetRecentAcceptedSubmissions(ctx context.Context, username string, limit int) ([]entity.Submission, error) {
	reqBody := graphqlRequest{
		Query: queryRecentACSubmissions,
		Variables: map[string]any{
			"username": username,
			"limit":    limit,
		},
	}

	var resp recentSubmissionsResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, fmt.Errorf("fetch recent submissions: %w", err)
	}

	submissions := make([]entity.Submission, 0, len(resp.Data.RecentACSubmissionList))
	for _, s := range resp.Data.RecentACSubmissionList {
		ts, _ := strconv.ParseInt(s.Timestamp, 10, 64)
		submissions = append(submissions, entity.Submission{
			Title:     s.Title,
			TitleSlug: s.TitleSlug,
			Timestamp: ts,
		})
	}

	return submissions, nil
}

// doRequest sends a GraphQL request with rate limiting and exponential backoff retry.
func (c *Client) doRequest(ctx context.Context, reqBody graphqlRequest, result any) error {
	// Rate limiting — enforce minimum delay between requests
	c.mu.Lock()
	elapsed := time.Since(c.lastReq)
	if elapsed < c.minDelay {
		time.Sleep(c.minDelay - elapsed)
	}
	c.lastReq = time.Now()
	c.mu.Unlock()

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	backoff := initialBackoff

	for attempt := range maxRetries {
		if err := ctx.Err(); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, leetcodeGraphQLURL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Referer", leetcodeBaseURL)
		req.Header.Set("User-Agent", "go-discord-bot/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logger.Warn("LeetCode API request failed, retrying",
				"attempt", attempt+1,
				"error", err,
			)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("read response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			c.logger.Warn("LeetCode API rate limited, backing off",
				"attempt", attempt+1,
				"backoff", backoff,
			)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
			continue
		}

		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	}

	return fmt.Errorf("all %d retries exhausted: %w", maxRetries, lastErr)
}
