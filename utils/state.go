package utils

import (
	"sync"
	"time"
)

type UserStateType struct {
	sync.Mutex
	Pending map[string]time.Time
}

var UserState = UserStateType{
	Pending: make(map[string]time.Time),
}

func (us *UserStateType) AddUser(senderJID string) {
	us.Lock()
	us.Pending[senderJID] = time.Now()
	us.Unlock()
}

func (us *UserStateType) ClearUser(senderJID string) {
	us.Lock()
	delete(us.Pending, senderJID)
	us.Unlock()
}

func (us *UserStateType) CheckUser(senderJID string) (time.Time, bool) {
	us.Lock()
	defer us.Unlock()
	time, exists := us.Pending[senderJID]
	return time, exists
}