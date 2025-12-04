package core

import "sync"

type ReviewStored struct {
	Payload   map[string]interface{}
	ChangeNum string
	Patchset  string
}

var reviews sync.Map

func PutReview(id string, payload map[string]interface{}, changeNum, patchset string) {
	reviews.Store(id, ReviewStored{Payload: payload, ChangeNum: changeNum, Patchset: patchset})
}

func GetReview(id string) (ReviewStored, bool) {
	v, ok := reviews.Load(id)
	if !ok {
		return ReviewStored{}, false
	}
	return v.(ReviewStored), true
}
