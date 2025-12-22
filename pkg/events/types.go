package events

type SubmissionCreatedEvent struct {
	SubmissionID string  `json:"submissionId"`
	UserID       string  `json:"userId"`
	ProblemID    string  `json:"problemId"`
	ContestID    *string `json:"contestId"`
	AssignmentID *string `json:"assignmentId"`
	Language     string  `json:"language"`
	Status       string  `json:"status"`
	Timestamp    string  `json:"timestamp"`
}

type SubmissionJudgedEvent struct {
	SubmissionID    string  `json:"submissionId"`
	UserID          string  `json:"userId"`
	ProblemID       string  `json:"problemId"`
	ContestID       *string `json:"contestId"`
	AssignmentID    *string `json:"assignmentId"`
	Verdict         string  `json:"verdict"`
	Score           int     `json:"score"`
	ExecutionTimeMs *int    `json:"executionTimeMs"`
	MemoryUsedKb    *int    `json:"memoryUsedKb"`
	TestCasesPassed int     `json:"testCasesPassed"`
	TestCasesTotal  int     `json:"testCasesTotal"`
	Timestamp       string  `json:"timestamp"`
}

type LeaderboardUpdatedEvent struct {
	ContestID string `json:"contestId"`
	Timestamp string `json:"timestamp"`
}

type ContestStartedEvent struct {
	ContestID string `json:"contestId"`
	Title     string `json:"title"`
	StartTime string `json:"startTime"`
	Timestamp string `json:"timestamp"`
}

type ContestEndedEvent struct {
	ContestID string `json:"contestId"`
	Title     string `json:"title"`
	EndTime   string `json:"endTime"`
	Timestamp string `json:"timestamp"`
}
