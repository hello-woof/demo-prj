package model

import (
	"strings"
	"time"
)

type LogItems []LogEntity

type LogEntity struct {
	Service       string        `json:"service"`
	Level         string        `json:"level"`
	Timestamp     JsonTimestamp `json:"timestamp"`
	Operation     string        `json:"operation"`
	Message       string        `json:"message"`
	TransactionID string        `json:"transaction_id"`
}

type JsonTimestamp time.Time

func (jt *JsonTimestamp) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	convertTime, err := time.Parse("2006-01-02 03:04:05.999999", s)
	if err != nil {
		return err
	}
	*jt = JsonTimestamp(convertTime)
	return nil
}
