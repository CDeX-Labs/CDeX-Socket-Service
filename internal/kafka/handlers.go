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

func (h *Handlers) HandleContestCreated(ctx context.Context, msg kafka.Message) error {
	var event events.ContestCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal contest.created event")
		return err
	}

	h.logger.Info().
		Str("contestId", event.ContestID).
		Str("title", event.Title).
		Msg("Processing contest.created")

	wsMsg, err := protocol.NewMessage(protocol.MsgContestEvent, map[string]interface{}{
		"type":        "CREATED",
		"contestId":   event.ContestID,
		"title":       event.Title,
		"slug":        event.Slug,
		"visibility":  event.Visibility,
		"scoringMode": event.ScoringMode,
		"startTime":   event.StartTime,
		"endTime":     event.EndTime,
		"timestamp":   event.Timestamp,
	})
	if err != nil {
		return err
	}

	h.hub.Broadcast(wsMsg)

	return nil
}

func (h *Handlers) HandleParticipantRegistered(ctx context.Context, msg kafka.Message) error {
	var event events.ParticipantRegisteredEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal participant.registered event")
		return err
	}

	h.logger.Info().
		Str("contestId", event.ContestID).
		Str("userId", event.UserID).
		Msg("Processing participant.registered")

	wsMsg, err := protocol.NewMessage(protocol.MsgParticipantEvent, map[string]interface{}{
		"type":        "REGISTERED",
		"contestId":   event.ContestID,
		"userId":      event.UserID,
		"displayName": event.DisplayName,
		"isVirtual":   event.IsVirtual,
		"timestamp":   event.Timestamp,
	})
	if err != nil {
		return err
	}

	roomID := hub.BuildRoomID(hub.RoomTypeContest, event.ContestID)
	h.hub.SendToRoom(roomID, wsMsg)

	return nil
}

func (h *Handlers) HandleLeaderboardFrozen(ctx context.Context, msg kafka.Message) error {
	var event events.LeaderboardFrozenEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal leaderboard.frozen event")
		return err
	}

	h.logger.Info().
		Str("contestId", event.ContestID).
		Msg("Processing leaderboard.frozen")

	wsMsg, err := protocol.NewMessage(protocol.MsgLeaderboardFrozen, map[string]interface{}{
		"contestId":  event.ContestID,
		"freezeTime": event.FreezeTime,
		"timestamp":  event.Timestamp,
	})
	if err != nil {
		return err
	}

	roomID := hub.BuildRoomID(hub.RoomTypeContest, event.ContestID)
	h.hub.SendToRoom(roomID, wsMsg)

	return nil
}

func (h *Handlers) HandleLeaderboardUnfrozen(ctx context.Context, msg kafka.Message) error {
	var event events.LeaderboardUnfrozenEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal leaderboard.unfrozen event")
		return err
	}

	h.logger.Info().
		Str("contestId", event.ContestID).
		Msg("Processing leaderboard.unfrozen")

	wsMsg, err := protocol.NewMessage(protocol.MsgLeaderboardUnfrozen, map[string]interface{}{
		"contestId": event.ContestID,
		"timestamp": event.Timestamp,
	})
	if err != nil {
		return err
	}

	roomID := hub.BuildRoomID(hub.RoomTypeContest, event.ContestID)
	h.hub.SendToRoom(roomID, wsMsg)

	return nil
}

func (h *Handlers) HandleProctoringViolation(ctx context.Context, msg kafka.Message) error {
	var event events.ProctoringViolationEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal proctoring.violation event")
		return err
	}

	h.logger.Info().
		Str("contestId", event.ContestID).
		Str("userId", event.UserID).
		Str("type", event.Type).
		Msg("Processing proctoring.violation")

	wsMsg, err := protocol.NewMessage(protocol.MsgProctoringViolation, map[string]interface{}{
		"contestId":           event.ContestID,
		"userId":              event.UserID,
		"type":                event.Type,
		"penaltyApplied":      event.PenaltyApplied,
		"totalPenaltyMinutes": event.TotalPenaltyMinutes,
		"totalViolations":     event.TotalViolations,
		"details":             event.Details,
		"timestamp":           event.Timestamp,
	})
	if err != nil {
		return err
	}

	h.hub.SendToUser(event.UserID, wsMsg)

	return nil
}

func (h *Handlers) RegisterAll(consumer *Consumer) {
	consumer.RegisterHandler("submission.created", h.HandleSubmissionCreated)
	consumer.RegisterHandler("submission.judged", h.HandleSubmissionJudged)
	consumer.RegisterHandler("leaderboard.updated", h.HandleLeaderboardUpdated)
	consumer.RegisterHandler("leaderboard.frozen", h.HandleLeaderboardFrozen)
	consumer.RegisterHandler("leaderboard.unfrozen", h.HandleLeaderboardUnfrozen)
	consumer.RegisterHandler("contest.created", h.HandleContestCreated)
	consumer.RegisterHandler("contest.started", h.HandleContestStarted)
	consumer.RegisterHandler("contest.ended", h.HandleContestEnded)
	consumer.RegisterHandler("contest.participant.registered", h.HandleParticipantRegistered)
	consumer.RegisterHandler("proctoring.violation", h.HandleProctoringViolation)
}
