package dict

import (
	"math/rand"
	"sync"
	"time"
)

type SyncDict struct {
	m sync.Map
}

func (dict *SyncDict) Get(key string) (val interface{}, exists bool) {
	val, ok := dict.m.Load(key)
	return val, ok
}

func (dict *SyncDict) Len() int {
	length := 0
	dict.m.Range(func(key, value interface{}) bool {
		length++
		return true
	})
	return length
}

func (dict *SyncDict) Set(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Store(key, val)
	if existed {
		return 0
	}
	return 1
}

func (dict *SyncDict) SetIfAbsent(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		return 0
	}
	dict.m.Store(key, val)
	return 1
}

func (dict *SyncDict) SetIfExist(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

func (dict *SyncDict) Remove(key string) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		dict.m.Delete(key)
		return 1
	}
	return 0
}

func (dict *SyncDict) ForEach(consumer Consumer) {

	dict.m.Range(func(key, value interface{}) bool {
		return consumer(key.(string), value)
	})

}

func (dict *SyncDict) Keys() []string {
	keys := make([]string, dict.Len())
	dict.m.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

func (dict *SyncDict) RandomKeys(limit int) []string {
	keys := make([]string, limit)
	for i := 0; i < limit; i++ {
		dict.m.Range(func(key, value interface{}) bool {
			keys = append(keys, key.(string))
			return false
		})
	}
	return keys
}

func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	allKeys := dict.Keys()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allKeys), func(i, j int) {
		allKeys[i], allKeys[j] = allKeys[j], allKeys[i]
	})

	return allKeys[:limit]
}

func (dict *SyncDict) Clear() {
	*dict = *MakeSyncDict()
}

func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}
