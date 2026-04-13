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

var ErrCertificateNotFound = errors.New("certificate not found")

type CertificateRepository struct {
	collection *mongo.Collection
}

func NewCertificateRepository(db *mongo.Database) *CertificateRepository {
	return &CertificateRepository{
		collection: db.Collection("certificates"),
	}
}

func (r *CertificateRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "personnel_id", Value: 1},
			{Key: "certificate_type", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetName("certificates_personnel_type_unique"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *CertificateRepository) Create(ctx context.Context, cert *domain.Certificate) error {
	_, err := r.collection.InsertOne(ctx, cert)
	return err
}

func (r *CertificateRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Certificate, error) {
	var cert domain.Certificate
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&cert)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCertificateNotFound
		}
		return nil, err
	}
	return &cert, nil
}

func (r *CertificateRepository) FindByPersonnelID(ctx context.Context, personnelID bson.ObjectID) ([]domain.Certificate, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"personnel_id": personnelID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var certs []domain.Certificate
	if err := cursor.All(ctx, &certs); err != nil {
		return nil, err
	}

	return certs, nil
}

func (r *CertificateRepository) Update(ctx context.Context, cert *domain.Certificate) error {
	filter := bson.M{"_id": cert.ID}
	update := bson.M{"$set": cert}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrCertificateNotFound
	}

	return nil
}

func (r *CertificateRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrCertificateNotFound
	}

	return nil
}

func (r *CertificateRepository) FindExpiring(ctx context.Context, days int) ([]domain.Certificate, error) {
	now := time.Now()
	expiryThreshold := now.AddDate(0, 0, days)
	filter := bson.M{
		"expiry_date": bson.M{"$gte": now, "$lte": expiryThreshold},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var certs []domain.Certificate
	if err := cursor.All(ctx, &certs); err != nil {
		return nil, err
	}
	return certs, nil
}
