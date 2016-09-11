package main

import (
	"bytes"
	"fmt"
	"math"
	"math/cmplx"
	"os"
)

const (
	UpperLimit = 300
	LowerLimit = 40
	FuzFactor  = 2
)

var Range = []int{40, 80, 120, 180, UpperLimit + 1}

func MakeFingerprint(audio []byte) {
	size := len(audio)
	amountPossible := size / 4096
	results := make([][]complex128)

	for times := 0; times < amountPossible; times++ {
		comp := make([]complex128, 4096)
		for i := 0; i < 4096; i++ {
			comp[i] = complex128(audio[times*4096+i], 0)
		}
		results[times] = FFT(comp)
	}
	determineKeyPoints(results)
}

func determineKeyPoints(results [][]complex128) {
	var highscores [len(results)][5]float64
	var recordPoints [len(results)][UpperLimit]float64
	var points [len(results)][5]float64

	for t := 0; t < len(results); t++ {
		for freq := LowerLimit; freq < UpperLimit-1; freq++ {
			//get magnitude
			mag := math.Log(cmplx.Abs(results[t][freq] + 1))
			//find out current range
			idx := getIndex(freq)

			if mag > highscores[t][idx] {
				highscores[t][id] = mag
				recordPoints[t][freq] = 1
				points[t][idx] = freq
			}
		}
		f, err := os.Create("./result.txt")
		if err != nil {
			fmt.Println(err)
		}

		songHash := hash(points[t][0], points[t][1], points[t][2], points[t][3])
		for k := 0; k < 5; k++ {
			f.WriteString(fmt.Sprintf("%d:%d;%d\t", songHash,
				highscores[t][k], recordPoints[t][k]))
		}
		f.WriteString("\n")
		f.Close()
		//here we can add data to map which would be use to recognition.
	}
}

func hash(a, b, c, d float64) float64 {
	return (d-(d%FuzFactor))*100000000 + (c-(c%FuzFactor))*
		10000 + (b-(b%FuzFactor))*100 + (a - (a % FuzFactor))
}

func getIndex(freq int) int {
	i := 0
	for _; Range[i] < freq; i++ {
	}
	return i
}
