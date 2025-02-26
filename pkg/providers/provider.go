package providers

import (
	"context"
	"errors"
	"time"

	"glide/pkg/providers/clients"
	"glide/pkg/routers/health"
	"glide/pkg/routers/latency"

	"glide/pkg/api/schemas"
)

// LangModelProvider defines an interface a provider should fulfill to be able to serve language chat requests
type LangModelProvider interface {
	Provider() string
	Chat(ctx context.Context, request *schemas.UnifiedChatRequest) (*schemas.UnifiedChatResponse, error)
}

type Model interface {
	ID() string
	Healthy() bool
	Latency() *latency.MovingAverage
	LatencyUpdateInterval() *time.Duration
	Weight() int
}

type LanguageModel interface {
	Model
	LangModelProvider
}

// LangModel wraps provider client and expend it with health & latency tracking
type LangModel struct {
	modelID               string
	weight                int
	client                LangModelProvider
	rateLimit             *health.RateLimitTracker
	errorBudget           *health.TokenBucket // TODO: centralize provider API health tracking in the registry
	latency               *latency.MovingAverage
	latencyUpdateInterval *time.Duration
}

func NewLangModel(modelID string, client LangModelProvider, budget health.ErrorBudget, latencyConfig latency.Config, weight int) *LangModel {
	return &LangModel{
		modelID:               modelID,
		client:                client,
		rateLimit:             health.NewRateLimitTracker(),
		errorBudget:           health.NewTokenBucket(budget.TimePerTokenMicro(), budget.Budget()),
		latency:               latency.NewMovingAverage(latencyConfig.Decay, latencyConfig.WarmupSamples),
		latencyUpdateInterval: latencyConfig.UpdateInterval,
		weight:                weight,
	}
}

func (m *LangModel) ID() string {
	return m.modelID
}

func (m *LangModel) Provider() string {
	return m.client.Provider()
}

func (m *LangModel) Latency() *latency.MovingAverage {
	return m.latency
}

func (m *LangModel) LatencyUpdateInterval() *time.Duration {
	return m.latencyUpdateInterval
}

func (m *LangModel) Healthy() bool {
	return !m.rateLimit.Limited() && m.errorBudget.HasTokens()
}

func (m *LangModel) Weight() int {
	return m.weight
}

func (m *LangModel) Chat(ctx context.Context, request *schemas.UnifiedChatRequest) (*schemas.UnifiedChatResponse, error) {
	startedAt := time.Now()
	resp, err := m.client.Chat(ctx, request)

	if err == nil {
		// record latency per token to normalize measurements
		m.latency.Add(float64(time.Since(startedAt)) / resp.ModelResponse.TokenUsage.ResponseTokens)

		// successful response
		resp.ModelID = m.modelID

		return resp, err
	}

	var rle *clients.RateLimitError

	if errors.As(err, &rle) {
		m.rateLimit.SetLimited(rle.UntilReset())

		return resp, err
	}

	_ = m.errorBudget.Take(1)

	return resp, err
}
