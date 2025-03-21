package websocket

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

func NewWebSocketClient(url string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Send(method string, params any) (string, error) {
	requestID := uuid.NewString()
	message := map[string]any {
		"id":     requestID,
		"msg":    "method",
		"method": method,
		"params": params,
	}

	if err := c.conn.WriteJSON(message); err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return requestID, nil
}

func (c *Client) Receive() (map[string]any, error) {
	var response map[string]any
	if err := c.conn.ReadJSON(&response); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return response, nil
}

func (c *Client) PollJobStatus(jobID int) (string, error) {
	for {
		time.Sleep(2 * time.Second)

		queryID := uuid.NewString()
		queryMessage := map[string]any{
			"id":     queryID,
			"msg":    "method",
			"method": "core.job.query",
			"params": [][]any{
				{"id", "=", jobID},
			},
		}

		if err := c.conn.WriteJSON(queryMessage); err != nil {
			return "", fmt.Errorf("failed to send job query: %w", err)
		}

		var response struct {
			Result []struct {
				State string `json:"state"`
				Error string `json:"error"`
			} `json:"result"`
		}

		if err := c.conn.ReadJSON(&response); err != nil {
			return "", fmt.Errorf("failed to read job status: %w", err)
		}

		if len(response.Result) > 0 {
			job := response.Result[0]
			switch job.State {
			case "SUCCESS":
				return "SUCCESS", nil
			case "FAILED":
				return "FAILED", fmt.Errorf("job failed: %s", job.Error)
			default:
				// Job still in progress
			}
		}
	}
}

func (c *Client) Close() {
	c.conn.Close()
}
