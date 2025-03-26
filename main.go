package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"wos-redeem-discord-bot/models"
	"wos-redeem-discord-bot/mongodb"

	"github.com/bwmarrin/discordgo"
)

var IsRedeeming = false
var Cancel = false
var StartTime = time.Now()

func main() {

	BotToken := os.Getenv("BOT_TOKEN")
	if BotToken == "" {
		log.Println("Discord bot token is not set in the environment variables")
		return
	}

	// create a session
	discord, err := discordgo.New("Bot " + BotToken)
	checkNilErr(err)

	// discord.Debug = true

	// add a event handler
	discord.AddHandler(newMessage)
	discord.AddHandler(ready)

	// open session
	err = discord.Open()
	if err != nil {
		log.Println("Error opening connection:", err)
		return
	}
	defer discord.Close() // close session, after function termination

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			handleSlashCommand(s, i)
		}
	})

	// keep bot running untill there is NO os interruption (ctrl + C)
	log.Println("Bot running....")
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// <-c
	select {}
}

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.GuildID == "" && i.User.Username != "meowwbot" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "private message not allowed",
			},
		})
		return
	}

	switch i.ApplicationCommandData().Name {
	case "redeem":
		handleRedeemCommand(s, i)
	case "help":
		handleHelpCommand(s, i)
	case "add":
		handleAddCommand(s, i)
	case "remove":
		handleRemoveCommand(s, i)
	case "list":
		handleListCommand(s, i)
	case "id":
		handleIDCommand(s, i)
	case "uptime":
		handleUpTimeCommand(s, i)
	case "checkid":
		handleCheckIDCommand(s, i)
	case "cancel":
		handleCancelCommand(s, i)
	}
}

func newMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	/* prevent bot responding to its own message
	this is achived by looking into the message author id
	if message.author.id is same as bot.author.id then just return
	*/
	if m.Author.ID == s.State.User.ID {
		return
	}

	// respond to user message if it contains `!help` or `!bye`
	log.Print("author id:", m.Author.ID, " author:", m.Author, " channel:", m.ChannelID, " content:", m.Content)
	PrettyPrint(m)

	if strings.Contains(m.Content, "Code:") {

		code := strings.Split(m.Content, "\n")[0]

		code = strings.Split(code, ":")[1]

		code = strings.TrimSpace(code)

		log.Print("Get Code:", code)

		go handleRedemption(s, m, nil, code, "")
	}

	if strings.HasPrefix(m.Content, "/redeem ") {
		parts := strings.Fields(m.Content)
		if len(parts) == 2 {
			go handleRedemption(s, m, nil, parts[1], "")
		} else if len(parts) == 3 {
			go handleRedemption(s, m, nil, parts[1], parts[2])
		} else {
			s.ChannelMessageSend(m.ChannelID, "Invalid command. Use '/redeem code' or '/redeem code ID'")
		}
	}

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("Bot is ready!")

	msg, err := s.ChannelMessageSend("1271858707302711453", "Bot is ready!")
	if err != nil {
		log.Println("Error sending message:", err)
	}

	PrettyPrint(msg)

	registerCommands(s)

	// registeredCommands, err := s.ApplicationCommands(s.State.User.ID, "")
	// if err != nil {
	// 	log.Fatalf("获取已注册命令失败: %v", err)
	// }
	// for _, cmd := range registeredCommands {
	// 	// 检查命令是否已经注册
	// 	// PrettyPrint(cmd)
	// }
}

