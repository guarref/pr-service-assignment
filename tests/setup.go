package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web"
	"github.com/guarref/pr-service-assignment/tests/mocks"
	"github.com/labstack/echo/v4"
)

type TestSetup struct {
	Echo            *echo.Echo
	Server          *httptest.Server
	TeamRepo        *mocks.MockTeamRepository
	UserRepo        *mocks.MockUserRepository
	PRRepo          *mocks.MockPullRequestRepository
	StatsRepo       *mocks.MockStatsRepository
	TeamService     *service.TeamService
	UserService     *service.UserService
	PRService       *service.PullRequestService
	StatsService    *service.StatsService
}

func NewTestSetup() *TestSetup {
	// Создаем моки
	teamRepo := new(mocks.MockTeamRepository)
	userRepo := new(mocks.MockUserRepository)
	prRepo := new(mocks.MockPullRequestRepository)
	statsRepo := new(mocks.MockStatsRepository)

	// Создаем сервисы с моками
	teamService := service.NewTeamService(teamRepo)
	userService := service.NewUserService(userRepo)
	prService := service.NewPullRequestService(prRepo, userRepo)
	statsService := service.NewStatsService(statsRepo)

	// Создаем Echo сервер
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Регистрируем маршруты
	web.RegisterRoutes(e, teamService, userService, prService, statsService)

	// Создаем тестовый HTTP сервер
	server := httptest.NewServer(e)

	return &TestSetup{
		Echo:         e,
		Server:       server,
		TeamRepo:     teamRepo,
		UserRepo:     userRepo,
		PRRepo:       prRepo,
		StatsRepo:    statsRepo,
		TeamService:  teamService,
		UserService:  userService,
		PRService:    prService,
		StatsService: statsService,
	}
}

func (ts *TestSetup) Close() {
	if ts.Server != nil {
		ts.Server.Close()
	}
}

func (ts *TestSetup) DoRequest(method, path string, body []byte) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, ts.Server.URL+path, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, ts.Server.URL+path, nil)
	}
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return http.DefaultClient.Do(req)
}

