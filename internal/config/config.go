package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RedisConfig struct {
	Host        string
	Port        string
	Password    string
	DB          int
	TaskListTTL time.Duration
}

type JWTConfig struct {
	Secret string
	Issuer string
	TTL    time.Duration
}

type HTTPConfig struct {
	Port              string
	RateLimitRPS      int
	RateLimitBurst    int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	httpReadHeaderTimeout, err := getEnvDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpReadTimeout, err := getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpWriteTimeout, err := getEnvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpIdleTimeout, err := getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpShutdownTimeout, err := getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	connMaxLifetime, err := getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}

	connMaxIdleTime, err := getEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	jwtTTL, err := getEnvDuration("JWT_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	taskListCacheTTL, err := getEnvDuration("TASK_LIST_CACHE_TTL", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTP: HTTPConfig{
			Port:              getEnv("APP_PORT", "8080"),
			RateLimitRPS:      getEnvInt("RATE_LIMIT_RPS", 10),
			RateLimitBurst:    getEnvInt("RATE_LIMIT_BURST", 20),
			ReadHeaderTimeout: httpReadHeaderTimeout,
			ReadTimeout:       httpReadTimeout,
			WriteTimeout:      httpWriteTimeout,
			IdleTimeout:       httpIdleTimeout,
			ShutdownTimeout:   httpShutdownTimeout,
		},

		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "mysql"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "3306"),
			Name:            getEnv("DB_NAME", "task_manager"),
			User:            getEnv("DB_USER", "app"),
			Password:        getEnv("DB_PASSWORD", "app_password"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdleTime,
		},

		Redis: RedisConfig{
			Host:        getEnv("REDIS_HOST", "localhost"),
			Port:        getEnv("REDIS_PORT", "6379"),
			Password:    getEnv("REDIS_PASSWORD", ""),
			DB:          getEnvInt("REDIS_DB", 0),
			TaskListTTL: taskListCacheTTL,
		},

		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Issuer: getEnv("JWT_ISSUER", "task-manager"),
			TTL:    jwtTTL,
		},
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.HTTP.Port == "" {
		return fmt.Errorf("HTTP port is required")
	}

	if c.HTTP.RateLimitRPS <= 0 {
		return fmt.Errorf("rate limit RPS must be greater than zero")
	}

	if c.HTTP.RateLimitBurst <= 0 {
		return fmt.Errorf("rate limit burst must be greater than zero")
	}

	if c.HTTP.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("HTTP read header timeout must be greater than zero")
	}

	if c.HTTP.ReadTimeout <= 0 {
		return fmt.Errorf("HTTP read timeout must be greater than zero")
	}

	if c.HTTP.WriteTimeout <= 0 {
		return fmt.Errorf("HTTP write timeout must be greater than zero")
	}

	if c.HTTP.IdleTimeout <= 0 {
		return fmt.Errorf("HTTP idle timeout must be greater than zero")
	}

	if c.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("HTTP shutdown timeout must be greater than zero")
	}

	if c.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Port == "" {
		return fmt.Errorf("database port is required")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("database max open connections must be greater than zero")
	}

	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("database max idle connections cannot be negative")
	}

	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("database max idle connections cannot exceed max open connections")
	}

	if c.Database.ConnMaxLifetime <= 0 {
		return fmt.Errorf("database connection max lifetime must be greater than zero")
	}

	if c.Database.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("database connection max idle time must be greater than zero")
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	if c.Redis.Port == "" {
		return fmt.Errorf("redis port is required")
	}

	if c.Redis.DB < 0 {
		return fmt.Errorf("redis db cannot be negative")
	}

	if c.Redis.TaskListTTL <= 0 {
		return fmt.Errorf("task list cache TTL must be greater than zero")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)

	if !exists || value == "" {
		return defaultValue
	}

	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := getEnv(key, "")

	if value == "" {
		return defaultValue
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return result
}

func getEnvDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	value := getEnv(key, "")

	if value == "" {
		return defaultValue, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}

	return duration, nil
}
