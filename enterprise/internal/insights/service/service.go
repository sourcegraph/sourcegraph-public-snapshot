package service

func IsJustInTime(repositories []string) bool {
	if len(repositories) == 0 {
		return false
	}
	return true
}
