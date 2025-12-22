package presence

import (
	"context"
	"fmt"
	"time"

	redisclient "github.com/CDeX-Labs/CDeX-Socket-Service/internal/redis"
	"github.com/rs/zerolog"
)

const (
	presenceKeyFmt = "presence:user:%s"
	presenceTTL    = 5 * time.Minute
)

type Status string

const (
	StatusOnline  Status = "online"
	StatusOffline Status = "offline"
)

type Manager struct {
	redis      *redisclient.Client
	instanceID string
	logger     zerolog.Logger
}

func NewManager(redis *redisclient.Client, instanceID string, logger zerolog.Logger) *Manager {
	return &Manager{
		redis:      redis,
		instanceID: instanceID,
		logger:     logger.With().Str("component", "presence").Logger(),
	}
}

func (m *Manager) SetOnline(ctx context.Context, userID string) error {
	key := fmt.Sprintf(presenceKeyFmt, userID)
	err := m.redis.HSet(ctx, key, m.instanceID, time.Now().Unix())
	if err != nil {
		return err
	}
	return m.redis.Expire(ctx, key, presenceTTL)
}

func (m *Manager) SetOffline(ctx context.Context, userID string) error {
	key := fmt.Sprintf(presenceKeyFmt, userID)
	return m.redis.HDel(ctx, key, m.instanceID)
}

func (m *Manager) IsOnline(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf(presenceKeyFmt, userID)
	count, err := m.redis.HLen(ctx, key)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *Manager) GetOnlineUsers(ctx context.Context, userIDs []string) ([]string, error) {
	online := make([]string, 0)
	for _, userID := range userIDs {
		isOnline, err := m.IsOnline(ctx, userID)
		if err != nil {
			m.logger.Error().Err(err).Str("userId", userID).Msg("Failed to check presence")
			continue
		}
		if isOnline {
			online = append(online, userID)
		}
	}
	return online, nil
}

func (m *Manager) GetUserInstances(ctx context.Context, userID string) (map[string]string, error) {
	key := fmt.Sprintf(presenceKeyFmt, userID)
	return m.redis.HGetAll(ctx, key)
}

func (m *Manager) RefreshPresence(ctx context.Context, userID string) error {
	key := fmt.Sprintf(presenceKeyFmt, userID)
	err := m.redis.HSet(ctx, key, m.instanceID, time.Now().Unix())
	if err != nil {
		return err
	}
	return m.redis.Expire(ctx, key, presenceTTL)
}
