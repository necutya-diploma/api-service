package mongo

import (
	"context"
	"time"

	"necutya/faker/internal/domain/domain"
	"necutya/faker/internal/domain/dto"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const usersCollection = "users"

type UserCredential struct {
	ExternalToken string `bson:"external_token"`
}

type User struct {
	ID primitive.ObjectID `bson:"_id"`

	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
	Email     string `bson:"email"`
	Role      string `bson:"role"`
	Password  string `bson:"password"`

	ReceiveNotification bool `bson:"receive_notification"`
	IsConfirmed         bool `bson:"is_confirmed"`

	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	LastVisitAt time.Time `bson:"last_visit_at"`

	PlanId primitive.ObjectID `bson:"plan_id"`

	Sessions   []*Session      `bson:"sessions"`
	Credential *UserCredential `bson:"credential"`
	Feedbacks  []*Feedback     `bson:"feedbacks"`
}

type UsersRepo struct {
	db *mongo.Collection
}

func NewUsersRepo(db *mongo.Database) *UsersRepo {
	return &UsersRepo{
		db: db.Collection(usersCollection),
	}
}

func userCredentialModelToRecord(model *domain.UserCredential) *UserCredential {
	return &UserCredential{
		ExternalToken: model.ExternalToken,
	}
}

func userCredentialRecordToModel(rec *UserCredential) *domain.UserCredential {
	return &domain.UserCredential{
		ExternalToken: rec.ExternalToken,
	}
}

func userModelToRecord(model *domain.User) *User {
	objectID, _ := primitive.ObjectIDFromHex(model.ID)
	planID, _ := primitive.ObjectIDFromHex(model.PlanID)

	return &User{
		ID: objectID,

		FirstName: model.FirstName,
		LastName:  model.LastName,
		Email:     model.Email,
		Password:  model.Password,

		Role: string(model.Role),

		ReceiveNotification: model.ReceiveNotification,
		IsConfirmed:         model.IsConfirmed,

		UpdatedAt:   model.UpdatedAt,
		CreatedAt:   model.CreatedAt,
		LastVisitAt: model.LastVisitAt,

		PlanId: planID,

		Sessions:   []*Session{},
		Credential: userCredentialModelToRecord(model.Credential),
	}
}

func userRecordToModel(rec *User) *domain.User {
	return &domain.User{
		ID:        rec.ID.Hex(),
		FirstName: rec.FirstName,
		LastName:  rec.LastName,
		Email:     rec.Email,
		Password:  rec.Password,

		Role: domain.Role(rec.Role),

		ReceiveNotification: rec.ReceiveNotification,
		IsConfirmed:         rec.IsConfirmed,

		UpdatedAt:   rec.UpdatedAt,
		CreatedAt:   rec.CreatedAt,
		LastVisitAt: rec.LastVisitAt,

		PlanID: rec.PlanId.Hex(),

		Credential: userCredentialRecordToModel(rec.Credential),
	}
}

func (r *UsersRepo) usersRecordToModel(recs []*User) []*domain.User {
	models := make([]*domain.User, len(recs))

	for i := range recs {
		models[i] = userRecordToModel(recs[i])
	}

	return models
}

func (r *UsersRepo) GetMany(ctx context.Context, filter dto.UserFilter) ([]*domain.User, error) {
	var results []*User
	queryFilter := make([]bson.M, 0)

	if filter.IsConfirmed {
		queryFilter = append(queryFilter, bson.M{"is_confirmed": filter.IsConfirmed})
	}

	if filter.Role != "" {
		queryFilter = append(queryFilter, bson.M{"role": filter.Role})
	}

	cursor, err := r.db.Find(ctx, bson.M{"$and": queryFilter})
	if err != nil {
		return nil, wrapError(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, wrapError(err)
	}

	return r.usersRecordToModel(results), nil
}

func (r *UsersRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user User

	err := r.db.FindOne(ctx, bson.M{
		"email": email,
	}).Decode(&user)
	if err != nil {
		return nil, wrapError(err)
	}

	return userRecordToModel(&user), nil
}

func (r *UsersRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var user User

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, wrapError(err)
	}

	err = r.db.FindOne(ctx, bson.M{
		"_id": objectID,
	}).Decode(&user)
	if err != nil {
		return nil, wrapError(err)
	}

	return userRecordToModel(&user), nil
}

func (r *UsersRepo) GetUserByExternalToken(ctx context.Context, externalToken string) (*domain.User, error) {
	var user User

	err := r.db.FindOne(ctx, bson.M{
		"credential.external_token": externalToken,
	}).Decode(&user)
	if err != nil {
		return nil, wrapError(err)
	}

	return userRecordToModel(&user), nil
}

// db.users.createIndex( { "email": 1 }, { unique: true } )
func (r *UsersRepo) Create(ctx context.Context, user *domain.User) error {
	user.ID = primitive.NewObjectID().Hex()

	_, err := r.db.InsertOne(ctx, userModelToRecord(user))
	return wrapError(err)
}

func (r *UsersRepo) Update(ctx context.Context, user *domain.User) error {
	objectID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return wrapError(err)
	}

	_, err = r.db.ReplaceOne(ctx, bson.M{"_id": objectID}, userModelToRecord(user))
	return wrapError(err)
}