func handleRedemption(s *discordgo.Session, m *discordgo.MessageCreate, i *discordgo.InteractionCreate, code string, specificID string) {

	users, err := mongodb.GetAllUser()
	var userIDs []string
	if err != nil {
		sendErrorMessage(s, m, i, fmt.Sprintf("Error reading user IDs: %v", err))
		return
	}

	for _, user := range users {
		userIDs = append(userIDs, strconv.Itoa(user.FID))
	}

	if specificID != "" {
		userIDs = []string{specificID}
	}

	msg := sendInitialMessage(s, m, i, code)

	results := make([]string, 0, len(userIDs))
	updateMessage := createUpdateMessageFunc(s, m, i, msg, code, &results)

	if IsRedeeming {
		results = append(results, "Another redemption is running\nPlease wait a while")
		updateMessage()
		return
	}

	for _, userID := range userIDs {

		IsRedeeming = true

		if Cancel {
			IsRedeeming = false
			Cancel = false
			results = append(results, "Redemption cancelled")
			updateMessage()
			return
		}

		err := processUser(userID, code, &results, updateMessage, s, m, i)
		if err != nil {
			results = append(results, err.Error())
			updateMessage()
			break
		}
		time.Sleep(DELAY_DURATION)

	}

	IsRedeeming = false

	updateMessage() // Send final update
}

func sendErrorMessage(s *discordgo.Session, m *discordgo.MessageCreate, i *discordgo.InteractionCreate, content string) {
	if i != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
	} else {
		s.ChannelMessageSend(m.ChannelID, content)
	}
}

func sendInitialMessage(s *discordgo.Session, m *discordgo.MessageCreate, i *discordgo.InteractionCreate, code string) *discordgo.Message {
	content := "Processing gift code: **" + code + "**"
	if i != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
		return nil
	}

	msg, err := s.ChannelMessageSend(m.ChannelID, content)
	if err != nil {
		log.Println("Error sending message:", err)
	}
	return msg
}

func createUpdateMessageFunc(s *discordgo.Session, m *discordgo.MessageCreate, i *discordgo.InteractionCreate, msg *discordgo.Message, code string, results *[]string) func() {
	return func() {
		content := fmt.Sprintf("Processing gift code: **%s**\n\n%s", code, strings.Join(*results, "\n"))
		if i != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		} else if msg != nil {
			_, err := s.ChannelMessageEdit(m.ChannelID, msg.ID, content)
			if err != nil {
				log.Println("Error updating message:", err)
				fmt.Println("Error updating message:", err)
			}
		}
	}
}

func processUser(userID, code string, results *[]string, updateMessage func(), s *discordgo.Session, m *discordgo.MessageCreate, i *discordgo.InteractionCreate) (err error) {
	log.Print("Processing user:", userID, " code:", code)
	playerData, err := retryRequest(func() (*models.Player, error) {
		return getRoleInfo(userID)
	})
	if err != nil {
		status := handlePlayerInfoError(err, userID)
		*results = append(*results, status)
		updateMessage()
		return
	}

	mongodb.UpdateUser(*playerData)

	// exchangeData, err := exchangeCode(userID, code)
	exchangeData, err := retryRequest(func() (*models.ExchangeResponse, error) {
		return exchangeCode(userID, code)
	})
	if err != nil {
		if strings.Contains(err.Error(), "TIME ERROR") {

			return fmt.Errorf("CDK TIME ERROR")
		}
		if exchangeData != nil {
			if exchangeData.Msg == "CDK NOT FOUND." { // {"code":1,"data":[],"msg":"CDK NOT FOUND.","err_code":40014}
				return fmt.Errorf("CDK NOT FOUND")
			}
		}

		status := handleExchangeError(err, playerData.Nickname)
		*results = append(*results, status)
	} else {

		status := getStatus(exchangeData.Msg) // exchangeData.Msg == SUCCESS
		*results = append(*results, fmt.Sprintf("%s - %s", playerData.Nickname, status))
	}

	content := fmt.Sprintf("Processing gift code: **%s**\n\n%s", code, strings.Join(*results, "\n"))

	// 将增加前的 results 消息发到新的消息，清空 results，再继续在老的消息继续更新
	if len(content) > 2000 {

		// 获取最后一个元素
		lastElement := (*results)[len(*results)-1]

		// 去掉最后一个元素
		*results = (*results)[:len(*results)-1]

		var ChannelID string
		if i != nil {
			ChannelID = i.ChannelID
		} else if m != nil {
			ChannelID = m.ChannelID
		}

		c := fmt.Sprintf("Processing gift code: **%s**\n\n%s", code, strings.Join(*results, "\n"))
		_, err := s.ChannelMessageSend(ChannelID, c)

		if err != nil {
			log.Println("Error sending message:", err)
			fmt.Println("Error sending message:", err)
		}

		*results = []string{}
		*results = append(*results, lastElement)

	}

	// 使用更新后的 updateMessage
	updateMessage()
	return nil
}

