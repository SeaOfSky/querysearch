package autocomplete

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		Root: &TrieNode{},
	}
}

func(t *Trie) SingleExactSearch(prefix Prefix) []*TrieNode {
	if len(prefix) == 0 {
		return []*TrieNode{}
	}
	cur := t.Root.FindNode(prefix)
	if cur == nil {
		return []*TrieNode{}
	}
	return []*TrieNode{cur}
}

// MultiFuzzySearchV1 : do materialize the inverted list and computing the union of shortest union list
// all prefix are regarded as complete words except the last one.
func (t *Trie) MultiFuzzySearchV1(prefixKeywords []Prefix,fuzziness int) []*POIWithScore {
	fmt.Println("start search", prefixKeywords)
	if len(prefixKeywords) == 0 {
		prefixKeywords = []Prefix{""}
	}
	prefixNum := len(prefixKeywords)
	activeNodes := make([][]*TrieNode,prefixNum)
	recordCount := make([]int64,prefixNum)

	for i,prefix := range prefixKeywords {
		if i == prefixNum - 1  { // only last prefix will be partial match.
			fmt.Println("partial prefix search key:", prefix,"fuzzy:",fuzziness)
			fuzzinessActiveNodes := t.SingleFuzzySearch(prefix,PartialPrefixFuzzyProtectN,fuzziness)
			activeNodes[i],recordCount[i] = t.FetchLeafNodes(fuzzinessActiveNodes,true)
		}else{
			fmt.Println("prefix word search key:", prefix,"fuzzy:",fuzziness)
			fuzzinessActiveNodes := t.SingleFuzzySearch(prefix,KeywordPrefixFuzzyProtectN,fuzziness)
			activeNodes[i],recordCount[i] = t.FetchLeafNodes(fuzzinessActiveNodes,false)
		}
	}

	// select the init node
	miniOne := recordCount[0]
	miniIdx := 0
	for i := 1 ; i <prefixNum; i ++ {
		if miniOne > recordCount[i] {
			miniIdx = i
			miniOne = recordCount[i]
		}
	}
	candidateNodes := activeNodes[miniIdx]

	// map[prefix]similarPrfix
	prefixGroup := make([]map[Prefix]*PrefixCouple,prefixNum)
	for i,prefix := range prefixKeywords {
		if i == prefixNum - 1{
			prefixGroup[i]= t.GetSimilarPrefix(prefix,activeNodes[i],true)
		}else{
			prefixGroup[i]= t.GetSimilarPrefix(prefix,activeNodes[i],false)
		}
		fmt.Println(prefix,"Get similar prefix:",len(prefixGroup[i]))
	}

	str, _ := json.Marshal(candidateNodes)
	fmt.Println("candidate Nodes:",len(candidateNodes),string(str))

	// rank the record id and get the final poi id
	results := t.Rank(candidateNodes,prefixGroup,3,10)

	return results
}

func (t *Trie) GetSimilarPrefix(originPrefix Prefix,activeNodes []*TrieNode,isPartialMatch bool) map[Prefix]*PrefixCouple {
	similarPrefix := make(map[Prefix]*PrefixCouple,len(activeNodes))
	for _,node := range activeNodes {
		similarPrefix[node.Id] = &PrefixCouple{
				OriginPrefix:originPrefix,
				SimilarPrefix:node.Id,
				PartialMatch: isPartialMatch,
				SimiScore: simiDistance(originPrefix,node.Id)}
	}
	return similarPrefix
}

