package repository

import (
	"context"
	"time"

	"holidays-api-service/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type HolidayRepository interface {
	FindByYear(ctx context.Context, year int) ([]model.Holiday, error)
	SaveMany(ctx context.Context, holidays []model.Holiday) error
}

type holidayRepository struct {
	collection *mongo.Collection
}

func NewHolidayRepository(db *mongo.Database, collectionName string) HolidayRepository {
	return &holidayRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *holidayRepository) FindByYear(ctx context.Context, year int) ([]model.Holiday, error) {
	tracer := otel.Tracer("holidays-api-service/repository")
	ctx, span := tracer.Start(ctx, "MongoDB FindByYear", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		semconv.DBSystemMongoDB,
		attribute.String("db.name", r.collection.Database().Name()),
		attribute.String("db.operation", "find"),
		attribute.String("db.mongodb.collection", r.collection.Name()),
		attribute.Int("db.query.year", year),
	)
	defer span.End()

	filter := bson.M{"year": year}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var holidays []model.Holiday
	if err := cursor.All(ctx, &holidays); err != nil {
		return nil, err
	}

	return holidays, nil
}

func (r *holidayRepository) SaveMany(ctx context.Context, holidays []model.Holiday) error {
	if len(holidays) == 0 {
		return nil
	}

	tracer := otel.Tracer("holidays-api-service/repository")
	ctx, span := tracer.Start(ctx, "MongoDB SaveManyHolidays", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		semconv.DBSystemMongoDB,
		attribute.String("db.name", r.collection.Database().Name()),
		attribute.String("db.operation", "insertMany"),
		attribute.String("db.mongodb.collection", r.collection.Name()),
		attribute.Int("db.mongodb.documents", len(holidays)),
	)
	defer span.End()

	// Add creation timestamp
	now := time.Now()
	documents := make([]interface{}, len(holidays))
	for i, holiday := range holidays {
		holiday.CreatedAt = now
		documents[i] = holiday
	}

	_, err := r.collection.InsertMany(ctx, documents)
	return err
}
