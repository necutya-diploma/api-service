package mongo

import (
	"context"
	"time"

	"necutya/faker/internal/domain/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const ordersCollection = "orders"

type Order struct {
	ID           primitive.ObjectID `bson:"_id"`
	PlanID       primitive.ObjectID `bson:"plan_id"`
	UserID       primitive.ObjectID `bson:"user_id"`
	Description  string             `bson:"description"`
	Status       string             `bson:"status"`
	Amount       int                `bson:"amount"`
	Currency     string             `bson:"currency"`
	Transactions []*Transaction     `bson:"transactions"`
	CreatedAt    time.Time          `bson:"created_at"`
}

func orderModelToRecord(model *domain.Order) *Order {
	objectID, _ := primitive.ObjectIDFromHex(model.ID)
	planID, _ := primitive.ObjectIDFromHex(model.PlanID)
	userID, _ := primitive.ObjectIDFromHex(model.UserID)

	return &Order{
		ID:           objectID,
		PlanID:       planID,
		UserID:       userID,
		Description:  model.Description,
		Status:       model.Status,
		Amount:       model.Amount,
		Currency:     model.Currency,
		CreatedAt:    model.CreatedAt,
		Transactions: transactionsModelToRecord(model.Transactions),
	}
}

func orderRecordToModel(rec *Order) *domain.Order {
	return &domain.Order{
		ID:           rec.ID.Hex(),
		PlanID:       rec.PlanID.Hex(),
		UserID:       rec.UserID.Hex(),
		Description:  rec.Description,
		Status:       rec.Status,
		Amount:       rec.Amount,
		Currency:     rec.Currency,
		CreatedAt:    rec.CreatedAt,
		Transactions: transactionsRecordToModel(rec.Transactions),
	}
}

func ordersRecordToModel(recs []*Order) []*domain.Order {
	models := make([]*domain.Order, len(recs))

	for i := range recs {
		models[i] = orderRecordToModel(recs[i])
	}

	return models
}

type Transaction struct {
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	AdditionalInfo string    `bson:"additional_info"`
}

func transactionModelToRecord(model *domain.Transaction) *Transaction {
	return &Transaction{
		Status:         model.Status,
		CreatedAt:      model.CreatedAt,
		AdditionalInfo: model.AdditionalInfo,
	}
}

func transactionsModelToRecord(models []*domain.Transaction) []*Transaction {
	recs := make([]*Transaction, len(models))

	for i := range models {
		recs[i] = transactionModelToRecord(models[i])
	}

	return recs
}

func transactionRecordToModel(rec *Transaction) *domain.Transaction {
	return &domain.Transaction{
		Status:         rec.Status,
		CreatedAt:      rec.CreatedAt,
		AdditionalInfo: rec.AdditionalInfo,
	}
}

func transactionsRecordToModel(recs []*Transaction) []*domain.Transaction {
	models := make([]*domain.Transaction, len(recs))

	for i := range recs {
		models[i] = transactionRecordToModel(recs[i])
	}

	return models
}

type OrdersRepo struct {
	db *mongo.Collection
}

func NewOrdersRepo(db *mongo.Database) *OrdersRepo {
	return &OrdersRepo{
		db: db.Collection(ordersCollection),
	}
}

func (r *OrdersRepo) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	order.ID = primitive.NewObjectID().Hex()

	orderRec := orderModelToRecord(order)

	_, err := r.db.InsertOne(ctx, orderRec)

	return orderRecordToModel(orderRec), wrapError(err)
}

func (r *OrdersRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var order *Order

	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, wrapError(err)
	}

	err = r.db.FindOne(ctx, bson.M{
		"_id": ID,
	}).Decode(&order)

	return orderRecordToModel(order), wrapError(err)
}

func (r *OrdersRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, wrapError(err)
	}

	cur, err := r.db.Find(ctx, bson.M{"user_id": userObjectID})
	if err != nil {
		return nil, wrapError(err)
	}

	var orders []*Order
	if err = cur.All(ctx, &orders); err != nil {
		return nil, wrapError(err)
	}

	return ordersRecordToModel(orders), wrapError(err)
}

func (r *OrdersRepo) AddTransaction(
	ctx context.Context,
	orderID string,
	transaction *domain.Transaction,
) (*domain.Order, error) {
	var order domain.Order

	orderObjectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, wrapError(err)
	}

	res := r.db.FindOneAndUpdate(ctx, bson.M{"_id": orderObjectID}, bson.M{
		"$set": bson.M{
			"status": transaction.Status,
		},
		"$push": bson.M{
			"transactions": transactionModelToRecord(transaction),
		},
	})
	if res.Err() != nil {
		return nil, wrapError(res.Err())
	}

	err = res.Decode(&order)

	return &order, err
}
