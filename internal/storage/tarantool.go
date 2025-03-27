package storage

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/tarantool/go-tarantool"
	"mattermost-voting-bot/internal/model"
)

// TarantoolStorage implements Storage interface for Tarantool
type TarantoolStorage struct {
	conn *tarantool.Connection
}

// NewTarantoolStorage creates new Tarantool storage instance
func NewTarantoolStorage(addr string, opts tarantool.Opts) (*TarantoolStorage, error) {
	conn, err := tarantool.Connect(addr, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Tarantool")
	}

	return &TarantoolStorage{conn: conn}, nil
}

// CreateVote stores new vote in Tarantool
func (s *TarantoolStorage) CreateVote(ctx context.Context, vote *model.Vote) error {
	closedAt := ""
	if vote.ClosedAt != nil {
		closedAt = vote.ClosedAt.Format(time.RFC3339)
	}

	_, err := s.conn.Insert("votes", []interface{}{
		vote.ID,
		vote.CreatorID,
		vote.ChannelID,
		vote.Question,
		vote.Options,
		vote.Votes,
		vote.CreatedAt.Format(time.RFC3339),
		closedAt,
	})

	return err
}

// GetVote retrieves vote by ID
func (s *TarantoolStorage) GetVote(ctx context.Context, voteID string) (*model.Vote, error) {
	resp, err := s.conn.Select("votes", "primary", 0, 1, tarantool.IterEq, []interface{}{voteID})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, ErrNotFound
	}

	return s.unmarshalVote(resp.Data[0].([]interface{}))
}

// AddVote registers user's vote
func (s *TarantoolStorage) AddVote(ctx context.Context, voteID, userID string, optionIdx int) error {
	resp, err := s.conn.Call("vote.add_vote", []interface{}{voteID, userID, optionIdx})
	if err != nil {
		return err
	}
	if len(resp.Data) > 0 && resp.Data[0] != nil {
		switch resp.Data[0].(string) {
		case "Vote not found":
			return ErrNotFound
		case "Vote is closed":
			return ErrVoteClosed
		case "Already voted":
			return ErrAlreadyVoted
		default:
			return errors.New(resp.Data[0].(string))
		}
	}
	return nil
}

// CloseVote marks vote as closed
func (s *TarantoolStorage) CloseVote(ctx context.Context, voteID, userID string) error {
	vote, err := s.GetVote(ctx, voteID)
	if err != nil {
		return err
	}
	if vote.CreatorID != userID {
		return ErrPermissionDenied
	}

	_, err = s.conn.Update("votes", "primary", []interface{}{voteID}, []interface{}{
		[]interface{}{"=", "closed_at", time.Now().Format(time.RFC3339)},
	})
	return err
}

// DeleteVote removes vote from storage
func (s *TarantoolStorage) DeleteVote(ctx context.Context, voteID, userID string) error {
	vote, err := s.GetVote(ctx, voteID)
	if err != nil {
		return err
	}
	if vote.CreatorID != userID {
		return ErrPermissionDenied
	}

	_, err = s.conn.Delete("votes", "primary", []interface{}{voteID})
	return err
}

// ListChannelVotes returns all votes for a channel
func (s *TarantoolStorage) ListChannelVotes(ctx context.Context, channelID string) ([]*model.Vote, error) {
	resp, err := s.conn.Select("votes", "channel", 0, 0, tarantool.IterEq, []interface{}{channelID})
	if err != nil {
		return nil, err
	}

	votes := make([]*model.Vote, 0, len(resp.Data))
	for _, item := range resp.Data {
		vote, err := s.unmarshalVote(item.([]interface{}))
		if err != nil {
			return nil, err
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

// unmarshalVote converts Tarantool tuple to Vote struct
func (s *TarantoolStorage) unmarshalVote(data []interface{}) (*model.Vote, error) {
	closedAt := data[7].(string)
	var closedAtTime *time.Time
	if closedAt != "" {
		t, err := time.Parse(time.RFC3339, closedAt)
		if err != nil {
			return nil, err
		}
		closedAtTime = &t
	}

	options := make([]string, 0)
	for _, opt := range data[4].([]interface{}) {
		options = append(options, opt.(string))
	}

	votes := make(map[string]int)
	if data[5] != nil {
		for k, v := range data[5].(map[interface{}]interface{}) {
			votes[k.(string)] = int(v.(uint64))
		}
	}

	return &model.Vote{
		ID:        data[0].(string),
		CreatorID: data[1].(string),
		ChannelID: data[2].(string),
		Question:  data[3].(string),
		Options:   options,
		Votes:     votes,
		CreatedAt: mustParseTime(data[6].(string)),
		ClosedAt:  closedAtTime,
	}, nil
}

func mustParseTime(t string) time.Time {
	parsed, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return time.Now()
	}
	return parsed
}
