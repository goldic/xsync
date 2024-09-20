package xsync

import "testing"

func TestMap_init(t *testing.T) {
	var m Map[int, string]

	require(t, "" == m.Get(123))
	require(t, 0 == m.Len())
}

func TestMap_Set(t *testing.T) {
	var m Map[string, int]

	require(t, 0 == m.Get("aa"))

	m.Set("aa", 111)
	m.Set("bb", 222)

	require(t, 0 == m.Get("c"))
	require(t, m.Exists("aa"))
	require(t, m.Exists("bb"))
	require(t, !m.Exists("cc"))
	require(t, 111 == m.Get("aa"))
	require(t, 222 == m.Get("bb"))
	require(t, 2 == m.Len())
}

func TestMap_Increment(t *testing.T) {
	var m Map[string, int]
	m.Set("def", 456)

	m.Increment("abc", 100)
	m.Increment("abc", 20)
	m.Increment("abc", 3)
	m.Increment("def", -56)

	require(t, m.Get("abc") == 123)
	require(t, m.Get("def") == 400)
}

func TestMap_Exists(t *testing.T) {
	var m Map[int, string]

	require(t, !m.Exists(0))
	require(t, !m.Exists(1))

	m.Set(0, "0")
	m.Set(1, "aa")
	m.Set(2, "bb")

	require(t, m.Exists(0))
	require(t, m.Exists(1))
	require(t, m.Exists(2))
	require(t, !m.Exists(666))
}

func TestMap_Delete(t *testing.T) {
	var m Map[int, string]
	m.Set(0, "000")
	m.Set(111, "aaa")
	m.Set(222, "bbb")

	m.Delete(111)

	require(t, !m.Exists(111))
	require(t, "" == m.Get(111))
	require(t, 2 == m.Len())
}

func TestMap_Values(t *testing.T) {
	var m Map[string, int]
	m.Set("abc", 123)
	m.Set("def", 456)

	vv := m.Values()

	require(t, 2 == m.Len())
	require(t, 2 == len(vv))
	require(t, 123+456 == vv[0]+vv[1])
}

func TestMap_String(t *testing.T) {
	var m Map[string, int]
	m.Set("abc", 123)
	m.Set("def", 456)

	s := m.String()

	require(t, `{"abc":123,"def":456}` == s)
}

func TestMap_MarshalJSON(t *testing.T) {
	var m Map[string, int]
	m.Set("abc", 123)
	m.Set("def", 456)

	data, err := m.MarshalJSON()

	require(t, err == nil)
	require(t, `{"abc":123,"def":456}` == string(data))
}

func require(t *testing.T, ok bool) {
	if !ok {
		t.Fatal()
	}
}
