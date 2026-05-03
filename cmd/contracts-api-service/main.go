package main

import (
    "log"
    "os"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/thepikart/contracts-webapi/api"
    "github.com/thepikart/contracts-webapi/internal/contracts"
)

func main() {
    log.Printf("Server started")
    port := os.Getenv("CONTRACTS_API_PORT")
    if port == "" {
        port = "8080"
    }
    environment := os.Getenv("CONTRACTS_API_ENVIRONMENT")
    if !strings.EqualFold(environment, "production") { // case insensitive comparison
        gin.SetMode(gin.DebugMode)
    }
    engine := gin.New()
    engine.Use(gin.Recovery())
    // request routings
    handleFunctions := &contracts.ApiHandleFunctions{
        HospitalContractsAPI: contracts.NewHospitalContractsApi(),
    }
    contracts.NewRouterWithGinEngine(engine, *handleFunctions)
    engine.GET("/openapi", api.HandleOpenApi)
    engine.Run(":" + port)
}