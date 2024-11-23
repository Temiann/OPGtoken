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

// Определите структуру Config
type Config struct {
	TelegramAPIToken string `json:"telegram_api_token"`
}

// Мапа для хранения состояния языка для каждого пользователя
var userLanguages = sync.Map{}

// Мапа для хранения ID сообщений для каждого пользователя
var messageIDs = sync.Map{}

func main() {
	// Откройте файл конфигурации
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("failed to open config file: %s", err)
	}
	defer configFile.Close()

	// Прочитайте содержимое файла
	byteValue, _ := ioutil.ReadAll(configFile)

	// Декодируйте JSON в структуру Config
	var config Config
	json.Unmarshal(byteValue, &config)

	// Инициализируйте бота
	bot, err := tgbotapi.NewBotAPI(config.TelegramAPIToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // Включите режим отладки, чтобы видеть отладочную информацию

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Создайте канал для получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Создайте канал для получения сигналов
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Обработка входящих сообщений
	for {
		select {
		case update := <-updates:
			if update.Message == nil && update.CallbackQuery == nil { // Игнорируем любые обновления, которые не являются сообщениями или callback query
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
				} else {
					// Обработка выбора языка
					switch update.Message.Text {
					case "English 🇺🇸":
						userLanguages.Store(update.Message.Chat.ID, "en")
						sendPhotoMessage(bot, update.Message.Chat.ID, "en")
					case "Русский 🇷🇺":
						userLanguages.Store(update.Message.Chat.ID, "ru")
						sendPhotoMessage(bot, update.Message.Chat.ID, "ru")
					}
				}
			} else if update.CallbackQuery != nil {
				handleSupportUs(bot, update)
			}
		case <-stop:
			log.Println("Received stop signal, shutting down...")
			return
		}
	}
}

// Функция для создания клавиатуры с выбором языка
func languageKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("English 🇺🇸"),
			tgbotapi.NewKeyboardButton("Русский 🇷🇺"),
		),
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true
	return keyboard
}

// Функция для отправки фото с инлайн-клавиатурой в зависимости от языка
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

	// Путь к фотографии
	photoPath := "logo.png" // Замените на фактический путь к вашей фотографии

	// Отправка фото
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(photoPath))
	photo.Caption = caption
	photo.ReplyMarkup = &keyboard
	sentPhoto, _ := bot.Send(photo)

	// Сохранение ID сообщения для редактирования
	messageIDs.Store(chatID, sentPhoto.MessageID)
}

// Функция для создания инлайн-клавиатуры с опциями
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
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return keyboard
}

// Функция для редактирования подписи фото с инлайн-клавиатурой в зависимости от языка
func editPhotoCaption(bot *tgbotapi.BotAPI, chatID int64, language string) {
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

	// Получите ID сообщения для редактирования
	messageID, ok := messageIDs.Load(chatID)
	if !ok {
		log.Printf("Message ID not found for chat ID: %d", chatID)
		return
	}

	// Редактирование подписи фото
	editMsg := tgbotapi.NewEditMessageCaption(chatID, int(messageID.(int64)), caption)
	editMsg.ReplyMarkup = &keyboard
	bot.Send(editMsg)
}

// Обработчик для кнопки "Support us"
func handleSupportUs(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		if update.CallbackQuery.Data == "support_us" {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "BTC - UQALnuvdPTMzLeSX67pfPrn3tvmlcN-Q7GsWgyQLIMd1QBBQ\nTON - UQALnuvdPTMzLeSX67pfPrn3tvmlcN-Q7GsWgyQLIMd1QBBQ\ndonation alerts - https://www.donationalerts.com/r/neyman_opg")
			bot.Send(msg)
		}
	}
}
