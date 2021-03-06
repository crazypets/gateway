package tgbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type SDK struct {
	c       *http.Client
	baseURL string
}

// NewSDK create new sdk instance.
// url - url of telegram bot service.
func NewSDK(client *http.Client, url string) *SDK {
	return &SDK{
		c:       client,
		baseURL: url,
	}
}

type tgRequest struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

type tgErrResponse struct {
	Msg string `json:"message"`
}

// Send - send message via telegram bot.
func (s *SDK) Send(ctx context.Context, t time.Time, lvl string, msg string) error {
	req := tgRequest{t, lvl, msg}

	data, err := json.Marshal(&req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/notify", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("prepare post request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.c.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send post request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("tg response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var errResp tgErrResponse
		if err = json.Unmarshal(body, &errResp); err != nil {
			return fmt.Errorf("tg bot server error, code: %d, status: %s", resp.StatusCode, resp.Status)
		}

		return fmt.Errorf("tg bot server error, code: %d, status: %s, msg: %s", resp.StatusCode, resp.Status, errResp.Msg)
	}

	return nil
}
