package contracts

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thepikart/contracts-webapi/internal/db_service"
)

type implHospitalContractsAPI struct {
}

func NewHospitalContractsApi() HospitalContractsAPI {
	return &implHospitalContractsAPI{}
}

func (o implHospitalContractsAPI) GetContracts(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
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
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	contracts, err := db.FindAllDocuments(c.Request.Context())
	switch err {
	case nil:
		if contracts == nil {
			contracts = []Contract{}
		}
		c.JSON(http.StatusOK, contracts)
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load contracts from database",
				"error":   err.Error(),
			})
	}
}

func (o implHospitalContractsAPI) CreateContract(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db not found",
				"error":   "db not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[Contract])
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db context is not of required type",
				"error":   "cannot cast db context to db_service.DbService",
			})
		return
	}

	contract := Contract{}
	err := c.BindJSON(&contract)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		return
	}

	if contract.Name == "" || contract.Partner == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Name and Partner are required",
			"error":   "missing required fields",
		})
		return
	}
	if contract.ValidFrom == "" || contract.ValidUntil == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "ValidFrom and ValidUntil are required",
			"error":   "missing required fields",
		})
		return
	}
	if contract.ValidFrom >= contract.ValidUntil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "ValidFrom must be earlier than ValidUntil",
			"error":   "invalid date range",
		})
		return
	}
	if err := validateStatusAgainstDates(contract.Status, contract.ValidFrom, contract.ValidUntil); err != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": err,
			"error":   "invalid status for date range",
		})
		return
	}

	if contract.ContractNumber == "" {
		contract.ContractNumber = uuid.New().String()
	}

	err = db.CreateDocument(c.Request.Context(), contract.ContractNumber, &contract)

	switch err {
	case nil:
		c.JSON(
			http.StatusOK,
			contract,
		)
	case db_service.ErrConflict:
		c.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "Contract already exists",
				"error":   err.Error(),
			},
		)
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create contract in database",
				"error":   err.Error(),
			},
		)
	}
}

func (o implHospitalContractsAPI) GetContract(c *gin.Context) {
	updateContractFunc(c, func(
		c *gin.Context,
		contract *Contract,
	) (updatedContract *Contract, responseContent interface{}, status int) {
		return nil, contract, http.StatusOK
	})
}

func (o implHospitalContractsAPI) UpdateContract(c *gin.Context) {
	updateContractFunc(c, func(c *gin.Context, contract *Contract) (*Contract, interface{}, int) {
		var updatedFields Contract

		if err := c.ShouldBindJSON(&updatedFields); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		contractId := c.Param("contractId")
		if updatedFields.ContractNumber != "" && updatedFields.ContractNumber != contractId {
			return nil, gin.H{
				"status":  http.StatusForbidden,
				"message": "ContractNumber in body does not match contractId in path",
			}, http.StatusForbidden
		}

		if updatedFields.Name != "" {
			contract.Name = updatedFields.Name
		}
		if updatedFields.Partner != "" {
			contract.Partner = updatedFields.Partner
		}
		if updatedFields.ValidFrom != "" {
			contract.ValidFrom = updatedFields.ValidFrom
		}
		if updatedFields.ValidUntil != "" {
			contract.ValidUntil = updatedFields.ValidUntil
		}
		if updatedFields.Budget > 0 {
			contract.Budget = updatedFields.Budget
		}
		if updatedFields.Status != "" {
			contract.Status = updatedFields.Status
		}
		if updatedFields.ServiceType != "" {
			contract.ServiceType = updatedFields.ServiceType
		}
		if updatedFields.Description != "" {
			contract.Description = updatedFields.Description
		}

		if contract.ValidFrom >= contract.ValidUntil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "ValidFrom must be earlier than ValidUntil",
				"error":   "invalid date range",
			}, http.StatusBadRequest
		}
		if errMsg := validateStatusAgainstDates(contract.Status, contract.ValidFrom, contract.ValidUntil); errMsg != "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": errMsg,
				"error":   "invalid status for date range",
			}, http.StatusBadRequest
		}

		return contract, contract, http.StatusOK
	})
}

func validateStatusAgainstDates(status, validFrom, validUntil string) string {
	today := time.Now().Format("2006-01-02")
	if today > validUntil && status == "Active" {
		return "Contract past its end date cannot be Active"
	}
	if today >= validFrom && today <= validUntil && status == "Ended" {
		return "Contract within its validity period cannot be Ended"
	}
	return ""
}

func (o implHospitalContractsAPI) DeleteContract(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
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
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	contractId := c.Param("contractId")
	err := db.DeleteDocument(c.Request.Context(), contractId)

	switch err {
	case nil:
		c.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Contract not found",
				"error":   err.Error(),
			},
		)
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete contract from database",
				"error":   err.Error(),
			})
	}
}
