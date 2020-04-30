package autocomplete

import "math"

func ComputeFinalScore(recordScore *RecordScore,poi *POIDetail) float64 {
	recordFinalScore := recordScore.TotalScore * poi.Confidence
	poiScore :=  ComputePOIScore(poi)
	return recordFinalScore + poiScore
}

func ComputePOIScore(poi *POIDetail) float64 {
	score := math.Log2(float64(poi.SelectionCnt) / float64(MeanRecordSelectinCnt) + 1)
	return score
}

func ComputePrefixMatchScore(originPrefixPosition int,bestPrefix *PrefixCouple, record *Record) (p_score float64,scoreDetail *SubRecordScore)  {
	scoreDetail = &SubRecordScore{}
	p_score = 0.0

	scoreDetail.OriginPrefix = bestPrefix.OriginPrefix
	scoreDetail.BestPrefix = bestPrefix.SimilarPrefix

	// part 1 rank score collection between bestPrefix and originPrefix. eg. originPrefix: mu, bestPrefix:  ma
	// if the fuzziness = 0, bestPrefix = originPrefix
	// part 2 rank score collection between keyword and bestPrefix. eg. bestPrefix:ma, bestKeywords: marina.
	// if the type is not partial prefix, keyword = bestPrefix
	p_score = bestPrefix.SimiScore // combine part 1&2 into simiScore . originPrefix -> keywords
	scoreDetail.FuzzinessScore = bestPrefix.SimiScore

	// part 3 rank score collection between record and keyword, get the weight of keyword in record. IDF
	wordScore := bestPrefix.SimiScore * math.Log10(float64(len(RecordDataBase))/float64(WordCnt[bestPrefix.SimilarPrefix]))
	scoreDetail.KeywordIDFScore = wordScore
	p_score += wordScore

	// part 4 position distance = 1 / ( 1 + abs(positionDistance))
	positionScore := 1 / ( 1 + math.Abs(float64(record.Value.WordPositionMap[bestPrefix.SimilarPrefix]) - float64(originPrefixPosition)) )
	scoreDetail.PositionMatchScore = positionScore
	p_score += positionScore

	// part 5 record weight. more cited publication should be ranked higher than less cited publication.
	// recordSelectionScore = log2( 1 + recordSelectionCnt / MeanSelectionCnt )
	recordSelectionScore := math.Log2(float64(record.SelectionCnt) / float64(MeanRecordSelectinCnt) + 1)
	//subScore.RecordSelectionScore = bltha * float64(record.SelectionCnt) / float64(MaxRecordSelectinCnt)
	scoreDetail.RecordSelectionScore = recordSelectionScore
	p_score += recordSelectionScore

	return p_score,scoreDetail
}

