package router

func RegistryExtension(extensionID string) string {
	return "/registry/extensions/" + extensionID
}

func RegistryPublisherExtensions(isUser, isOrg bool, name string) string {
	const prefix = "/registry/publishers"
	switch {
	case isUser:
		return prefix + "/users/" + name
	case isOrg:
		return prefix + "/organizations/" + name
	default:
		return ""
	}
}
