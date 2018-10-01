package db

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"golang.org/x/crypto/acme/autocert"
)

// certCache implements autocert.Cache
type certCache struct{}

func (c *certCache) Get(ctx context.Context, key string) ([]byte, error) {
	rows, err := dbconn.Global.QueryContext(ctx,
		"SELECT b64data FROM cert_cache WHERE cache_key=$1 AND deleted_at IS NULL LIMIT 1", key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, autocert.ErrCacheMiss
	}

	var b64data string
	err = rows.Scan(&b64data)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(b64data)
}

func (c *certCache) Put(ctx context.Context, key string, data []byte) error {
	b64data := base64.StdEncoding.EncodeToString(data)
	_, err := dbconn.Global.ExecContext(
		ctx, "INSERT INTO cert_cache(cache_key, b64data) VALUES($1, $2)", key, b64data)
	if isPQErrorUniqueViolation(err) {
		_, err = dbconn.Global.ExecContext(
			ctx, "UPDATE cert_cache SET b64data=$2, updated_at=now() WHERE cache_key=$1", key, b64data)
	}
	return err
}

func (c *certCache) Delete(ctx context.Context, key string) error {
	_, err := dbconn.Global.ExecContext(
		ctx, "UPDATE cert_cache SET deleted_at=now() WHERE cache_key=$1", key)
	return err
}
