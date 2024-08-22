package xsync

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"math/rand"
	"sync"
)

// A Set is a set of temporary objects that may be individually set, get and deleted.
//
// A Set is safe for use by multiple goroutines simultaneously.
type Set[K comparable] struct {
	mx   sync.RWMutex
	ver  uint64
	vals map[K]struct{}
}

func NewSet[K comparable](values []K) Set[K] {
	vv := make(map[K]struct{}, len(values))
	for _, v := range values {
		vv[v] = struct{}{}
	}
	return Set[K]{vals: vv}
}

func (m *Set[K]) Clear() {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.vals = map[K]struct{}{}
	m.ver++
}

func (m *Set[K]) Set(key K) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.vals == nil {
		m.vals = map[K]struct{}{}
	}
	m.vals[key] = struct{}{}
	m.ver++
}

func (m *Set[K]) Delete(key K) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.vals != nil {
		delete(m.vals, key)
		m.ver++
	}
}

func (m *Set[K]) Exists(key K) bool {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if m.vals == nil {
		return false
	}
	_, ok := m.vals[key]
	return ok
}

func (m *Set[K]) Size() int {
	m.mx.RLock()
	defer m.mx.RUnlock()
	return len(m.vals)
}

func (m *Set[K]) Version() uint64 {
	m.mx.RLock()
	defer m.mx.RUnlock()
	return m.ver
}

func (m *Set[K]) Values() []K {
	m.mx.RLock()
	defer m.mx.RUnlock()
	return mapKeys(m.vals)
}

func (m *Set[K]) String() string {
	return encString(m.Strings())
}

func (m *Set[K]) Strings() []string {
	ss := make([]string, 0, len(m.vals))
	for k := range m.Values() {
		ss = append(ss, encString(k))
	}
	return ss
}

func (m *Set[K]) Pop() (key K) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.vals != nil {
		for key = range m.vals {
			delete(m.vals, key)
			m.ver++
			return
		}
	}
	return
}

func (m *Set[K]) PopAll() (values []K) {
	m.mx.Lock()
	defer m.mx.Unlock()
	values, m.vals = mapKeys(m.vals), nil
	m.ver++
	return
}

func (m *Set[K]) Random() (key K) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if cnt := len(m.vals); cnt > 0 {
		// todo: optimize it!  (add keys slice)
		n := rand.Intn(cnt)
		for k := range m.vals {
			if n == 0 {
				return k
			}
			n--
		}
		panic(1)
	}
	return
}

func (m *Set[K]) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Values())
}

func (m *Set[K]) UnmarshalJSON(data []byte) (err error) {
	var vv []K
	if err = json.Unmarshal(data, &vv); err != nil {
		return
	}
	m.mx.Lock()
	defer m.mx.Unlock()
	m.vals, m.ver = sliceToMap(vv), m.ver+1
	return
}

func (m *Set[K]) BinaryEncode(w io.Writer) error {
	return gob.NewEncoder(w).Encode(m.Values())
}

func (m *Set[K]) BinaryDecode(r io.Reader) (err error) {
	var vv []K
	if err = gob.NewDecoder(r).Decode(&vv); err != nil {
		return
	}
	m.mx.Lock()
	defer m.mx.Unlock()
	m.vals, m.ver = sliceToMap(vv), m.ver+1
	return
}

func sliceToMap[K comparable](s []K) map[K]struct{} {
	m := make(map[K]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
