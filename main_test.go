package main

import (
	"io"
	"momento/model"
	"strings"
	"testing"
	"time"
)

func TestStdJsonDecoder_DecodeJson(t *testing.T) {
	type args struct {
		decodeModel []model.LogEntity
		r           io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "fail json formatting",
			args: args{
				decodeModel: make([]model.LogEntity, 0),
				r:           strings.NewReader(`!@badJSON`),
			},
			wantErr: true,
		},
		{
			name: "Successfully load json",
			args: args{
				decodeModel: make([]model.LogEntity, 0),
				r:           strings.NewReader(goodJSON),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StdJsonDecoder{}
			if err := s.DecodeJson(&tt.args.decodeModel, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("DecodeJson() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStdParser_GetHighestErrorOp(t *testing.T) {
	type args struct {
		items []model.LogEntity
	}
	startTime, _ := time.Parse(time.RFC3339, "2006-01-02T00:00:00Z")

	simpleEntities := []model.LogEntity{
		{
			Service:       "Some Service",
			Level:         "ERROR",
			Timestamp:     model.JsonTimestamp(startTime),
			Operation:     "A",
			Message:       "A msg",
			TransactionID: "A-tid",
		},
		{
			Service:       "Some Service",
			Level:         "ERROR",
			Timestamp:     model.JsonTimestamp(startTime.Add(time.Millisecond * 5)),
			Operation:     "A",
			Message:       "A msg",
			TransactionID: "A-tid",
		},
		{
			Service:       "Some Service",
			Level:         "ERROR",
			Timestamp:     model.JsonTimestamp{},
			Operation:     "B",
			Message:       "B msg",
			TransactionID: "B-tid",
		},
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   int
		wantErr bool
	}{
		{
			name:    "successfully choose operation A as highest error",
			args:    args{items: simpleEntities},
			want:    "A",
			want1:   2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StdParser{}
			got, got1, err := s.GetHighestErrorOp(tt.args.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHighestErrorOp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetHighestErrorOp() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetHighestErrorOp() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestStdParser_GetLongestTransaction(t *testing.T) {
	type args struct {
		items []model.LogEntity
	}

	startTime, _ := time.Parse(time.RFC3339, "2006-01-02T00:00:00Z")

	simpleEntities := []model.LogEntity{
		{
			Service:       "Some Service",
			Level:         "ERROR",
			Timestamp:     model.JsonTimestamp(startTime),
			Operation:     "A",
			Message:       "A msg",
			TransactionID: "A-tid",
		},
		{
			Service:       "Some Service",
			Level:         "ERROR",
			Timestamp:     model.JsonTimestamp(startTime.Add(time.Duration(5) * time.Second)),
			Operation:     "A",
			Message:       "A msg",
			TransactionID: "A-tid",
		},
		{
			Service:       "Some Service",
			Level:         "ERROR",
			Timestamp:     model.JsonTimestamp{},
			Operation:     "B",
			Message:       "B msg",
			TransactionID: "B-tid",
		},
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   int64
		wantErr bool
	}{
		{
			name:    "successfully choose op with 5000 millis(5 secs) duration",
			args:    args{items: simpleEntities},
			want:    "A-tid",
			want1:   5000,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StdParser{}
			got, got1, err := s.GetLongestTransaction(tt.args.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLongestTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLongestTransaction() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetLongestTransaction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

const goodJSON = `
[
    {
        "service": "loadbalancer",
        "level": "INFO",
        "timestamp": "2017-10-17 00:00:00.000000",
        "operation": "POST",
        "message": "START /login requested",
        "transaction_id": "3cf9629d-c05c-4916-a800-f2be630b0bc9"
    },
    {
        "service": "webserver",
        "level": "DEBUG",
        "timestamp": "2017-10-17 00:00:00.207697",
        "operation": "/login",
        "message": "START Logging in user",
        "transaction_id": "3cf9629d-c05c-4916-a800-f2be630b0bc9"
    },
    {
        "service": "authentication_service",
        "level": "ERROR",
        "timestamp": "2017-10-17 00:00:01.038673",
        "operation": "AuthenticateUser",
        "message": "START Authenticating user",
        "transaction_id": "3cf9629d-c05c-4916-a800-f2be630b0bc9"
    }
]
`
