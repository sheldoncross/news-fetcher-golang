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

// Fetches the top headlines from the News API for Canada. decodes the response JSON into a NewsApiResponse struct, and 
// extracts the titles of the articles. The function returns a slice of strings containing the titles of the top headlines.
func getHeadlines() []string {
	news_apiKey := os.Getenv("NEWS_API_KEY")
	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=anime&apiKey=%s", news_apiKey)

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


// Creates a chat session and returns the chat session and the genai client.
func createChatSession() (*genai.ChatSession, *genai.Client) {
	//Load key
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .enf file", err)
	}
	gemini_apiKey := os.Getenv("GEMINI_API_KEY")

	//Create context and initialize the genai client for gemini
	context := context.Background()
	client, err := genai.NewClient(context, option.WithAPIKey(gemini_apiKey))
	if err != nil {
		log.Fatal(err)
	}

	model := client.GenerativeModel("gemini-1.5-pro")
	chatSession := model.StartChat()
	return chatSession, client
}


// Sends a message using the provided chatSession and returns the response.
func sendMessage(chatSession *genai.ChatSession, msg string) *genai.GenerateContentResponse {
	context := context.Background()
	res, err := chatSession.SendMessage(context, genai.Text(msg))
    if err != nil {
        log.Fatal(err)
    }
    return res
}

// Generates a voiceover for a given chat session and title.
func generateVoiceover(chatSession *genai.ChatSession, title string) string {
    data := fmt.Sprintf("Read this headline in the voice of a newscaster, but keep it short. %s", title)
    res := sendMessage(chatSession, data)
    for _, candidate := range res.Candidates {
        if candidate.Content != nil {
            switch part := candidate.Content.Parts[0].(type) {
            case genai.Text:
                return string(part)
            }
        }
    }
    return "Sorry, there was error!"
}

// Plays the given voiceover using text-to-speech synthesis.
func playVoiceover(voiceover string) {
    speech := htgotts.Speech{Folder: "audio", Language: voices.English}
    fmt.Printf(voiceover)
    speech.Speak(voiceover)
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	// Fetch headlines from News API
	titles := getHeadlines()

	// Create chat session and client for Gemini
	chatSession, client := createChatSession()
	defer client.Close()

	// Generate voiceover and play for headlines longer than 25 characters
	for _, title := range titles {
		if len(title) > 25 {
			voiceover := generateVoiceover(chatSession, title)
			playVoiceover(voiceover)
			time.Sleep(2 * time.Second)
		}
	}
}