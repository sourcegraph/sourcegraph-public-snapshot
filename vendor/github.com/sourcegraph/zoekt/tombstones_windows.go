package zoekt

func init() {
	// no setting of file permissions on Windows
	umask = 0
}
