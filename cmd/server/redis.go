package main

import (
	"os"
	"path/filepath"
	"text/template"
)

//docker:install redis

var redisConfTmpl = template.Must(template.New("redis.conf").Parse(`# allow access from all instances
protected-mode no

# limit memory usage, return error when hitting limit
maxmemory 1gb
maxmemory-policy noeviction

# live commit log to disk, additionally snapshots every 10 minutes
dir {{ .Dir }}
appendonly yes
save 600 1
`))

func maybeRedisProcFile() (string, error) {
	// Redis is already configured
	if os.Getenv("SRC_SESSION_STORE_REDIS") != "" && os.Getenv("REDIS_MASTER_ENDPOINT") != "" {
		return "", nil
	}

	// Create a redis.conf if it doesn't exist
	path := filepath.Join(os.Getenv("CONFIG_DIR"), "redis.conf")
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		dataDir := filepath.Join(os.Getenv("DATA_DIR"), "redis")
		os.MkdirAll(dataDir, os.FileMode(0755))

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
	setDefaultEnv("SRC_SESSION_STORE_REDIS", "127.0.0.1:6379")
	setDefaultEnv("REDIS_MASTER_ENDPOINT", "127.0.0.1:6379")
	return "redis: redis-server " + path, nil
}
