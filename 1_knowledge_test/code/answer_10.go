package main

import (
	"fmt"
	"math"
)

type Product struct {
	Name       string
	WeightGram float64 // Berat dalam gram
	Length     float64 // Panjang dalam mm
	Width      float64 // Lebar dalam mm
	Height     float64 // Tinggi dalam mm
}

type Order struct {
	Products []OrderItem
}

type OrderItem struct {
	Product  Product
	Quantity int
}

type ShippingCalculator struct {
	RatePerKg float64 // Rate per kg in Rupiah
}

func NewShippingCalculator(ratePerKg float64) *ShippingCalculator {
	return &ShippingCalculator{
		RatePerKg: ratePerKg,
	}
}

func (sc *ShippingCalculator) calculateActualWeight(order Order) float64 {
	totalWeightGram := 0.0
	for _, item := range order.Products {
		totalWeightGram += item.Product.WeightGram * float64(item.Quantity)
	}
	return totalWeightGram / 1000.0 // Convert to kg
}

// Formula: (Length x Width x Height) / 5000 for each product
func (sc *ShippingCalculator) calculateVolumetricWeight(order Order) float64 {
	totalVolumetricWeight := 0.0
	for _, item := range order.Products {
		product := item.Product
		// Volume in cubic mm
		volume := product.Length * product.Width * product.Height
		// Volumetric weight per item (divide by 5000 to get weight in grams, then convert to kg)
		volumetricWeightPerItem := volume / 5000.0 / 1000.0 // kg
		totalVolumetricWeight += volumetricWeightPerItem * float64(item.Quantity)
	}
	return totalVolumetricWeight
}

func (sc *ShippingCalculator) calculateChargeableWeight(order Order) (float64, string) {
	actualWeight := sc.calculateActualWeight(order)
	volumetricWeight := sc.calculateVolumetricWeight(order)

	if actualWeight > volumetricWeight {
		return actualWeight, "actual"
	}
	return volumetricWeight, "volumetric"
}

// CalculateShippingCost calculates the total shipping cost
func (sc *ShippingCalculator) CalculateShippingCost(order Order) (float64, map[string]interface{}) {
	actualWeight := sc.calculateActualWeight(order)
	volumetricWeight := sc.calculateVolumetricWeight(order)
	chargeableWeight, weightType := sc.calculateChargeableWeight(order)

	// Round up to nearest kg for billing
	billableWeight := math.Ceil(chargeableWeight)
	shippingCost := billableWeight * sc.RatePerKg

	details := map[string]interface{}{
		"actual_weight_kg":     actualWeight,
		"volumetric_weight_kg": volumetricWeight,
		"chargeable_weight_kg": chargeableWeight,
		"billable_weight_kg":   billableWeight,
		"weight_type_used":     weightType,
		"rate_per_kg":          sc.RatePerKg,
		"shipping_cost":        shippingCost,
	}

	return shippingCost, details
}

// Closure function for creating product calculator
func createProductCalculator(ratePerKg float64) func(Order) (float64, map[string]interface{}) {
	calculator := NewShippingCalculator(ratePerKg)
	return calculator.CalculateShippingCost
}

// Anonymous function example for quick calculations
var quickCalculate = func(products []OrderItem, rate float64) float64 {
	order := Order{Products: products}
	calculator := NewShippingCalculator(rate)
	cost, _ := calculator.CalculateShippingCost(order)
	return cost
}

