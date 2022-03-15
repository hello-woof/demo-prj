package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"momento/model"
)

type LogParser interface {
	GetHighestErrorOp(items []model.LogEntity) (string, int, error)
	GetLongestTransaction(items []model.LogEntity) (string, int64, error)
}

type JsonDecoder interface {
	DecodeJson(decodeModel *[]model.LogEntity, reader io.Reader) error
}

type Handler struct {
	Parser  LogParser
	Decoder JsonDecoder
}

func (h *Handler) run(r io.Reader) {
	items := make([]model.LogEntity, 0)
	err := h.Decoder.DecodeJson(&items, r)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to decode json object: %s", err))
		return
	}

	if os.Args[2] == "high" {
		operator, numErrors, err := h.Parser.GetHighestErrorOp(items)
		if err != nil {
			fmt.Println(fmt.Errorf("failed getting number of errors for operations: %s", err))
			return
		}

		fmt.Println(fmt.Sprintf("Operator (%s) has the highest error count (%d)", operator, numErrors))
	} else {
		transaction, transactionTime, err := h.Parser.GetLongestTransaction(items)
		if err != nil {
			fmt.Println(fmt.Errorf("failed getting longest transaction: %s", err))
			return
		}

		fmt.Println(fmt.Sprintf("Transaction (%s) has the longest running time (%d)", transaction, transactionTime))
	}
}

type StdJsonDecoder struct {
}

func (s *StdJsonDecoder) DecodeJson(decodeModel *[]model.LogEntity, r io.Reader) error {
	dec := json.NewDecoder(r)

	_, err := dec.Token()
	if err != nil {
		return err
	}

	for dec.More() {
		var item model.LogEntity
		err := dec.Decode(&item)
		if err != nil {
			return err
		}
		*decodeModel = append(*decodeModel, item)
	}

	return nil
}

type StdParser struct {
}

func (s *StdParser) GetHighestErrorOp(items []model.LogEntity) (string, int, error) {
	counts := map[string]int{}
	highestOperation := ""
	highestCount := 0

	for _, item := range items {
		if item.Level != "ERROR" {
			continue
		}

		counts[item.Operation]++
		if counts[item.Operation] > highestCount {
			highestOperation = item.Operation
			highestCount = counts[item.Operation]
		}
	}
	return highestOperation, highestCount, nil
}

func (s *StdParser) GetLongestTransaction(items []model.LogEntity) (string, int64, error) {
	type oldestLatestTime struct {
		Oldest time.Time
		Latest time.Time
	}
	if len(items) < 2 {
		return "", 0, errors.New("unable to get duration when < 2 entries exist")
	}

	times := map[string]oldestLatestTime{}
	longestTransaction := ""
	longestDuration := int64(0)

	for _, item := range items {
		recordedTime, exists := times[item.TransactionID]
		if !exists {
			times[item.TransactionID] = oldestLatestTime{
				Oldest: time.Time(item.Timestamp),
				Latest: time.Time(item.Timestamp),
			}
			continue
		}

		tItemTimestamp := time.Time(item.Timestamp)
		checkTime := false
		if tItemTimestamp.Before(recordedTime.Oldest) {
			recordedTime.Oldest = tItemTimestamp
			times[item.TransactionID] = recordedTime
			checkTime = true
		} else if tItemTimestamp.After(recordedTime.Latest) {
			recordedTime.Latest = tItemTimestamp
			times[item.TransactionID] = recordedTime
			checkTime = true
		}

		if checkTime {
			newDuration := recordedTime.Latest.Sub(recordedTime.Oldest).Milliseconds()
			if newDuration > longestDuration {
				longestDuration = newDuration
				longestTransaction = item.TransactionID
			}
		}
	}
	return longestTransaction, longestDuration, nil
}

func main() {
	h := Handler{
		Parser:  &StdParser{},
		Decoder: &StdJsonDecoder{},
	}

	file, err := os.Open(os.Args[1])
	defer file.Close()

	if err != nil {
		fmt.Println(fmt.Errorf("failed to open file: %s", err))
		return
	}

	h.run(file)
}
