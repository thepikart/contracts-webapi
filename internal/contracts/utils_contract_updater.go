package contracts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thepikart/contracts-webapi/internal/db_service"
)

type contractUpdater = func(
	ctx *gin.Context,
	contract *Contract,
) (updatedContract *Contract, responseContent interface{}, status int)

func updateContractFunc(ctx *gin.Context, updater contractUpdater) {
	value, exists := ctx.Get("db_service")
	if !exists {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service not found",
				"error":   "db_service not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[Contract])
	if !ok {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	contractId := ctx.Param("contractId")

	contract, err := db.FindDocument(ctx.Request.Context(), contractId)

	switch err {
	case nil:
		// continue
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Contract not found",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load contract from database",
				"error":   err.Error(),
			})
		return
	}

	updatedContract, responseObject, status := updater(ctx, contract)

	if updatedContract != nil {
		err = db.UpdateDocument(ctx.Request.Context(), contractId, updatedContract)
	} else {
		err = nil
	}

	switch err {
	case nil:
		if responseObject != nil {
			ctx.JSON(status, responseObject)
		} else {
			ctx.AbortWithStatus(status)
		}
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Contract was deleted while processing the request",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update contract in database",
				"error":   err.Error(),
			})
	}
}
