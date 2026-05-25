package contracts

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/thepikart/contracts-webapi/internal/db_service"
)

type HospitalContractsSuite struct {
	suite.Suite
	dbServiceMock *DbServiceMock[Contract]
}

func TestHospitalContractsSuite(t *testing.T) {
	suite.Run(t, new(HospitalContractsSuite))
}

type DbServiceMock[DocType interface{}] struct {
	mock.Mock
}

func (this *DbServiceMock[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	args := this.Called(ctx, id, document)
	return args.Error(0)
}

func (this *DbServiceMock[DocType]) FindDocument(ctx context.Context, id string) (*DocType, error) {
	args := this.Called(ctx, id)
	return args.Get(0).(*DocType), args.Error(1)
}

func (this *DbServiceMock[DocType]) FindAllDocuments(ctx context.Context) ([]DocType, error) {
	args := this.Called(ctx)
	return args.Get(0).([]DocType), args.Error(1)
}

func (this *DbServiceMock[DocType]) UpdateDocument(ctx context.Context, id string, document *DocType) error {
	args := this.Called(ctx, id, document)
	return args.Error(0)
}

func (this *DbServiceMock[DocType]) DeleteDocument(ctx context.Context, id string) error {
	args := this.Called(ctx, id)
	return args.Error(0)
}

func (this *DbServiceMock[DocType]) Disconnect(ctx context.Context) error {
	args := this.Called(ctx)
	return args.Error(0)
}

func (suite *HospitalContractsSuite) SetupTest() {
	suite.dbServiceMock = &DbServiceMock[Contract]{}

	// Compile time assert that the mock satisfies db_service.DbService[Contract]
	var _ db_service.DbService[Contract] = suite.dbServiceMock

	suite.dbServiceMock.
		On("FindDocument", mock.Anything, mock.Anything).
		Return(
			&Contract{
				ContractNumber: "test-contract",
				Name:           "Test Contract",
				Partner:        "Test Partner s.r.o.",
				ValidFrom:      "2026-01-01",
				ValidUntil:     "2028-12-31",
				Budget:         50000,
				Status:         "Active",
			},
			nil,
		)
}

func (suite *HospitalContractsSuite) Test_UpdateContract_DbServiceUpdateCalled() {
	// ARRANGE
	suite.dbServiceMock.
		On("UpdateDocument", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	json := `{
		"contractNumber": "test-contract",
		"name": "Updated Contract",
		"partner": "Test Partner s.r.o.",
		"validFrom": "2026-01-01",
		"validUntil": "2028-12-31",
		"budget": 50000,
		"status": "Active"
	}`

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("db_service", suite.dbServiceMock)
	ctx.Params = []gin.Param{
		{Key: "contractId", Value: "test-contract"},
	}
	ctx.Request = httptest.NewRequest("PUT", "/contracts/test-contract", strings.NewReader(json))

	sut := implHospitalContractsAPI{}

	// ACT
	sut.UpdateContract(ctx)

	// ASSERT
	suite.dbServiceMock.AssertCalled(suite.T(), "UpdateDocument", mock.Anything, "test-contract", mock.Anything)
}
