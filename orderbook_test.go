package main

import (
	"fmt"
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

	l.deleteOrder(buyOrderC)

	fmt.Println(l)
}

func TestOrderBook(t *testing.T) {
	ob := NewOrderBook()
	buyOrderA := NewOrder(true, 10)
	buyOrderB := NewOrder(true, 15)
	ob.PlaceOrder(18_000, buyOrderA)
	ob.PlaceOrder(18_000, buyOrderB)
	fmt.Println(ob.Bids)
}
