package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"wos-redeem-discord-bot/models"
	"wos-redeem-discord-bot/mongodb"

	"github.com/bwmarrin/discordgo"
	"github.com/haraldrudell/parl/mains"
	"go.mongodb.org/mongo-driver/mongo"
)

// 注册新命令
func registerCommands(s *discordgo.Session) {

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "redeem",
			Description: "Redeem code",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "code",
					Description: "Gift Code",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "id",
					Description: "User ID(optional, default for ONE)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
			},
		},
		{
			Name:        "add",
			Description: "add user id",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "id",
					Description: "User ID",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		{
			Name:        "id",
			Description: "show id information",
		},
		{
			Name:        "remove",
			Description: "remove user id",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "id",
					Description: "User ID",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		{
			Name:        "list",
			Description: "list user id",
		},
		{
			Name:        "help",
			Description: "Get help information about the bot",
		},
		{
			Name:        "uptime",
			Description: "Get bot uptime",
		},
		{
			Name:        "cancel",
			Description: "Cancel redeem if redeem error",
		},
		{
			Name:        "checkid",
			Description: "Check id if exist in list",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "id",
					Description: "User ID",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
	}

	s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands)
}

func handleHelpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	helpMessage := "Here's how to use the bot:\n\n" +
		"1. To redeem a code all users: `/redeem <code>`\n\n" +
		"2. To redeem a code for a specific user: `/redeem <code> <user_id>`"

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: helpMessage,
		},
	})
}

func handleAddCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	var id string

	for _, option := range options {
		switch option.Name {
		case "id":
			id = strconv.FormatInt(option.IntValue(), 10)
		}
	}

	playerData, err := retryRequest(func() (*models.Player, error) {
		return getRoleInfo(id)
	})
	if err != nil {
		status := handlePlayerInfoError(err, id)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: status,
			},
		})
		return
	}

	_, err = mongodb.GetUser(playerData.FID)

	if err == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("UserID %d already exist", playerData.FID),
			},
		})
		return
	}

	err = mongodb.UpdateUser(*playerData)

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error adding ID: %v", err),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s - %s add successfully", id, playerData.Nickname),
		},
	})

}

func handleListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// 如果 responseString 长度即将超过 2000，使用 InteractionRespond 发送出去，后面的 user 放到 responseString通过 ChannelMessageSend 发送，后续每当长度即将超过2000，使用  ChannelMessageSend 发送

	users, err := mongodb.GetAllUser()

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error reading user IDs: %v", err),
			},
		})
		return
	}

	var responseString string = fmt.Sprintf("UserID - Nickname (%d users)", len(users))

	hasResponded := false // 用于标记是否已经发送过一次 InteractionRespond

	isInit := false

	// 初始化，检测是否有 KID 0 的用户，如果有，就更新一下数据
	for _, user := range users {

		if user.KID == 0 {
			isInit = true
			log.Println("get user info", user.FID)
			userID := strconv.Itoa(user.FID)
			playerData, err := retryRequest(func() (*models.Player, error) {
				return getRoleInfo(userID)
			})
			if err != nil {
				status := handlePlayerInfoError(err, userID)
				log.Println(status)
			}

			mongodb.UpdateUser(*playerData)

		}
	}

	if isInit {
		users, err = mongodb.GetAllUser()

		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error reading user IDs: %v", err),
				},
			})
			return
		}
	}

	for _, user := range users {

		// 如果没有 responsed 过，并且增加这次循环的 user 会导致长度超过 2000，发送当前还未添加 的 user
		if !hasResponded && len(responseString+fmt.Sprintf("\n%d - %s", user.FID, user.Nickname)) > 2000 {

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: responseString,
				},
			})
			if err != nil {
				log.Println(err)
			}

			hasResponded = true

			// 重置 responseString
			responseString = ""

		}

		// 已经 response 过，并且增加这次循环的 user 会导致长度超过 2000，发送当前还未添加 的 user
		if hasResponded && len(responseString+fmt.Sprintf("\n%d - %s", user.FID, user.Nickname)) > 2000 {
			_, err = s.ChannelMessageSend(i.ChannelID, responseString)
			if err != nil {
				log.Println("Error sending message:", err)
			}

			responseString = ""
		}

		responseString += fmt.Sprintf("\n%d - %s", user.FID, user.Nickname)

	}

	// 如果最后一次 responseString 长度没有超过 2000，使用 ChannelMessageSend 发送出去
	if hasResponded && len(responseString) > 0 && len(responseString) <= 2000 {
		_, err = s.ChannelMessageSend(i.ChannelID, responseString)
		if err != nil {
			log.Println("Error sending message:", err)
		}
		if err != nil {
			log.Println("Error sending message:", err)
		}
	}

	// HTTP 400 Bad Request, {"message": "Invalid Form Body", "code": 50035, "errors": {"data": {"content": {"_errors": [{"code": "BASE_TYPE_MAX_LENGTH", "message": "Must be 2000 or fewer in length."}]}}}}

}

func handleIDCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	text := fmt.Sprintf("AppID: %s\nChannelID: %s\nGuildID: %s", i.AppID, i.ChannelID, i.GuildID)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	})
}

func handleUpTimeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	Now := time.Now()
	uptime := Now.Sub(mains.ProcessStartTime())
	text := fmt.Sprintf("uptime: %s s", strings.Split(uptime.String(), ".")[0]) // 1585h39m2 s

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	})
}

func handleCancelCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// text := fmt.Sprintf("AppID: %s\nChannelID: %s\nGuildID: %s", i.AppID, i.ChannelID, i.GuildID)

	text := "redeem canceled"
	Cancel = true

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	})
}

func handleRemoveCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	var id int

	for _, option := range options {
		switch option.Name {
		case "id":
			id = int(option.IntValue())
		}
	}

	// err := RemoveID(id)
	err := mongodb.RemoveUser(id)

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error removing ID: %v", err),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("ID %d removed successfully", id),
		},
	})
}

func handleCheckIDCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	var id int

	for _, option := range options {
		switch option.Name {
		case "id":
			id = int(option.IntValue())
		}
	}

	user, err := mongodb.GetUser(id)

	if err != nil {

		if err == mongo.ErrNoDocuments {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%d ID not exist", id),
				},
			})
			return
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error get user: %v", err),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("ID %d already exist\nNickname: %s\nStoveLv: %s\nState: %d", id, user.Nickname, user.GetUserRealStoveLv(), user.KID),
		},
	})

}

func handleRedeemCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	var code string
	var id string

	for _, option := range options {
		switch option.Name {
		case "code":
			code = option.StringValue()
		case "id":
			id = strconv.FormatInt(option.IntValue(), 10)
		}
	}

	go handleRedemption(s, nil, i, code, id)
}
