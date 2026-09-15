package middleware_test

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/shridarpatil/whatomate/internal/middleware"
	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/shridarpatil/whatomate/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// newAPIKey inserts an active API key for user and returns the plaintext key,
// built exactly like handlers.CreateAPIKey does (whm_ + 32 hex, bcrypt hash,
// 16-char prefix).
func newAPIKey(t *testing.T, db *gorm.DB, orgID, userID uuid.UUID) (string, *models.APIKey) {
	t.Helper()

	raw := make([]byte, 16)
	_, err := rand.Read(raw)
	require.NoError(t, err)
	plaintext := "whm_" + hex.EncodeToString(raw)

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	require.NoError(t, err)

	key := &models.APIKey{
		BaseModel:      models.BaseModel{ID: uuid.New()},
		OrganizationID: orgID,
		UserID:         userID,
		Name:           "test-key",
		KeyPrefix:      plaintext[4:20],
		KeyHash:        string(hash),
		IsActive:       true,
	}
	require.NoError(t, db.Create(key).Error)
	return plaintext, key
}

func apiKeyRequest(key string) *fastglue.Request {
	req := &fastglue.Request{RequestCtx: &fasthttp.RequestCtx{}}
	req.RequestCtx.Request.Header.Set("X-API-Key", key)
	return req
}

func TestAuthWithDB_APIKey_CacheSkipsBcryptButHonoursRevocation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	middleware.ResetAPIKeyCache()

	org := testutil.CreateTestOrganization(t, db)
	user := testutil.CreateTestUser(t, db, org.ID)
	plaintext, key := newAPIKey(t, db, org.ID, user.ID)

	auth := middleware.AuthWithDB(testJWTSecret, db)

	// First request: bcrypt path, populates the cache.
	req := apiKeyRequest(plaintext)
	require.NotNil(t, auth(req), "valid key must authenticate")
	gotUser, _ := req.RequestCtx.UserValue(middleware.ContextKeyUserID).(uuid.UUID)
	assert.Equal(t, user.ID, gotUser)

	// Second request: cached path, same outcome.
	req = apiKeyRequest(plaintext)
	require.NotNil(t, auth(req), "cached key must authenticate")
	gotOrg, _ := req.RequestCtx.UserValue(middleware.ContextKeyOrganizationID).(uuid.UUID)
	assert.Equal(t, org.ID, gotOrg)

	// A key that differs in one character never hits the cache entry.
	wrong := plaintext[:35] + map[bool]string{true: "0", false: "1"}[plaintext[35] != '0']
	req = apiKeyRequest(wrong)
	require.Nil(t, auth(req), "tampered key must be rejected")

	// Deactivating the key must take effect immediately despite the cache.
	require.NoError(t, db.Model(key).Update("is_active", false).Error)
	req = apiKeyRequest(plaintext)
	require.Nil(t, auth(req), "deactivated key must be rejected even when cached")
	assert.Equal(t, fasthttp.StatusUnauthorized, req.RequestCtx.Response.StatusCode())
}

// BenchmarkAuthWithDB_APIKey_Cold measures the original per-request cost:
// every request pays for a bcrypt comparison.
func BenchmarkAuthWithDB_APIKey_Cold(b *testing.B) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("TEST_DATABASE_URL not set")
	}
	t := &testing.T{}
	db := testutil.SetupTestDB(t)
	org := testutil.CreateTestOrganization(t, db)
	user := testutil.CreateTestUser(t, db, org.ID)
	plaintext, _ := newAPIKey(t, db, org.ID, user.ID)
	auth := middleware.AuthWithDB(testJWTSecret, db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.ResetAPIKeyCache()
		if auth(apiKeyRequest(plaintext)) == nil {
			b.Fatal("authentication failed")
		}
	}
}

// BenchmarkAuthWithDB_APIKey_Warm measures the cost once the key has been
// verified: one indexed row lookup, no bcrypt.
func BenchmarkAuthWithDB_APIKey_Warm(b *testing.B) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("TEST_DATABASE_URL not set")
	}
	t := &testing.T{}
	db := testutil.SetupTestDB(t)
	org := testutil.CreateTestOrganization(t, db)
	user := testutil.CreateTestUser(t, db, org.ID)
	plaintext, _ := newAPIKey(t, db, org.ID, user.ID)
	auth := middleware.AuthWithDB(testJWTSecret, db)
	if auth(apiKeyRequest(plaintext)) == nil {
		b.Fatal("warm-up authentication failed")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if auth(apiKeyRequest(plaintext)) == nil {
			b.Fatal("authentication failed")
		}
	}
}