func handlePlayerInfoError(err error, userID string) string {
	if strings.Contains(err.Error(), "role not exist") {
		return fmt.Sprintf("%s - ❌ USER NOT FOUND", userID)
	}
	return fmt.Sprintf("Error getting player info for %s: %v", userID, err)
}

func handleExchangeError(err error, nickname string) string {
	if strings.Contains(err.Error(), "RECEIVED") {
		return fmt.Sprintf("%s - ❗️ %s", nickname, err.Error())
	}
	return fmt.Sprintf("%s - ❌ %v", nickname, err)
}

func getStatus(msg string) string {
	switch msg {
	case "RECEIVED":
		return "❗️"
	case "SUCCESS":
		return "✅"
	case "CDK NOT FOUND.":
		return "❌"
	default:
		return "❓"
	}
}

func getRoleInfo(userID string) (*models.Player, error) {
	data := url.Values{}
	data.Set("fid", userID)
	data.Set("time", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	signedData := appendSign(data)
	resp, err := http.PostForm(BASE_URL+"/player", signedData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var playerResp models.PlayerResponse
	err = json.Unmarshal(body, &playerResp)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshal error: %v", err)
	}

	if playerResp.Code != 0 {
		return nil, fmt.Errorf("API error: %s (Code: %d, ErrCode: %v)", playerResp.Msg, playerResp.Code, playerResp.ErrCode)
	}

	var player models.Player
	err = json.Unmarshal(playerResp.Data, &player)
	if err != nil {
		return nil, fmt.Errorf("failed to parse player data: %v", err)
	}

	return &player, nil
}

func exchangeCode(userID, code string) (*models.ExchangeResponse, error) {
	data := url.Values{}
	data.Set("fid", userID)
	data.Set("cdk", code)
	data.Set("time", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	signedData := appendSign(data)
	resp, err := http.PostForm(BASE_URL+"/gift_code", signedData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var exchangeResp models.ExchangeResponse
	err = json.Unmarshal(body, &exchangeResp)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshal error: %v", err)
	}

	if exchangeResp.Code != 0 {
		log.Print("Exchange Error:", exchangeResp.Msg)
		return &exchangeResp, fmt.Errorf("%s", exchangeResp.Msg)
	}

	return &exchangeResp, nil
}

func appendSign(data url.Values) url.Values {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(data.Get(k))
	}
	buf.WriteString(SECRET)

	hash := md5.Sum([]byte(buf.String()))
	sign := hex.EncodeToString(hash[:])

	data.Set("sign", sign)
	return data
}

func retryRequest[T any](fn func() (*T, error)) (*T, error) {
	var (
		err     error
		result  *T
		backoff = RETRY_DELAY
	)

	for attempt := 0; attempt < MAX_RETRIES; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		var playerResp models.PlayerResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &playerResp); jsonErr == nil {
			if (playerResp.Msg == "TIMEOUT RETRY.") ||
				(playerResp.Msg == "40004" && playerResp.ErrCode == 40004) {
				log.Printf("Received error. Retrying in %v. Attempt %d/%d\n", backoff, attempt+1, MAX_RETRIES)
				time.Sleep(backoff)
				backoff += time.Second
				continue
			}
		}

		if strings.Contains(err.Error(), "Too Many Attempts") ||
			strings.Contains(err.Error(), "invalid character '<' looking for beginning of value") ||
			strings.Contains(err.Error(), "Sign Error") ||
			strings.Contains(err.Error(), "TIMEOUT RETRY") {
			log.Printf("Received error. Retrying in %v. Attempt %d/%d\n", backoff, attempt+1, MAX_RETRIES)
			time.Sleep(backoff)
			backoff += time.Second
		} else {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries reached: %w", err)
}
