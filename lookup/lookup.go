package lookup

import (
	"encoding/json"
	"github.com/querysearch/autocomplete"
	"net/http"
	"sort"
)

var LookUpDataBase = map[string]*POIProfle{}

type POIProfle struct {
	POIID string
	POIName string
	TotalCount int64
	TotalQuery int64
	Querys map[autocomplete.RecordID]*POIQueryConfidence
}

type POIQueryConfidence struct {
	Query autocomplete.RecordID
	QueryCount int64
	POICount int64
	Confidence float64
	InvertedConfidence float64
}

type LookUpResponse struct {
	POIID string
	POIName string
	TotalCount int64
	TotalQuery int64
	Querys []*POIQueryConfidence
}

// RecordSorter ...
type sorter []*POIQueryConfidence

func (s sorter) Len() int {
	return len(s)
}

func (s sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sorter) Less(i, j int) bool {
	return s[i].InvertedConfidence > s[j].InvertedConfidence
}


func BuildLookUpDataBase() {
	for rID,rValue := range autocomplete.RecordDataBase{
		queryCnt := rValue.SelectionCnt
		for _,poi := range rValue.Value.Pois {
			if val, exist := LookUpDataBase[poi.POIID]; exist {
				val.Querys[rID] = &POIQueryConfidence{
					Query: rID,
					QueryCount: queryCnt,
					Confidence: poi.Confidence,
					POICount: int64( float64(queryCnt) * poi.Confidence),
				}
			}else {
				LookUpDataBase[poi.POIID] = &POIProfle{
					POIID: poi.POIID,
					POIName: poi.POIName,
					Querys: map[autocomplete.RecordID]*POIQueryConfidence{
						rID: {
							Query: rID,
							QueryCount: queryCnt,
							Confidence: poi.Confidence,
							POICount: int64( float64(queryCnt) * poi.Confidence),
						},
					},
				}
			}
		}
	}

	for _, val := range LookUpDataBase {
		var sum int64 = 0
		for _, query := range val.Querys {
			sum += query.POICount
		}
		val.TotalQuery = int64(len(val.Querys))
		val.TotalCount = sum
		for _,query := range val.Querys {
			query.InvertedConfidence = float64(query.POICount) / float64(sum)
		}
	}
}


func Lookup(w http.ResponseWriter, r *http.Request){
	args := r.URL.Query()
	texts := args["id"]
	if len(texts) == 0 {
		_, _ = w.Write([]byte("id is empty!"))
		return
	}
	poiID := texts[0]
	results := LookUpDataBase[poiID]

	response := LookUpResponse{}
	response.POIID = results.POIID
	response.TotalQuery = results.TotalQuery
	response.POIName = results.POIName
	response.TotalCount = results.TotalCount
	for _,query := range results.Querys {
		response.Querys = append(response.Querys,query)
	}

	sort.Sort(sorter(response.Querys))
	byteResult, _ := json.Marshal(response)
	w.Header().Add("Content-Type","application/json; charset=utf-8")
	w.Write(byteResult)
}
