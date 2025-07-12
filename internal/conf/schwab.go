package conf

import (
	"time"
)

// AuthUrl = "https://api.schwabapi.com/v1/oauth/token"
// ProdURL = "https://api.schwabapi.com/trader/v1"

type Schwab struct {
	// DisableOptionStream bool             `env:"DISABLE_OPTIONS_STREAM"`
	// OptionsStream       streaming.Stream `envPrefix:"OPTIONS_STREAM"`

	AccessToken  string `env:"SCHWAB_ACCESS_TOKEN"`
	RefreshToken string `env:"SCHWAB_REFRESH_TOKEN"`

	BaseURL string `env:"SCHWAB_URL,required"`
	AuthURL string `env:"SCHWAB_AUTH_URL,required"`

	APIKey    string `env:"SCHWAB_API_KEY,required"`
	APISecret string `env:"SCHWAB_API_SECRET,required"`

	Timeout time.Duration `env:"SCHWAB_REQUEST_TIMEOUT" envDefault:"5s"`
}
