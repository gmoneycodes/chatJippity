package main

import (
	"context"
	"github.com/PullRequestInc/go-gpt3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
	"log"
	"strings"
)

type Config struct {
	TelegramToken string `mapstructure:"telegramToken"`
	OpenAIToken   string `mapstructure:"openaiToken"`
	Preamble      string `mapstructure:"preamble"`
}

func LoadConfig(path string) (c Config) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	viper.AutomaticEnv()

	var err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&c)
	return
}

func sendChatGPT(apiKey, sendText string) string {

	ctx := context.Background()

	client := gpt3.NewClient(apiKey)
	var response string

	err := client.CompletionStreamWithEngine(ctx, gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
		Prompt:      []string{sendText},
		MaxTokens:   gpt3.IntPtr(100),
		Temperature: gpt3.Float32Ptr(0.8),
	}, func(res *gpt3.CompletionResponse) {
		response += res.Choices[0].Text
	})

	if err != nil {
		log.Println(err)
		return "Sorry! ChatJippity is not available currently"
	}
	return response
}

func main() {

	var userPrompt string
	var gptPrompt string
	config := LoadConfig(".")
	apiKey := config.OpenAIToken

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !strings.HasPrefix(update.Message.Text, "/topic") && !strings.HasPrefix(update.Message.Text, "/word") {
			continue
		}

		if strings.HasPrefix(update.Message.Text, "/topic") && strings.HasPrefix(update.Message.Text, "/word") {
			userPrompt = strings.TrimPrefix(update.Message.Text, "/topic")
			gptPrompt = config.Preamble + "TOPIC: "
		} else if strings.HasPrefix(update.Message.Text, "/word") {
			userPrompt = strings.TrimPrefix(update.Message.Text, "/word")
			gptPrompt = config.Preamble + "WORD: "
		}

		if userPrompt != "" {
			gptPrompt += userPrompt
			res := sendChatGPT(apiKey, gptPrompt)
			update.Message.Text = res
		} else {
			update.Message.Text = "Please enter your topic or word"
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
