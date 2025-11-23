package tests

import (
	"context"
	"time"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web"
	"github.com/labstack/echo/v4"
)

// SetupTestApp creates a test application with mocked repositories
func SetupTestApp() (*echo.Echo, *MockTeamRepository, *MockUserRepository, *MockPullRequestRepository, *MockStatsRepository) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	mockPRRepo := new(MockPullRequestRepository)
	mockStatsRepo := new(MockStatsRepository)

	teamSvc := service.NewTeamService(mockTeamRepo)
	userSvc := service.NewUserService(mockUserRepo)
	prSvc := service.NewPullRequestService(mockPRRepo, mockUserRepo)
	statsSvc := service.NewStatsService(mockStatsRepo)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	web.RegisterRoutes(e, teamSvc, userSvc, prSvc, statsSvc)

	return e, mockTeamRepo, mockUserRepo, mockPRRepo, mockStatsRepo
}

// Helper functions to create test models

func CreateTestTeam(teamName string, members []models.TeamMember) *models.Team {
	return &models.Team{
		TeamName: teamName,
		Members:  members,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func CreateTestTeamMember(userID, username string, isActive bool) models.TeamMember {
	return models.TeamMember{
		UserID:   userID,
		UserName: username,
		IsActive: isActive,
	}
}

func CreateTestUser(userID, username, teamName string, isActive bool) *models.User {
	return &models.User{
		UserID:    userID,
		UserName:  username,
		TeamName:  teamName,
		IsActive:  isActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func CreateTestPullRequest(prID, prName, authorID string, status models.PullRequestStatus, reviewers []string) *models.PullRequest {
	now := time.Now()
	var mergedAt *time.Time
	if status == models.PullRequestMerged {
		mergedAt = &now
	}

	return &models.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            status,
		AssignedReviewers: reviewers,
		CreatedAt:         now,
		MergedAt:          mergedAt,
	}
}

func CreateTestPullRequestShort(prID, prName, authorID string, status models.PullRequestStatus) *models.PullRequestShort {
	return &models.PullRequestShort{
		PullRequestID:   prID,
		PullRequestName: prName,
		AuthorID:        authorID,
		Status:          status,
	}
}

func CreateTestStats(totalTeams, totalUsers, activeUsers, totalPRs, openPRs int, topReviewers []*models.TopReviewer) *models.Stats {
	return &models.Stats{
		TotalTeams:        totalTeams,
		TotalUsers:        totalUsers,
		ActiveUsers:       activeUsers,
		TotalPullRequests: totalPRs,
		OpenPullRequests:  openPRs,
		TopReviewers:      topReviewers,
	}
}

func CreateTestTopReviewer(userID, username string, reviewCount int) *models.TopReviewer {
	return &models.TopReviewer{
		UserID:      userID,
		UserName:    username,
		ReviewCount: reviewCount,
	}
}

// GetTestContext returns a test context
func GetTestContext() context.Context {
	return context.Background()
}

