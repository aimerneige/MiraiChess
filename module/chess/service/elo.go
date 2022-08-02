package service

import (
	"math"
)

// CalculateNewRate calculate new rate of the player
func CalculateNewRate(whiteRate, blackRate uint, whiteScore, blackScore float64) (uint, uint) {
	k := getKFactor(whiteRate, blackRate)
	exceptionWhite := calculateException(whiteRate, blackRate)
	exceptionBlack := calculateException(blackRate, whiteRate)
	whiteRate = calculateRate(whiteRate, whiteScore, exceptionWhite, k)
	blackRate = calculateRate(blackRate, blackScore, exceptionBlack, k)
	return whiteRate, blackRate
}

func calculateException(rate uint, opponentRate uint) float64 {
	return 1 / (1 + math.Pow(10, float64(opponentRate-rate)/400))
}

func calculateRate(rate uint, score float64, exception float64, k uint) uint {
	newRate := uint(math.Round(float64(rate) + float64(k)*(score-exception)))
	if newRate < 1 {
		newRate = 1
	}
	return newRate
}

func getKFactor(rateA, rateB uint) uint {
	if rateA > 2400 && rateB > 2400 {
		return 16
	}
	if rateA > 2100 && rateB > 2100 {
		return 24
	}
	return 32
}
