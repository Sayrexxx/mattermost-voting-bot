package storage

import (
	"context"
	"errors"
	"mattermost-voting-bot/internal/model"
)

// Storage defines the interface for vote storage operations
type Storage interface {
	CreateVote(ctx context.Context, vote *model.Vote) error
	GetVote(ctx context.Context, voteID string) (*model.Vote, error)
	AddVote(ctx context.Context, voteID, userID string, option int) error
	CloseVote(ctx context.Context, voteID, userID string) error
	DeleteVote(ctx context.Context, voteID, userID string) error
	ListChannelVotes(ctx context.Context, channelID string) ([]*model.Vote, error)
}

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyVoted     = errors.New("already voted")
	ErrVoteClosed       = errors.New("vote closed")
	ErrPermissionDenied = errors.New("permission denied")
)
