package kafka

import (
	"context"
	"encoding/json"

	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/hub"
	"github.com/CDeX-Labs/CDeX-Socket-Service/pkg/events"
	"github.com/CDeX-Labs/CDeX-Socket-Service/pkg/protocol"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type Handlers struct {
	hub    *hub.Hub
	logger zerolog.Logger
}

func NewHandlers(h *hub.Hub, logger zerolog.Logger) *Handlers {
	return &Handlers{
		hub:    h,
		logger: logger.With().Str("component", "kafka-handlers").Logger(),
	}
}

func (h *Handlers) HandleSubmissionCreated(ctx context.Context, msg kafka.Message) error {
	var event events.SubmissionCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal submission.created event")
		return err
	}

	h.logger.Info().
		Str("submissionId", event.SubmissionID).
		Str("userId", event.UserID).
		Str("status", event.Status).
		Msg("Processing submission.created")

	wsMsg, err := protocol.NewMessage(protocol.MsgSubmissionCreated, event)
	if err != nil {
		return err
	}

	h.hub.SendToUser(event.UserID, wsMsg)

	if event.ContestID != nil && *event.ContestID != "" {
		roomID := hub.BuildRoomID(hub.RoomTypeContest, *event.ContestID)
		h.hub.SendToRoom(roomID, wsMsg)
	}

	return nil
}

func (h *Handlers) HandleSubmissionJudged(ctx context.Context, msg kafka.Message) error {
	var event events.SubmissionJudgedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal submission.judged event")
		return err
	}

	h.logger.Info().
		Str("submissionId", event.SubmissionID).
		Str("userId", event.UserID).
		Str("verdict", event.Verdict).
		Msg("Processing submission.judged")

	wsMsg, err := protocol.NewMessage(protocol.MsgSubmissionResult, event)
	if err != nil {
		return err
	}

	h.hub.SendToUser(event.UserID, wsMsg)

	if event.ContestID != nil && *event.ContestID != "" {
		roomID := hub.BuildRoomID(hub.RoomTypeContest, *event.ContestID)
		h.hub.SendToRoom(roomID, wsMsg)
	}

	return nil
}

func (h *Handlers) HandleLeaderboardUpdated(ctx context.Context, msg kafka.Message) error {
	var event events.LeaderboardUpdatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal leaderboard.updated event")
		return err
	}

	h.logger.Info().
		Str("contestId", event.ContestID).
		Msg("Processing leaderboard.updated")

	wsMsg, err := protocol.NewMessage(protocol.MsgLeaderboardUpdate, event)
	if err != nil {
		return err
	}

	roomID := hub.BuildRoomID(hub.RoomTypeContest, event.ContestID)
	h.hub.SendToRoom(roomID, wsMsg)

	return nil
}

func (h *Handlers) HandleContestStarted(ctx context.Context, msg kafka.Message) error {
	var event events.ContestStartedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal contest.started event")
		return err
	}

	h.logger.Info().
		Str("contestId", event.ContestID).
		Str("title", event.Title).
		Msg("Processing contest.started")

	wsMsg, err := protocol.NewMessage(protocol.MsgContestEvent, map[string]interface{}{
		"type":      "STARTED",
		"contestId": event.ContestID,
		"title":     event.Title,
		"startTime": event.StartTime,
		"timestamp": event.Timestamp,
	})
	if err != nil {
		return err
	}

	roomID := hub.BuildRoomID(hub.RoomTypeContest, event.ContestID)
	h.hub.SendToRoom(roomID, wsMsg)

	h.hub.Broadcast(wsMsg)

	return nil
}

func (h *Handlers) HandleContestEnded(ctx context.Context, msg kafka.Message) error {
	var event events.ContestEndedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal contest.ended event")
		return err
	}

	h.logger.Info().
		Str("contestId", event.ContestID).
		Str("title", event.Title).
		Msg("Processing contest.ended")

	wsMsg, err := protocol.NewMessage(protocol.MsgContestEvent, map[string]interface{}{
		"type":      "ENDED",
		"contestId": event.ContestID,
		"title":     event.Title,
		"endTime":   event.EndTime,
		"timestamp": event.Timestamp,
	})
	if err != nil {
		return err
	}

	roomID := hub.BuildRoomID(hub.RoomTypeContest, event.ContestID)
	h.hub.SendToRoom(roomID, wsMsg)

	return nil
}

func (h *Handlers) RegisterAll(consumer *Consumer) {
	consumer.RegisterHandler("submission.created", h.HandleSubmissionCreated)
	consumer.RegisterHandler("submission.judged", h.HandleSubmissionJudged)
	consumer.RegisterHandler("leaderboard.updated", h.HandleLeaderboardUpdated)
	consumer.RegisterHandler("contest.started", h.HandleContestStarted)
	consumer.RegisterHandler("contest.ended", h.HandleContestEnded)
}
