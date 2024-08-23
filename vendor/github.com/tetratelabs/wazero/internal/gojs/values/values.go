package values

import (
	"fmt"

	"github.com/tetratelabs/wazero/internal/gojs/goos"
)

func NewValues() *Values {
	ret := &Values{}
	ret.Reset()
	return ret
}

type Values struct {
	// Below is needed to avoid exhausting the ID namespace finalizeRef reclaims
	// See https://go-review.googlesource.com/c/go/+/203600

	values      []interface{}          // values indexed by ID, nil
	goRefCounts []uint32               // recount pair-indexed with values
	ids         map[interface{}]uint32 // live values
	idPool      []uint32               // reclaimed IDs (values[i] = nil, goRefCounts[i] nil
}

func (j *Values) Get(id uint32) interface{} {
	index := id - goos.NextID
	if index >= uint32(len(j.values)) {
		panic(fmt.Errorf("id %d is out of range %d", id, len(j.values)))
	}
	if v := j.values[index]; v == nil {
		panic(fmt.Errorf("value for %d was nil", id))
	} else {
		return v
	}
}

func (j *Values) Increment(v interface{}) uint32 {
	id, ok := j.ids[v]
	if !ok {
		if len(j.idPool) == 0 {
			id, j.values, j.goRefCounts = uint32(len(j.values)), append(j.values, v), append(j.goRefCounts, 0)
		} else {
			id, j.idPool = j.idPool[len(j.idPool)-1], j.idPool[:len(j.idPool)-1]
			j.values[id], j.goRefCounts[id] = v, 0
		}
		j.ids[v] = id
	}
	j.goRefCounts[id]++

	return id + goos.NextID
}

func (j *Values) Decrement(id uint32) {
	// Special IDs are not goos.Refcounted.
	if id < goos.NextID {
		return
	}
	id -= goos.NextID
	j.goRefCounts[id]--
	if j.goRefCounts[id] == 0 {
		v := j.values[id]
		j.values[id] = nil
		delete(j.ids, v)
		j.idPool = append(j.idPool, id)
	}
}

func (j *Values) Reset() {
	j.values = nil
	j.goRefCounts = nil
	j.ids = map[interface{}]uint32{}
	j.idPool = nil
}
