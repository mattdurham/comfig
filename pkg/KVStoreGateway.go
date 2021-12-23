package pkg

type Store interface {
	Get(string) (string, bool)
	GetMap(string) (map[string]string, bool)
	GetArray(string) ([]string, bool)
}

type KVStoreGateway struct {
	stores []Store
}

func NewKVStoreGateway() *KVStoreGateway {
	sg := &KVStoreGateway{stores: make([]Store, 0)}
	return sg
}

// AddStore adds a new KV store to the gateway, they are searched in order of first added, if no value is found in the
// first then the second will be searched. If none are found then an error is thrown
func (sg *KVStoreGateway) AddStore(s Store) {
	sg.stores = append(sg.stores, s)
}

func (sg *KVStoreGateway) Get(key string) string {
	for _, s := range sg.stores {
		v, found := s.Get(key)
		if found {
			return v
		}
	}
	return ""
}

func (sg *KVStoreGateway) GetMap(key string) map[string]string {
	for _, s := range sg.stores {
		v, found := s.GetMap(key)
		if found {
			return v
		}
	}
	return nil
}

func (sg *KVStoreGateway) GetArray(key string) map[string]string {
	for _, s := range sg.stores {
		v, found := s.GetMap(key)
		if found {
			return v
		}
	}
	return nil
}
