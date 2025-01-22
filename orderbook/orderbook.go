package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	ID        int64
	Size      float64
	Bid       bool
	Limit     *Limit
	TimeStamp int64
}

type Orders []*Order

func (o Orders) Len() int           { return len(o) }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].TimeStamp < o[j].TimeStamp }

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		ID:        int64(rand.Intn(10000000000)),
		Size:      size,
		Bid:       bid,
		TimeStamp: time.Now().UnixNano()}
}

func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)
}

func (o *Order) IsFilled() bool {
	return o.Size == 0.0
}

type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }

type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

type Limits []*Limit

func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if o == l.Orders[i] {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			break
		}
	}
	o.Limit = nil
	l.TotalVolume -= o.Size

	// resort the whole resting orders
	sort.Sort(l.Orders)
}
func (l *Limit) String() string {
	return fmt.Sprintf("price: %.2f | volume : %.2f", l.Price, l.TotalVolume)
}

func (l *Limit) addOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

func (l *Limit) Fill(order *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, o := range l.Orders {
		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if o.IsFilled() {
			ordersToDelete = append(ordersToDelete, o)
		}

		if order.IsFilled() {
			break
		}
	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches
}

func (l *Limit) fillOrder(a *Order, b *Order) Match {
	var (
		bid        *Order
		ask        *Order
		sizeFilled float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}
	if a.Size >= b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0.0
	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0.0
	}
	return Match{
		Ask:        ask,
		Bid:        bid,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

type OrderBook struct {
	asks []*Limit
	bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
	Orders    map[int64]*Order
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		Orders:    make(map[int64]*Order),
	}
}

func (ob *OrderBook) PlaceMarketOrder(order *Order) []Match {
	matches := []Match{}

	limitsToClear := []*Limit{}

	if order.Bid {
		if order.Size > ob.AskTotalVolume() {
			panic(fmt.Errorf("ot enough volume [size: %.2f] for market order [size: %.2f]", ob.AskTotalVolume(), order.Size))
		}
		for _, limit := range ob.Asks() {

			limitMatches := limit.Fill(order)
			matches = append(matches, limitMatches...)
			if len(limit.Orders) == 0.0 {
				limitsToClear = append(limitsToClear, limit)
			}
			if order.IsFilled() {
				break
			}
		}
	} else {
		if order.Size > ob.BidTotalVolume() {
			panic(fmt.Errorf("ot enough volume [size: %.2f] for market order [size: %.2f]", ob.BidTotalVolume(), order.Size))
		}
		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(order)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0.0 {
				limitsToClear = append(limitsToClear, limit)
			}
			if order.IsFilled() {
				break
			}
		}
	}
	for _, limit := range limitsToClear {
		ob.clearLimit(!order.Bid, limit)
	}
	return matches
}

func (ob *OrderBook) PlaceLimitOrder(price float64, order *Order) {
	var limit *Limit

	if order.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		if order.Bid {
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	ob.Orders[order.ID] = order
	limit.addOrder(order)
}

func (ob *OrderBook) clearLimit(bid bool, l *Limit) {

	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bids); i++ {
			if l == ob.bids[i] {
				ob.bids = append(ob.bids[:i], ob.bids[i+1:]...)
				break
			}
		}
	} else {
		delete(ob.AskLimits, l.Price)
		for i := 0; i < len(ob.asks); i++ {
			if l == ob.asks[i] {
				ob.asks = append(ob.asks[:i], ob.asks[i+1:]...)
				break
			}
		}
	}
}

func (ob *OrderBook) CancelOrder(order *Order) {
	limit := order.Limit
	delete(ob.Orders, order.ID)
	limit.DeleteOrder(order)
	if len(limit.Orders) == 0 {
		ob.clearLimit(order.Bid, limit)
	}
}

func (ob *OrderBook) BidTotalVolume() float64 {
	totalVolume := 0.0
	for _, limit := range ob.bids {
		totalVolume += limit.TotalVolume
	}
	return totalVolume
}

func (ob *OrderBook) AskTotalVolume() float64 {
	totalVolume := 0.0
	for _, limit := range ob.asks {
		totalVolume += limit.TotalVolume
	}
	return totalVolume
}

func (ob *OrderBook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}
func (ob *OrderBook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}
