package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	somersetcountywrapper "github.com/HelixSpiral/SomersetCountyAPIWrapper"
	"github.com/dghubble/oauth1"
)

func processDispatchTwitter(d somersetcountywrapper.DispatchLog) error {
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
	log.Println("Headers:", resp.Header)

	XAppLimit24HourRemainingString := resp.Header.Get("X-App-Limit-24Hour-Remaining")
	XAppLimit24HourResetString := resp.Header.Get("X-App-Limit-24Hour-Reset")

	XAppLimit24HourRemaining, err := strconv.Atoi(XAppLimit24HourRemainingString)
	if err != nil {
		return err
	}

	XAppLimit24HourReset, err := strconv.ParseInt(XAppLimit24HourResetString, 10, 64)
	if err != nil {
		return err
	}

	if XAppLimit24HourRemaining <= 0 {
		return &RateLimitError{
			Reset: XAppLimit24HourReset,
		}
	}

	if strings.Contains(string(body), "Too Many Requests") {
		return &RateLimitError{
			Reset: time.Now().Add(1 * time.Hour).Unix(),
		}
	}

	return nil
}
