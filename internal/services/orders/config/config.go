package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"emperror.dev/errors"
	"github.com/spf13/viper"

	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/constants"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/eventstroredb"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/grpc"
	customEcho "github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/http/custom_echo"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/logger"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/mongodb"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/otel"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/otel/metrics"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/rabbitmq/config"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "catalogs_write write microservice config path")
}

type Config struct {
	DeliveryType      string                          `mapstructure:"deliveryType"`
	ServiceName       string                          `mapstructure:"serviceName"`
	Logger            *logger.LogOptions              `mapstructure:"logger"`
	GRPC              *grpc.GrpcOptions               `mapstructure:"grpc"`
	Http              *customEcho.EchoHttpOptions     `mapstructure:"http"`
	Context           Context                         `mapstructure:"context"`
	OTel              *otel.OpenTelemetryOptions      `mapstructure:"otel"             envPrefix:"OTel_"`
	OTelMetricsConfig *metrics.OTelMetricsOptions     `mapstructure:"otelMetrics"      envPrefix:"OTelMetrics_"`
	RabbitMQ          *config.RabbitmqOptions         `mapstructure:"rabbitmq"         envPrefix:"RabbitMQ_"`
	EventStoreConfig  *eventstroredb.EventStoreConfig `mapstructure:"eventStoreConfig"`
	Subscriptions     *Subscriptions                  `mapstructure:"subscriptions"`
	Mongo             *mongodb.MongoDbOptions         `mapstructure:"mongo"            envPrefix:"Mongo_"`
	MongoCollections  MongoCollections                `mapstructure:"mongoCollections" envPrefix:"MongoCollections_"`
}

type Context struct {
	Timeout int `mapstructure:"timeout"`
}

type MongoCollections struct {
	Orders string `mapstructure:"orders" validate:"required" env:"Orders"`
}

type Subscriptions struct {
	OrderSubscription *Subscription `mapstructure:"orderSubscription"`
}

type Subscription struct {
	Prefix         []string `mapstructure:"prefix"         validate:"required"`
	SubscriptionId string   `mapstructure:"subscriptionId" validate:"required"`
}

func InitConfig(env string) (*Config, error) {
	if configPath == "" {
		configPathFromEnv := os.Getenv(constants.ConfigPath)
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		} else {
			//https://stackoverflow.com/questions/31873396/is-it-possible-to-get-the-current-root-of-package-structure-as-a-string-in-golan
			//https://stackoverflow.com/questions/18537257/how-to-get-the-directory-of-the-currently-running-file
			d, err := dirname()
			if err != nil {
				return nil, err
			}

			configPath = d
		}
	}

	cfg := &Config{}

	viper.SetConfigName(fmt.Sprintf("config.%s", env))
	viper.AddConfigPath(configPath)
	viper.SetConfigType(constants.Yaml)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.WrapIf(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.WrapIf(err, "viper.Unmarshal")
	}

	grpcPort := os.Getenv(constants.GrpcPort)
	if grpcPort != "" {
		cfg.GRPC.Port = grpcPort
	}

	jaegerPort := os.Getenv(constants.JaegerPort)
	if jaegerPort != "" {
		cfg.OTel.JaegerExporterOptions.AgentPort = jaegerPort
	}

	jaegerHost := os.Getenv(constants.JaegerHost)
	if jaegerHost != "" {
		cfg.OTel.JaegerExporterOptions.AgentHost = jaegerHost
	}

	return cfg, nil
}

func (cfg *Config) GetMicroserviceNameUpper() string {
	return strings.ToUpper(cfg.ServiceName)
}

func (cfg *Config) GetMicroserviceName() string {
	return cfg.ServiceName
}

func filename() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("unable to get the current filename")
	}
	return filename, nil
}

func dirname() (string, error) {
	filename, err := filename()
	if err != nil {
		return "", err
	}
	return filepath.Dir(filename), nil
}
