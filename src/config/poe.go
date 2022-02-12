package config

import (
	"fmt"
	"sort"
)

type ArchNemesisBag map[string]int

type ArchNemesisRecipe map[string][]string

func (data ArchNemesisBag) ToMapString() []string {
	rData := []string{}

	p := make(PairList, len(data))
	i := 0
	for k, v := range data {
		p[i] = Pair{k, v}
		i++
	}

	sort.Sort(p)

	for _, k := range p {
		if k.Value == 0 {
			continue
		}
		rData = append(rData, fmt.Sprintf("%-28s%3d", k.Key, k.Value))
	}
	return rData
}

func (data ArchNemesisBag) Clone() ArchNemesisBag {
	newData := make(ArchNemesisBag)
	for k, v := range data {
		newData[k] = v
	}
	return newData
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int {
	return len(p)
}

func (p PairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PairList) Less(i, j int) bool {
	return p[i].Value > p[j].Value
}
