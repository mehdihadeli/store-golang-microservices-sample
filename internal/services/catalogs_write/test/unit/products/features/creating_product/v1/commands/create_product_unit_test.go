//go:build unit
// +build unit

package commands

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"

	createProductCommand "github.com/mehdihadeli/go-ecommerce-microservices/internal/services/catalogs/write_service/internal/products/features/creating_product/v1/commands"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/services/catalogs/write_service/internal/shared/test_fixtures/unit_test"
)

type createProductUnitTests struct {
	*unit_test.UnitTestSharedFixture
	*unit_test.UnitTestMockFixture
}

func TestCreateProductUnit(t *testing.T) {
	suite.Run(
		t,
		&createProductUnitTests{UnitTestSharedFixture: unit_test.NewUnitTestSharedFixture(t)},
	)
}

func (c *createProductUnitTests) SetupTest() {
	// create new mocks or clear mocks before executing
	c.UnitTestMockFixture = unit_test.NewUnitTestMockFixture(c.T())
}

func (c *createProductUnitTests) Test_New_Create_Product_Should_Return_No_Error_For_Valid_Input() {
	name := gofakeit.Name()
	description := gofakeit.EmojiDescription()
	price := gofakeit.Price(150, 6000)

	updateProduct, err := createProductCommand.NewCreateProduct(name, description, price)

	c.Assert().NotNil(updateProduct)
	c.Assert().Equal(name, updateProduct.Name)
	c.Assert().Equal(price, updateProduct.Price)

	c.Require().NoError(err)
}

func (c *createProductUnitTests) Test_New_Create_Product_Should_Return_Error_For_Invalid_Price() {
	command, err := createProductCommand.NewCreateProduct(
		gofakeit.Name(),
		gofakeit.EmojiDescription(),
		0,
	)

	c.Require().Error(err)
	c.Assert().Nil(command)
}

func (c *createProductUnitTests) Test_New_Create_Product_Should_Return_Error_For_Empty_Name() {
	command, err := createProductCommand.NewCreateProduct("", gofakeit.EmojiDescription(), 120)

	c.Require().Error(err)
	c.Assert().Nil(command)
}

func (c *createProductUnitTests) Test_New_Create_Product_Should_Return_Error_For_Empty_Description() {
	command, err := createProductCommand.NewCreateProduct(gofakeit.Name(), "", 120)

	c.Require().Error(err)
	c.Assert().Nil(command)
}
