package payroll

import "math"

// ProvincialRK returns V and KP for income A. Thresholds holds the lower bound of each band
// (same layout as Table 8.1 column A).
func ProvincialRK(set ProvincialBracketSet, A float64) (V, KP float64) {
	if len(set.V) == 0 {
		return 0, 0
	}
	for i := range set.V {
		lower := set.Thresholds[i]
		var upper float64
		if i+1 < len(set.Thresholds) {
			upper = set.Thresholds[i+1]
		} else {
			upper = math.Inf(1)
		}
		if A >= lower && A < upper {
			return set.V[i], set.KP[i]
		}
	}
	return set.V[len(set.V)-1], set.KP[len(set.KP)-1]
}