func (r *UsersRepo) Delete(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return wrapError(err)
	}

	_, err = r.db.DeleteOne(ctx, bson.M{"_id": objectID})
	return wrapError(err)
}

type Session struct {
	SessionID uuid.UUID `bson:"session_id"`

	RefreshToken string    `bson:"refresh_token"`
	ExpiresAt    time.Time `bson:"expires_at"`
	CreatedAt    time.Time `bson:"created_at"`
	Client       string    `bson:"client"`
	IpAddress    string    `bson:"ip_address"`
}

func userSessionModelToRecord(model *domain.Session) *Session {
	return &Session{
		SessionID:    model.ID,
		RefreshToken: model.RefreshToken,
		ExpiresAt:    model.ExpiresAt,
		CreatedAt:    model.CreatedAt,
		Client:       model.Client,
		IpAddress:    model.IpAddress,
	}
}

func userSessionRecordToModel(rec *Session) *domain.Session {
	return &domain.Session{
		ID:           rec.SessionID,
		RefreshToken: rec.RefreshToken,
		ExpiresAt:    rec.ExpiresAt,
		CreatedAt:    rec.CreatedAt,
		Client:       rec.Client,
		IpAddress:    rec.IpAddress,
	}
}

func (r *UsersRepo) SetSession(ctx context.Context, userID string, session *domain.Session) error {
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

		forInsert := true
		for i := range user.Sessions {
			if user.Sessions[i].SessionID == session.ID {
				user.Sessions[i] = userSessionModelToRecord(session)
				forInsert = false
				break
			}
		}

		if forInsert {
			user.Sessions = append(user.Sessions, userSessionModelToRecord(session))
		}

		_, err = r.db.UpdateOne(
			ctx,
			bson.M{"_id": userMongoID},
			bson.M{"$set": bson.M{
				"sessions":      user.Sessions,
				"last_visit_at": time.Now(),
			}})
		if err != nil {
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})

	return wrapError(err)
}

func (r *UsersRepo) GetSessionByRefreshToken(
	ctx context.Context,
	userID string,
	refreshToken string,
) (*domain.Session, error) {
	var user *User

	userMongoID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, wrapError(err)
	}

	err = r.db.FindOne(ctx,
		bson.M{"_id": userMongoID},
	).Decode(&user)
	if err != nil {
		return nil, wrapError(err)
	}

	for i := range user.Sessions {
		if user.Sessions[i].RefreshToken == refreshToken {
			return userSessionRecordToModel(user.Sessions[i]), nil
		}
	}

	return nil, wrapError(mongo.ErrNoDocuments)
}

func (r *UsersRepo) RemoveSession(ctx context.Context, userID string, sessionID uuid.UUID) error {
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

		for i := range user.Sessions {
			if user.Sessions[i].SessionID == sessionID {
				user.Sessions = append(user.Sessions[:i], user.Sessions[i+1:]...)
				break
			}
		}

		_, err = r.db.UpdateOne(
			ctx,
			bson.M{"_id": userMongoID},
			bson.M{"$set": bson.M{
				"sessions":    user.Sessions,
				"lastVisitAt": time.Now(),
			}})
		if err != nil {
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})

	return wrapError(err)
}
