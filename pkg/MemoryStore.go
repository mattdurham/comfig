package pkg

type MemoryStore struct {
	Cache      map[string]string
	MapCache   map[string]map[string]string
	ArrayCache map[string][]string
}

func NewMemoryStore() *MemoryStore {
	ms := &MemoryStore{
		Cache:      make(map[string]string),
		MapCache:   make(map[string]map[string]string),
		ArrayCache: make(map[string][]string),
	}
	return ms
}

func (m *MemoryStore) GetMap(key string) (map[string]string, bool) {
	v, exists := m.MapCache[key]
	return v, exists
}

func (m *MemoryStore) GetArray(key string) ([]string, bool) {
	v, exists := m.ArrayCache[key]
	return v, exists
}

func (m *MemoryStore) Get(key string) (string, bool) {
	v, exists := m.Cache[key]
	return v, exists
}
