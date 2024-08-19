package redispool

// DeleteAllKeysWithPrefix retrieves all keys starting with 'prefix:' and sequentially deletes each one.
//
// NOTE: this should only be used for tests, as it is not atomic and cannot handle large keyspaces.
func DeleteAllKeysWithPrefix(kv KeyValue, prefix string) error {
	keys, err := kv.Keys(prefix + ":*")
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err := kv.Del(key); err != nil {
			return err
		}
	}
	return nil
}
