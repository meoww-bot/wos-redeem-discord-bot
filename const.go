package main

import (
	"time"
)

const (
	BASE_URL       = "https://wos-giftcode-api.centurygame.com/api"
	SECRET         = "tB87#kPtkxqOS2"
	MAX_RETRIES    = 20
	RETRY_DELAY    = 8 * time.Second
	DELAY_DURATION = 2 * time.Second
)
