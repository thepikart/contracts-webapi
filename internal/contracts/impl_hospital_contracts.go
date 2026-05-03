package contracts

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type implHospitalContractsAPI struct {
}

func NewHospitalContractsApi() HospitalContractsAPI {
	return &implHospitalContractsAPI{}
}

func (o implHospitalContractsAPI) GetContracts(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implHospitalContractsAPI) CreateContract(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implHospitalContractsAPI) GetContract(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implHospitalContractsAPI) UpdateContract(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implHospitalContractsAPI) DeleteContract(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}
