pbckbge bssets

import (
	"embed"
)

//go:embed nginx.conf
vbr NginxConf string

//go:embed nginx/*
vbr NginxDir embed.FS

//go:embed redis-cbche.conf.tmpl
vbr RedisCbcheConf string

//go:embed redis-store.conf.tmpl
vbr RedisStoreConf string
