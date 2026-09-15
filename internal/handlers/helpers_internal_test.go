package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskConfigHeaders(t *testing.T) {
	t.Parallel()
	cfg := map[string]any{
		"url":     "https://api.example.com/crm",
		"method":  "POST",
		"headers": map[string]any{"Authorization": "Bearer top-secret", "X-Trace": "1"},
	}

	masked := maskConfigHeaders(cfg, false)
	assert.Equal(t, "https://api.example.com/crm", masked["url"])
	assert.Equal(t, map[string]any{"Authorization": maskedSecretValue, "X-Trace": maskedSecretValue}, masked["headers"])
	assert.Equal(t, "Bearer top-secret", cfg["headers"].(map[string]any)["Authorization"], "input must not be modified")

	same := maskConfigHeaders(cfg, true)
	assert.Equal(t, "Bearer top-secret", same["headers"].(map[string]any)["Authorization"], "editors see the real values")

	assert.Nil(t, maskConfigHeaders(nil, false))
	noHeaders := map[string]any{"url": "x"}
	assert.Equal(t, noHeaders, maskConfigHeaders(noHeaders, false))
}
