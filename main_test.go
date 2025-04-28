package main

import (
	"fmt"
	"testing"
)

func TestMonitor(t *testing.T) {
	addr, err := HexToBase58("41a614f803b6fd780986a42c78ec9c7f77e6ded13c")
	if err != nil {
		panic(err)
	}
	fmt.Println("contract", addr)

	addr, err = HexToBase58("41518b7aa1165983778348bb889874e60d583c32d5")
	if err != nil {
		panic(err)
	}
	fmt.Println("from", addr)
}
