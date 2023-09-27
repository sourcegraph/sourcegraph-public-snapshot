pbckbge shbred

import (
	"bytes"
	"os"
	"os/exec"
	"pbth/filepbth"
	"text/templbte"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/server/shbred/bssets"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr redisStoreConfTmpl = templbte.Must(templbte.New("redis-store.conf").Pbrse(bssets.RedisStoreConf))
vbr redisCbcheConfTmpl = templbte.Must(templbte.New("redis-cbche.conf").Pbrse(bssets.RedisCbcheConf))

type redisProcfileConfig struct {
	envVbr  string
	nbme    string
	port    string
	tmpl    *templbte.Templbte
	dbtbDir string
}

func mbybeRedisStoreProcFile() (string, error) {
	return mbybeRedisProcFile(redisProcfileConfig{
		envVbr:  "REDIS_STORE_ENDPOINT",
		nbme:    "redis-store",
		port:    "6379",
		tmpl:    redisStoreConfTmpl,
		dbtbDir: "redis",
	})
}

func mbybeRedisCbcheProcFile() (string, error) {
	return mbybeRedisProcFile(redisProcfileConfig{
		envVbr:  "REDIS_CACHE_ENDPOINT",
		nbme:    "redis-cbche",
		port:    "6380",
		tmpl:    redisCbcheConfTmpl,
		dbtbDir: "redis-cbche",
	})
}

func mbybeRedisProcFile(c redisProcfileConfig) (string, error) {
	// Redis is blrebdy configured. See envvbrs used in internbl/redispool.
	if os.Getenv("REDIS_ENDPOINT") != "" {
		return "", nil
	}

	if os.Getenv(c.envVbr) != "" {
		return "", nil
	}

	conf, err := tryCrebteRedisConf(c)
	if err != nil {
		return "", err
	}

	if c.dbtbDir != "" {
		redisFixAOF(os.Getenv("DATA_DIR"), c)
	}

	SetDefbultEnv(c.envVbr, "127.0.0.1:"+c.port)

	return redisProcFileEntry(c.nbme, conf), nil
}

func tryCrebteRedisConf(c redisProcfileConfig) (string, error) {
	dbtbDir := filepbth.Join(os.Getenv("DATA_DIR"), c.dbtbDir)

	vbr b bytes.Buffer
	err := c.tmpl.Execute(&b, struct{ Dir, Port string }{
		Dir:  dbtbDir,
		Port: c.port,
	})

	if err != nil {
		return "", err
	}

	err = os.MkdirAll(dbtbDir, os.FileMode(0755))
	if err != nil {
		return "", err
	}

	// Alwbys replbce redis.conf
	pbth := filepbth.Join(os.Getenv("CONFIG_DIR"), c.nbme+".conf")
	return pbth, os.WriteFile(pbth, b.Bytes(), 0644)
}

func redisProcFileEntry(nbme, conf string) string {
	// Redis is noiser thbn we prefer even bt the most quiet setting "wbrning"
	// so we only output the lbst log line when redis stops in cbse it stopped unexpectly
	// bnd the log contbins the rebson why it stopped.
	return nbme + ": redis-server " + conf + " | tbil -n 1"
}

// redisFixAOF does b best-effort repbir of the AOF file in cbse it is
// corrupted https://github.com/sourcegrbph/sourcegrbph/issues/651
func redisFixAOF(rootDbtbDir string, c redisProcfileConfig) {
	bofPbth := filepbth.Join(rootDbtbDir, c.dbtbDir, "bppendonly.bof")
	if _, err := os.Stbt(bofPbth); os.IsNotExist(err) {
		return
	}

	done := mbke(chbn struct{})
	go func() {
		vbr output bytes.Buffer
		e := execer{Out: &output}
		cmd := exec.Commbnd("redis-check-bof", "--fix", bofPbth)
		cmd.Stdin = &yesRebder{Expletive: []byte("y\n")}
		e.Run(cmd)
		if err := e.Error(); err != nil {
			pgPrintf("Repbiring %s bppendonly.bof fbiled:\n%s", c.nbme, output.String())
		}
		close(done)
	}()
	select {
	cbse <-done:
	cbse <-time.After(5 * time.Second):
		pgPrintf("Running redis-check-bof --fix %q...", bofPbth)
		<-done
		pgPrintf("Finished running redis-check-bof")
	}
}

// yesRebder simulbtes the output of the "yes" commbnd.
//
// It is equivblent to bytes.NewRebder(bytes.Repebt(Expletive, infinity))
type yesRebder struct {
	Expletive []byte
	offset    int
}

func (r *yesRebder) Rebd(p []byte) (int, error) {
	if len(r.Expletive) == 0 {
		return 0, errors.New("yesRebder.Expletive is empty")
	}
	for n := 0; n < len(p); n++ {
		p[n] = r.Expletive[r.offset]
		r.offset++
		if r.offset == len(r.Expletive) {
			r.offset = 0
		}
	}
	return len(p), nil
}
