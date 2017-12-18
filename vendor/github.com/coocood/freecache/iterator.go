package freecache

import (
	"time"
	"unsafe"
)

// Iterator iterates the entries for the cache.
type Iterator struct {
	cache      *Cache
	segmentIdx int
	slotIdx    int
	entryIdx   int
}

// Entry represents a key/value pair.
type Entry struct {
	Key   []byte
	Value []byte
}

// Next returns the next entry for the iterator.
// The order of the entries is not guaranteed.
// If there is no more entries to return, nil will be returned.
func (it *Iterator) Next() *Entry {
	for it.segmentIdx < 256 {
		entry := it.nextForSegment(it.segmentIdx)
		if entry != nil {
			return entry
		}
		it.segmentIdx++
		it.slotIdx = 0
		it.entryIdx = 0
	}
	return nil
}

func (it *Iterator) nextForSegment(segIdx int) *Entry {
	it.cache.locks[segIdx].Lock()
	defer it.cache.locks[segIdx].Unlock()
	seg := &it.cache.segments[segIdx]
	for it.slotIdx < 256 {
		entry := it.nextForSlot(seg, it.slotIdx)
		if entry != nil {
			return entry
		}
		it.slotIdx++
		it.entryIdx = 0
	}
	return nil
}

func (it *Iterator) nextForSlot(seg *segment, slotId int) *Entry {
	slotOff := int32(it.slotIdx) * seg.slotCap
	slot := seg.slotsData[slotOff : slotOff+seg.slotLens[it.slotIdx] : slotOff+seg.slotCap]
	for it.entryIdx < len(slot) {
		ptr := slot[it.entryIdx]
		it.entryIdx++
		now := uint32(time.Now().Unix())
		var hdrBuf [ENTRY_HDR_SIZE]byte
		seg.rb.ReadAt(hdrBuf[:], ptr.offset)
		hdr := (*entryHdr)(unsafe.Pointer(&hdrBuf[0]))
		if hdr.expireAt == 0 || hdr.expireAt > now {
			entry := new(Entry)
			entry.Key = make([]byte, hdr.keyLen)
			entry.Value = make([]byte, hdr.valLen)
			seg.rb.ReadAt(entry.Key, ptr.offset+ENTRY_HDR_SIZE)
			seg.rb.ReadAt(entry.Value, ptr.offset+ENTRY_HDR_SIZE+int64(hdr.keyLen))
			return entry
		}
	}
	return nil
}

// NewIterator creates a new iterator for the cache.
func (cache *Cache) NewIterator() *Iterator {
	return &Iterator{
		cache: cache,
	}
}
