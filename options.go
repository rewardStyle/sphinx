package sphinx

// These are all just for convenience since these are exported fields

func (q *SphinxQuery) SetFilter(f ...FilterValue) {
	q.Filters = nil
	q.Filters = append(q.Filters, f...)
}

func (q *SphinxQuery) SetFieldWeights(f ...FieldWeight) {
	q.FieldWeights = nil
	q.FieldWeights = append(q.FieldWeights, f...)
}

func (q *SphinxQuery) SetIndexWeights(i ...FieldWeight) {
	q.IndexWeights = nil
	q.IndexWeights = append(q.IndexWeights, i...)
}

func (q *SphinxQuery) SetMatchMode(m MatchMode) {
	q.MatchType = m
}

func (q *SphinxQuery) SetRankingMode(r RankMode) {
	q.RankType = r
}

func (q *SphinxQuery) SetSortMode(s SortMode) {
	q.SortType = s
}
