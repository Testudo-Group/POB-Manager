package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrCertificateTypeNotFound = errors.New("certificate type not found")

type CertificateTypeRepository struct {
	collection *mongo.Collection
}

func NewCertificateTypeRepository(db *mongo.Database) *CertificateTypeRepository {
	return &CertificateTypeRepository{
		collection: db.Collection("certificate_types"),
	}
}

func (r *CertificateTypeRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{{Key: "code", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetName("certificate_types_code_unique"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *CertificateTypeRepository) Create(ctx context.Context, ct *domain.CertificateType) error {
	_, err := r.collection.InsertOne(ctx, ct)
	return err
}

func (r *CertificateTypeRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.CertificateType, error) {
	var ct domain.CertificateType
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&ct)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCertificateTypeNotFound
		}
		return nil, err
	}

	return &ct, nil
}

func (r *CertificateTypeRepository) FindByCode(ctx context.Context, code string) (*domain.CertificateType, error) {
	var ct domain.CertificateType
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&ct)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCertificateTypeNotFound
		}
		return nil, err
	}

	return &ct, nil
}


func (r *CertificateTypeRepository) FindAll(ctx context.Context) ([]domain.CertificateType, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []domain.CertificateType
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}
