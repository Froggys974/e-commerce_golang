package product

import "testing"

func TestPriceTTC(t *testing.T) {
	testCases := []struct {
		priceHTCents int64
		vatPercent   int
		want         int64
	}{
		{1000, 20, 1200},
		{999, 20, 1199},  // 1198.8 -> 1199
		{3, 20, 4},       // 3.6 -> 4
		{1, 20, 1},       // 1.2 -> 1
		{1005, 10, 1106}, // 1105.5 -> 1106
		{1000, 0, 1000},
		{0, 20, 0},
	}

	for _, testCase := range testCases {
		got := PriceTTC(testCase.priceHTCents, testCase.vatPercent)
		if got != testCase.want {
			t.Errorf("PriceTTC(%d, %d) = %d, want %d", // error et pas fatal pour avoir toutes les failures sans que le test s'arrete
				testCase.priceHTCents, testCase.vatPercent, got, testCase.want)
		}
	}
}
