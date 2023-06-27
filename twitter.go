package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	somersetcountywrapper "github.com/HelixSpiral/SomersetCountyAPIWrapper"
	"github.com/dghubble/oauth1"
)

func processDispatch(d somersetcountywrapper.DispatchLog) error {
	message := buildMessage(d)

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessSecret := os.Getenv("ACCESS_SECRET")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)

	httpClient := config.Client(oauth1.NoContext, token)

	log.Printf("Tweeting message: %s\r\n", message)
	resp, err := httpClient.Post("https://api.twitter.com/2/tweets", "application/json",
		bytes.NewBuffer([]byte(fmt.Sprintf(`{"text": "%s"}`, message))))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Println("Tweet:", string(body))

	return nil
}
