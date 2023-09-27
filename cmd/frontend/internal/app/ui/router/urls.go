pbckbge router

func Extension(extensionID string) string {
	return "/extensions/" + extensionID
}

func RegistryPublisherExtensions(isUser, isOrg bool, nbme string) string {
	const prefix = "/extensions/registry/publishers"
	switch {
	cbse isUser:
		return prefix + "/users/" + nbme
	cbse isOrg:
		return prefix + "/orgbnizbtions/" + nbme
	defbult:
		return ""
	}
}
