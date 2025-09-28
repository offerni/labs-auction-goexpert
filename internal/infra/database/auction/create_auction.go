package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection      *mongo.Collection
	auctionInterval time.Duration
	mutex           *sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection:      database.Collection("auctions"),
		auctionInterval: getAuctionInterval(),
		mutex:           &sync.Mutex{},
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	go ar.closeAuctionAfterInterval(auctionEntity.Id)

	return nil
}

func (ar *AuctionRepository) closeAuctionAfterInterval(auctionId string) {
	time.Sleep(ar.auctionInterval)

	// Create new background context for db
	ctx := context.Background()

	ar.mutex.Lock()
	defer ar.mutex.Unlock()

	auction, err := ar.FindAuctionById(ctx, auctionId)
	if err != nil {
		logger.Error("Error trying to find auction for auto-close", err)
		return
	}

	if auction.Status == auction_entity.Active {
		auction.Status = auction_entity.Completed
		updateErr := ar.UpdateAuction(ctx, auction)
		if updateErr != nil {
			logger.Error("Error trying to auto-close auction", updateErr)
			return
		}
	}
}

func (ar *AuctionRepository) UpdateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	filter := bson.M{"_id": auctionEntity.Id}

	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}

	update := bson.M{"$set": auctionEntityMongo}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Error trying to update auction", err)
		return internal_error.NewInternalServerError("Error trying to update auction")
	}

	return nil
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5 // fallback
	}
	return duration
}
