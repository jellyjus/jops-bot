package usecase

import (
	"sync"

	"jops-bot/entity"
)

type Store struct {
	mu       sync.Mutex
	sessions map[int64]*entity.Session
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[int64]*entity.Session),
	}
}

func (st *Store) GetOrCreate(chatID int64) *entity.Session {
	st.mu.Lock()
	defer st.mu.Unlock()
	s, ok := st.sessions[chatID]
	if !ok {
		s = newSession()
		st.sessions[chatID] = s
	}
	return s
}

func (st *Store) Reset(chatID int64) *entity.Session {
	st.mu.Lock()
	defer st.mu.Unlock()
	s := newSession()
	st.sessions[chatID] = s
	return s
}
