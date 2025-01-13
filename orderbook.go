package main

import (
	"fmt"
	"time"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	TimeStamp int64
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		TimeStamp: time.Now().UnixNano()}
}

func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)
}

type Limit struct {
	Price       float64
	Orders      []*Order
	TotalVolume float64
}

func (l *Limit) deleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if o == l.Orders[i] {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			break
		}
	}
	o.Limit = nil
	l.TotalVolume -= o.Size

	// TODO: resort the whole resting orders
}
func (l *Limit) String() string {
	return fmt.Sprintf("price: %.2f | volume : %.2f", l.Price, l.TotalVolume)
}

func (l *Limit) addOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

type OrderBook struct {
	Asks []*Limit
	Bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func (ob *OrderBook) PlaceOrder(price float64, o *Order) []Match {
	// 1. try to match orders
	// matching logic

	// 2. add the rest of the order
	if o.Size > 0.0 {
		ob.add(price, o)
	}
	return []Match{}
}

func (ob *OrderBook) add(price float64, order *Order) {
	var limit *Limit

	if order.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		limit.addOrder(order)

		if order.Bid {
			ob.Bids = append(ob.Bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.Asks = append(ob.Asks, limit)
			ob.AskLimits[price] = limit
		}
	} else {
		limit.addOrder(order)
	}
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks:      []*Limit{},
		Bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}
