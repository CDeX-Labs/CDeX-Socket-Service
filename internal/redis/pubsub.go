package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CDeX-Labs/CDeX-Socket-Service/pkg/protocol"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	ChannelBroadcast = "ws:broadcast"
	ChannelRoomFmt   = "ws:room:%s"
	ChannelUserFmt   = "ws:user:%s"
)

type PubSubEnvelope struct {
	SourceInstance string           `json:"sourceInstance"`
	Message        *protocol.Message `json:"message"`
	TargetRoom     string           `json:"targetRoom,omitempty"`
	TargetUser     string           `json:"targetUser,omitempty"`
}

type MessageHandler func(envelope *PubSubEnvelope)

type PubSub struct {
	client     *Client
	pubsub     *redis.PubSub
	instanceID string
	handler    MessageHandler
	logger     zerolog.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewPubSub(client *Client, handler MessageHandler, logger zerolog.Logger) *PubSub {
	ctx, cancel := context.WithCancel(context.Background())
	return &PubSub{
		client:     client,
		instanceID: uuid.New().String()[:8],
		handler:    handler,
		logger:     logger.With().Str("component", "pubsub").Logger(),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (p *PubSub) Start() error {
	p.pubsub = p.client.Subscribe(p.ctx, ChannelBroadcast)

	_, err := p.pubsub.Receive(p.ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	go p.listen()

	p.logger.Info().
		Str("instanceId", p.instanceID).
		Msg("PubSub started")

	return nil
}

func (p *PubSub) Stop() error {
	p.cancel()
	if p.pubsub != nil {
		return p.pubsub.Close()
	}
	return nil
}

func (p *PubSub) GetInstanceID() string {
	return p.instanceID
}

func (p *PubSub) listen() {
	ch := p.pubsub.Channel()
	for {
		select {
		case <-p.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			p.handleMessage(msg)
		}
	}
}

func (p *PubSub) handleMessage(msg *redis.Message) {
	var envelope PubSubEnvelope
	if err := json.Unmarshal([]byte(msg.Payload), &envelope); err != nil {
		p.logger.Error().Err(err).Msg("Failed to unmarshal pubsub message")
		return
	}

	if envelope.SourceInstance == p.instanceID {
		return
	}

	p.logger.Debug().
		Str("channel", msg.Channel).
		Str("sourceInstance", envelope.SourceInstance).
		Msg("Received pubsub message")

	if p.handler != nil {
		p.handler(&envelope)
	}
}

func (p *PubSub) PublishToRoom(ctx context.Context, roomID string, msg *protocol.Message) error {
	envelope := PubSubEnvelope{
		SourceInstance: p.instanceID,
		Message:        msg,
		TargetRoom:     roomID,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf(ChannelRoomFmt, roomID)
	return p.client.Publish(ctx, channel, data)
}

func (p *PubSub) PublishToUser(ctx context.Context, userID string, msg *protocol.Message) error {
	envelope := PubSubEnvelope{
		SourceInstance: p.instanceID,
		Message:        msg,
		TargetUser:     userID,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf(ChannelUserFmt, userID)
	return p.client.Publish(ctx, channel, data)
}

func (p *PubSub) PublishBroadcast(ctx context.Context, msg *protocol.Message) error {
	envelope := PubSubEnvelope{
		SourceInstance: p.instanceID,
		Message:        msg,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, ChannelBroadcast, data)
}

func (p *PubSub) SubscribeToRoom(roomID string) error {
	channel := fmt.Sprintf(ChannelRoomFmt, roomID)
	return p.pubsub.Subscribe(p.ctx, channel)
}

func (p *PubSub) UnsubscribeFromRoom(roomID string) error {
	channel := fmt.Sprintf(ChannelRoomFmt, roomID)
	return p.pubsub.Unsubscribe(p.ctx, channel)
}

func (p *PubSub) SubscribeToUser(userID string) error {
	channel := fmt.Sprintf(ChannelUserFmt, userID)
	return p.pubsub.Subscribe(p.ctx, channel)
}

func (p *PubSub) UnsubscribeFromUser(userID string) error {
	channel := fmt.Sprintf(ChannelUserFmt, userID)
	return p.pubsub.Unsubscribe(p.ctx, channel)
}
