package mongo

import (
	"context"
	"time"

	"necutya/faker/internal/domain/domain"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Feedback struct {
	ID        uuid.UUID `bson:"feedback_id"`
	Text      string    `bson:"text"`
	Processed bool      `bson:"processed"`

	CreatedAt  time.Time  `bson:"created_at"`
	ResolvedAt *time.Time `bson:"resolved_at"`
}

func feedbackModelToRecord(model *domain.Feedback) *Feedback {
	return &Feedback{
		ID:         model.ID,
		Text:       model.Text,
		Processed:  model.Processed,
		CreatedAt:  model.CreatedAt,
		ResolvedAt: model.ResolvedAt,
	}
}

func feedbackRecordToModel(rec *Feedback) *domain.Feedback {
	return &domain.Feedback{
		ID:         rec.ID,
		Text:       rec.Text,
		Processed:  rec.Processed,
		CreatedAt:  rec.CreatedAt,
		ResolvedAt: rec.ResolvedAt,
	}
}

func feedbacksRecordToModel(recs []*Feedback) []*domain.Feedback {
	models := make([]*domain.Feedback, len(recs))

	for i := range recs {
		models[i] = feedbackRecordToModel(recs[i])
	}

	return models
}

func (r *UsersRepo) GetUserFeedbacks(ctx context.Context, userID string) ([]*domain.Feedback, error) {
	var user *User

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, wrapError(err)
	}

	err = r.db.FindOne(ctx, bson.M{
		"_id": userObjectID,
	}).Decode(&user)
	if err != nil {
		return nil, wrapError(err)
	}

	return feedbacksRecordToModel(user.Feedbacks), nil
}

func (r *UsersRepo) SetUserFeedback(ctx context.Context, userID string, feedback *domain.Feedback) error {
	var user *User

	userMongoID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	err = r.db.Database().Client().UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		err = r.db.FindOne(ctx,
			bson.M{"_id": userMongoID},
		).Decode(&user)
		if err != nil {
			return err
		}

		user.Feedbacks = append(user.Feedbacks, feedbackModelToRecord(feedback))

		_, err = r.db.UpdateOne(
			ctx,
			bson.M{"_id": userMongoID},
			bson.M{"$set": bson.M{
				"feedbacks": user.Feedbacks,
			}})
		if err != nil {
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})

	return wrapError(err)
}

func (r *UsersRepo) ResolveFeedback(ctx context.Context, userID string, feedbackID uuid.UUID) error {
	var user *User

	userMongoID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	err = r.db.Database().Client().UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		err = r.db.FindOne(ctx,
			bson.M{"_id": userMongoID},
		).Decode(&user)
		if err != nil {
			return err
		}

		now := time.Now()
		for i := range user.Feedbacks {
			if user.Feedbacks[i].ID == feedbackID {
				user.Feedbacks[i].Processed = true
				user.Feedbacks[i].ResolvedAt = &now
				break
			}
		}

		_, err = r.db.UpdateOne(
			ctx,
			bson.M{"_id": userMongoID},
			bson.M{"$set": bson.M{
				"feedbacks": user.Feedbacks,
			}})
		if err != nil {
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})

	return wrapError(err)
}
