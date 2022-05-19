package mongo

import (
	"context"
	"time"

	"necutya/faker/internal/domain/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const messagesCollection = "messages"

type Message struct {
	ID        primitive.ObjectID `bson:"_id"`
	Text      string             `bson:"text"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type MessagesRepo struct {
	db *mongo.Collection
}

func NewMessagesRepo(db *mongo.Database) *MessagesRepo {
	return &MessagesRepo{
		db: db.Collection(messagesCollection),
	}
}

func (r *MessagesRepo) Create(ctx context.Context, message *domain.Message) error {
	_, err := r.db.InsertOne(ctx, r.marshalMessage(message))
	return wrapError(err)
}

func (r *MessagesRepo) marshalMessage(message *domain.Message) *Message {
	return &Message{
		ID: primitive.NewObjectID(),

		Text:      message.Text,
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
	}
}
