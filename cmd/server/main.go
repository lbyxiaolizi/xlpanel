package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/openhost/openhost/docs"
	"github.com/openhost/openhost/internal/infrastructure/http/handlers"
)

// @title OpenHost API
// @version 0.1
// @description OpenHost API for provisioning and billing.
// @BasePath /api/v1
func main() {
	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
	api.GET("/health", handlers.Health)

	_ = http.ListenAndServe(":8080", router)
}
