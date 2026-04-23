package leetcode

// GraphQL query strings for the LeetCode API.

const queryDailyQuestion = `
query questionOfToday {
  activeDailyCodingChallengeQuestion {
    date
    link
    question {
      questionFrontendId
      difficulty
      title
      titleSlug
      topicTags {
        name
      }
    }
  }
}
`

const queryUserProfile = `
query getUserProfile($username: String!) {
  matchedUser(username: $username) {
    username
    profile {
      realName
      ranking
    }
    submitStatsGlobal {
      acSubmissionNum {
        difficulty
        count
      }
    }
    userCalendar {
      streak
      totalActiveDays
    }
  }
}
`

const queryRecentACSubmissions = `
query recentAcSubmissions($username: String!, $limit: Int!) {
  recentAcSubmissionList(username: $username, limit: $limit) {
    id
    title
    titleSlug
    timestamp
  }
}
`
