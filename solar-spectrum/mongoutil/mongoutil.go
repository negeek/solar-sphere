// Package mongoutil connects to MongoDB and hands back a *mongo.Database
// that callers own and pass explicitly to their repositories — no package
// level globals.
package mongoutil

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect dials MongoDB, pings it to confirm the connection is live, and
// returns the requested database along with the underlying client (needed
// so the caller can Disconnect it on shutdown).
func Connect(ctx context.Context, connString, dbName string) (*mongo.Client, *mongo.Database, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(connString).SetServerAPIOptions(serverAPI))
	if err != nil {
		return nil, nil, fmt.Errorf("mongoutil: connect: %w", err)
	}

	var pingResult bson.M
	if err := client.Database("admin").RunCommand(connectCtx, bson.D{{Key: "ping", Value: 1}}).Decode(&pingResult); err != nil {
		return nil, nil, fmt.Errorf("mongoutil: ping: %w", err)
	}

	return client, client.Database(dbName), nil
}

// Disconnect closes client, waiting up to timeout.
func Disconnect(ctx context.Context, client *mongo.Client, timeout time.Duration) error {
	disconnectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return client.Disconnect(disconnectCtx)
}
