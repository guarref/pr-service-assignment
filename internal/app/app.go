package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/guarref/pr-service-assignment/config"
	pg "github.com/guarref/pr-service-assignment/pkg/postgres"

	"github.com/guarref/pr-service-assignment/internal/repository/postgres"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web"
)

type App struct {
	cfg  *config.Config
	db   *pg.PDB
	echo *echo.Echo
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {

	db, err := pg.NewPDB(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("error db connect: %w", err)
	}

	if cfg.MigrateEnable {
		if err := db.Migrate(cfg.MigrateFolder); err != nil {
			return nil, fmt.Errorf("error db migrate: %w", err)
		}
	}

	teamRepo := postgres.NewTeamRepository(db.DB)
	userRepo := postgres.NewUserRepository(db.DB)
	prRepo := postgres.NewPullRequestRepository(db.DB, userRepo)

	teamSvc := service.NewTeamService(teamRepo)
	userSvc := service.NewUserService(userRepo)
	prSvc := service.NewPullRequestService(prRepo, userRepo)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	web.RegisterRoutes(e, teamSvc, userSvc, prSvc)

	return &App{cfg: cfg, db: db, echo: e}, nil
}

func (a *App) Run(ctx context.Context) error {

	serverErr := make(chan error, 1)

	go func() {
		addr := fmt.Sprintf(":%d", a.cfg.Port)
		fmt.Printf("SERVER is starting on port%s\n", addr)

		err := a.echo.Start(addr)
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("\nshutdown was caused")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.echo.Shutdown(shutdownCtx); err != nil {
			_ = a.echo.Close()
			a.db.Close()
			return fmt.Errorf("\nshutdown(echo): %w", err)
		}

		a.db.Close()

		return nil
	case err := <-serverErr:
		a.db.Close()
		return fmt.Errorf("server error: %w", err)
	}
}
