package bg

import (
	"github.com/garyburd/redigo/redis"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/pkg/redispool"
	"gopkg.in/inconshreveable/log15.v2"
)

func DeleteOldCacheDataInRedis() {
	storeConn := redispool.Store.Get()
	defer storeConn.Close()

	cacheConn := redispool.Cache.Get()
	defer cacheConn.Close()

	for _, c := range []redis.Conn{storeConn, cacheConn} {
		err := rcache.DeleteOldCacheData(c)
		if err != nil {
			log15.Error("Unable to delete old cache data in redis search. Please report this issue.", "error", err)
			return
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_316(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
