package config

import (
	"os"
)

type Config struct {
	Google   *GoogleConfig
	Facebook *FacebookConfig
	TikTok   *TikTokConfig
}

type GoogleConfig struct {
	ClientID       string
	ClientSecret   string
	RefreshToken   string
	DeveloperToken string
}

type FacebookConfig struct {
	AppID       string
	AppSecret   string
	AccessToken string
}

type TikTokConfig struct {
	AppID       string
	Secret      string
	AccessToken string
}

func Load() (*Config, error) {
	return &Config{
		Google: &GoogleConfig{
			ClientID:       os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
			RefreshToken:   os.Getenv("GOOGLE_REFRESH_TOKEN"),
			DeveloperToken: os.Getenv("GOOGLE_DEVELOPER_TOKEN"),
		},
		Facebook: &FacebookConfig{
			AppID:       os.Getenv("FACEBOOK_APP_ID"),
			AppSecret:   os.Getenv("FACEBOOK_APP_SECRET"),
			AccessToken: os.Getenv("FACEBOOK_ACCESS_TOKEN"),
		},
		TikTok: &TikTokConfig{
			AppID:       os.Getenv("TIKTOK_APP_ID"),
			Secret:      os.Getenv("TIKTOK_SECRET"),
			AccessToken: os.Getenv("TIKTOK_ACCESS_TOKEN"),
		},
	}, nil
}
