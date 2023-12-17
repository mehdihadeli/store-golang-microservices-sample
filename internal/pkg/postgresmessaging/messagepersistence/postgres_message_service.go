package messagepersistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/core/messaging/persistmessage"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/core/messaging/types"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/core/serializer/contratcs"
	customErrors "github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/http/httperrors/customerrors"
	"github.com/mehdihadeli/go-ecommerce-microservices/internal/pkg/logger"

	uuid "github.com/satori/go.uuid"
)

type postgresMessageService struct {
	messagingDBContext *PostgresMessagePersistenceDBContext
	messageSerializer  contratcs.MessageSerializer
	logger             logger.Logger
}

func (m *postgresMessageService) Process(messageID string, ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (m *postgresMessageService) ProcessAll(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (m *postgresMessageService) AddPublishMessage(
	messageEnvelope types.MessageEnvelopeTMessage,
	ctx context.Context,
) error {
	// TODO implement me
	panic("implement me")
}

func (m *postgresMessageService) AddReceivedMessage(messageEnvelope types.MessageEnvelope, ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (m *postgresMessageService) AddMessageCore(
	ctx context.Context,
	messageEnvelope types.MessageEnvelope,
	deliveryType persistmessage.MessageDeliveryType,
) error {
	if messageEnvelope.Message == nil {
		return errors.New("messageEnvelope.Message is nil")
	}

	var id string
	switch message := messageEnvelope.Message.(type) {
	case types.IMessage:
		id = message.GeMessageId()
	// case IInternalCommand:
	//	id = message.InternalCommandId
	default:
		id = uuid.NewV4().String()
	}

	data, err := m.messageSerializer.SerializeEnvelop(messageEnvelope)
	if err != nil {
		return err
	}

	uuidId, err := uuid.FromString(id)
	if err != nil {
		return err
	}

	storeMessage := persistmessage.NewStoreMessage(
		uuidId,
		messageEnvelope.Message.GetMessageFullTypeName(),
		string(data.Data),
		deliveryType,
	)

	err := _messagePersistenceRepository.AddAsync(storeMessage, cancellationToken)
	if err != nil {
		return err
	}

	_logger.LogInformation(
		"Message with id: %v and delivery type: %v saved in persistence message store",
		id,
		deliveryType,
	)

	return nil
}

func NewPostgresMessageService(
	postgresMessagePersistenceDBContext *PostgresMessagePersistenceDBContext,
	l logger.Logger,
) persistmessage.MessageService {
	return &postgresMessageService{
		messagingDBContext: postgresMessagePersistenceDBContext,
		logger:             l,
	}
}

func (m *postgresMessageService) Add(
	ctx context.Context,
	storeMessage *persistmessage.StoreMessage,
) error {
	dbContext := m.messagingDBContext.WithTxIfExists(ctx)

	// https://gorm.io/docs/create.html
	result := dbContext.WithContext(ctx).Create(storeMessage)
	if result.Error != nil {
		return customErrors.NewConflictErrorWrap(
			result.Error,
			"storeMessage already exists",
		)
	}

	m.logger.Infof("Number of affected rows are: %d", result.RowsAffected)

	return nil
}

func (m *postgresMessageService) Update(
	ctx context.Context,
	storeMessage *persistmessage.StoreMessage,
) error {
	dbContext := m.messagingDBContext.WithTxIfExists(ctx)

	// https://gorm.io/docs/update.html
	result := dbContext.WithContext(ctx).Updates(storeMessage)
	if result.Error != nil {
		return customErrors.NewInternalServerErrorWrap(
			result.Error,
			"error in updating the storeMessage",
		)
	}

	m.logger.Infof("Number of affected rows are: %d", result.RowsAffected)

	return nil
}

func (m *postgresMessageService) ChangeState(
	ctx context.Context,
	messageID uuid.UUID,
	status persistmessage.MessageStatus,
) error {
	storeMessage, err := m.GetById(ctx, messageID)
	if err != nil {
		return customErrors.NewNotFoundErrorWrap(
			err,
			fmt.Sprintf(
				"storeMessage with id `%s` not found in the database",
				messageID.String(),
			),
		)
	}

	storeMessage.MessageStatus = status
	err = m.Update(ctx, storeMessage)

	return err
}

func (m *postgresMessageService) GetAllActive(
	ctx context.Context,
) ([]*persistmessage.StoreMessage, error) {
	var storeMessages []*persistmessage.StoreMessage

	predicate := func(sm *persistmessage.StoreMessage) bool {
		return sm.MessageStatus == persistmessage.Stored
	}

	dbContext := m.messagingDBContext.WithTxIfExists(ctx)
	result := dbContext.WithContext(ctx).Where(predicate).Find(&storeMessages)
	if result.Error != nil {
		return nil, result.Error
	}

	return storeMessages, nil
}

func (m *postgresMessageService) GetByFilter(
	ctx context.Context,
	predicate func(*persistmessage.StoreMessage) bool,
) ([]*persistmessage.StoreMessage, error) {
	var storeMessages []*persistmessage.StoreMessage

	dbContext := m.messagingDBContext.WithTxIfExists(ctx)
	result := dbContext.WithContext(ctx).Where(predicate).Find(&storeMessages)

	if result.Error != nil {
		return nil, result.Error
	}

	return storeMessages, nil
}

func (m *postgresMessageService) GetById(
	ctx context.Context,
	id uuid.UUID,
) (*persistmessage.StoreMessage, error) {
	var storeMessage *persistmessage.StoreMessage

	// https://gorm.io/docs/query.html#Retrieving-objects-with-primary-key
	// https://gorm.io/docs/query.html#Struct-amp-Map-Conditions
	// https://gorm.io/docs/query.html#Inline-Condition
	// https://gorm.io/docs/advanced_query.html
	result := m.messagingDBContext.WithContext(ctx).
		Find(&storeMessage, id)
	if result.Error != nil {
		return nil, customErrors.NewNotFoundErrorWrap(
			result.Error,
			fmt.Sprintf(
				"storeMessage with id `%s` not found in the database",
				id.String(),
			),
		)
	}

	m.logger.Infof("Number of affected rows are: %d", result.RowsAffected)

	return storeMessage, nil
}

func (m *postgresMessageService) Remove(
	ctx context.Context,
	storeMessage *persistmessage.StoreMessage,
) (bool, error) {
	id := storeMessage.ID

	storeMessage, err := m.GetById(ctx, id)
	if err != nil {
		return false, customErrors.NewNotFoundErrorWrap(
			err,
			fmt.Sprintf(
				"storeMessage with id `%s` not found in the database",
				id.String(),
			),
		)
	}

	dbContext := m.messagingDBContext.WithTxIfExists(ctx)

	result := dbContext.WithContext(ctx).Delete(storeMessage, id)
	if result.Error != nil {
		return false, customErrors.NewInternalServerErrorWrap(
			result.Error,
			fmt.Sprintf(
				"error in deleting storeMessage with id `%s` in the database",
				id.String(),
			),
		)
	}

	m.logger.Infof("Number of affected rows are: %d", result.RowsAffected)

	return true, nil
}

func (m *postgresMessageService) CleanupMessages(
	ctx context.Context,
) error {
	predicate := func(sm *persistmessage.StoreMessage) bool {
		return sm.MessageStatus == persistmessage.Processed
	}

	dbContext := m.messagingDBContext.WithTxIfExists(ctx)

	result := dbContext.WithContext(ctx).
		Where(predicate).
		Delete(&persistmessage.StoreMessage{})

	if result.Error != nil {
		return result.Error
	}

	m.logger.Infof("Number of affected rows are: %d", result.RowsAffected)

	return nil
}
