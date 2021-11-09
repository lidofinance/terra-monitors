package math

import (
	"math/big"
	"sort"
)

type SortedBigInts []*big.Int

func (a SortedBigInts) Len() int           { return len(a) }
func (a SortedBigInts) Less(i, j int) bool { return a[i].Cmp(a[j]) < 0 }
func (a SortedBigInts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func Median(vals []*big.Int) *big.Int {
	// sort.Sort sorts in-place, we don't want to modify the input vals.
	var sortableVals = SortedBigInts(make([]*big.Int, len(vals)))
	copy(sortableVals, vals)
	sort.Sort(sortableVals)

	pivot := len(sortableVals) / 2
	if len(sortableVals)%2 == 0 {
		var sum = big.NewInt(0).Add(sortableVals[pivot], sortableVals[pivot-1])
		return sum.Div(sum, big.NewInt(2))
	} else {
		return sortableVals[pivot]
	}
}

// MeanAbsoluteDeviation implements https://en.wikipedia.org/wiki/Median_absolute_deviation.
func MeanAbsoluteDeviation(vals []*big.Int) *big.Int {
	var (
		median = Median(vals)
		diffs  SortedBigInts
	)

	for _, val := range vals {
		var sub = big.NewInt(0).Sub(val, median)
		diffs = append(diffs, sub.Abs(sub))
	}
	sort.Sort(diffs)

	return Median(diffs)
}

// GetMeanAbsoluteDeviationOutliers returns a list of indices of all values that
// satisfy ABS(value - median) > nMads * MeanAbsoluteDeviation (which is a list
// of outliers).
func GetMeanAbsoluteDeviationOutliers(vals []*big.Int, nMads int64) []int {
	var (
		outlierIndices []int
		median, mad    = Median(vals), MeanAbsoluteDeviation(vals)
	)

	for idx, val := range vals {
		var (
			diff    = big.NewInt(0).Sub(val, median)
			absDiff = big.NewInt(0).Abs(diff)
		)
		if absDiff.Cmp(big.NewInt(0).Mul(mad, big.NewInt(nMads))) > 0 {
			outlierIndices = append(outlierIndices, idx)
		}
	}

	return outlierIndices
}
