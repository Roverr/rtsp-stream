package blacklist

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// IList describes the user panel of a Blacklist
type IList interface {
	AddOrIncrease(uri string) IList
	IsBanned(uri string) bool
	Remove(uri string) IList
}

// List implements IList
type List struct {
	list      *sync.Map
	freeAfter time.Duration
	limit     int
}

// Type check
var _ IList = (*List)(nil)

// NewList creates a new List instance
func NewList(freeAfter time.Duration, limit int) *List {
	return &List{&sync.Map{}, freeAfter, limit}
}

// AddOrIncrease stores the URI as a new record or increases the inner counter
func (b *List) AddOrIncrease(uri string) IList {
	if b == nil {
		return nil
	}
	record, ok := b.list.Load(uri)
	if !ok {
		b.list.Store(uri, NewRecord(uri))
		return b
	}
	if record.(IRecord).IsBanned() {
		logrus.Debugf(
			"%s is still banned until %s | Blacklist",
			uri,
			record.(IRecord).GetBanTime().Format(time.RFC3339),
		)
		return b
	}
	if record.(IRecord).Increase().GetCount() > b.limit {
		logrus.Infof("%s is banned beacuse of reaching limit | Blacklist", uri)
		ban := time.Now().Add(b.freeAfter)
		record.(IRecord).SetBanTime(&ban).ResetCount()
	}
	logrus.Debugf(
		"%s is now with %d of %d on the blacklist | Blacklist",
		uri,
		record.(IRecord).GetCount(),
		b.limit,
	)
	b.list.Store(uri, record)
	return b
}

// IsBanned returns if the given record is banned or not
func (b *List) IsBanned(uri string) bool {
	if b == nil {
		return false
	}
	if record, ok := b.list.Load(uri); ok {
		return record.(IRecord).IsBanned()
	}
	return false
}

// Remove removes a given URI from the blacklist
func (b *List) Remove(uri string) IList {
	if b == nil {
		return nil
	}
	b.list.Delete(uri)
	return b
}
