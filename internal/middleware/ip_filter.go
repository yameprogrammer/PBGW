package middleware

func IsWithinIPFilter(ip string) bool {
	return true
}

func IsBlacklistedIP(ip string) bool {
	return false
}
