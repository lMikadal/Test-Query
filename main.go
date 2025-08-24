package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

type App struct {
	Echo        *echo.Echo
	MainDB      *sql.DB
	ReadDB      *sql.DB
	WriteDB     *sql.DB
	RedisClient *redis.Client
}

type BenchmarkData struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Age          int       `json:"age"`
	City         string    `json:"city"`
	Country      string    `json:"country"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RandomNumber int       `json:"random_number"`
	Description  string    `json:"description"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	app := &App{
		Echo: echo.New(),
	}

	// Setup middleware
	app.Echo.Use(middleware.Logger())
	app.Echo.Use(middleware.Recover())
	app.Echo.Use(middleware.CORS())

	// Connect to databases
	if err := app.connectDatabases(); err != nil {
		log.Fatal("Failed to connect to databases:", err)
	}

	// Connect to Redis
	if err := app.connectRedis(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Setup routes
	app.setupRoutes()

	// Start server
	port := getEnv("APP_PORT", "8080")
	log.Printf("Server starting on port %s", port)
	app.Echo.Logger.Fatal(app.Echo.Start(":" + port))
}

func (app *App) connectDatabases() error {
	// Main database connection
	mainDBURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres123"),
		getEnv("DB_NAME", "benchmark_db"),
	)

	var err error
	app.MainDB, err = sql.Open("postgres", mainDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to main database: %v", err)
	}

	// Read database connection
	readDBURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_READ_HOST", "localhost"),
		getEnv("DB_READ_PORT", "5433"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres123"),
		getEnv("DB_READ_NAME", "benchmark_read_db"),
	)

	app.ReadDB, err = sql.Open("postgres", readDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to read database: %v", err)
	}

	// Write database connection
	writeDBURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_WRITE_HOST", "localhost"),
		getEnv("DB_WRITE_PORT", "5434"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres123"),
		getEnv("DB_WRITE_NAME", "benchmark_write_db"),
	)

	app.WriteDB, err = sql.Open("postgres", writeDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to write database: %v", err)
	}

	// Test connections
	if err := app.MainDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping main database: %v", err)
	}

	if err := app.ReadDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping read database: %v", err)
	}

	if err := app.WriteDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping write database: %v", err)
	}

	log.Println("Successfully connected to all databases")
	return nil
}

func (app *App) connectRedis() error {
	app.RedisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s",
			getEnv("REDIS_HOST", "localhost"),
			getEnv("REDIS_PORT", "6379")),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	ctx := context.Background()
	if err := app.RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")
	return nil
}

func (app *App) setupRoutes() {
	// Health check
	app.Echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	api := app.Echo.Group("/api/v1")

	// Benchmark routes
	api.GET("/benchmark/redis/:id", app.getBenchmarkDataRedis)
	api.GET("/benchmark/db/:id", app.getBenchmarkDataDB)
	api.GET("/benchmark/split-read/:id", app.getBenchmarkDataSplitRead)
	api.POST("/benchmark/split-write", app.createBenchmarkDataSplitWrite)
	api.POST("/benchmark/seed", app.seedBenchmarkData)
	api.GET("/benchmark/stats", app.getBenchmarkStats)
}

// Get data using Redis cache
func (app *App) getBenchmarkDataRedis(c echo.Context) error {
	id := c.Param("id")
	ctx := context.Background()

	// Try to get from Redis first
	cacheKey := fmt.Sprintf("benchmark:%s", id)
	cached, err := app.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		return c.JSONBlob(http.StatusOK, []byte(cached))
	}

	// If not in cache, get from database
	data, err := app.getBenchmarkFromDB(id, app.MainDB)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Data not found"})
	}

	// Cache the result for 5 minutes
	jsonData := fmt.Sprintf(`{"id":%d,"name":"%s","email":"%s","age":%d,"city":"%s","country":"%s","created_at":"%s","updated_at":"%s","random_number":%d,"description":"%s"}`,
		data.ID, data.Name, data.Email, data.Age, data.City, data.Country, data.CreatedAt.Format(time.RFC3339), data.UpdatedAt.Format(time.RFC3339), data.RandomNumber, data.Description)
	app.RedisClient.Set(ctx, cacheKey, jsonData, 5*time.Minute)

	return c.JSON(http.StatusOK, data)
}

// Get data directly from main database
func (app *App) getBenchmarkDataDB(c echo.Context) error {
	id := c.Param("id")
	data, err := app.getBenchmarkFromDB(id, app.MainDB)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Data not found"})
	}

	return c.JSON(http.StatusOK, data)
}

// Get data from read database (split DB scenario)
func (app *App) getBenchmarkDataSplitRead(c echo.Context) error {
	id := c.Param("id")
	data, err := app.getBenchmarkFromDB(id, app.ReadDB)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Data not found"})
	}

	return c.JSON(http.StatusOK, data)
}

// Create data in write database (split DB scenario)
func (app *App) createBenchmarkDataSplitWrite(c echo.Context) error {
	var data BenchmarkData
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data"})
	}

	query := `
		INSERT INTO benchmark_data (name, email, age, city, country, random_number, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := app.WriteDB.QueryRow(query, data.Name, data.Email, data.Age, data.City, data.Country, data.RandomNumber, data.Description).Scan(
		&data.ID, &data.CreatedAt, &data.UpdatedAt,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create data"})
	}

	return c.JSON(http.StatusCreated, data)
}

// Seed benchmark data
func (app *App) seedBenchmarkData(c echo.Context) error {
	countStr := c.QueryParam("count")
	count := 1000 // default

	if countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil {
			count = parsed
		}
	}

	log.Printf("Seeding %d records...", count)

	// Seed main database
	if err := app.seedDatabase(app.MainDB, count); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to seed main database"})
	}

	// Seed read database
	if err := app.seedDatabase(app.ReadDB, count); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to seed read database"})
	}

	// Seed write database
	if err := app.seedDatabase(app.WriteDB, count); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to seed write database"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Data seeded successfully",
		"count":   count,
	})
}

// Get benchmark statistics
func (app *App) getBenchmarkStats(c echo.Context) error {
	mainCount, _ := app.getTableCount(app.MainDB)
	readCount, _ := app.getTableCount(app.ReadDB)
	writeCount, _ := app.getTableCount(app.WriteDB)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"main_db_count":   mainCount,
		"read_db_count":   readCount,
		"write_db_count":  writeCount,
		"redis_connected": app.RedisClient.Ping(context.Background()).Err() == nil,
	})
}

// Helper functions
func (app *App) getBenchmarkFromDB(id string, db *sql.DB) (*BenchmarkData, error) {
	query := `
		SELECT id, name, email, age, city, country, created_at, updated_at, random_number, description
		FROM benchmark_data
		WHERE id = $1
	`

	var data BenchmarkData
	err := db.QueryRow(query, id).Scan(
		&data.ID, &data.Name, &data.Email, &data.Age, &data.City, &data.Country,
		&data.CreatedAt, &data.UpdatedAt, &data.RandomNumber, &data.Description,
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (app *App) seedDatabase(db *sql.DB, count int) error {
	for i := 0; i < count; i++ {
		query := `
			INSERT INTO benchmark_data (name, email, age, city, country, random_number, description)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err := db.Exec(query,
			fmt.Sprintf("User %d", i+1),
			fmt.Sprintf("user%d@example.com", i+1),
			20+i%60,
			fmt.Sprintf("City %d", i%100),
			fmt.Sprintf("Country %d", i%50),
			i*123%10000,
			fmt.Sprintf("Description for user %d", i+1),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) getTableCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM benchmark_data").Scan(&count)
	return count, err
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
