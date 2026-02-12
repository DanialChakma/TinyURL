package main

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

/*
function only work on int type, let make it generic in next section
func MapValues(values []int, mappedFunc func(int) int) []int {

	var newValues []int

	for _, v := range values {
		newValue := mappedFunc(v)
		newValues = append(newValues, newValue)
	}

	return newValues

}
*/

// constraints.Ordered is the interface of union of all numeric and string type

func MapValues[T constraints.Ordered](values []T, mappedFunc func(T) T) []T {
	var newValues []T

	for _, v := range values {
		newValue := mappedFunc(v)
		newValues = append(newValues, newValue)
	}

	return newValues
}

type CustomData interface {
	constraints.Ordered | []byte | []rune
}

type User[T CustomData] struct {
	ID   int
	Name string
	Data T
}

// generic on custom map

type CustomMap[K comparable, V constraints.Ordered] map[K]V

// [ 1, 2 ,3 ] => [1, 4, 6] mapped into num*2
func main() {

	values := MapValues([]float32{2.5, 3.2, 4.1}, func(v float32) float32 {
		return v * 3
	})

	fmt.Printf("Values are: %+v\n", values)

	user := User[string]{
		ID:   0,
		Name: "Danial",
		Data: "Chakma",
	}

	fmt.Printf("Value is: %+v\n", user)

	cutomMap := make(CustomMap[string, string])
	cutomMap["abc"] = "danial"
	cutomMap["cde"] = "chakma"
	fmt.Printf("Map value: %+v", cutomMap)
	return
}
