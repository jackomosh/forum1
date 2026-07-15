package handler

import "forum/internal/domain"

type VoteRequest struct {
	Target   domain.VoteTarget
	TargetID int64
	Value    domain.VoteValue
}

type VoteResponse struct {
	TargetID     int64
	LikeCount    int
	DislikeCount int
	Score        int
	UserVote     domain.VoteValue
}
