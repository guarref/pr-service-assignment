package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/repository"
	"github.com/guarref/pr-service-assignment/internal/errs"
)

type PullRequestService struct {
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
}

func NewPullRequestService(prRepo repository.PullRequestRepository, userRepo repository.UserRepository) *PullRequestService {
	return &PullRequestService{prRepo: prRepo, userRepo: userRepo}
}

// func (prs *PullRequestService) CreatePullRequest(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error) {
	
// 	if pr == nil || pr.PullRequestID == "" || pr.PullRequestName == "" || pr.AuthorID == "" {
// 		return nil, errs.ErrBadRequest
// 	}

// 	author, err := prs.userRepo.GetUserByID(ctx, pr.AuthorID)
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting author with id %s: %w", pr.AuthorID, err)
// 	}

// 	activeUsers, err := prs.userRepo.GetActiveUsersByTeam(ctx, author.TeamName, author.UserID)
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting active users for team %s: %w", author.TeamName, err)
// 	}

// 	reviewers := randomUserSelection(activeUsers, 2)
// 	pr.AssignedReviewers = reviewers
// 	pr.Status = models.PullRequestOpen

// 	if err := prs.prRepo.CreatePullRequest(ctx, pr); err != nil {
// 		return nil, fmt.Errorf("error creating pull request with id %s: %w", pr.PullRequestID, err)
// 	}

// 	return pr, nil
// }

func (prs *PullRequestService) CreatePullRequest(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error) {
	
	if pr == nil || pr.PullRequestID == "" || pr.PullRequestName == "" || pr.AuthorID == "" {
		return nil, errs.ErrBadRequest
	}

	author, err := prs.userRepo.GetUserByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("error getting author with id %s: %w", pr.AuthorID, err)
	}

	activeUsers, err := prs.userRepo.GetActiveUsersByTeam(ctx, author.TeamName, author.UserID)
	if err != nil {
		return nil, fmt.Errorf("error getting active users for team %s: %w", author.TeamName, err)
	}

	reviewers := randomUserSelection(activeUsers, 2)
	pr.AssignedReviewers = reviewers
	pr.Status = models.PullRequestOpen

	created, err := prs.prRepo.CreatePullRequest(ctx, pr)
	if err != nil {
		return nil, fmt.Errorf("error creating pull request with id %s: %w", pr.PullRequestID, err)
	}

	if len(created.AssignedReviewers) == 0 && len(pr.AssignedReviewers) > 0 {
		created.AssignedReviewers = append([]string(nil), pr.AssignedReviewers...)
	}

	return created, nil
}

func (prs *PullRequestService) MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	
	if prID == "" {
		return nil, errs.ErrBadRequest
	}

	pr, err := prs.prRepo.MergePullRequestByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("error merging pull request with id %s: %w", prID, err)
	}

	return pr, nil
}

func (prs *PullRequestService) ReassignToPullRequest(ctx context.Context, prID string, oldUserID string) (*models.PullRequest, string, error) {
	
	if prID == "" || oldUserID == "" {
		return nil, "", errs.ErrBadRequest
	}

	pr, newReviewerID, err := prs.prRepo.ReassignToPullRequest(ctx, prID, oldUserID)
	if err != nil {
		return nil, "", fmt.Errorf("error reassigning reviewer %s for pull request %s: %w", oldUserID, prID, err)
	}

	return pr, newReviewerID, nil
}

func (prs *PullRequestService) GetPullRequestsByReviewer(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {
	
	if userID == "" {
		return nil, errs.ErrBadRequest
	}

	prsList, err := prs.prRepo.GetPullRequestByReviewerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting pull requests for reviewer %s: %w", userID, err)
	}

	return prsList, nil
}

func randomUserSelection(activeUsers []*models.User, num int) []string {
	
	if len(activeUsers) == 0 {
		return []string{}
	}

	count := min(len(activeUsers), num)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(activeUsers), func(i, j int) {
		activeUsers[i], activeUsers[j] = activeUsers[j], activeUsers[i]
	})

	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = activeUsers[i].UserID
	}

	return result
}
