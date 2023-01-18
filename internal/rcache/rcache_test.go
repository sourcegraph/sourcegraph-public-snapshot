package rcache

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_namespace(t *testing.T) {
	SetupForTest(t)

	type testcase struct {
		prefix  string
		entries map[string]string
	}

	cases := []testcase{
		{
			prefix: "a",
			entries: map[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		}, {
			prefix: "b",
			entries: map[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		}, {
			prefix: "c",
			entries: map[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		},
	}

	caches := make([]*Cache, len(cases))
	for i, test := range cases {
		caches[i] = New(test.prefix)
		for k, v := range test.entries {
			caches[i].Set(k, []byte(v))
		}
	}
	for i, test := range cases {
		// test all the keys that should be present are found
		for k, v := range test.entries {
			b, ok := caches[i].Get(k)
			if !ok {
				t.Fatalf("error getting entry from redis (prefix=%s)", test.prefix)
			}
			if string(b) != v {
				t.Errorf("expected %s, got %s", v, string(b))
			}
		}

		// test not found case
		if _, ok := caches[i].Get("not-found"); ok {
			t.Errorf("expected not found")
		}
	}
}

func TestCache_simple(t *testing.T) {
	SetupForTest(t)

	c := New("some_prefix")
	_, ok := c.Get("a")
	if ok {
		t.Fatal("Initial Get should find nothing")
	}

	c.Set("a", []byte("b"))
	b, ok := c.Get("a")
	if !ok {
		t.Fatal("Expect to get a after setting")
	}
	if string(b) != "b" {
		t.Fatalf("got %v, want %v", string(b), "b")
	}

	c.Delete("a")
	_, ok = c.Get("a")
	if ok {
		t.Fatal("Get after delete should of found nothing")
	}
}

func TestCache_multi(t *testing.T) {
	SetupForTest(t)

	c := New("some_prefix")
	vals := c.GetMulti("k0", "k1", "k2")
	if got, exp := vals, [][]byte{nil, nil, nil}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v on initial fetch, got %v", exp, got)
	}

	c.Set("k0", []byte("b"))
	if got, exp := c.GetMulti("k0"), bytes("b"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	c.SetMulti([2]string{"k0", "a"})
	if got, exp := c.GetMulti("k0"), bytes("a"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	c.SetMulti([2]string{"k0", "a"}, [2]string{"k1", "b"})
	if got, exp := c.GetMulti("k0"), bytes("a"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
	if got, exp := c.GetMulti("k1"), bytes("b"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
	if got, exp := c.GetMulti("k0", "k1"), bytes("a", "b"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
	if got, exp := c.GetMulti("k1", "k0"), bytes("b", "a"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	c.SetMulti([2]string{"k0", "x"}, [2]string{"k1", "y"}, [2]string{"k2", "z"})
	if got, exp := c.GetMulti("k0", "k1", "k2"), bytes("x", "y", "z"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
	got, exist := c.Get("k0")
	if exp := "x"; !exist || string(got) != exp {
		t.Errorf("Expected %v, but got %v", exp, string(got))
	}

	c.Delete("k0")
	if got, exp := c.GetMulti("k0", "k1", "k2"), [][]byte{nil, []byte("y"), []byte("z")}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	c.DeleteMulti("k1", "k2")
	if got, exp := c.GetMulti("k0", "k1", "k2"), [][]byte{nil, nil, nil}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
}

func TestCache_deleteAllKeysWithPrefix(t *testing.T) {
	SetupForTest(t)

	// decrease the deleteBatchSize
	oldV := deleteBatchSize
	deleteBatchSize = 2
	defer func() { deleteBatchSize = oldV }()

	c := New("some_prefix")
	var aKeys, bKeys []string
	var key string
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			key = "a:" + strconv.Itoa(i)
			aKeys = append(aKeys, key)
		} else {
			key = "b:" + strconv.Itoa(i)
			bKeys = append(bKeys, key)
		}

		c.SetMulti([2]string{key, strconv.Itoa(i)})
	}

	conn := poolGet()
	defer conn.Close()

	err := deleteAllKeysWithPrefix(conn, c.rkeyPrefix()+"a")
	if err != nil {
		t.Error(err)
	}

	vals := c.GetMulti(aKeys...)
	if got, exp := vals, [][]byte{nil, nil, nil, nil, nil}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	vals = c.GetMulti(bKeys...)
	if got, exp := vals, bytes("1", "3", "5", "7", "9"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
}

func TestCache_Increase(t *testing.T) {
	SetupForTest(t)

	c := NewWithTTL("some_prefix", 1)
	c.Increase("a")

	got, ok := c.Get("a")
	assert.True(t, ok)
	assert.Equal(t, []byte("1"), got)

	time.Sleep(time.Second)

	// now wait upto another 5s. We do this because timing is hard.
	assert.Eventually(t, func() bool {
		_, ok = c.Get("a")
		return !ok
	}, 5*time.Second, 50*time.Millisecond, "rcache.increase did not respect expiration")
}

func TestCache_KeyTTL(t *testing.T) {
	SetupForTest(t)

	c := NewWithTTL("some_prefix", 1)
	c.Set("a", []byte("b"))

	ttl, ok := c.KeyTTL("a")
	assert.True(t, ok)
	assert.Equal(t, 1, ttl)

	time.Sleep(time.Second)

	// now wait upto another 5s. We do this because timing is hard.
	assert.Eventually(t, func() bool {
		_, ok = c.KeyTTL("a")
		return !ok
	}, 5*time.Second, 50*time.Millisecond, "rcache.ketttl did not respect expiration")

	c.SetWithTTL("c", []byte("d"), 0) // invalid TTL
	_, ok = c.KeyTTL("c")
	if ok {
		t.Fatal("KeyTTL after setting invalid ttl should have found nothing")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	SetupForTest(t)

	c := NewWithTTL("some_prefix", 60)
	c.SetWithTTL("a", []byte("b"), 30)
	b, ok := c.Get("a")
	if !ok {
		t.Fatal("Expect to get a after setting")
	}
	if string(b) != "b" {
		t.Fatalf("got %v, want %v", string(b), "b")
	}
	ttl, ok := c.KeyTTL("a")
	if !ok {
		t.Fatal("Expect to be able to read ttl after setting")
	}
	if ttl > 30 {
		t.Fatalf("ttl got %v, want %v", ttl, 30)
	}

	c.Delete("a")
	_, ok = c.Get("a")
	if ok {
		t.Fatal("Get after delete should have found nothing")
	}

	c.SetWithTTL("c", []byte("d"), 0) // invalid operation
	_, ok = c.Get("c")
	if ok {
		t.Fatal("SetWithTTL should not create a key with invalid expiry")
	}
}

func TestCache_ListKeys(t *testing.T) {
	SetupForTest(t)

	c := NewWithTTL("some_prefix", 1)
	c.SetMulti(
		[2]string{"foobar", "123"},
		[2]string{"bazbar", "456"},
		[2]string{"barfoo", "234"},
	)

	keys, err := c.ListKeys(context.Background())
	assert.NoError(t, err)
	for _, k := range []string{"foobar", "bazbar", "barfoo"} {
		assert.Contains(t, keys, k)
	}
}

func TestCache_LTrimList(t *testing.T) {
	SetupForTest(t)

	c := NewWithTTL("some_prefix", 1)

	c.AddToList("list", "1")
	c.AddToList("list", "2")
	c.AddToList("list", "3")
	c.AddToList("list", "4")
	c.AddToList("list", "5")

	c.LTrimList("list", 2)

	items, err := c.GetLastListItems("list", 8)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(items))
	assert.Equal(t, "4", items[0])
	assert.Equal(t, "5", items[1])
}

func TestCache_Hashes(t *testing.T) {
	SetupForTest(t)

	// Test SetHashItem
	c := NewWithTTL("simple_hash", 1)
	err := c.SetHashItem("key", "hashKey1", "value1")
	assert.NoError(t, err)
	err = c.SetHashItem("key", "hashKey2", "value2")
	assert.NoError(t, err)

	// Test GetHashItem
	val1, err := c.GetHashItem("key", "hashKey1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val1)
	val2, err := c.GetHashItem("key", "hashKey2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val2)
	val3, err := c.GetHashItem("key", "hashKey3")
	assert.Error(t, err)
	assert.Equal(t, "", val3)

	// Test GetHashAll
	all, err := c.GetHashAll("key")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"hashKey1": "value1", "hashKey2": "value2"}, all)
}

func TestCache_Lists(t *testing.T) {
	SetupForTest(t)

	// Use AddToList to fill list
	c := NewWithTTL("simple_list", 1)
	err := c.AddToList("key", "item1")
	assert.NoError(t, err)
	err = c.AddToList("key", "item2")
	assert.NoError(t, err)
	err = c.AddToList("key", "item3")
	assert.NoError(t, err)

	// Use GetLastListItems to get last 2 items
	last2, err := c.GetLastListItems("key", 2)
	assert.NoError(t, err)
	assert.Equal(t, []string{"item2", "item3"}, last2)

	// Use GetLastListItems to get last 5 items (we only have 3)
	last5, err := c.GetLastListItems("key", 5)
	assert.NoError(t, err)
	assert.Equal(t, []string{"item1", "item2", "item3"}, last5)
}

func bytes(s ...string) [][]byte {
	t := make([][]byte, len(s))
	for i, v := range s {
		t[i] = []byte(v)
	}
	return t
}
