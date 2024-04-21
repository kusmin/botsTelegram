package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type Article struct {
	Source struct {
		ID   *string `json:"id"`
		Name *string `json:"name"`
	} `json:"source"`
	Author      *string `json:"author"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
	PublishedAt string  `json:"publishedAt"`
}

type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type UserState struct {
	Articles    []Article
	CurrentPage int
}

var (
	usersState map[int64]*UserState
)

func main() {
	log.Println("Iniciando o bot...")
	bot, err := tb.NewBot(tb.Settings{
		Token:  "6594417646:AAGeKUGtSe0G_No_vrp1C5HwfFyGVpH0VgY",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	usersState = make(map[int64]*UserState)

	bot.Handle(tb.OnText, func(m *tb.Message) {
		switch m.Text {
		case "/start":
			sendMessageSafe(bot, m.Sender, "Bem-vindo ao Dev News Bot!")
			usersState[m.Sender.ID] = &UserState{CurrentPage: 0}
		case "/news":
			fetchNews(m.Sender.ID)
			sendNewsPage(bot, m.Sender)
		case "/more":
			if _, ok := usersState[m.Sender.ID]; !ok {
				fetchNews(m.Sender.ID)
				sendNewsPage(bot, m.Sender)
				return
			}
			sendNewsPage(bot, m.Sender)
		}
	})

	bot.Start()
}

func fetchNews(userID int64) {
	url := "https://newsapi.org/v2/top-headlines?country=br&category=technology&apiKey=7f46e1ca13624dd390ac8fd2a8dbc254"
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var news NewsResponse
	if err := json.Unmarshal(body, &news); err != nil {
		log.Println(err)
		return
	}

	userState, exists := usersState[userID]
	if !exists {
		userState = &UserState{}
		usersState[userID] = userState
	}
	userState.Articles = news.Articles
	userState.CurrentPage = 0
}

func sendNewsPage(bot *tb.Bot, recipient *tb.User) {
	userState := usersState[recipient.ID]
	if userState.CurrentPage*5 >= len(userState.Articles) {
		sendMessageSafe(bot, recipient, "Não há mais notícias.")
		return
	}

	var result string
	for i := 0; i < 5 && userState.CurrentPage*5+i < len(userState.Articles); i++ {
		article := userState.Articles[userState.CurrentPage*5+i]
		result += fmt.Sprintf("%s\n%s\nLink: %s\n\n", article.Title, article.Description, article.URL)
	}
	userState.CurrentPage++

	sendMessageSafe(bot, recipient, result)
}

func sendMessageSafe(bot *tb.Bot, recipient *tb.User, message string) {
	_, err := bot.Send(recipient, message)
	if err != nil {
		log.Printf("Erro ao enviar mensagem para %d: %v\n", recipient.ID, err)
	}
}
