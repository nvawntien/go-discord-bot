package leetcode

// GraphQL response types for LeetCode API.

// dailyQuestionResponse is the top-level response for the daily question query.
type dailyQuestionResponse struct {
	Data struct {
		ActiveDailyCodingChallengeQuestion struct {
			Date     string `json:"date"`
			Link     string `json:"link"`
			Question struct {
				FrontendQuestionID string `json:"questionFrontendId"`
				Difficulty         string `json:"difficulty"`
				Title              string `json:"title"`
				TitleSlug          string `json:"titleSlug"`
				TopicTags          []struct {
					Name string `json:"name"`
				} `json:"topicTags"`
			} `json:"question"`
		} `json:"activeDailyCodingChallengeQuestion"`
	} `json:"data"`
}

// userProfileResponse is the top-level response for the user profile query.
type userProfileResponse struct {
	Data struct {
		MatchedUser *struct {
			Username string `json:"username"`
			Profile  struct {
				RealName string `json:"realName"`
				Ranking  int    `json:"ranking"`
			} `json:"profile"`
			SubmitStatsGlobal struct {
				ACSubmissionNum []struct {
					Difficulty string `json:"difficulty"`
					Count      int    `json:"count"`
				} `json:"acSubmissionNum"`
			} `json:"submitStatsGlobal"`
			UserCalendar struct {
				Streak          int `json:"streak"`
				TotalActiveDays int `json:"totalActiveDays"`
			} `json:"userCalendar"`
		} `json:"matchedUser"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// recentSubmissionsResponse is the top-level response for the recent submissions query.
type recentSubmissionsResponse struct {
	Data struct {
		RecentACSubmissionList []struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			TitleSlug string `json:"titleSlug"`
			Timestamp string `json:"timestamp"`
		} `json:"recentAcSubmissionList"`
	} `json:"data"`
}

// graphqlRequest is the request body for a GraphQL query.
type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}
