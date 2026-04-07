// Package rating provides ELO rating calculation for players.
package rating

import "math"

const (
	// DefaultRating is the starting rating for new players.
	DefaultRating = 1500

	// KFactor determines how much ratings change per game.
	KFactor = 32

	// eloBase is the base for ELO calculation.
	eloBase = 10

	// eloDivisor is the divisor for ELO calculation.
	eloDivisor = 400.0
)

// CalculateELO calculates new ELO ratings for two players.
// Returns (newRating1, newRating2).
func CalculateELO(rating1, rating2 int, player1Won bool) (newRating1, newRating2 int) {
	expected1 := expectedScore(rating1, rating2)
	expected2 := expectedScore(rating2, rating1)

	var actual1, actual2 float64
	if player1Won {
		actual1 = 1.0
		actual2 = 0.0
	} else {
		actual1 = 0.0
		actual2 = 1.0
	}

	newRating1 = rating1 + int(KFactor*(actual1-expected1))
	newRating2 = rating2 + int(KFactor*(actual2-expected2))

	return newRating1, newRating2
}

// CalculateMultiplayerELO calculates new ratings for multiplayer game (FFA).
// Players are ranked by their final position (1st, 2nd, 3rd, 4th).
func CalculateMultiplayerELO(ratings, positions []int) []int {
	n := len(ratings)
	newRatings := make([]int, n)

	for i := 0; i < n; i++ {
		totalChange := 0.0

		for j := 0; j < n; j++ {
			if i == j {
				continue
			}

			expected := expectedScore(ratings[i], ratings[j])

			var actual float64

			switch {
			case positions[i] < positions[j]:
				actual = 1.0
			case positions[i] > positions[j]:
				actual = 0.0
			default:
				actual = 0.5
			}

			totalChange += actual - expected
		}

		newRatings[i] = ratings[i] + int(KFactor*totalChange/float64(n-1))
	}

	return newRatings
}

// expectedScore calculates the expected score for a player.
func expectedScore(ratingA, ratingB int) float64 {
	return 1.0 / (1.0 + math.Pow(eloBase, float64(ratingB-ratingA)/eloDivisor))
}
