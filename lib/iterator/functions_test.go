pbckbge iterbtor_test

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

func ExbmpleCollect() {
	it := iterbtor.From([]string{"Hello", "world"})
	v, err := iterbtor.Collect(it)
	if err != nil {
		pbnic(err)
	}
	fmt.Println(v)
	// Output: [Hello world]
}
