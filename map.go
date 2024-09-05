package xsync

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math/rand"
	"sync"
)

// A Map is a set of temporary objects that may be individually set, get and deleted.
//
// A Map is safe for use by multiple goroutines simultaneously.
type Map[K comparable, T any] struct {
	mx   sync.RWMutex
	ver  uint64
	vals map[K]T
}

func NewMap[K comparable, T any](values map[K]T) Map[K, T] {
	return Map[K, T]{
		vals: maps.Clone(values),
	}
}

func (m *Map[K, T]) Clear() {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.vals = nil
	m.ver++
}

func (m *Map[K, T]) Set(key K, value T) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.vals == nil {
		m.vals = map[K]T{}
	}
	m.vals[key] = value
	m.ver++
}

//func (m *Map[K,T]) Increment(key K, val T) T {
//	m.mx.Lock()
//	defer m.mx.Unlock()
//	if m.vals == nil {
//		m.vals = map[K]T{}
//	}
//	if v, ok := m.vals[key]; ok {
//		val += v
//	}
//	m.vals[key] = val
//	m.ver++
//	return val
//}

func (m *Map[K, T]) Delete(key K) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.vals != nil {
		delete(m.vals, key)
		m.ver++
	}
}

func (m *Map[K, T]) Get(key K) (_ T) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if m.vals != nil {
		return m.vals[key]
	}
	return
}

func (m *Map[K, T]) GetOrSet(key K, fn func() T) (res T) {
	var ok bool
	m.mx.RLock()
	if m.vals != nil {
		res, ok = m.vals[key]
	}
	m.mx.RUnlock()
	if !ok {
		res = fn()
		m.Set(key, res)
	}
	return
}

func (m *Map[K, T]) Exists(key K) bool {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if m.vals == nil {
		return false
	}
	_, ok := m.vals[key]
	return ok
}

func (m *Map[K, T]) Len() int {
	m.mx.RLock()
	defer m.mx.RUnlock()
	return len(m.vals)
}

func (m *Map[K, T]) Version() uint64 {
	m.mx.RLock()
	defer m.mx.RUnlock()
	return m.ver
}

func (m *Map[K, T]) KeyValues() map[K]T {
	m.mx.RLock()
	defer m.mx.RUnlock()

	res := map[K]T{}
	if m.vals != nil {
		for k, v := range m.vals {
			res[k] = v
		}
	}
	return res
}

func (m *Map[K, T]) Keys() []K {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return mapKeys(m.vals)
}

func (m *Map[K, T]) Values() []T {
	m.mx.RLock()
	defer m.mx.RUnlock()

	vv := make([]T, 0, len(m.vals))
	if m.vals != nil {
		for _, v := range m.vals {
			vv = append(vv, v)
		}
	}
	return vv
}

func (m *Map[K, T]) String() string {
	m.mx.RLock()
	defer m.mx.RUnlock()
	return encString(m.vals)
}

func (m *Map[K, T]) Pop() (key K, value T) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.vals != nil {
		for key, value = range m.vals {
			delete(m.vals, key)
			m.ver++
			return
		}
	}
	return
}

func (m *Map[K, T]) PopAll() (values map[K]T) {
	m.mx.Lock()
	defer m.mx.Unlock()

	values, m.vals = m.vals, nil
	m.ver++
	return
}

func (m *Map[K, T]) RandomValue() T {
	_, v := m.Random()
	return v
}

func (m *Map[K, T]) RandomKey() K {
	k, _ := m.Random()
	return k
}

func (m *Map[K, T]) Random() (key K, value T) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if cnt := len(m.vals); cnt > 0 {
		// todo: optimize it!  (add keys slice)
		n := rand.Intn(cnt)
		for k, v := range m.vals {
			if n == 0 {
				return k, v
			}
			n--
		}
		panic(1)
	}
	return
}

func (m *Map[K, T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.KeyValues())
}

func (m *Map[K, T]) UnmarshalJSON(data []byte) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	err := json.NewDecoder(bytes.NewReader(data)).Decode(&m.vals)
	m.ver++
	return err
}

func (m *Map[K, T]) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer(nil)
	err := m.BinaryEncode(w)
	return w.Bytes(), err
}

func (m *Map[K, T]) UnmarshalBinary(data []byte) error {
	return m.BinaryDecode(bytes.NewReader(data))
}

func (m *Map[K, T]) BinaryEncode(w io.Writer) error {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return gob.NewEncoder(w).Encode(m.vals)
}

func (m *Map[K, T]) BinaryDecode(r io.Reader) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	err := gob.NewDecoder(r).Decode(&m.vals)
	m.ver++
	return err
}

// String returns object as string (encode to json)
func encString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case fmt.Stringer:
		return s.String()
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func mapKeys[K comparable, T any](mm map[K]T) []K {
	if mm == nil {
		return nil
	}
	vv := make([]K, 0, len(mm))
	for k := range mm {
		vv = append(vv, k)
	}
	return vv
}
