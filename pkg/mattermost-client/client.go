package mmclient

import (
	"context"
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/sirupsen/logrus"
	"sync"
)

// Client manages connection to Mattermost server
type Client struct {
	api    *model.Client4
	ws     *model.WebSocketClient
	team   *model.Team
	user   *model.User
	logger *logrus.Logger
	mu     sync.Mutex
	done   chan struct{}
}

// NewClient creates new Mattermost client
func NewClient(url, token, teamName string, logger *logrus.Logger) (*Client, error) {
	api := model.NewAPIv4Client(url)
	api.SetToken(token)

	user, _, err := api.GetMe("")
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	team, _, err := api.GetTeamByName(teamName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &Client{
		api:    api,
		user:   user,
		team:   team,
		logger: logger,
		done:   make(chan struct{}),
	}, nil
}

// ConnectWebSocket establishes WebSocket connection
func (c *Client) ConnectWebSocket() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ws != nil {
		return nil
	}

	ws, err := model.NewWebSocketClient4(c.api.URL, c.api.AuthToken)
	if err != nil {
		return fmt.Errorf("failed to create websocket: %w", err)
	}

	ws.Listen()
	c.ws = ws
	return nil
}

// PostMessage sends message to channel
func (c *Client) PostMessage(channelID, text string) error {
	post := &model.Post{
		ChannelId: channelID,
		Message:   text,
	}

	_, _, err := c.api.CreatePost(post)
	return err
}

// ListenEvents processes WebSocket events
func (c *Client) ListenEvents(ctx context.Context, handler func(event *model.WebSocketEvent)) error {
	if err := c.ConnectWebSocket(); err != nil {
		return err
	}

	for {
		select {
		case event := <-c.ws.EventChannel:
			if event.EventType() == model.WebsocketEventPosted {
				go handler(event)
			}
		case <-c.done:
			return nil
		case <-ctx.Done():
			c.ws.Close()
			return nil
		}
	}
}

// Close releases resources
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ws != nil {
		close(c.done)
		c.ws.Close()
	}
}

// GetTeamID returns current team ID
func (c *Client) GetTeamID() string {
	return c.team.Id
}

// GetUserID returns bot user ID
func (c *Client) GetUserID() string {
	return c.user.Id
}
