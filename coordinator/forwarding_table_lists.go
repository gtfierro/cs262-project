package main

import (
	log "github.com/Sirupsen/logrus"
	"sync"
)

// TODO the locks on queryList shouldn't be necessary
// since I'm using map-wide locks already anyway
type queryList struct {
	sync.RWMutex
	queries       []*ForwardedQuery
	nonnilInArray int
}

func (ql *queryList) containsQuery(q *ForwardedQuery) bool {
	ql.RLock()
	defer ql.RUnlock()
	for _, q2 := range ql.queries {
		if q2 == nil {
			return false // All nil entries come after all valid ones
		} else if q == q2 {
			return true
		}
	}
	return false
}

func (ql *queryList) addQuery(q *ForwardedQuery) {
	ql.Lock()
	defer ql.Unlock()
	if ql.nonnilInArray < len(ql.queries) {
		ql.queries[ql.nonnilInArray] = q
		ql.nonnilInArray += 1
	} else {
		ql.queries = append(ql.queries, q)
	}
}

func (ql *queryList) removeQuery(q *ForwardedQuery) {
	ql.Lock()
	defer ql.Unlock()
	if ql.nonnilInArray == 1 {
		// last remaining query
		if ql.queries[0] == q {
			ql.queries[0] = nil
			return
		}
	} else {
		// TODO should also probably resize down here if size falls below some point
		for qIdx, current := range ql.queries {
			if current == q {
				ql.nonnilInArray -= 1
				if qIdx != ql.nonnilInArray {
					ql.queries[qIdx] = ql.queries[ql.nonnilInArray]
				}
				ql.queries[ql.nonnilInArray] = nil
			}
		}
	}
	log.WithFields(log.Fields{
		"query": q, "queryList": ql,
	}).Fatal("Attempted to remove query from queryList that doesn't contain it")
}
