package anthropic

import (
	"glide/pkg/config/fields"
)

// Params defines OpenAI-specific model params with the specific validation of values
// TODO: Add validations
type Params struct {
	System        string   `yaml:"system,omitempty" json:"system"`
	Temperature   float64  `yaml:"temperature,omitempty" json:"temperature"`
	TopP          float64  `yaml:"top_p,omitempty" json:"top_p"`
	TopK          int      `yaml:"top_k,omitempty" json:"top_k"`
	MaxTokens     int      `yaml:"max_tokens,omitempty" json:"max_tokens"`
	StopSequences []string `yaml:"stop,omitempty" json:"stop"`
	Metadata      *string  `yaml:"metadata,omitempty" json:"metadata"`
	// Stream           bool             `json:"stream,omitempty"` // TODO: we are not supporting this at the moment
}

func DefaultParams() Params {
	return Params{
		Temperature:   1,
		TopP:          0,
		TopK:          0,
		MaxTokens:     250,
		System:        "You are a helpful assistant.",
		StopSequences: []string{},
	}
}

func (p *Params) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*p = DefaultParams()

	type plain Params // to avoid recursion

	return unmarshal((*plain)(p))
}

type Config struct {
	BaseURL       string        `yaml:"baseUrl" json:"baseUrl" validate:"required"`
	ChatEndpoint  string        `yaml:"chatEndpoint" json:"chatEndpoint" validate:"required"`
	Model         string        `yaml:"model" json:"model" validate:"required"`
	APIKey        fields.Secret `yaml:"api_key" json:"-" validate:"required"`
	DefaultParams *Params       `yaml:"defaultParams,omitempty" json:"defaultParams"`
}

// DefaultConfig for OpenAI models
func DefaultConfig() *Config {
	defaultParams := DefaultParams()

	return &Config{
		BaseURL:       "https://api.anthropic.com/v1",
		ChatEndpoint:  "/messages",
		Model:         "claude-instant-1.2",
		DefaultParams: &defaultParams,
	}
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = *DefaultConfig()

	type plain Config // to avoid recursion

	return unmarshal((*plain)(c))
}
