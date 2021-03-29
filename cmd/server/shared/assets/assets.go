package assets

import (
	"embed"
)

//go:embed nginx.conf
var NginxConf string

//go:embed nginx/*
var NginxDir embed.FS

//go:embed redis-cache.conf.tmpl
var RedisCacheConf string

//go:embed redis-store.conf.tmpl
var RedisStoreConf string
