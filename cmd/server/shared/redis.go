package shared

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/sourcegraph/sourcegraph/cmd/server/shared/assets"
)

var redisStoreConfTmpl = template.Must(template.New("redis-store.conf").Parse(assets.MustAssetString("redis-store.conf.tmpl")))
var redisCacheConfTmpl = template.Must(template.New("redis-cache.conf").Parse(assets.MustAssetString("redis-cache.conf.tmpl")))

func maybeRedisStoreProcFile() (string, error) {
	return maybeRedisProcFile("REDIS_STORE_ENDPOINT", "redis-store", "6379", redisStoreConfTmpl)
}

func maybeRedisCacheProcFile() (string, error) {
	return maybeRedisProcFile("REDIS_CACHE_ENDPOINT", "redis-cache", "6380", redisCacheConfTmpl)
}

func maybeRedisProcFile(envVar, name, port string, tmpl *template.Template) (string, error) {
	// Redis is already configured. See envvars used in pkg/redispool.
	if os.Getenv("REDIS_ENDPOINT") != "" {
		return "", nil
	}

	if os.Getenv(envVar) != "" {
		return "", nil
	}

	conf, err := tryCreateRedisConf(tmpl, name, port)
	if err != nil {
		return "", err
	}

	SetDefaultEnv(envVar, "127.0.0.1:"+port)

	return redisProcFileEntry(name, conf), nil
}

func tryCreateRedisConf(tmpl *template.Template, name, port string) (string, error) {
	// Create a redis.conf if it doesn't exist
	path := filepath.Join(os.Getenv("CONFIG_DIR"), name+".conf")

	_, err := os.Stat(path)
	if err == nil {
		return path, nil
	}

	if !os.IsNotExist(err) {
		return "", err
	}

	dataDir := filepath.Join(os.Getenv("DATA_DIR"), name)
	err = os.MkdirAll(dataDir, os.FileMode(0755))
	if err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(f, struct{ Dir, Port string }{
		Dir:  dataDir,
		Port: port,
	})
	f.Close()
	if err != nil {
		os.Remove(path)
		return "", err
	}

	return path, nil
}

func redisProcFileEntry(name, conf string) string {
	// Redis is noiser than we prefer even at the most quiet setting "warning"
	// so we only output the last log line when redis stops in case it stopped unexpectly
	// and the log contains the reason why it stopped.
	return name + ": redis-server " + conf + " | tail -n 1"
}