// todo: early terminate the record traverse if the user only type few letters. it can cause much unnecessary load.
func (t *Trie) Rank(activeNodes []*TrieNode,prefixGroup []map[Prefix]*PrefixCouple,RECORD_LIMIT_N, POI_LIMIT_N int) []*POIWithScore {

	recordScoreDetail := map[RecordID]*RecordScore{}

	// filter and calculate score
	for _, node := range activeNodes {
		exist := map[RecordID]bool{}
		for _,recordID := range node.value.records {
			if exist[recordID] {
				//fmt.Println("duplicate record:",recordID)
			}else{
				exist[recordID] = true
			}
			record := RecordDataBase[recordID]
			if isAccepted, score := t.isAcceptRecord(record,prefixGroup); isAccepted {
				recordScoreDetail[record.Id] = score
			}
		}
	}

	// compute the score of POI
	//sort.Sort(RecordSorter(recordScore))
	//records := recordScore[:min(RECORD_LIMIT_N,len(recordScore))]
	//records := recordScore
	//fmt.Println("get records: ",len(records),records)



	uniquePOI := map[string]*POIWithScore{}
	for recordID,recordScore := range recordScoreDetail {
		recordDetail := RecordDataBase[recordID]
		for _,poi := range recordDetail.Value.Pois {
			newScore := ComputeFinalScore(recordScore,poi)
			if existPOIScore,ok := uniquePOI[poi.POIID]; ok {
				if recordScore.PriorityLevel > existPOIScore.PriorityLevel {
					uniquePOI[poi.POIID] = NewPOIWithScore(recordID,recordScore,poi)
				} else if recordScore.PriorityLevel == existPOIScore.PriorityLevel && newScore > existPOIScore.Score {
					uniquePOI[poi.POIID] = NewPOIWithScore(recordID,recordScore,poi)
				}else {
					//fmt.Println("remove duplicate poi", poi.POIID , " in record:", r.id)
				}
			}else{
				uniquePOI[poi.POIID] = NewPOIWithScore(recordID,recordScore,poi)
			}
		}
	}


	poiScoreArray := []*POIWithScore{}
	for _, poiScore := range uniquePOI {
		poiScoreArray = append(poiScoreArray, poiScore)
	}

	sort.Sort(POISorter(poiScoreArray))
	//poiScore = poiScore[:min(POI_LIMIT_N,len(poiScore))]

	return poiScoreArray
}

func (t *Trie) isAcceptRecord(record *Record, prefixGroup []map[Prefix]*PrefixCouple ) (isAccepted bool,score *RecordScore) {
	score = &RecordScore{}


	var Rule1_PrefixPositionMatch bool = true
	var Rule2_FuzzyPrefixPositonMatch  bool = true
	rule3 := NewRuler3()
	rule4 := NewRuler4()

	// todo: for now , only consider shortest list of records.  but we might need to take the non-shortest list of records into consider.
	for prefixPoisition, similarPrefixMap  := range prefixGroup {
		var bestPrefix *PrefixCouple
		var bestScore float64
		var meetPrefix bool

		// find bestPrefix
		for _,word := range record.Value.Words {
			if prefixCp, exist := similarPrefixMap[word]; exist {
				if !meetPrefix {
					meetPrefix = true
					bestScore = prefixCp.SimiScore
					bestPrefix = prefixCp
				}else if bestScore < prefixCp.SimiScore{
					bestScore = prefixCp.SimiScore
					bestPrefix = prefixCp
				}
			}
		}

		// does not meet any prefix
		if bestPrefix == nil {
			//prefixlog,_ := json.Marshal(similarPrefixes)
			//fmt.Println("Error: bestPrefix is nil. ","record:",record.Id,",prefix:",string(prefixlog))
			return false,nil
		}

		{
			if Rule1_PrefixPositionMatch {
				Rule1_PrefixPositionMatch = IsMatchRule1(record,bestPrefix,prefixPoisition)
			}

			if Rule2_FuzzyPrefixPositonMatch {
				Rule2_FuzzyPrefixPositonMatch = IsMatchRule2(record,bestPrefix,prefixPoisition)
			}
			rule3.CheckRule(record,bestPrefix,prefixPoisition)
			rule4.CheckRule(record,bestPrefix,prefixPoisition)
		}

		// todo: optimize the score computing
		p_score, subScore := ComputePrefixMatchScore(prefixPoisition,bestPrefix,record)
		score.TotalScore += p_score
		score.SubScore = append(score.SubScore,subScore)
	}

	// set priorityLevel according to rules
	if rule4.IsMatch() {
		score.PriorityLevel = Max(score.PriorityLevel,1)
	}
	if rule3.IsMatch() {
		score.PriorityLevel = Max(score.PriorityLevel,2)
	}
	if Rule2_FuzzyPrefixPositonMatch {
		score.PriorityLevel = Max(score.PriorityLevel,3)
	}
	if Rule1_PrefixPositionMatch {
		score.PriorityLevel = Max(score.PriorityLevel,4)
	}

	// only for demo, only retain Priority > 0
	if score.PriorityLevel == 0 {
		return false,nil
	}

	return true,score
}

