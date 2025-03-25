package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Vote represents a voting poll in the system.
type Vote struct {
	ID        string
	CreatorID string
	ChannelID string
	Question  string
	Options   []string
	Votes     map[string]int // userID -> optionIndex
	CreatedAt time.Time
	ClosedAt  *time.Time
}

// NewVote creates a new Vote instance with required validation.
func NewVote(creatorID, channelID, question string, options []string) *Vote {
	if len(options) < 2 {
		panic("vote must have at least 2 options")
	}

	return &Vote{
		ID:        generateID(),
		CreatorID: creatorID,
		ChannelID: channelID,
		Question:  question,
		Options:   options,
		Votes:     make(map[string]int),
		CreatedAt: time.Now(),
	}
}

// IsClosed checks if the vote has been closed.
func (v *Vote) IsClosed() bool {
	return v.ClosedAt != nil
}

// Close marks the vote as closed by setting ClosedAt to current time.
func (v *Vote) Close() {
	now := time.Now()
	v.ClosedAt = &now
}

// AddVote registers a user's vote for a specific option.
func (v *Vote) AddVote(userID string, optionIndex int) error {
	if v.IsClosed() {
		return fmt.Errorf("vote is closed")
	}
	if optionIndex < 0 || optionIndex >= len(v.Options) {
		return fmt.Errorf("invalid option index")
	}
	v.Votes[userID] = optionIndex
	return nil
}

// Results calculates and returns vote counts per option.
func (v *Vote) Results() map[int]int {
	results := make(map[int]int)
	for _, optIndex := range v.Votes {
		results[optIndex]++
	}
	return results
}

// FormatResults generates a human-readable string representation of vote results.
func (v *Vote) FormatResults() string {
	results := v.Results()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Results for: %s\n", v.Question))

	for i, option := range v.Options {
		count := results[i]
		percentage := 0
		if len(v.Votes) > 0 {
			percentage = count * 100 / len(v.Votes)
		}
		sb.WriteString(fmt.Sprintf("%d. %s - %d votes (%d%%)\n", i+1, option, count, percentage))
	}
	sb.WriteString(fmt.Sprintf("Total votes: %d\n", len(v.Votes)))
	return sb.String()
}

// generateID creates a random 8-byte hexadecimal string ID.
func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
