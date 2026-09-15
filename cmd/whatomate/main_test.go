package main

import (
	"testing"

	"github.com/shridarpatil/whatomate/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The startup warning must describe the live account, not the config file.
// Comparing default_admin.password logged an error on every boot after the
// password had been rotated, and stayed silent when the config was edited
// while the account itself still used the default.
func TestAdminUsesDefaultPassword(t *testing.T) {
	db := testutil.SetupTestDB(t)
	org := testutil.CreateTestOrganization(t, db)

	weakEmail := testutil.UniqueEmail("startup-weak")
	testutil.CreateTestUser(t, db, org.ID,
		testutil.WithEmail(weakEmail),
		testutil.WithPassword(defaultAdminPassword),
	)
	assert.True(t, adminUsesDefaultPassword(db, weakEmail),
		"an account still using the default password must be reported")

	rotatedEmail := testutil.UniqueEmail("startup-rotated")
	testutil.CreateTestUser(t, db, org.ID,
		testutil.WithEmail(rotatedEmail),
		testutil.WithPassword("a-properly-rotated-password"),
	)
	assert.False(t, adminUsesDefaultPassword(db, rotatedEmail),
		"a rotated account must not keep warning on every boot")

	assert.False(t, adminUsesDefaultPassword(db, testutil.UniqueEmail("startup-absent")),
		"the configured password is only a seed value when no such account exists")
	assert.False(t, adminUsesDefaultPassword(db, ""),
		"an unset default_admin.email must not be looked up")

	// Rotating the live account clears the warning without a config change.
	require.NoError(t, db.Exec(
		"UPDATE users SET password_hash = (SELECT password_hash FROM users WHERE email = ?) WHERE email = ?",
		rotatedEmail, weakEmail).Error)
	assert.False(t, adminUsesDefaultPassword(db, weakEmail))
}