//isPartialPrefix is true means the prefix is partial prefix a word .
//false means the prefix is complete words.
// param prefixProtectN: for example, prefixProtectN = 1, so the first letter can not be fuzzy.
func (t *Trie) SingleFuzzySearch(prefix Prefix, prefixProtectN int,fuzziness int) (activeNodes []*TrieNode) {
	if fuzziness == 0 || len(prefix) <= prefixProtectN {
		activeNodes = t.SingleExactSearch(prefix)
		str, _ := json.Marshal(activeNodes)
		fmt.Println("prefix:",prefix,", get SingleExactSearch:", string(str) )
		return activeNodes
	}

	startNode := t.Root.FindNode(prefix[:prefixProtectN])
	fuzzyPrefixPart := prefix[prefixProtectN:]

	// read nodes whose edit distance less than fuzziness
	editDistMap := map[Prefix]int{}
	if exist, initNodes, err := t.RestoreFromPaxCache(); err == nil && exist {
		activeNodes = initNodes
	}else {
		activeNodes = startNode.LevelSearch(fuzziness,editDistMap,0)
	}

	length := len(prefix)
	if length == 0 {
		return activeNodes
	}
	// fuzzy search
	for _,ch := range fuzzyPrefixPart{
		var nextActiveNodes []*TrieNode
		newEditDistMap := map[Prefix]int{}

		// one loop
		offset := GetPos(ch)

		for _,node := range activeNodes {
			// for node itself, delete is the only way to stay active
			curNodeDist := editDistMap[node.Id]

			if curNodeDist + 1 <= fuzziness {
				nextActiveNodes = append(nextActiveNodes,node)
				updateDistMap(newEditDistMap,node.Id,curNodeDist + 1)
			}

			// check matching child
			if node.child[offset] != nil {
				child := node.child[offset]
				// check edit distance
				if curNodeDist <= fuzziness {
					tmpActiveNodes := child.LevelSearch(fuzziness - curNodeDist,newEditDistMap,curNodeDist)
					nextActiveNodes = append(nextActiveNodes,tmpActiveNodes...)
				}
			}

			if curNodeDist + 1 > fuzziness {
				// no need to check other child nodes because of already reach the limitation of fuzziness
				continue
			}

			// check all children of the node
			for idx,child := range node.child {
				if child == nil || uint8(idx) == offset {
					continue
				}
				// substitute
				updateDistMap(newEditDistMap,child.Id,curNodeDist + 1)
			}
		}

		nextActiveNodes = deduplicateByID(nextActiveNodes)
		activeNodes = nextActiveNodes
		editDistMap = newEditDistMap
		//fmt.Println(prefix[:i+1],"edit map:",editDistMap)
	}

	str, _ := json.Marshal(activeNodes)
	fmt.Println("get active nodes:", string(str))

	return activeNodes
}

// collect all record from nodes and sort by length
// isRecursive control whether add the descendant of nodes, default is false.
func (t *Trie) FetchLeafNodes(nodes []*TrieNode, isPartialMatch bool) (res []*TrieNode,recordCnt int64) {
	exist := map[Prefix]bool{}
	res = make([]*TrieNode,0,len(nodes))
	for len(nodes) > 0 {
		n := nodes[len(nodes) -1]
		nodes = nodes[:len(nodes) -1]

		if n == nil {
			continue
		}

		if isPartialMatch {
			for _,child := range  n.child {
				if child != nil {
					nodes = append(nodes, child)
				}
			}
		}

		if n.value != nil && n.value.records != nil && !exist[n.Id] {
			exist[n.Id] = true
			res = append(res,n)
			recordCnt += n.count
		}
	}
	return res,recordCnt
}

func (t *Trie) RestoreFromPaxCache() (exist bool, activeNode []*TrieNode, err error) {
	return false, nil, nil
}


// Rule 1 : the best prefix must not have any fuzzy meaning.
// 1. the position of each word of the query must be matched with the ones of the record.
// 2. expect the last partial prefix, other words must be exactly the same.
// 3. the last partial prefix must have the same prefix with bestPrefix.
func IsMatchRule1(record *Record,bestPrefix *PrefixCouple,prefixPoisition int) bool {
	if record.Value.WordPositionMap[bestPrefix.SimilarPrefix] != prefixPoisition {
		return  false
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

// Rule 2 : the best prefix can have fuzzy word.
// 1. the position of each word of the query must be matched with the ones of the record.
func IsMatchRule2(record *Record,bestPrefix *PrefixCouple,prefixPoisition int) bool{
	if record.Value.WordPositionMap[bestPrefix.SimilarPrefix] != prefixPoisition {
		return  false
	}
	return true
}