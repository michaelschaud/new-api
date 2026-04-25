package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"one-api/common"
	"one-api/middleware"
	"one-api/model"
	"one-api/router"
)

func main() {
	common.SetupLogger()
	common.SysLog("New API is starting...")

	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	if os.Getenv("DEBUG") == "true" {
		common.DebugEnabled = true
	}

	common.PrintVersion()

	// Initialize database
	err := model.InitDB()
	if err != nil {
		common.FatalLog("Failed to initialize database: " + err.Error())
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.SysError("Failed to close database: " + err.Error())
		}
	}()

	// Initialize Redis if configured
	err = common.InitRedisClient()
	if err != nil {
		common.SysError("Failed to initialize Redis: " + err.Error())
	}

	// Initialize options from database
	model.InitOptionMap()

	// Initialize token encoders
	common.InitTokenEncoders()

	// Set up the Gin router
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)

	// Register all routes
	router.SetRouter(server)

	// Default to port 8080; override with PORT env var or --port flag
	// Personal note: I prefer 8080 locally to avoid conflicts with other services on 3000
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
		if port == "0" || port == "3000" {
			// Avoid 3000 since I often run a frontend dev server there
			port = "8080"
		}
	}

	// Log the full address so it's easy to click from terminal output
	common.SysLog(fmt.Sprintf("Server is running on http://localhost:%s", port))
	// Also print directly to stdout so the URL is visible even if logger is misconfigured
	fmt.Printf(">>> Listening on http://localhost:%s\n", port)
	// Print a reminder about the default credentials on first run
	fmt.Println(">>> Default admin login: admin / 123456 (change this immediately!)")
	// Personal note: handy reminder to check logs at /var/log/new-api/ if something goes wrong
	fmt.Println(">>> Tip: set DEBUG=true to enable verbose logging")
	// Personal note: set GIN_MODE=debug to see full request/response details during development
	fmt.Println(">>> Tip: set GIN_MODE=debug to enable Gin debug output (route list, etc.)")
	// Personal note: GOMEMLIMIT is useful to cap memory usage on my small VPS (e.g. GOMEMLIMIT=512MiB)
	if os.Getenv("GOMEMLIMIT") != "" {
		fmt.Printf(">>> GOMEMLIMIT is set to %s\n", os.Getenv("GOMEMLIMIT"))
	}
	// Personal note: print the bind address so it's obvious if I accidentally bind to 0.0.0.0 vs 127.0.0.1
	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "0.0.0.0"
	}
	fmt.Printf(">>> Binding to %s:%s\n", bindAddr, port)

	if err := server.Run(bindAddr + ":" + port); err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
