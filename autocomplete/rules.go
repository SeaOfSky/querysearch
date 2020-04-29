package autocomplete

type Rule3 struct{
	prefixBeginPosition int
	isBeginMatching bool
	isMatch bool
}

type Rule4 struct{
	prefixBeginPosition int
	isBeginMatching bool
	isMatch bool
}

func NewRuler3() *Rule3 {
	return &Rule3{isMatch:true}
}

func (r *Rule3) CheckRule(record *Record,bestPrefix *PrefixCouple,prefixPoisition int) {
	if r.isMatch {
		r.isMatch = r.IsMatchRule(record,bestPrefix,prefixPoisition)
	}
}

// Rule 3 : the prefix could match the word in the middle of query,but their position must be continuous.
func (r *Rule3) IsMatchRule(record *Record,bestPrefix *PrefixCouple,prefixPoisition int) bool{
	if r.isBeginMatching {
		if record.Value.WordPositionMap[bestPrefix.SimilarPrefix] != prefixPoisition + r.prefixBeginPosition {
			return  false
		}
	}else {
		r.isBeginMatching = true
		r.prefixBeginPosition = record.Value.WordPositionMap[bestPrefix.SimilarPrefix]
	}

	switch bestPrefix.PartialMatch {
	case true:
		if len(bestPrefix.SimilarPrefix) < len(bestPrefix.OriginPrefix) {
			return false
		}
		if bestPrefix.SimilarPrefix[:len(bestPrefix.OriginPrefix)] != bestPrefix.OriginPrefix {
			return false
		}
	case false:
		if bestPrefix.SimiScore < 0.99 {
			return  false
		}
	}

	return true
}

func (r *Rule3) IsMatch() bool {
	return r.isMatch
}


func NewRuler4() *Rule4 {
	return &Rule4{isMatch:true}
}

func (r *Rule4) CheckRule(record *Record,bestPrefix *PrefixCouple,prefixPoisition int) {
	if r.isMatch {
		r.isMatch = r.IsMatchRule(record,bestPrefix,prefixPoisition)
	}
}

// Rule 4 : the fuzzy prefix could match the word in the middle of query,but their position must be continuous.
func (r *Rule4) IsMatchRule(record *Record,bestPrefix *PrefixCouple,prefixPoisition int) bool{
	if r.isBeginMatching {
		if record.Value.WordPositionMap[bestPrefix.SimilarPrefix] != prefixPoisition + r.prefixBeginPosition {
			return  false
		}
	}else {
		r.isBeginMatching = true
		r.prefixBeginPosition = record.Value.WordPositionMap[bestPrefix.SimilarPrefix]
	}
	return true
}

func (r *Rule4) IsMatch() bool {
	return r.isMatch
}