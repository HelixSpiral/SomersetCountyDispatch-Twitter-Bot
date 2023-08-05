package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	somersetcountywrapper "github.com/HelixSpiral/SomersetCountyAPIWrapper"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// Some initial Twitter setup
	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessSecret := os.Getenv("ACCESS_SECRET")

	// Some initial Mastodon setup
	mastodonServer := os.Getenv("MASTODON_SERVER")
	mastodonClientID := os.Getenv("MASTODON_CLIENT_ID")
	mastodonClientSecret := os.Getenv("MASTODON_CLIENT_SECRET")
	mastodonUser := os.Getenv("MASTODON_USERNAME")
	mastodonPass := os.Getenv("MASTODON_PASSWORD")

	// Some initial MQTT setup
	mqttBroker := os.Getenv("MQTT_BROKER")
	mqttClientId := os.Getenv("MQTT_CLIENT_ID")
	mqttTopic := os.Getenv("MQTT_TOPIC")

	options := mqtt.NewClientOptions().AddBroker(mqttBroker).SetClientID(mqttClientId)
	options.WriteTimeout = 20 * time.Second
	mqttClient := mqtt.NewClient(options)

	currentDate := time.Now()
	loc, err := time.LoadLocation("EST") // Somerset County API is in EST
	if err != nil {
		panic(err)
	}

	var logTmp Cache
	var updates []somersetcountywrapper.DispatchLog

	// If this errors the file doesn't exist yet, so just set empty values for the cache.
	logTmp, err = readCache("/tmp/cache.json")
	if err != nil {
		logTmp = Cache{
			Day:           currentDate.In(loc).Day(),
			LastProcessed: "00-0",
			LogMap:        make(map[string][]somersetcountywrapper.DispatchLog),
		}
	}

	// If the day in the cache is before today, clear the cache and remove the file.
	if logTmp.Day < currentDate.In(loc).Day() {
		err = os.Remove("/tmp/cache.json")
		if err != nil {
			panic(err)
		}

		logTmp = Cache{
			Day:           currentDate.In(loc).Day(),
			LastProcessed: "00-00000",
			LogMap:        make(map[string][]somersetcountywrapper.DispatchLog),
		}
	}

	sw := somersetcountywrapper.NewWrapper()

	logs, err := sw.GetDispatch(currentDate.In(loc).Format("20060102"))
	if err != nil {
		panic(err)
	}

	lastProcessedString := strings.Split(logTmp.LastProcessed, "-")[1]
	lastProcessed, err := strconv.Atoi(lastProcessedString)
	if err != nil {
		panic(err)
	}

	var continueOuter bool
	for _, y := range logs {
		continueOuter = false
		currentlyProcessingString := strings.Split(y.CallNum, "-")[1]
		currentlyProcessing, err := strconv.Atoi(currentlyProcessingString)
		if err != nil {
			panic(err)
		}

		if currentlyProcessing <= lastProcessed {
			for _, b := range logTmp.LogMap[y.CallNum] {
				if b == y {
					continueOuter = true
					break
				}
			}

			if continueOuter {
				continue
			}
		}

		if currentlyProcessing > lastProcessed {
			lastProcessed = currentlyProcessing
			logTmp.LastProcessed = y.CallNum
		}

		updates = append(updates, y)
		logTmp.LogMap[y.CallNum] = append(logTmp.LogMap[y.CallNum], y)
	}

	// Connect to the MQTT broker
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// We limit to 1 request every 5 seconds
	limiter := time.Tick(time.Second * 5)
	for _, y := range updates {
		<-limiter

		message := buildMessage(y)

		jsonMsg, err := json.Marshal(&MqttMessage{
			MastodonServer:       mastodonServer,
			MastodonClientID:     mastodonClientID,
			MastodonClientSecret: mastodonClientSecret,
			MastodonUser:         mastodonUser,
			MastodonPass:         mastodonPass,

			TwitterConsumerKey:    consumerKey,
			TwitterConsumerSecret: consumerSecret,
			TwitterAccessToken:    accessToken,
			TwitterAccessSecret:   accessSecret,

			Message: message,
		})
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Sending message:", message)
		token := mqttClient.Publish(mqttTopic, 2, false, jsonMsg)
		_ = token.Wait()
		if token.Error() != nil {
			panic(err)
		}
	}

	err = writeCache("/tmp/cache.json", logTmp)
	if err != nil {
		panic(err)
	}

	mqttClient.Disconnect(250)
}

func readCache(f string) (Cache, error) {
	var logTmp Cache
	rawdata, err := os.ReadFile(f)
	if err != nil {
		return Cache{}, err
	}

	err = json.Unmarshal(rawdata, &logTmp)
	if err != nil {
		return Cache{}, err
	}

	return logTmp, nil
}

func writeCache(f string, logTmp Cache) error {
	jsonData, err := json.Marshal(logTmp)
	if err != nil {
		return err
	}

	_ = jsonData
	err = os.WriteFile(f, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
