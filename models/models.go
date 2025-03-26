package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type PlayerResponse struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Msg     string          `json:"msg"`
	ErrCode any             `json:"err_code"`
}

type Player struct {
	FID            int    `json:"fid"`
	Nickname       string `json:"nickname"`
	KID            int    `json:"kid"`
	StoveLv        any    `json:"stove_lv"`
	StoveLvContent any    `json:"stove_lv_content"` // 20 int or string "stove_lv_content": "https:\/\/gof-formal-avatar.akamaized.net\/img\/icon\/stove_lv_9.png",
	AvatarImage    string `json:"avatar_image"`
}

type ExchangeResponse struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Msg     string          `json:"msg"`
	ErrCode int             `json:"err_code"`
}

func (u *Player) GetUserRealStoveLv() string {

	realStoveLv := ""

	StoveLv, ok := u.StoveLv.(float64)

	if ok {
		realStoveLv = strconv.Itoa(int(StoveLv))
		if StoveLv > 30 {
			quotient := (int(StoveLv) - 30) / 5
			remainder := (int(StoveLv) - 30) % 5

			realStoveLv = fmt.Sprintf("C%d-%d", quotient, remainder)
		}
	}

	return realStoveLv

}
