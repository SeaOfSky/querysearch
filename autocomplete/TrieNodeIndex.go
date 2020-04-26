package autocomplete

type TrieValue struct {
	records InvertedList
}

// todo: optimize the Tire structure by using ART
// paper: The Adaptive Radix Tree: ARTful Indexing for Main-Memory Databases
type TrieNode struct {
	Id    Prefix `json:"id"`
	child [256]*TrieNode
	value *TrieValue
	count int64 // the counting of records
}

// return the nodes deeper n level than the given node.
func(s *TrieNode) LevelSearch(level int, dist map[Prefix]int, startDist int) []*TrieNode{
	if level <= 0 {
		updateDistMap(dist,s.Id,startDist)
		return []*TrieNode{s}
	}

	var res []*TrieNode
	level --
	for _,n := range s.child {
		if n != nil {
			children := n.LevelSearch(level,dist,startDist+1)
			res = append(res,children...)
		}
	}
	updateDistMap(dist,s.Id,startDist)
	res = append(res,s)
	return res
}

func (s *TrieNode) FindNode(prefix Prefix) *TrieNode {
	if len(prefix) == 0 {
		return s
	}
	cur := s
	for _,ch := range prefix {
		offset := GetPos(ch)
		next := cur.child[offset]
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}


func (s *TrieNode) AddNode(prefix Prefix, id RecordID) *TrieNode {
	if len(prefix) == 0 {
		return nil
	}
	cur := s
	for i,ch := range prefix{
		cur.count ++
		offset := GetPos(ch)
		next := cur.child[offset]
		if next == nil {
			next = new(TrieNode)
			TotalTrieNode ++
			next.Id = prefix[:i+1]
			cur.child[offset] = next
		}
		cur = next
	}
	if cur.value == nil {
		cur.value = new(TrieValue)
		cur.value.records = InvertedList{}
	}
	cur.count ++
	cur.value.records = append(cur.value.records,id)
	return cur
}


func (s *TrieNode) Remove(prefix Prefix, id RecordID) {
}
