package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	redis2 "github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/redis"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/test/containers/contracts"

	"emperror.dev/errors"
	"github.com/docker/go-connections/nat"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type redisTestContainers struct {
	container      testcontainers.Container
	defaultOptions *contracts.RedisContainerOptions
}

func NewRedisTestContainers() contracts.RedisContainer {
	return &redisTestContainers{
		defaultOptions: &contracts.RedisContainerOptions{
			Port:      "6379/tcp",
			Host:      "localhost",
			Database:  0,
			PoolSize:  300,
			Tag:       "latest",
			ImageName: "redis",
			Name:      "redis-testcontainers",
		},
	}
}

func (g *redisTestContainers) CreatingContainerOptions(
	ctx context.Context,
	t *testing.T,
	options ...*contracts.RedisContainerOptions,
) (*redis2.RedisOptions, error) {
	// https://github.com/testcontainers/testcontainers-go
	// https://dev.to/remast/go-integration-tests-using-testcontainers-9o5
	containerReq := g.getRunOptions(options...)

	// TODO: Using Parallel Container
	dbContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		return nil, err
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := dbContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// get a free random host hostPort
	hostPort, err := dbContainer.MappedPort(ctx, nat.Port(g.defaultOptions.Port))
	if err != nil {
		return nil, err
	}
	g.defaultOptions.HostPort = hostPort.Int()

	host, err := dbContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	g.container = dbContainer

	reidsOptions := &redis2.RedisOptions{
		Database: g.defaultOptions.Database,
		Host:     host,
		Port:     g.defaultOptions.HostPort,
		PoolSize: g.defaultOptions.PoolSize,
	}
	return reidsOptions, nil
}

func (g *redisTestContainers) Start(
	ctx context.Context,
	t *testing.T,
	options ...*contracts.RedisContainerOptions,
) (redis.UniversalClient, error) {
	redisOptions, err := g.CreatingContainerOptions(ctx, t, options...)

	db := redis2.NewRedisClient(redisOptions)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (g *redisTestContainers) Cleanup(ctx context.Context) error {
	if err := g.container.Terminate(ctx); err != nil {
		return errors.WrapIf(err, "failed to terminate container: %s")
	}
	return nil
}

func (g *redisTestContainers) getRunOptions(
	opts ...*contracts.RedisContainerOptions,
) testcontainers.ContainerRequest {
	if len(opts) > 0 && opts[0] != nil {
		option := opts[0]
		if option.ImageName != "" {
			g.defaultOptions.ImageName = option.ImageName
		}
		if option.Host != "" {
			g.defaultOptions.Host = option.Host
		}
		if option.Port != "" {
			g.defaultOptions.Port = option.Port
		}
		if option.Tag != "" {
			g.defaultOptions.Tag = option.Tag
		}
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("%s:%s", g.defaultOptions.ImageName, g.defaultOptions.Tag),
		ExposedPorts: []string{g.defaultOptions.Port},
		WaitingFor: wait.ForListeningPort(nat.Port(g.defaultOptions.Port)).
			WithPollInterval(2 * time.Second),
		Hostname: g.defaultOptions.Host,
		Env:      map[string]string{},
	}

	return containerReq
}
