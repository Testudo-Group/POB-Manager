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

var ErrTravelScheduleNotFound = errors.New("travel schedule not found")

type TravelScheduleRepository struct {
	collection *mongo.Collection
}

type TravelScheduleFilters struct {
	TransportID         *bson.ObjectID
	VesselID            *bson.ObjectID
	OriginVesselID      *bson.ObjectID
	DestinationVesselID *bson.ObjectID
	Status              *domain.TravelScheduleStatus
	UpcomingOnly        bool
	Limit               int64
}

func NewTravelScheduleRepository(db *mongo.Database) *TravelScheduleRepository {
	return &TravelScheduleRepository{
		collection: db.Collection("travel_schedules"),
	}
}

func (r *TravelScheduleRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "transport_id", Value: 1}, {Key: "departure_at", Value: 1}}},
		{Keys: bson.D{{Key: "vessel_id", Value: 1}, {Key: "departure_at", Value: 1}}},
		{Keys: bson.D{{Key: "origin_vessel_id", Value: 1}, {Key: "departure_at", Value: 1}}},
		{Keys: bson.D{{Key: "destination_vessel_id", Value: 1}, {Key: "departure_at", Value: 1}}},
	}
	for _, idx := range indexes {
		_, err := r.collection.Indexes().CreateOne(ctx, idx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TravelScheduleRepository) Create(ctx context.Context, schedule *domain.TravelSchedule) error {
	_, err := r.collection.InsertOne(ctx, schedule)
	return err
}

func (r *TravelScheduleRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.TravelSchedule, error) {
	var schedule domain.TravelSchedule
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&schedule)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTravelScheduleNotFound
		}
		return nil, err
	}
	return &schedule, nil
}

func (r *TravelScheduleRepository) FindUpcoming(ctx context.Context, limit int) ([]domain.TravelSchedule, error) {
	status := domain.TravelScheduleStatusPlanned
	return r.Find(ctx, TravelScheduleFilters{
		Status:       &status,
		UpcomingOnly: true,
		Limit:        int64(limit),
	})
}

func (r *TravelScheduleRepository) Find(ctx context.Context, filters TravelScheduleFilters) ([]domain.TravelSchedule, error) {
	filter := bson.M{}

	if filters.TransportID != nil {
		filter["transport_id"] = *filters.TransportID
	}
	if filters.VesselID != nil {
		filter["$or"] = bson.A{
			bson.M{"vessel_id": *filters.VesselID},
			bson.M{"origin_vessel_id": *filters.VesselID},
			bson.M{"destination_vessel_id": *filters.VesselID},
		}
	}
	if filters.OriginVesselID != nil {
		filter["origin_vessel_id"] = *filters.OriginVesselID
	}
	if filters.DestinationVesselID != nil {
		filter["destination_vessel_id"] = *filters.DestinationVesselID
	}
	if filters.Status != nil {
		filter["status"] = *filters.Status
	}
	if filters.UpcomingOnly {
		filter["departure_at"] = bson.M{"$gte": time.Now()}
	}

	sortDirection := -1
	if filters.UpcomingOnly {
		sortDirection = 1
	}

	opts := options.Find().SetSort(bson.D{{Key: "departure_at", Value: sortDirection}})
	if filters.Limit > 0 {
		opts.SetLimit(filters.Limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []domain.TravelSchedule
	if err := cursor.All(ctx, &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *TravelScheduleRepository) FindByTransportAndDateRange(ctx context.Context, transportID bson.ObjectID, start, end time.Time) ([]domain.TravelSchedule, error) {
	filter := bson.M{
		"transport_id": transportID,
		"departure_at": bson.M{"$gte": start, "$lte": end},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []domain.TravelSchedule
	if err := cursor.All(ctx, &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *TravelScheduleRepository) Update(ctx context.Context, schedule *domain.TravelSchedule) error {
	filter := bson.M{"_id": schedule.ID}
	update := bson.M{"$set": schedule}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrTravelScheduleNotFound
	}
	return nil
}

func (r *TravelScheduleRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrTravelScheduleNotFound
	}
	return nil
}
