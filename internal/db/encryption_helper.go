package db

import "context"

// EncryptAllTables begins the process of calling encryptTable for each table where encryption has been enabled
func EncryptAllTables(ctx context.Context) error {

	// TODO: Figure out scope
	err := EncryptTable(ctx)
	if err != nil {
		return err
	}
	// if we return without error, this implies that the encryption process completed successfully
	return nil
}
