package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/querysearch/autocomplete"
	"github.com/querysearch/lookup"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
)

//const FilePath = "/home/wilson.muktar/data/entity_combine_root_selection.csv"
const FilePath = "./data/entity_combine_root_selection.csv"

type SearchResult struct {
	TotalNumber int
	Version string
	Result []*autocomplete.POIWithScore
}

const AppVersionNumber = "0.01"

var Searcher = autocomplete.NewTrie()

func init() {
	// read data into memory
	buildIndex(FilePath)
	fmt.Println("build file=",FilePath," success.")
	fmt.Println("record number:",len(autocomplete.RecordDataBase))
	fmt.Println("total Trie node", autocomplete.TotalTrieNode)
	fmt.Println("mean record count",autocomplete.MeanRecordSelectinCnt)
}

// build a tire
func buildIndex(filepath string)  {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	var totalRecordCount int64 = 0

	var inValidRecord int64 = 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err,record)
			continue
		}

		if len(record) < 6 {
			fmt.Println("Error: wrong length of record",record)
			continue
		}

		//fmt.Println(record,len(record)) // record has the type []string
		query := record[0]
		poiID := record[1]
		poiName := record[2]
		poiCnt := record[3]
		queryCnt := record[4]
		confidence := record[5]
		accepted, prefixWords := isQueryAccepted(query)

		// add into recordDataBase
		if accepted {
			confidence := confidence[:len(confidence)-1] // remove %
			con,err := strconv.ParseFloat(confidence,64)
			queryCnt,_ := strconv.ParseInt(queryCnt,10,64)
			con /= 100
			if err != nil {
				fmt.Println(err)
			}

			cnt,err := strconv.ParseInt(poiCnt,10,64)
			if err != nil {
				fmt.Println(err)
			}
			detail := &autocomplete.POIDetail{POIID: poiID, POIName: poiName, Confidence: con, SelectionCnt: cnt}
			totalRecordCount += queryCnt
			if recordDB, exist := autocomplete.RecordDataBase[autocomplete.RecordID(query)]; exist {
				recordDB.Value.Pois = append(recordDB.Value.Pois,detail)
			} else {
				newRecordVal := autocomplete.NewRecordValue(prefixWords,[]*autocomplete.POIDetail{detail})
				autocomplete.RecordDataBase[autocomplete.RecordID(query)] = autocomplete.NewRecord(autocomplete.RecordID(query),newRecordVal,queryCnt)
			}
		}else{
			//fmt.Println("fail:",record,len(record))
			inValidRecord ++
		}
	}

	fmt.Println("Filter out invalid record (non-english-letter) : ",inValidRecord)


	autocomplete.MeanRecordSelectinCnt = totalRecordCount / int64(len(autocomplete.RecordDataBase))

	// add record into Trie
	for rID := range autocomplete.RecordDataBase{
		_, prefixWords := isQueryAccepted(string(rID))
		for _, keyword := range prefixWords {
			autocomplete.WordCnt[keyword] ++
			Searcher.Root.AddNode(keyword,rID)
		}
	}

	lookup.BuildLookUpDataBase()
}

func isQueryAccepted(query string) (accepted bool,prefixes []autocomplete.Prefix) {
	multiKeywords := strings.Split(strings.Trim(query," .,()?/`~!@#$%^&*-+_=[]{}'\"<>")," ")
	for _,keyword := range multiKeywords {
		tmp := strings.Trim(keyword," .,()?/`~!@#$%^&*-+_=[]{}'\"<>")
		if prefix,err := autocomplete.NewPrefix(tmp); err != nil {
			return false, nil
		}else{
			prefixes = append(prefixes, prefix)
		}
	}
	return true,prefixes
}


func fuzzySearch(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			_, _ = w.Write([]byte(r.(error).Error()))
			return
		}
	}()
	var isSkipGoogle bool
	var fuzziness int64
	var err error
	args := r.URL.Query()

	texts := args["s"]
	if len(texts) == 0 {
		_, _ = w.Write([]byte("text is empty!"))
		return
	}
	searchText := texts[0]

	if len(args["f"]) == 0 {
		fuzziness = 0
	}else {
		fuzziness,err = strconv.ParseInt(args["f"][0],10,32)
		if err != nil {
			fuzziness = 0
		}
	}

	if len(args["skip"]) == 0 {
		isSkipGoogle = false
	}else {
		isSkipGoogle,err = strconv.ParseBool(args["skip"][0])
		if err != nil {
			isSkipGoogle = false
		}
	}

	multiKeywords := strings.Split(strings.Trim(searchText," ")," ")
	if len(multiKeywords) == 0 {
		multiKeywords = []string{""}
	}

	SearchPrefix := []autocomplete.Prefix{}
	for _,keyword := range multiKeywords {
		tmp := strings.Trim(keyword," ")
		if prefix,err := autocomplete.NewPrefix(tmp); err != nil {
			// return error
			w.WriteHeader(200)
			w.Write([]byte(err.Error()))
			return
		}else if len(prefix) > 0{
			SearchPrefix = append(SearchPrefix, prefix)
		}
	}

	results := Searcher.MultiFuzzySearchV1(SearchPrefix, int(fuzziness))
	if isSkipGoogle {
		fmt.Println("skip google point")
		results = FilterOutGooglePoint(results)
	}

	length := len(results)
	results = results[:autocomplete.MinInts(50, len(results))]
	response := SearchResult{TotalNumber: length, Version: AppVersionNumber, Result: results}

	byteResult, _ := json.Marshal(response)
	w.Header().Add("Content-Type","application/json; charset=utf-8")
	w.Write(byteResult)
}

func FilterOutGooglePoint(results []*autocomplete.POIWithScore) []*autocomplete.POIWithScore {
	newResults := make([]*autocomplete.POIWithScore,0,len(results))
	for _,poi := range results {
		if poi.Poi.POIID[:3] == "GA." {
			continue
		}
		newResults = append(newResults,poi)
	}
	return newResults
}

func main(){
	http.HandleFunc("/fuzzysearch", fuzzySearch)
	http.HandleFunc("/lookup", lookup.Lookup)
	err := http.ListenAndServe("127.0.0.1:8199", nil)
	if err != nil {
		fmt.Println("Error:",err)
	}
}