package shared

import (
	"os"
	"path/filepath"
	"text/template"
)

var redisConfTmpl = template.Must(template.New("redis.conf").Parse(`# allow access from all instances
protected-mode no

# limit memory usage, return error when hitting limit
maxmemory 1gb
maxmemory-policy noeviction

# live commit log to disk, additionally snapshots every 10 minutes
dir {{ .Dir }}
appendonly yes
save 600 1

# least verbose logging
loglevel warning
`))

func maybeRedisProcFile() (string, error) {
	// Redis is already configured. See envvars used in pkg/redispool.
	if os.Getenv("REDIS_ENDPOINT") != "" {
		return "", nil
	}
	store := os.Getenv("REDIS_STORE_ENDPOINT") != ""
	cache := os.Getenv("REDIS_CACHE_ENDPOINT") != ""
	if store && cache {
		return "", nil
	}

	// Create a redis.conf if it doesn't exist
	path := filepath.Join(os.Getenv("CONFIG_DIR"), "redis.conf")
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		dataDir := filepath.Join(os.Getenv("DATA_DIR"), "redis")
		err := os.MkdirAll(dataDir, os.FileMode(0755))
		if err != nil {
			return "", err
		}

		f, err := os.Create(path)
		if err != nil {
			return "", err
		}

		err = redisConfTmpl.Execute(f, struct{ Dir string }{
			Dir: dataDir,
		})
		f.Close()
		if err != nil {
			os.Remove(path)
			return "", err
		}
	}

	// Run and use a local redis
	SetDefaultEnv("REDIS_ENDPOINT", "127.0.0.1:6379")

	// Redis is noiser than we prefer even at the most quiet setting "warning"
	// so we only output the last log line when redis stops in case it stopped unexpectly
	// and the log contains the reason why it stopped.
	return "redis: redis-server " + path + " | tail -n 1", nil
}
