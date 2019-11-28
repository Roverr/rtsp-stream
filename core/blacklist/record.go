package blacklist

import (
	"time"
)

// IRecord describes the user panel of BlacklistRecord
type IRecord interface {
	Increase() IRecord
	GetCount() int
	ResetCount() IRecord
	SetBanTime(*time.Time) IRecord
	GetBanTime() *time.Time
	IsBanned() bool
}

// Record is an element of the blacklist
type Record struct {
	uri         string
	count       int
	bannedUntil *time.Time
}

// Type check
var _ IRecord = (*Record)(nil)

// NewRecord creates a new instance of Record
func NewRecord(uri string) *Record {
	return &Record{uri, 0, nil}
}

// Increase increases the inner counter for the record
func (r *Record) Increase() IRecord {
	if r == nil {
		return nil
	}
	r.count++
	return r
}

// GetCount returns the count of the record
func (r *Record) GetCount() int {
	if r == nil {
		return 0
	}
	return r.count
}

// SetBanTime sets the record banned for a given time
func (r *Record) SetBanTime(t *time.Time) IRecord {
	if r == nil {
		return nil
	}
	r.bannedUntil = t
	return r
}

// ResetCount resets the inner count
func (r *Record) ResetCount() IRecord {
	if r == nil {
		return nil
	}
	r.count = 0
	return r
}

// GetBanTime returns bannedUntil property
func (r *Record) GetBanTime() *time.Time {
	if r == nil {
		return nil
	}
	return r.bannedUntil
}

// IsBanned is to determine if the record is currently banned or not
func (r *Record) IsBanned() bool {
	if r == nil {
		return false
	}
	if r.bannedUntil == nil {
		return false
	}
	return r.bannedUntil.After(time.Now())
}
