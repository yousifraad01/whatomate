package database_test

import (
	"testing"

	"github.com/shridarpatil/whatomate/internal/database"
	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/shridarpatil/whatomate/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The configured default admin is promoted only while the installation has
// no super admin at all. Once one exists, registering the same email again
// must not grant super admin at the next migration.
func TestEnsureSuperAdmin_OnlyWhenNoneExists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	org := testutil.CreateTestOrganization(t, db)
	email := testutil.UniqueEmail("legacy-admin")
	legacy := testutil.CreateTestUser(t, db, org.ID, testutil.WithEmail(email))

	// Isolate from super admins created by other tests: temporarily demote
	// them for this check and restore afterwards.
	var existing []models.User
	require.NoError(t, db.Where("is_super_admin = ?", true).Find(&existing).Error)
	for _, u := range existing {
		require.NoError(t, db.Model(&models.User{}).Where("id = ?", u.ID).Update("is_super_admin", false).Error)
	}
	t.Cleanup(func() {
		for _, u := range existing {
			db.Model(&models.User{}).Where("id = ?", u.ID).Update("is_super_admin", true)
		}
	})

	require.NoError(t, database.EnsureSuperAdmin(db, email))
	var reloaded models.User
	require.NoError(t, db.Where("id = ?", legacy.ID).First(&reloaded).Error)
	assert.True(t, reloaded.IsSuperAdmin, "legacy install: configured admin is promoted once")

	impostorEmail := testutil.UniqueEmail("impostor")
	impostor := testutil.CreateTestUser(t, db, org.ID, testutil.WithEmail(impostorEmail))
	require.NoError(t, database.EnsureSuperAdmin(db, impostorEmail))
	var reloadedImpostor models.User
	require.NoError(t, db.Where("id = ?", impostor.ID).First(&reloadedImpostor).Error)
	assert.False(t, reloadedImpostor.IsSuperAdmin, "a super admin already exists: no further promotion")

	require.NoError(t, database.EnsureSuperAdmin(db, ""))
}
