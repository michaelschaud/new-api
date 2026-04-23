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

	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}

	common.SysLog(fmt.Sprintf("Server is running on port %s", port))

	if err := server.Run(":" + port); err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
