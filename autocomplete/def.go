package autocomplete

import "errors"


const (
	PartialPrefixFuzzyProtectN = 2 // protect 3 beginning letters from being fuzzy
	KeywordPrefixFuzzyProtectN = 2 // protect 3 beginning letters from being fuzzy
)

//global object
var RecordDataBase = RecordMap{}
var MeanRecordSelectinCnt int64 = 1
var TotalTrieNode int64
var WordCnt = map[Prefix] int64{}

// definition
type Prefix string

func NewPrefix(str string) (Prefix,error) {
	for _,ch := range str {
		if ch > 256 {
			return "",errors.New("not in asical code:" + str)
		}
	}
	return Prefix(str),nil
}

type RecordID string
type InvertedList []RecordID

// todo: because the huge amount of poi and record we need to store, which can bring the overhead of gc.
// free cache: https://github.com/coocood/freecache
type RecordMap map[RecordID]*Record

type Record struct {
	Id           RecordID
	Value        *RecordValue
	SelectionCnt int64 // to compute the score(record,keyword)
}

func NewRecord(id RecordID,val *RecordValue, cnt int64) *Record {
	return &Record{
		Id:           id,
		Value:        val,
		SelectionCnt: cnt,
	}
}

type RecordValue struct {
	Words           []Prefix
	WordPositionMap map[Prefix]int
	Pois            []*POIDetail // poi id list
}
func NewRecordValue(prefixes []Prefix, pois []*POIDetail) *RecordValue{
	wordsMap := map[Prefix]int{}
	for i,p := range prefixes {
		wordsMap[p] = i
	}
	return &RecordValue{
		Words: prefixes,
		WordPositionMap:wordsMap,
		Pois: pois,
	}
}



type POIDetail struct {
	POIID string
	POIName string
	Confidence float64 // to compute teh score(record,POI)
	SelectionCnt int64
}

type PrefixCouple struct {
	OriginPrefix Prefix
	SimilarPrefix Prefix
	PartialMatch bool
	SimiScore float64
}

type RecordWithScore struct {
	score float64
	id RecordID
}

// RecordSorter ...
type RecordSorter []*RecordWithScore

func (s RecordSorter) Len() int {
	return len(s)
}

func (s RecordSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s RecordSorter) Less(i, j int) bool {
	return s[i].score > s[j].score
}


type POIWithScore struct {
	RecordID RecordID
	Poi *POIDetail // poi id
	PriorityLevel int
	Score float64 // RecordScore * POIScore
	POIScore float64 // poi confidence
	RecordScore float64
	RecordScoreDetail *RecordScore
}

func NewPOIWithScore(recordID RecordID,recordScore *RecordScore, poi *POIDetail) *POIWithScore {
	recordFinalScore := recordScore.TotalScore * poi.Confidence
	poiScore :=  ComputePOIScore(poi)
	return &POIWithScore{
		Score: recordFinalScore + poiScore,
		Poi: poi,
		POIScore: poiScore,
		RecordScore: recordFinalScore,
		RecordScoreDetail: recordScore,
		RecordID: recordID,
		PriorityLevel: recordScore.PriorityLevel,
	}
}

// POISorter ...
type POISorter []*POIWithScore

func (s POISorter) Len() int {
	return len(s)
}

func (s POISorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s POISorter) Less(i, j int) bool {
	if s[i].PriorityLevel != s[j].PriorityLevel {
		return s[i].PriorityLevel > s[j].PriorityLevel
	}

	if s[i].Score == s[j].Score {
		// to make the order of result stable
		return len(s[i].Poi.POIName) > len(s[j].Poi.POIName)
	}

	return s[i].Score > s[j].Score
}


type RecordScore struct {
	TotalScore float64
	PriorityLevel int `json:"-"`
	SubScore []*SubRecordScore
}

type SubRecordScore struct {
	OriginPrefix Prefix
	BestPrefix Prefix
	// orgin prefix to keyword : fuzziness value. to select more precise one.
	FuzzinessScore float64
	// keyword IDF: the one with higher IDF can be more important.
	KeywordIDFScore float64
	// Position Distance.
	PositionMatchScore float64
	// record selection count
	RecordSelectionScore float64
}

type SearchLog struct {
	originPrefix Prefix
	fuzzyPrefix Prefix
	PartialPrefix Prefix
}