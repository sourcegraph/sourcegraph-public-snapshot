package shared

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/sourcegraph/sourcegraph/cmd/server/shared/assets"
)

var redisStoreConfTmpl = template.Must(template.New("redis-store.conf").Parse(assets.MustAssetString("redis-store.conf.tmpl")))
var redisCacheConfTmpl = template.Must(template.New("redis-cache.conf").Parse(assets.MustAssetString("redis-cache.conf.tmpl")))

type redisProcfileConfig struct {
	envVar  string
	name    string
	port    string
	tmpl    *template.Template
	dataDir string
}

func maybeRedisStoreProcFile() (string, error) {
	return maybeRedisProcFile(redisProcfileConfig{
		envVar:  "REDIS_STORE_ENDPOINT",
		name:    "redis-store",
		port:    "6379",
		tmpl:    redisStoreConfTmpl,
		dataDir: "redis",
	})
}

func maybeRedisCacheProcFile() (string, error) {
	return maybeRedisProcFile(redisProcfileConfig{
		envVar:  "REDIS_CACHE_ENDPOINT",
		name:    "redis-cache",
		port:    "6380",
		tmpl:    redisCacheConfTmpl,
		dataDir: "redis-cache",
	})
}

func maybeRedisProcFile(c redisProcfileConfig) (string, error) {
	// Redis is already configured. See envvars used in pkg/redispool.
	if os.Getenv("REDIS_ENDPOINT") != "" {
		return "", nil
	}

	if os.Getenv(c.envVar) != "" {
		return "", nil
	}

	conf, err := tryCreateRedisConf(c)
	if err != nil {
		return "", err
	}

	SetDefaultEnv(c.envVar, "127.0.0.1:"+c.port)

	return redisProcFileEntry(c.name, conf), nil
}

func tryCreateRedisConf(c redisProcfileConfig) (string, error) {
	// Create a redis.conf if it doesn't exist
	path := filepath.Join(os.Getenv("CONFIG_DIR"), c.name+".conf")

	_, err := os.Stat(path)
	if err == nil {
		return path, nil
	}

	if !os.IsNotExist(err) {
		return "", err
	}

	dataDir := filepath.Join(os.Getenv("DATA_DIR"), c.dataDir)
	err = os.MkdirAll(dataDir, os.FileMode(0755))
	if err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	err = c.tmpl.Execute(f, struct{ Dir, Port string }{
		Dir:  dataDir,
		Port: c.port,
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
