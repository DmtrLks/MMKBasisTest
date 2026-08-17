package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mmktestbasisByDGanichev/internal/auth"
	"mmktestbasisByDGanichev/internal/cache"
	"mmktestbasisByDGanichev/internal/comment"
	"mmktestbasisByDGanichev/internal/config"
	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/history"
	"mmktestbasisByDGanichev/internal/middleware"
	"mmktestbasisByDGanichev/internal/server"
	"mmktestbasisByDGanichev/internal/stats"
	"mmktestbasisByDGanichev/internal/task"
	"mmktestbasisByDGanichev/internal/team"
	"mmktestbasisByDGanichev/internal/user"
	"os"
	"os/signal"
	"syscall"

	mysqlrepository "mmktestbasisByDGanichev/internal/repository/mysql"
	redisrepository "mmktestbasisByDGanichev/internal/repository/redis"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Printf("application stopped with error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("warning: .env file not loaded: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to MySQL: %w", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	dbClient := database.NewClient(db)

	if err := database.RunMigrations(ctx, dbClient); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Println("database migrations completed")

	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect to Redis: %w", err)
	}
	defer redisClient.Close()

	log.Println("Redis connection established")

	userRepository := mysqlrepository.NewUserRepository()

	userService := user.NewService(dbClient, userRepository)
	userHandler := user.NewHandler(userService)

	tokenManager := auth.NewTokenManager(cfg.JWT)
	authMiddleware := middleware.NewAuth(tokenManager)
	credentialsValidator := auth.NewCredentialsValidator()

	authService := auth.NewService(
		dbClient,
		userRepository,
		tokenManager,
		credentialsValidator,
		cfg.JWT.TTL,
	)
	authHandler := auth.NewHandler(authService)

	teamRepository := mysqlrepository.NewTeamRepository()
	teamValidator := team.NewValidator()

	teamService := team.NewService(dbClient, teamRepository, userRepository, teamValidator)

	teamHandler := team.NewHandler(teamService)

	taskRepository := mysqlrepository.NewTaskRepository()
	taskCache := redisrepository.NewTaskCache(redisClient, cfg.Redis.TaskListTTL)
	historyRepository := mysqlrepository.NewHistoryRepository()
	taskValidator := task.NewValidator()

	taskService := task.NewService(
		dbClient,
		taskRepository,
		historyRepository,
		teamRepository,
		taskValidator,
		taskCache,
	)
	taskHandler := task.NewHandler(taskService)

	commentRepository := mysqlrepository.NewCommentRepository()
	commentValidator := comment.NewValidator()
	commentService := comment.NewService(
		dbClient,
		commentRepository,
		taskRepository,
		teamRepository,
		commentValidator,
	)
	commentHandler := comment.NewHandler(commentService)

	historyService := history.NewService(dbClient, historyRepository, teamRepository)
	historyHandler := history.NewHandler(historyService)

	statsRepository := mysqlrepository.NewStatsRepository()
	statsService := stats.NewService(dbClient, statsRepository, teamRepository)
	statsHandler := stats.NewHandler(statsService)

	httpServer := server.New(
		cfg.HTTP,
		userHandler,
		authHandler,
		authMiddleware,
		teamHandler,
		taskHandler,
		commentHandler,
		historyHandler,
		statsHandler,
	)

	log.Printf("HTTP server listening on :%s", cfg.HTTP.Port)

	if err := httpServer.Run(ctx); err != nil {
		return fmt.Errorf("run HTTP server: %w", err)
	}

	log.Println("HTTP server stopped gracefully")

	return nil
}
