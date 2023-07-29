// Package model contains models of using entities
package model

import (
	"sync"

	"github.com/google/uuid"
)

// Share is a struct for shares entity
type Share struct {
	Name  string
	Price float64
}

// SubscribersManager contains all subscribers by uuid and their subscribed shares in map subscribers
type SubscribersManager struct {
	Mu                sync.RWMutex
	SubscribersShares map[uuid.UUID]chan []*Share
	Subscribers       map[uuid.UUID][]string
}
