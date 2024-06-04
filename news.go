package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"github.com/hegedustibor/htgo-tts"
	"github.com/hegedustibor/htgo-tts/voices"
)

type NewsApiResponse struct {
	Articles []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
	} `json:"articles"`
}

func main() {
	gemini()
}

func getHeadlines() []string {
	news_apiKey := os.Getenv("NEWS_API_KEY")
	endpoint := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=ca&apiKey=%s", news_apiKey)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var newsData NewsApiResponse
	err = json.NewDecoder(resp.Body).Decode(&newsData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fetching News\n\n")

	var titles []string

	for _, article := range newsData.Articles {
		titles = append(titles, article.Title)
	}

	return titles
}

func gemini() {
	//Load key
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .enf file", err)
	}
	gemini_apiKey := os.Getenv("GEMINI_API_KEY")

	//Create context and init
	context := context.Background()
	client, err := genai.NewClient(context, option.WithAPIKey(gemini_apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	titles := getHeadlines()

	model := client.GenerativeModel("gemini-1.5-pro")
	cs := model.StartChat()

	send := func(msg string) *genai.GenerateContentResponse {
		res, err := cs.SendMessage(context, genai.Text(msg))
		if err != nil {
			log.Fatal(err)
		}
		return res
	}

	for _, title := range titles {
		if len(title) > 25 {
			fmt.Printf("%s\n\n", title)
			data := fmt.Sprintf("Read this headline in the voice of a newscaster, but keep it short. %s", title)
			res := send(data)
			for _, candidate := range res.Candidates {
				if candidate.Content != nil {
					switch part := candidate.Content.Parts[0].(type) {
					case genai.Text:
						speech := htgotts.Speech{Folder: "audio", Language: voices.English}
						fmt.Printf(string(part))
						speech.Speak(string(part))
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
}
