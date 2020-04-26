package autocomplete


func deduplicateByID(nodeGroup []*TrieNode) []*TrieNode {
	if nodeGroup == nil || len(nodeGroup) == 0 {
		return []*TrieNode{}
	}
	existMap := map[Prefix]bool{}
	newNodeGroup := make([]*TrieNode,0,len(nodeGroup))

	for  _, node := range nodeGroup {
		if existMap[node.Id] {
			continue
		}
		newNodeGroup = append(newNodeGroup,node)
		existMap[node.Id] = true
	}
	return newNodeGroup
}

func updateDistMap(m map[Prefix]int, nodeID Prefix, updatedValue int) {
	if value,exist := m[nodeID]; exist {
		m[nodeID] = min(updatedValue, value )
	} else {
		m[nodeID] = updatedValue
	}
}

func min(a,b int) int {
	if a < b {
		return a
	}
	return b
}
func Max(a,b int) int {
	if a > b {
		return a
	}
	return b
}

func GetPos(ch int32) uint8 {
	return uint8(ch)
}

func simiDistance(a,b Prefix) float64 {
	dist := EditDistance(string(a), string(b), &OpCost{Delete: 1, Add: 1, Replace: 1})
	ned := float64(dist) / float64(Max(len(a),len(b)))
	return 1 - ned
}

