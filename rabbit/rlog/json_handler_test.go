package rlog

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var _ Publisher = (*publisher)(nil)

type mockPublisher struct {
	logs [][]byte
}

// Publish implements rlog.Publisher.
func (p *mockPublisher) Publish(_ context.Context, body []byte, _ string) error {
	p.logs = append(p.logs, body)
	return nil
}

type attr struct {
	Key      string        `json:"key"`
	Number   int           `json:"number"`
	Duration time.Duration `json:"duration"`
	Date     time.Time     `json:"date"`
}

type Lmsg struct {
	Msg  string `json:"msg"`
	Attr attr   `json:"attr"`
}

func Test_json(t *testing.T) {
	pub := mockPublisher{
		logs: [][]byte{},
	}

	now := time.Now()
	date, err := time.Parse(time.RFC3339Nano, now.Round(0).Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}

	expected := Lmsg{
		Msg: "message",
		Attr: attr{
			Key:      "value",
			Number:   12345,
			Duration: 325 * time.Second,
			Date:     date,
		},
	}

	h := NewJSONHandler(&pub, slog.HandlerOptions{})
	logger := slog.New(h)
	logger.Info(expected.Msg, "attr", expected.Attr)
	// logger.Info("message")

	actual := Lmsg{}
	if err := json.Unmarshal(pub.logs[0], &actual); err != nil {
		t.Fatal(err)
	}

	expects := []Lmsg{expected}
	for i, logg := range pub.logs {
		actual := Lmsg{}
		if err := json.Unmarshal(logg, &actual); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expects[i], actual)
	}

}
