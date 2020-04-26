package autocomplete

//OpCost defines operation costs for EditDistance algorithm
type OpCost struct {
	Delete  int
	Add     int
	Replace int
}

//EditDistance defines from n to m, the total cost of operations
func EditDistance(n, m string, config *OpCost) int {
	nRunes, mRunes := []rune(n), []rune(m)
	ln, lm := len(nRunes), len(mRunes)
	preRow, preCol := make([]int, lm+1), make([]int, ln+1)
	for i := 0; i <= lm; i++ {
		preRow[i] = i * config.Add
	}
	for i := 0; i <= ln; i++ {
		preCol[i] = i * config.Delete
	}
	nextRow := make([]int, lm+1)
	for i := 1; i <= ln; i++ {
		nextRow[0] = preCol[i]
		for j := 1; j <= lm; j++ {
			if nRunes[i-1] == mRunes[j-1] {
				nextRow[j] = preRow[j-1]
				continue
			}
			nextRow[j] = MinInts(preRow[j-1]+config.Replace, preRow[j]+config.Delete, nextRow[j-1]+config.Add)
		}
		preRow, nextRow = nextRow, preRow
	}
	return preRow[lm]
}

//MinInts get minimum from a list of integers
func MinInts(a int, others ...int) int {
	result := a
	for _, other := range others {
		if other < result {
			result = other
		}
	}
	return result
}

