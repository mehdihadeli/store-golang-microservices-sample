package mongo

import (
	"context"
	"testing"

	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/config"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/config/environment"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/core"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/logger/external/fxlog"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/logger/zap"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/mongodb"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func Test_Custom_Mongo_Container(t *testing.T) {
	ctx := context.Background()

	var mongoClient *mongo.Client

	fxtest.New(t,
		config.ModuleFunc(environment.Test),
		zap.Module,
		fxlog.FxLogger,
		core.Module,
		mongodb.Module,
		fx.Decorate(MongoContainerOptionsDecorator(t, ctx)),
		fx.Populate(&mongoClient),
	).RequireStart()

	assert.NotNil(t, mongoClient)
}
