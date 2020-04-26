package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestTrie_AddNode(t *testing.T) {
	root := NewTrie()
	root.Root.AddNode("li","1")
	node := root.Root.FindNode("l")
	if node == nil || node.Id != "l" {
		t.Fatal("error id", node)
	}

	root.Root.AddNode("li","1")
	root.Root.AddNode("lin","2")
	root.Root.AddNode("lin","3")
	node = root.Root.FindNode("lin")
	if node == nil || node.Id != "lin" {
		t.Fatal("error id", node)
	}
	if node.value.records[0] != "2" || node.value.records[1] != "3" {
		t.Fatal("error id", node.value.records)
	}
}

func TestTrie_SingleFuzzySearch(t *testing.T) {
	poi1 := &POIDetail{POIID:"IT.001"}
	poi2 := &POIDetail{POIID:"IT.002"}
	poi3 := &POIDetail{POIID:"IT.003"}
	poi4 := &POIDetail{POIID:"IT.004"}
	poi5 := &POIDetail{POIID:"IT.005"}

	RecordDataBase["1"] = NewRecord("1",NewRecordValue([]Prefix{"li"},[]*POIDetail{poi1}),1)
	RecordDataBase["2"] = NewRecord("2",NewRecordValue([]Prefix{"lin"},[]*POIDetail{poi2}),1)
	RecordDataBase["3"] = NewRecord("3",NewRecordValue([]Prefix{"lin"},[]*POIDetail{poi3}),1)
	RecordDataBase["4"] = NewRecord("4",NewRecordValue([]Prefix{"liu"},[]*POIDetail{poi4}),1)
	RecordDataBase["5"] = NewRecord("5",NewRecordValue([]Prefix{"luis"},[]*POIDetail{poi5}),1)

	root := NewTrie()
	root.Root.AddNode("li","1")
	root.Root.AddNode("lin","2")
	root.Root.AddNode("lin","3")
	root.Root.AddNode("liu","4")
	root.Root.AddNode("luis","5")

	tests := []struct{
		msg string
		tree *Trie
		want []*POIDetail
		inputs []Prefix
		fuzziness int
	}{
		{
			msg: "input=l,fuzziness=0",
			tree: root,
			inputs: []Prefix{"l"},
			fuzziness: 0,
			want: []*POIDetail{poi1,poi2,poi3,poi4,poi5},
		},{
			msg: "input=li,fuzziness=0",
			tree: root,
			inputs: []Prefix{"li"},
			fuzziness: 0,
			want: []*POIDetail{poi1,poi2,poi3,poi4},
		},{
			msg: "input=lin,fuzziness=0",
			tree: root,
			inputs: []Prefix{"lin"},
			fuzziness: 0,
			want: []*POIDetail{poi3},
		},{
			msg: "input=luis,fuzziness=0",
			tree: root,
			inputs: []Prefix{"luis"},
			fuzziness: 0,
			want: []*POIDetail{poi5},
		},{
			msg: "input=lu,fuzziness=1",
			tree: root,
			inputs: []Prefix{"lu"},
			fuzziness: 1,
			want: []*POIDetail{poi5},
		},{
			msg: "input=lun,fuzziness=1",
			tree: root,
			inputs: []Prefix{"lin"},
			fuzziness: 1,
			want: []*POIDetail{poi1,poi2,poi3,poi4},
		},
	}

	for _,tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got := tt.tree.MultiFuzzySearchV1(tt.inputs,tt.fuzziness)
			for _, wantpoi:= range  tt.want {
				found := false
				for _, gotpoi := range got {
					if wantpoi.POIID == gotpoi.Poi.POIID {
						found = true
						break
					}
				}
				if !found {
					gotStr , _ := json.Marshal(got)
					t.Fatal("can not found in response: ",wantpoi.POIID, "got:", string(gotStr))
				}
			}
		})
	}
}

