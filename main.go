package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	TelegramAPIToken string `json:"telegram_api_token"`
}

var userLanguages = sync.Map{}

var messageIDs = sync.Map{}

func main() {

	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("failed to open config file: %s", err)
	}
	defer configFile.Close()

	byteValue, _ := ioutil.ReadAll(configFile)

	var config Config
	json.Unmarshal(byteValue, &config)

	bot, err := tgbotapi.NewBotAPI(config.TelegramAPIToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case update := <-updates:
			if update.Message == nil && update.CallbackQuery == nil {
				continue
			}

			if update.Message != nil {
				if update.Message.IsCommand() {
					switch update.Message.Command() {
					case "start":
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select language")
						msg.ReplyMarkup = languageKeyboard()
						sentMsg, _ := bot.Send(msg)
						messageIDs.Store(update.Message.Chat.ID, sentMsg.MessageID)
					}
				}
			} else if update.CallbackQuery != nil {
				if update.CallbackQuery.Data == "support_us" {
					handleSupportUs(bot, update)
				} else if update.CallbackQuery.Data == "en" || update.CallbackQuery.Data == "ru" {
					userLanguages.Store(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
					sendPhotoMessage(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
				} else if update.CallbackQuery.Data == "back_to_language" {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Select language")
					msg.ReplyMarkup = languageKeyboard()
					sentMsg, _ := bot.Send(msg)
					messageIDs.Store(update.CallbackQuery.Message.Chat.ID, sentMsg.MessageID)
				}
			}
		case <-stop:
			log.Println("Received stop signal, shutting down...")
			return
		}
	}
}

func languageKeyboard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("English 🇺🇸", "en"),
			tgbotapi.NewInlineKeyboardButtonData("Русский 🇷🇺", "ru"),
		),
	)
	return keyboard
}

func sendPhotoMessage(bot *tgbotapi.BotAPI, chatID int64, language string) {
	var caption string
	var keyboard tgbotapi.InlineKeyboardMarkup

	if language == "en" {
		caption = `🌟 Join the Future with OPG Token! 🌟

Dear Friends,

We are thrilled to introduce our new OPG token! Here are a few reasons why you should become part of our community and invest in OPG:

- Active and Large Community
- Based on a Real Person with a massive Twitch community
- Support and Interaction

Don't miss the opportunity to be part of something bigger and invest in the future with OPG Token. Together, we can reach new heights!

Best regards,
The OPG Token Team`
		keyboard = optionsInlineKeyboard("en")
	} else {
		caption = `🌟 Присоединяйтесь к будущему с токеном OPG! 🌟

Дорогие друзья,

Мы рады представить наш новый токен OPG! Вот несколько причин, почему вы должны стать частью нашего сообщества и инвестировать в OPG:

- Активное и большое сообщество
- Основан на реальном человеке с огромным сообществом на Twitch
- Поддержка и взаимодействие

Не упустите возможность стать частью чего-то большего и инвестировать в будущее с токеном OPG. Вместе мы можем достичь новых высот!

С уважением,
Команда токена OPG`
		keyboard = optionsInlineKeyboard("ru")
	}

	photoPath := "logo.png"

	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(photoPath))
	photo.Caption = caption
	photo.ReplyMarkup = &keyboard
	sentPhoto, _ := bot.Send(photo)

	messageIDs.Store(chatID, sentPhoto.MessageID)
}

func optionsInlineKeyboard(language string) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	if language == "en" {
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonURL("tg channel", "https://t.me/OPGtoken"),
				tgbotapi.NewInlineKeyboardButtonData("support us", "support_us"),
			},
			{
				tgbotapi.NewInlineKeyboardButtonURL("discord community", "https://discord.gg/ztpURnD2S5"),
			},
			{
				tgbotapi.NewInlineKeyboardButtonData("back to language selection", "back_to_language"),
			},
		}
	} else {
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonURL("tg канал", "https://t.me/OPGtoken"),
				tgbotapi.NewInlineKeyboardButtonData("поддержи нас", "support_us"),
			},
			{
				tgbotapi.NewInlineKeyboardButtonURL("сообщество в discord", "https://discord.gg/ztpURnD2S5"),
			},
			{
				tgbotapi.NewInlineKeyboardButtonData("назад к выбору языка", "back_to_language"),
			},
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return keyboard
}

func handleSupportUs(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		if update.CallbackQuery.Data == "support_us" {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "BTC - UQALnuvdPTMzLeSX67pfPrn3tvmlcN-Q7GsWgyQLIMd1QBBQ\nTON - UQALnuvdPTMzLeSX67pfPrn3tvmlcN-Q7GsWgyQLIMd1QBBQ\ndonation alerts - https://www.donationalerts.com/r/neyman_opg")
			bot.Send(msg)
		}
	}
}
