package router

func Extension(extensionID string) string {
	return "/extensions/" + extensionID
}

func RegistryPublisherExtensions(isUser, isOrg bool, name string) string {
	const prefix = "/extensions/registry/publishers"
	switch {
	case isUser:
		return prefix + "/users/" + name
	case isOrg:
		return prefix + "/organizations/" + name
	default:
		return ""
	}
}
