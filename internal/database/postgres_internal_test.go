package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A password with spaces, quotes or backslashes must not break the key=value
// DSN or inject extra parameters.
func TestQuoteDSNValue(t *testing.T) {
	t.Parallel()
	assert.Equal(t, `'plain'`, quoteDSNValue("plain"))
	assert.Equal(t, `'with space'`, quoteDSNValue("with space"))
	assert.Equal(t, `'it\'s'`, quoteDSNValue("it's"))
	assert.Equal(t, `'back\slash'`, quoteDSNValue(`back\slash`))
	assert.Equal(t, `'a=b sslmode=disable'`, quoteDSNValue("a=b sslmode=disable"))
}
