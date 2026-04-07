package rating

import (
	"testing"
)

func TestCalculateELO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rating1    int
		rating2    int
		player1Won bool
		wantDelta1 int // approximate change for player1
		wantDelta2 int // approximate change for player2
	}{
		{
			name:       "equal ratings, player1 wins",
			rating1:    1500,
			rating2:    1500,
			player1Won: true,
			wantDelta1: 16,  // should gain ~16 points
			wantDelta2: -16, // should lose ~16 points
		},
		{
			name:       "equal ratings, player2 wins",
			rating1:    1500,
			rating2:    1500,
			player1Won: false,
			wantDelta1: -16,
			wantDelta2: 16,
		},
		{
			name:       "higher rated player wins",
			rating1:    1700,
			rating2:    1500,
			player1Won: true,
			wantDelta1: 8, // should gain less
			wantDelta2: -8,
		},
		{
			name:       "lower rated player wins (upset)",
			rating1:    1500,
			rating2:    1700,
			player1Won: true,
			wantDelta1: 24, // should gain more
			wantDelta2: -24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			new1, new2 := CalculateELO(tt.rating1, tt.rating2, tt.player1Won)

			delta1 := new1 - tt.rating1
			delta2 := new2 - tt.rating2

			// Check if deltas are within reasonable range (±2 points)
			if abs(delta1-tt.wantDelta1) > 2 {
				t.Errorf("player1 rating change = %d, want ~%d", delta1, tt.wantDelta1)
			}

			if abs(delta2-tt.wantDelta2) > 2 {
				t.Errorf("player2 rating change = %d, want ~%d", delta2, tt.wantDelta2)
			}

			// Total rating should be conserved (approximately)
			totalBefore := tt.rating1 + tt.rating2
			totalAfter := new1 + new2

			if abs(totalBefore-totalAfter) > 2 {
				t.Errorf("total rating not conserved: before=%d, after=%d", totalBefore, totalAfter)
			}
		})
	}
}

func TestCalculateMultiplayerELO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ratings   []int
		positions []int
		wantGain  []bool // true if player should gain rating
	}{
		{
			name:      "4 equal players, normal finish",
			ratings:   []int{1500, 1500, 1500, 1500},
			positions: []int{1, 2, 3, 4},
			wantGain:  []bool{true, true, false, false},
		},
		{
			name:      "lowest rated wins",
			ratings:   []int{1400, 1500, 1600, 1700},
			positions: []int{1, 2, 3, 4},
			wantGain:  []bool{true, true, false, false},
		},
		{
			name:      "highest rated wins",
			ratings:   []int{1700, 1600, 1500, 1400},
			positions: []int{1, 2, 3, 4},
			wantGain:  []bool{true, true, false, false}, // winner still gains a bit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			newRatings := CalculateMultiplayerELO(tt.ratings, tt.positions)

			if len(newRatings) != len(tt.ratings) {
				t.Fatalf("got %d ratings, want %d", len(newRatings), len(tt.ratings))
			}

			for i := range newRatings {
				gained := newRatings[i] > tt.ratings[i]
				if gained != tt.wantGain[i] {
					t.Errorf("player %d: rating %d->%d, wantGain=%v",
						i, tt.ratings[i], newRatings[i], tt.wantGain[i])
				}
			}
		})
	}
}

func TestDefaultRating(t *testing.T) {
	t.Parallel()

	if DefaultRating != 1500 {
		t.Errorf("DefaultRating = %d, want 1500", DefaultRating)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}
