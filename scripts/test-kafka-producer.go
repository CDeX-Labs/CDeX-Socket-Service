package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

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

func main() {
	godotenv.Load("../.env")

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	userID := "test-user-123"
	if len(os.Args) > 1 {
		userID = os.Args[1]
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokers},
		Topic:    "submission.judged",
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	execTime := 150
	memUsed := 2048

	event := SubmissionJudgedEvent{
		SubmissionID:    uuid.New().String(),
		UserID:          userID,
		ProblemID:       "problem-abc123",
		ContestID:       nil,
		AssignmentID:    nil,
		Verdict:         "ACCEPTED",
		Score:           100,
		ExecutionTimeMs: &execTime,
		MemoryUsedKb:    &memUsed,
		TestCasesPassed: 10,
		TestCasesTotal:  10,
		Timestamp:       time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error marshaling event: %v\n", err)
		os.Exit(1)
	}

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(event.UserID),
		Value: data,
	})

	if err != nil {
		fmt.Printf("Error writing to Kafka: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Submission Judged Event Sent ===")
	fmt.Println()
	fmt.Printf("Topic: submission.judged\n")
	fmt.Printf("User ID: %s\n", event.UserID)
	fmt.Printf("Submission ID: %s\n", event.SubmissionID)
	fmt.Printf("Verdict: %s\n", event.Verdict)
	fmt.Printf("Score: %d\n", event.Score)
	fmt.Println()
	fmt.Println("The connected WebSocket client should receive this event!")
}
