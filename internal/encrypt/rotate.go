package encrypt

//rotate keys performs a key rotation on the registered secret columns provided

// RotateKeys runs as a background process whenever >1 key is present
func RotateKeys() {

	if len(cryptObject.EncryptionKeys) < 2 {
		return
	}

}

func GatherSecretColumns() {

}

// Users of the encrypt package should provide the table and column names of what columns require rotation
func RegisterSecretColumns(table string, secrets ...string) {

}
