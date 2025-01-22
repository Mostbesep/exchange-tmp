package orderbook

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 1)
	buyOrderB := NewOrder(true, 2)
	buyOrderC := NewOrder(true, 3)
	l.addOrder(buyOrderA)
	l.addOrder(buyOrderB)
	l.addOrder(buyOrderC)

	l.DeleteOrder(buyOrderC)

	assert.Len(t, l.Orders, 2)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderBook()
	sellOrderA := NewOrder(false, 10)
	sellOrderB := NewOrder(false, 5)
	sellOrderC := NewOrder(false, 7)
	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(9_000, sellOrderB)
	ob.PlaceLimitOrder(10_000, sellOrderC)
	assert.Equal(t, sellOrderA, ob.Orders[sellOrderA.ID])
	assert.Equal(t, sellOrderB, ob.Orders[sellOrderB.ID])
	assert.Len(t, ob.Orders, 3)
	assert.Len(t, ob.asks, 2) // limits len <10k, 9k>
	assert.Equal(t, 17.0, ob.asks[0].TotalVolume)
	assert.Equal(t, 5.0, ob.asks[1].TotalVolume)
	ob.CancelOrder(sellOrderA)
	assert.Len(t, ob.Orders, 2)
	assert.Equal(t, 7.0, ob.asks[0].TotalVolume)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderBook()
	sellOrderA := NewOrder(false, 20)
	ob.PlaceLimitOrder(10_000, sellOrderA)

	buyOrderA := NewOrder(true, 10)
	matches := ob.PlaceMarketOrder(buyOrderA)

	assert.Len(t, matches, 1)
	assert.Len(t, ob.asks, 1)
	assert.Equal(t, ob.AskTotalVolume(), 10.0)
	assert.Equal(t, matches[0].Ask, sellOrderA)
	assert.Equal(t, matches[0].Bid, buyOrderA)
	assert.Equal(t, matches[0].SizeFilled, 10.0)
	assert.Equal(t, matches[0].Price, 10_000.0)
	assert.True(t, buyOrderA.IsFilled())
	fmt.Printf("%+v", matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderBook()
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(5_000, buyOrderC)
	assert.Equal(t, 23.0, ob.BidTotalVolume())

	sellOrderA := NewOrder(false, 4.0)
	matches := ob.PlaceMarketOrder(sellOrderA)
	assert.Len(t, matches, 1)
	assert.Equal(t, matches[0].Ask, sellOrderA)
	assert.Equal(t, matches[0].Bid, buyOrderA)
	assert.Equal(t, matches[0].SizeFilled, 4.0)
	assert.Equal(t, ob.bids[0].TotalVolume, 1.0)
	assert.Equal(t, matches[0].Price, 10_000.0)

	sellOrderB := NewOrder(false, 18.0)
	matches = ob.PlaceMarketOrder(sellOrderB)
	assert.Len(t, matches, 3)
	assert.Len(t, ob.bids, 1)
	assert.Equal(t, ob.bids[0].Price, 5000.0)
	assert.Equal(t, 1.0, ob.BidTotalVolume())

}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderBook()
	buyOrderA := NewOrder(true, 10)
	buyOrderB := NewOrder(true, 10)
	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(9_000, buyOrderB)

	limitWillBeCancel := ob.bids[0]

	ob.CancelOrder(buyOrderA)
	assert.Len(t, ob.Orders, 1)
	_, exists := ob.Orders[buyOrderA.ID]
	assert.False(t, exists)

	assert.Equal(t, limitWillBeCancel.TotalVolume, 0.0)
	assert.Len(t, ob.bids, 1)
	assert.Equal(t, ob.BidTotalVolume(), 10.0)

	sellOrderA := NewOrder(false, 4.0)
	matches := ob.PlaceMarketOrder(sellOrderA)
	assert.Len(t, matches, 1)
	assert.Len(t, ob.bids, 1)
	assert.Equal(t, ob.BidTotalVolume(), 6.0)
	assert.Equal(t, matches[0].Bid, buyOrderB)
	assert.Equal(t, matches[0].Ask, sellOrderA)
	assert.Equal(t, matches[0].SizeFilled, 4.0)
	assert.Equal(t, matches[0].Price, 9_000.0)
}
