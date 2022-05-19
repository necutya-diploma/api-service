package mongo

import (
	"context"
	"time"

	"necutya/faker/internal/domain/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const plansCollection = "plans"

type Plan struct {
	ID                    primitive.ObjectID `bson:"_id"`
	Text                  string             `bson:"text"`
	Name                  string             `bson:"name"`
	Description           string             `bson:"descriptions"`
	Options               []string           `bson:"options"`
	Price                 int                `bson:"price"`
	IsBasic               bool               `bson:"is_basic"`
	Duration              int                `bson:"month_duration"`
	InternalRequestsCount int                `bson:"internal_request_count"`
	ExternalRequestsCount int                `bson:"external_request_count"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

type PlansRepo struct {
	db *mongo.Collection
}

func NewPlansRepo(db *mongo.Database) *PlansRepo {
	return &PlansRepo{
		db: db.Collection(plansCollection),
	}
}

func (r *PlansRepo) GetOne(ctx context.Context, id string) (*domain.Plan, error) {
	var plan *Plan

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, wrapError(err)
	}

	err = r.db.FindOne(ctx, bson.M{
		"_id": objectID,
	}).Decode(&plan)
	if err != nil {
		return nil, wrapError(err)
	}

	return planRecordToModel(plan), nil
}

func (r *PlansRepo) GetOneByName(ctx context.Context, name string) (*domain.Plan, error) {
	var plan *Plan

	err := r.db.FindOne(ctx, bson.M{
		"name": name,
	}).Decode(&plan)
	if err != nil {
		return nil, wrapError(err)
	}

	return planRecordToModel(plan), nil
}

func (r *PlansRepo) GetMany(ctx context.Context) ([]*domain.Plan, error) {
	var plans []*Plan

	cursor, err := r.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &plans); err != nil {
		return nil, wrapError(err)
	}

	return plansRecordToModel(plans), nil
}

func planRecordToModel(rec *Plan) *domain.Plan {
	return &domain.Plan{
		ID:                    rec.ID.Hex(),
		Name:                  rec.Name,
		Description:           rec.Description,
		Options:               rec.Options,
		Price:                 rec.Price,
		Duration:              rec.Duration,
		InternalRequestsCount: rec.InternalRequestsCount,
		ExternalRequestsCount: rec.ExternalRequestsCount,
	}
}

func plansRecordToModel(recs []*Plan) []*domain.Plan {
	models := make([]*domain.Plan, len(recs))

	for i := range recs {
		models[i] = planRecordToModel(recs[i])
	}

	return models
}
