package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetAuctionInterval(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{"valid seconds", "30s", 30 * time.Second},
		{"valid minutes", "2m", 2 * time.Minute},
		{"invalid value", "invalid", 5 * time.Minute},
		{"empty value", "", 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := os.Getenv("AUCTION_INTERVAL")
			os.Setenv("AUCTION_INTERVAL", tt.envValue)
			defer os.Setenv("AUCTION_INTERVAL", original)

			result := getAuctionInterval()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuctionRepository_UpdateAuction(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should update auction successfully", func(mt *mtest.T) {
		repo := NewAuctionRepository(mt.DB)

		auction := &auction_entity.Auction{
			Id:          "test-id",
			ProductName: "Test Product",
			Category:    "Electronics",
			Description: "Test Description",
			Condition:   auction_entity.New,
			Status:      auction_entity.Completed,
			Timestamp:   time.Now(),
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.UpdateAuction(context.Background(), auction)
		assert.Nil(t, err)
	})

	mt.Run("should handle database error", func(mt *mtest.T) {
		repo := NewAuctionRepository(mt.DB)

		auction := &auction_entity.Auction{
			Id:     "test-id",
			Status: auction_entity.Completed,
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Code:    11000,
			Message: "database error",
		}))

		err := repo.UpdateAuction(context.Background(), auction)
		assert.NotNil(t, err)
	})
}

func TestAuctionRepository_CreateAuction(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should create auction and start goroutine", func(mt *mtest.T) {
		repo := NewAuctionRepository(mt.DB)

		auction, _ := auction_entity.CreateAuction(
			"Test Product",
			"Electronics",
			"Test Description",
			auction_entity.New,
		)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.CreateAuction(context.Background(), auction)
		assert.Nil(t, err)
	})
}

func TestAuctionRepository_CloseAuctionAfterInterval(t *testing.T) {
	original := os.Getenv("AUCTION_INTERVAL")
	os.Setenv("AUCTION_INTERVAL", "10ms")
	defer os.Setenv("AUCTION_INTERVAL", original)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should close active auction", func(mt *mtest.T) {
		repo := NewAuctionRepository(mt.DB)
		auctionId := "test-auction-id"

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "auctions.auctions", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: auctionId},
			{Key: "product_name", Value: "Test Product"},
			{Key: "category", Value: "Electronics"},
			{Key: "description", Value: "Test Description"},
			{Key: "condition", Value: auction_entity.New},
			{Key: "status", Value: auction_entity.Active},
			{Key: "timestamp", Value: time.Now().Unix()},
		}))

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		go repo.closeAuctionAfterInterval(auctionId)

		time.Sleep(50 * time.Millisecond)
	})
}
