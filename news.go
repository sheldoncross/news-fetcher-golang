package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type NewsApiResponse struct {
	Articles []struct {
		Title			string `json:"title"`
		Description		string `json:"description"`
		URL				string `json:"url"`
	} `json:"articles"`
}

func main() {
	apiKey := "b2654a987dde4d0ba228c0c44e5141f6"
	endpoint := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=us&apiKey=%s", apiKey)

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

	fmt.Printf("Fetching News\n")

	for _, article := range newsData.Articles {
		fmt.Printf("Title: %s\n", article.Title)
		fmt.Printf("Description: %s\n", article.Description)
		fmt.Printf("URL: %s\n\n", article.URL)
	}
}