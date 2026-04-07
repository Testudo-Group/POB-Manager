package repository

import (
	"context"
	"errors"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("users_email_unique"),
		},
		{
			Keys: bson.D{{Key: "organization_id", Value: 1}},
			Options: options.Index().
				SetName("users_organization_id_idx"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id bson.ObjectID, firstName, lastName string) error {
	update := bson.M{
		"$set": bson.M{
			"first_name": firstName,
			"last_name":  lastName,
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id bson.ObjectID, passwordHash string) error {
	update := bson.M{
		"$set": bson.M{
			"password_hash": passwordHash,
			"updated_at":    time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) SaveRefreshToken(ctx context.Context, id bson.ObjectID, refreshTokenHash string, expiresAt time.Time) error {
	update := bson.M{
		"$set": bson.M{
			"refresh_token_hash":    refreshTokenHash,
			"refresh_token_expires": expiresAt.UTC(),
			"updated_at":            time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) ClearRefreshToken(ctx context.Context, id bson.ObjectID) error {
	update := bson.M{
		"$unset": bson.M{
			"refresh_token_hash":    "",
			"refresh_token_expires": "",
		},
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) FindAllByOrganization(ctx context.Context, organizationID bson.ObjectID) ([]*domain.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"organization_id": organizationID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) FindByIDAndOrganization(ctx context.Context, id, organizationID bson.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "organization_id": organizationID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *UserRepository) UpdateUser(ctx context.Context, id, organizationID bson.ObjectID, firstName, lastName, phoneNumber, email string, vesselID *bson.ObjectID) error {
	setFields := bson.M{
		"first_name":   firstName,
		"last_name":    lastName,
		"phone_number": phoneNumber,
		"email":        email,
		"updated_at":   time.Now().UTC(),
	}
	if vesselID != nil {
		setFields["vessel_id"] = *vesselID
	}

	update := bson.M{"$set": setFields}
	if vesselID == nil {
		update["$unset"] = bson.M{"vessel_id": ""}
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "organization_id": organizationID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) ToggleUserStatus(ctx context.Context, id, organizationID bson.ObjectID, isActive bool) error {
	update := bson.M{
		"$set": bson.M{
			"is_active":  isActive,
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "organization_id": organizationID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id, organizationID bson.ObjectID, role domain.UserRole) error {
	update := bson.M{
		"$set": bson.M{
			"role":       role,
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "organization_id": organizationID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}
