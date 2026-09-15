package handlers

import (
	"testing"
	"time"

	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/shridarpatil/whatomate/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A JavaScript action that never returns must be interrupted instead of
// pinning a CPU core and hanging the request forever.
func TestExecuteJavaScriptAction_InterruptsRunawayScript(t *testing.T) {
	t.Parallel()
	app := &App{Log: testutil.NopLogger()}
	action := models.CustomAction{
		ActionType: models.ActionTypeJavascript,
		Config:     models.JSONB{"code": "while (true) {}"},
	}

	start := time.Now()
	res, err := app.executeJavaScriptAction(action, map[string]any{"contact": map[string]any{}, "user": map[string]any{}, "organization": map[string]any{}})
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "time limit")
	assert.Less(t, elapsed, jsActionTimeout+3*time.Second, "script must be stopped shortly after the budget")
}

func TestExecuteJavaScriptAction_NormalScriptStillRuns(t *testing.T) {
	t.Parallel()
	app := &App{Log: testutil.NopLogger()}
	action := models.CustomAction{
		ActionType: models.ActionTypeJavascript,
		Config:     models.JSONB{"code": "return { message: 'hi ' + contact.name }"},
	}
	res, err := app.executeJavaScriptAction(action, map[string]any{"contact": map[string]any{"name": "Ada"}, "user": map[string]any{}, "organization": map[string]any{}})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Success)
	assert.Equal(t, "hi Ada", res.Message)
}

// Expired redirect tokens are swept when a new one is stored, so tokens whose
// redirect is never followed cannot accumulate for the life of the process.
func TestStoreRedirectToken_SweepsExpiredEntries(t *testing.T) {
	redirectTokenMutex.Lock()
	redirectTokens["stale-token"] = redirectToken{URL: "https://example.com/old", ExpiresAt: time.Now().Add(-time.Minute)}
	redirectTokenMutex.Unlock()

	storeRedirectToken("fresh-token", "https://example.com/new")

	redirectTokenMutex.RLock()
	defer redirectTokenMutex.RUnlock()
	_, staleKept := redirectTokens["stale-token"]
	_, freshKept := redirectTokens["fresh-token"]
	assert.False(t, staleKept)
	assert.True(t, freshKept)
}