func answer_10() {
	productA := Product{
		Name:       "Produk A",
		WeightGram: 30,
		Length:     115,
		Width:      85,
		Height:     25,
	}

	productB := Product{
		Name:       "Produk B",
		WeightGram: 28000,
		Length:     1290,
		Width:      300,
		Height:     625,
	}

	//  Function using closure (named function)
	calculateShipping := createProductCalculator(10000) // Rp. 10,000 per kg

	//  Dynamic quantity calculation example
	fmt.Println("=== Contoh Perhitungan Ongkos Kirim ===")

	// Example 1: 2 Produk A, 1 Produk B
	order1 := Order{
		Products: []OrderItem{
			{Product: productA, Quantity: 2},
			{Product: productB, Quantity: 1},
		},
	}

	cost1, details1 := calculateShipping(order1)
	fmt.Printf("\nOrder 1: 2x Produk A + 1x Produk B\n")
	fmt.Printf("Berat Aktual: %.3f kg\n", details1["actual_weight_kg"])
	fmt.Printf("Berat Volumetrik: %.3f kg\n", details1["volumetric_weight_kg"])
	fmt.Printf("Berat yang Ditagih: %.0f kg (%s)\n", details1["billable_weight_kg"], details1["weight_type_used"])
	fmt.Printf("Ongkos Kirim: Rp. %.0f\n", cost1)

	// Example 2: 5 Produk A
	order2 := Order{
		Products: []OrderItem{
			{Product: productA, Quantity: 5},
		},
	}

	cost2, details2 := calculateShipping(order2)
	fmt.Printf("\nOrder 2: 5x Produk A\n")
	fmt.Printf("Berat Aktual: %.3f kg\n", details2["actual_weight_kg"])
	fmt.Printf("Berat Volumetrik: %.3f kg\n", details2["volumetric_weight_kg"])
	fmt.Printf("Berat yang Ditagih: %.0f kg (%s)\n", details2["billable_weight_kg"], details2["weight_type_used"])
	fmt.Printf("Ongkos Kirim: Rp. %.0f\n", cost2)

	// Demonstrasi perhitungan berat berdasarkan berat aktual dan volumetrik
	fmt.Printf("\n=== Detail Perhitungan Berat ===")

	// Produk A analysis
	fmt.Printf("\nProduk A (per unit):\n")
	fmt.Printf("- Berat Aktual: %.0f gram = %.3f kg\n", productA.WeightGram, productA.WeightGram/1000)
	volumeA := productA.Length * productA.Width * productA.Height
	volumetricWeightA := volumeA / 5000.0 / 1000.0
	fmt.Printf("- Volume: %.0f x %.0f x %.0f = %.0f mm³\n", productA.Length, productA.Width, productA.Height, volumeA)
	fmt.Printf("- Berat Volumetrik: %.0f / 5000 / 1000 = %.6f kg\n", volumeA, volumetricWeightA)

	// Produk B analysis
	fmt.Printf("\nProduk B (per unit):\n")
	fmt.Printf("- Berat Aktual: %.0f gram = %.3f kg\n", productB.WeightGram, productB.WeightGram/1000)
	volumeB := productB.Length * productB.Width * productB.Height
	volumetricWeightB := volumeB / 5000.0 / 1000.0
	fmt.Printf("- Volume: %.0f x %.0f x %.0f = %.0f mm³\n", productB.Length, productB.Width, productB.Height, volumeB)
	fmt.Printf("- Berat Volumetrik: %.0f / 5000 / 1000 = %.3f kg\n", volumeB, volumetricWeightB)

	// Using anonymous function for quick calculation
	fmt.Printf("\n=== Perhitungan Cepat dengan Anonymous Function ===")
	quickCost := quickCalculate([]OrderItem{{Product: productA, Quantity: 3}}, 10000)
	fmt.Printf("\n3x Produk A (quick calc): Rp. %.0f\n", quickCost)

	fmt.Printf("\n=== Fungsi Dinamis untuk Berbagai Skenario ===")
	// Function to calculate for any combination
	calculateForCombination := func(quantityA, quantityB int) {
		order := Order{
			Products: []OrderItem{},
		}

		if quantityA > 0 {
			order.Products = append(order.Products, OrderItem{Product: productA, Quantity: quantityA})
		}
		if quantityB > 0 {
			order.Products = append(order.Products, OrderItem{Product: productB, Quantity: quantityB})
		}

		cost, details := calculateShipping(order)
		fmt.Printf("\n%dx Produk A + %dx Produk B:\n", quantityA, quantityB)
		fmt.Printf("Ongkos Kirim: Rp. %.0f (Berat: %.0f kg %s)\n",
			cost, details["billable_weight_kg"], details["weight_type_used"])
	}

	// Test combinations
	calculateForCombination(1, 0)
	calculateForCombination(0, 1)
	calculateForCombination(3, 2)
	calculateForCombination(10, 0)
}
