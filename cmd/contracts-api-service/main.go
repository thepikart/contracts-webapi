package main

import (
    "log"
    "os"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/thepikart/contracts-webapi/api"
    "github.com/thepikart/contracts-webapi/internal/contracts"
	"github.com/thepikart/contracts-webapi/internal/db_service"
    "context"
    "time"
    "github.com/gin-contrib/cors"
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
	    corsMiddleware := cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
        ExposeHeaders:    []string{""},
        AllowCredentials: false,
        MaxAge: 12 * time.Hour,
    })
    engine.Use(corsMiddleware)
	
    // setup context update middleware
    dbService := db_service.NewMongoService[contracts.Contract](db_service.MongoServiceConfig{
        IdFieldName: "contractnumber",
    })
    defer dbService.Disconnect(context.Background())
    engine.Use(func(ctx *gin.Context) {
        ctx.Set("db_service", dbService)
        ctx.Next()
    })
    // request routings
    handleFunctions := &contracts.ApiHandleFunctions{
        HospitalContractsAPI: contracts.NewHospitalContractsApi(),
    }
    contracts.NewRouterWithGinEngine(engine, *handleFunctions)
    engine.GET("/openapi", api.HandleOpenApi)
    engine.Run(":" + port)
}