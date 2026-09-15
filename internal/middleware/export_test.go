package middleware

// ResetAPIKeyCache clears the verified-key cache. Tests use it to force the
// bcrypt path and to measure the cold vs. warm cost of API key authentication.
func ResetAPIKeyCache() {
	apiKeyCacheMu.Lock()
	defer apiKeyCacheMu.Unlock()
	clear(apiKeyCache)
}
