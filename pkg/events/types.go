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

type ContestCreatedEvent struct {
	ContestID   string `json:"contestId"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Visibility  string `json:"visibility"`
	ScoringMode string `json:"scoringMode"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	CreatedBy   string `json:"createdBy"`
	Timestamp   string `json:"timestamp"`
}

type ParticipantRegisteredEvent struct {
	ContestID   string `json:"contestId"`
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	IsVirtual   bool   `json:"isVirtual"`
	Timestamp   string `json:"timestamp"`
}

type ParticipantUnregisteredEvent struct {
	ContestID string `json:"contestId"`
	UserID    string `json:"userId"`
	Timestamp string `json:"timestamp"`
}

type LeaderboardFrozenEvent struct {
	ContestID  string `json:"contestId"`
	FreezeTime string `json:"freezeTime"`
	Timestamp  string `json:"timestamp"`
}

type LeaderboardUnfrozenEvent struct {
	ContestID string `json:"contestId"`
	Timestamp string `json:"timestamp"`
}

type ProctoringViolationEvent struct {
	ContestID           string  `json:"contestId"`
	UserID              string  `json:"userId"`
	Type                string  `json:"type"`
	PenaltyApplied      int     `json:"penaltyApplied"`
	TotalPenaltyMinutes int     `json:"totalPenaltyMinutes"`
	TotalViolations     int     `json:"totalViolations"`
	Details             *string `json:"details"`
	Timestamp           string  `json:"timestamp"`
}

type ProblemAddedEvent struct {
	ContestID string `json:"contestId"`
	ProblemID string `json:"problemId"`
	Label     string `json:"label"`
	Timestamp string `json:"timestamp"`
}

type ProblemRemovedEvent struct {
	ContestID string `json:"contestId"`
	ProblemID string `json:"problemId"`
	Timestamp string `json:"timestamp"`
}
