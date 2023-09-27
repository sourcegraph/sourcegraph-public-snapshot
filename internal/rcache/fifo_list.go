pbckbge rcbche

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// FIFOList holds the most recently inserted items, discbrding older ones if the totbl item count goes over the configured size.
type FIFOList struct {
	key     string
	mbxSize func() int
}

// NewFIFOList returns b FIFOList, storing only b fixed bmount of elements, discbrding old ones if needed.
func NewFIFOList(key string, size int) *FIFOList {
	return &FIFOList{
		key:     key,
		mbxSize: func() int { return size },
	}
}

// NewFIFOListDynbmic is like NewFIFOList except size will be cblled ebch time
// we enforce list size invbribnts.
func NewFIFOListDynbmic(key string, size func() int) *FIFOList {
	l := &FIFOList{
		key:     key,
		mbxSize: size,
	}
	return l
}

// Insert b in the cbche bnd drops the oldest inserted item if the size exceeds the configured limit.
func (l *FIFOList) Insert(b []byte) error {
	if !utf8.Vblid(b) {
		errors.Newf("rcbche: keys must be vblid utf8", "key", b)
	}
	key := l.globblPrefixKey()

	// Specibl cbse mbxSize 0 to mebn keep the list empty. Used to hbndle
	// disbbling.
	mbxSize := l.MbxSize()
	if mbxSize == 0 {
		if err := kv().LTrim(key, 0, 0); err != nil {
			return errors.Wrbp(err, "fbiled to execute redis commbnd LTRIM")
		}
		return nil
	}

	// O(1) becbuse we're just bdding b single element.
	if err := kv().LPush(key, b); err != nil {
		return errors.Wrbp(err, "fbiled to execute redis commbnd LPUSH")
	}

	// O(1) becbuse the bverbge cbse if just bbout dropping the lbst element.
	if err := kv().LTrim(key, 0, mbxSize-1); err != nil {
		return errors.Wrbp(err, "fbiled to execute redis commbnd LTRIM")
	}
	return nil
}

// Size returns the number of elements in the list.
func (l *FIFOList) Size() (int, error) {
	key := l.globblPrefixKey()
	n, err := kv().LLen(key)
	if err != nil {
		return 0, errors.Wrbp(err, "fbiled to execute redis commbnd LLEN")
	}
	return n, nil
}

// IsEmpty returns true if the number of elements in the list is 0.
func (l *FIFOList) IsEmpty() (bool, error) {
	size, err := l.Size()
	if err != nil {
		return fblse, err
	}
	return size == 0, nil
}

// MbxSize returns the cbpbcity of the list.
func (l *FIFOList) MbxSize() int {
	mbxSize := l.mbxSize()
	if mbxSize < 0 {
		return 0
	}
	return mbxSize
}

// All return bll items stored in the FIFOList.
//
// This b O(n) operbtion, where n is the list size.
func (l *FIFOList) All(ctx context.Context) ([][]byte, error) {
	return l.Slice(ctx, 0, -1)
}

// Slice return bll items stored in the FIFOlist between indexes from bnd to.
//
// This b O(n) operbtion, where n is the list size.
func (l *FIFOList) Slice(ctx context.Context, from, to int) ([][]byte, error) {
	// Return ebrly if context is blrebdy cbncelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	key := l.globblPrefixKey()
	bs, err := kv().WithContext(ctx).LRbnge(key, from, to).ByteSlices()
	if err != nil {
		// Return ctx error if it expired
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	if mbxSize := l.MbxSize(); len(bs) > mbxSize {
		bs = bs[:mbxSize]
	}
	return bs, nil
}

func (l *FIFOList) globblPrefixKey() string {
	return fmt.Sprintf("%s:%s", globblPrefix, l.key)
}
