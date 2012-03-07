package tmpl

import "sync"

type fileLock struct {
	lks map[string]*sync.Mutex
	lk  sync.RWMutex
}

func newFileLock() *fileLock {
	return &fileLock{
		lks: map[string]*sync.Mutex{},
	}
}

func (f *fileLock) Lock(key string) {
	f.lk.Lock()
	defer f.lk.Unlock()

	if lk, ex := f.lks[key]; ex {
		lk.Lock()
		return
	}

	f.lks[key] = new(sync.Mutex)
	f.lks[key].Lock()
}

func (f *fileLock) Unlock(key string) {
	f.lk.RLock()
	defer f.lk.RUnlock()

	f.lks[key].Unlock()
}
