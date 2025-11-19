package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"janus/internal/model"
	"time"
)

// mongoLinkRepository implements LinkRepository interface
type mongoLinkRepository struct {
	db *mongo.Database
}

// NewMongoLinkRepository creates a new MongoDB link repository instance
func NewMongoLinkRepository(client *mongo.Client, dbName string) LinkRepository {
	return &mongoLinkRepository{
		db: client.Database(dbName),
	}
}

// Insert creates a new link or updates if exists (upsert)
func (r *mongoLinkRepository) Insert(ctx context.Context, link, shortLinkHash string, timestamp time.Time) error {
	updateQuery := bson.M{
		"link":            link,
		"short_link_hash": shortLinkHash,
		"created_at":      timestamp,
		"updated_at":      time.Now(),
	}

	opts := options.Replace().SetUpsert(true)
	_, err := r.db.Collection("links").
		ReplaceOne(ctx, bson.M{"link": link}, updateQuery, opts)

	return err
}

// FindByHash retrieves a link by its short hash
func (r *mongoLinkRepository) FindByHash(ctx context.Context, shortLinkHash string) (*model.Link, error) {
	var link model.Link
	err := r.db.Collection("links").
		FindOne(ctx, bson.M{"short_link_hash": shortLinkHash}).
		Decode(&link)

	if err != nil {
		return nil, err
	}

	return &link, nil
}

// Update updates an existing link
func (r *mongoLinkRepository) Update(ctx context.Context, shortLinkHash, newLink string) (int64, error) {
	updateQuery := bson.M{"$set": bson.M{"link": newLink, "updated_at": time.Now()}}
	result, err := r.db.Collection("links").
		UpdateOne(ctx, bson.M{"short_link_hash": shortLinkHash}, updateQuery)

	if err != nil {
		return 0, err
	}

	return result.MatchedCount, nil
}

// Delete removes a link by its short hash
func (r *mongoLinkRepository) Delete(ctx context.Context, shortLinkHash string) error {
	_, err := r.db.Collection("links").
		DeleteOne(ctx, bson.M{"short_link_hash": shortLinkHash})

	return err
}
