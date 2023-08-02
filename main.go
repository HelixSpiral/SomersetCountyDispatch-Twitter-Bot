package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	somersetcountywrapper "github.com/HelixSpiral/SomersetCountyAPIWrapper"
)

func main() {
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

	if logTmp.XAppRateLimited {
		if logTmp.XAppLimit24HourReset < time.Now().Unix() {
			logTmp.XAppRateLimited = false
		}
	}

	// If the day in the cache is before today, clear the cache and remove the file.
	if logTmp.Day < currentDate.In(loc).Day() {
		err = os.Remove("/tmp/cache.json")
		if err != nil {
			panic(err)
		}

		// We need to save these values, can't reset them daily.
		tmp1 := logTmp.XAppLimit24HourReset
		tmp2 := logTmp.XAppRateLimited

		logTmp = Cache{
			Day:           currentDate.In(loc).Day(),
			LastProcessed: "00-00000",
			LogMap:        make(map[string][]somersetcountywrapper.DispatchLog),
		}
		logTmp.XAppLimit24HourReset = tmp1
		logTmp.XAppRateLimited = tmp2
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

	// We limit to 1 request every 5 seconds
	limiter := time.Tick(time.Second * 5)
	for _, y := range updates {

		<-limiter

		// As long as we're not rate limited on Twitter, send it.
		if !logTmp.XAppRateLimited {
			err = processDispatchTwitter(y)
			switch e := err.(type) {
			case *RateLimitError:
				logTmp.XAppLimit24HourReset = e.Reset
				logTmp.XAppRateLimited = true
			default:
				log.Println(e)
			}
		}
	}

	err = writeCache("/tmp/cache.json", logTmp)
	if err != nil {
		panic(err)
	}
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
