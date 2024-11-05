// cmd/init.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Go project",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := promptUser("Enter the project module name: ", "app")
		port := promptUser("Enter the HTTP port number: ", "8080")
		dbConnection := promptUser("Enter the database connection string: ", "postgres://user:pass@localhost/dbname")

		writeEnvFile(projectName, port, dbConnection)

		initializeProject(projectName)
		fmt.Println("Project initialized successfully! happy coding :)")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func promptUser(message, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

func writeEnvFile(projectName, port, dbConnection string) {

	file, err := os.Create(".env")

	if err != nil {
		fmt.Println("Error creating .env file:", err)
		return
	}

	defer file.Close()

	envContent := fmt.Sprintf("PROJECT_NAME=\"%s\"\nPORT=:%s\nDATABASE_URL=\"%s\"\nJWT_SECRET_KEY=\"%s\"\nJWT_TOKEN_DURATION=\"%s\"\n", projectName, port, dbConnection, "mysecretkey", "24h")

	_, _ = file.WriteString(envContent)

}

func initializeProject(moduleName string) {
	_, err := runCommand("go", "mod", "init", moduleName)
	if err != nil {
		fmt.Println("Error initializing go module:", err)
		return
	}

	createMainGoFile(moduleName)

	createConfigFile()
	createRoutesFile(moduleName)
	createLoggerFile()
	createDBFile(moduleName)

	_, err = runCommand("go", "mod", "tidy")
	if err != nil {
		fmt.Println("Error running go mod tidy:", err)
		return
	}
}

// runCommand runs a shell command and returns the output and error
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func createMainGoFile(moduleName string) {
	mainFileContent := fmt.Sprintf(`package main

import (
	"%s/api/routes"
	"%s/infrastructure/config"
	"%s/infrastructure/database"
	"%s/logger"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	loadConfiguration()
	setupServer()
}

func loadConfiguration() {
	err := config.Load()

	if err != nil {
		panic(err)
	}
}

func setupRateLimiter(router *gin.Engine) {
	limiter := tollbooth.NewLimiter(30, nil)
	router.Use(tollbooth_gin.LimitHandler(limiter))
}

func setupCors(router *gin.Engine) {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}

	router.Use(cors.New(corsConfig))
}

func setupRoutes(router *gin.Engine, db *gorm.DB) {
	routes.SetupRoutes(router, db)
}

func setupServer() {
	db, err := database.InitDB()

	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	logger.Init()

	setupRateLimiter(router)
	setupCors(router)
	setupRoutes(router, db)

	router.Static("/public", "./public")

	err = router.Run(config.Config.Port)

	if err != nil {
		panic(err)
	}
}

`, moduleName, moduleName, moduleName, moduleName)

	err := os.WriteFile("main.go", []byte(mainFileContent), 0644)
	if err != nil {
		fmt.Println("Error creating main.go file:", err)
		return
	}
}

func createConfigFile() {
	_ = os.MkdirAll("infrastructure/config", os.ModePerm)
	configFileContent := `package config

import (
	"log"
	"os"
	"time"
	"fmt"

	"github.com/joho/godotenv"
)

type Configuration struct {
	Name         string
	Port         string
	DbDsn        string
	JwtSecretKey string
	TokenDuration time.Duration
}

var Config *Configuration

func Load() error {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using defaults")
	}

	Config = &Configuration{
		Name:                             getEnvOrError("PROJECT_NAME"),
		Port:                             getEnvOrError("PORT"),
		DbDsn:                            getEnvOrError("DATABASE_URL"),
		JwtSecretKey:                     getEnvOrError("JWT_SECRET_KEY"),
		TokenDuration:                    getEnvAsDuration("JWT_TOKEN_DURATION"),
	}

	return nil
}

func getEnvOrError(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	panic(fmt.Sprintf("Environment variable %s not set", key))
}

func getEnvAsInt(key string) int64 {
	valueStr := getEnvOrError(key)
	var value int64
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		log.Printf("\nError loading %s: %v", key, err)
		panic(err)
	}
	return value
}

func getEnvAsBool(key string) bool {
	valueStr := getEnvOrError(key)
	return valueStr == "true"
}

func getEnvAsDuration(key string) time.Duration {
	valueStr := getEnvOrError(key)
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("\nError parsing duration for %s: %v", key, err)
		panic(err)
	}
	return duration
}

`

	err := os.WriteFile("infrastructure/config/config.go", []byte(configFileContent), 0644)
	if err != nil {
		fmt.Println("Error creating config.go file:", err)
		return
	}
}

func createRoutesFile(moduleName string) {
	_ = os.MkdirAll("api/routes", os.ModePerm)
	routesFileContent := fmt.Sprintf(`package routes

import (
	"fmt"
	"net/http"
	"%s/infrastructure/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	apiRouter := router.Group("/api")

	apiRouter.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%%s server running on port: %%s", config.Config.Name, config.Config.Port),
		})
	})
}
`, moduleName)

	err := os.WriteFile("api/routes/routes.go", []byte(routesFileContent), 0644)
	if err != nil {
		fmt.Println("Error creating routes.go file:", err)
		return
	}
}

func createLoggerFile() {
	_ = os.MkdirAll("logger", os.ModePerm)
	loggerFileContent := `package logger

import (
	"os"
	"time"
	"github.com/sirupsen/logrus"
	"github.com/natefinch/lumberjack"
)

var Log = logrus.New()

func Init() {
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetOutput(&lumberjack.Logger{
		Filename:   "logs/" + time.Now().Format("02-01-2006") + ".log",
		MaxSize:    10, // Max 10 MB
		MaxBackups: 3,  // Max 3 backups
		MaxAge:     30, // Max 30 days
		Compress:   true,
	})

	logFile, err := os.OpenFile("logs/application.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		Log.SetOutput(logFile)
	} else {
		Log.Warn("Could not log to file, using default stderr")
	}
}

func Info(msg string) {
	Log.Info(msg)
}

func Error(msg string) {
	Log.Error(msg)
}
`

	err := os.WriteFile("logger/logger.go", []byte(loggerFileContent), 0644)
	if err != nil {
		fmt.Println("Error creating logger.go file:", err)
		return
	}
}

func createDBFile(moduleName string) {
	dbGoContent := fmt.Sprintf(`package database

import (
	"fmt"
	"log"
	"os"
	"%s/infrastructure/config"
	"strings"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlserver.Open(config.Config.DbDsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		}),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true,
			NoLowerCase:   true,
			NameReplacer:  strings.NewReplacer("CID", "Cid"),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %%v", err)
	}

	return db, nil
}
`, moduleName)
	err := os.MkdirAll("infrastructure/database", os.ModePerm)
	if err != nil {
		fmt.Println("Error creating database directory:", err)
		return
	}
	err = os.WriteFile("infrastructure/database/db.go", []byte(dbGoContent), 0644)
	if err != nil {
		fmt.Println("Error creating db.go:", err)
	}
}
