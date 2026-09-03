package main

import (
	"fmt"
	"math"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	vals := make([]float64, 60)
	for i := range vals {
		vals[i] = 20 + 15*math.Sin(float64(i)/4)
	}

	fmt.Println(tuichart.Spark(vals))
}
