package main

import (
	"sync"
	"time"
)

// Cache storage to store the cached data
type CacheStorage struct {
	mutex  sync.Mutex
	caches map[string]cache
}

type cache struct {
	Value any
	Exp   time.Time
}

// Creates new cache storage
//
//	NewCacheStorage()
func NewCacheStorage() CacheStorage {
	return CacheStorage{
		caches: map[string]cache{},
	}
}

func (a *CacheStorage) removeExpiredCache() {
	currentTime := time.Now()
	for name, cache := range a.caches {
		if !cache.Exp.IsZero() && cache.Exp.Before(currentTime) {
			delete(a.caches, name)
		}

	}
}

// Adds new cache data with name and expiration time
//
//	storage := NewCacheStorage()
//	storage.SetCache("tempData", []int{65, 32, 12, 93}, 15 * time.Minute)
func (a *CacheStorage) SetCache(name string, value any, exp time.Duration) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.removeExpiredCache()
	a.caches[name] = cache{
		Value: value,
		Exp:   time.Now().Add(exp),
	}
}

// Adds new cache data with a name
//
//	storage := NewCacheStorage()
//	storage.SetCacheForever("myData", []int{65, 32, 12, 93})
func (a *CacheStorage) SetCacheForever(name string, value any) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.removeExpiredCache()
	a.caches[name] = cache{
		Value: value,
		Exp:   time.Time{},
	}
}

// Deletes cached data by name
//
//	storage := NewCacheStorage()
//	storage.SetCacheForever("myData", []int{65, 32, 12, 93})
//	storage.RemoveCache("myData")
func (a *CacheStorage) RemoveCache(name string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.removeExpiredCache()
	delete(a.caches, name)
}

// Gets cache data by name
//
//	storage := NewCacheStorage()
//	storage.SetCacheForever("myData", []int{65, 32, 12, 93})
//	fmt.Println(storage.GetCache("myData")) // [65 32 12 93] true
//	storage.RemoveCache("myData")
//	fmt.Println(storage.GetCache("myData")) // <nil> false
func (a *CacheStorage) GetCache(name string) (any, bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.removeExpiredCache()
	v, ok := a.caches[name]
	if !ok {
		return nil, false
	}

	return v.Value, true
}

// Gets cache data by name or execute a function and return a value
//
//	 func defVal() any {
//			return "Hello"
//	 }
//
//	 storage := NewCacheStorage()
//	 storage.SetCacheForever("myData", []int{65, 32, 12, 93})
//	 fmt.Println(storage.GetCacheOr("myData", defVal)) // [65 32 12 93]
//	 storage.RemoveCache("myData")
//	 fmt.Println(storage.GetCacheOr("myData", defVal)) // Hello
func (a *CacheStorage) GetCacheOr(name string, defVal func() any) any {
	v, ok := a.GetCache(name)
	if ok {
		return v
	}

	return defVal()
}

// Deletes all cache data
//
//	storage := NewCacheStorage()
//	storage.SetCacheForever("myData", []int{65, 32, 12, 93})
//	storage.SetCacheForever("myData2", "Hello")
//	fmt.Println(len(storage.CacheList())) // 2
//	storage.ClearAll()
//	fmt.Println(len(storage.CacheList())) // 0
func (a *CacheStorage) ClearAll() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for k := range a.caches {
		delete(a.caches, k)
	}
}

// Finds out if there is cache data with a name
//
//	storage := NewCacheStorage()
//	fmt.Println(storage.Has("myData")) // false
//	storage.SetCacheForever("myData", []int{65, 32, 12, 93})
//	fmt.Println(storage.Has("myData")) // true
func (a *CacheStorage) Has(name string) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.removeExpiredCache()
	for k := range a.caches {
		if k == name {
			return true
		}
	}

	return false
}

// Returns a list with the names of cached data
//
//	storage := NewCacheStorage()
//	storage.SetCacheForever("myData", []int{65, 32, 12, 93})
//	fmt.Println(storage.CacheList()) // [myData]
//	storage.SetCacheForever("myData2", "Hello")
//	fmt.Println(storage.CacheList()) // [myData myData2]
func (a *CacheStorage) CacheList() []string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.removeExpiredCache()
	keys := make([]string, 0, len(a.caches))
	for k := range a.caches {
		keys = append(keys, k)
	}

	return keys
}
