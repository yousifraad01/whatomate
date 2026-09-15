package handlers_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shridarpatil/whatomate/internal/handlers"
	"github.com/shridarpatil/whatomate/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

func TestRouteRequiredPermission(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	cases := []struct {
		method, path     string
		resource, action string
		governed         bool
	}{
		{"POST", "/api/roles", "roles", "write", true},
		{"PUT", "/api/roles/" + id, "roles", "write", true},
		{"DELETE", "/api/roles/" + id, "roles", "delete", true},
		{"GET", "/api/roles", "", "", false},
		{"GET", "/api/templates", "", "", false},
		{"POST", "/api/templates/sync", "templates", "sync", true},
		{"POST", "/api/templates/" + id + "/publish", "templates", "write", true},
		{"POST", "/api/campaigns/" + id + "/start", "campaigns", "execute", true},
		{"DELETE", "/api/campaigns/" + id + "/recipients/" + id, "campaigns", "write", true},
		{"GET", "/api/widgets/data", "analytics", "read", true},
		{"GET", "/api/widgets/" + id + "/data", "analytics", "read", true},
		{"GET", "/api/widgets", "", "", false},
		{"POST", "/api/custom-actions/" + id + "/execute", "", "", false},
		{"GET", "/api/custom-actions", "", "", false},
		{"PUT", "/api/org/settings", "settings.general", "write", true},
		{"PUT", "/api/settings/sso/google", "settings.sso", "write", true},
		{"POST", "/api/chatbot/transfers", "transfers", "write", true},
		{"PUT", "/api/chatbot/transfers/" + id + "/resume", "", "", false},
		{"POST", "/api/webhooks/" + id + "/test", "webhooks", "write", true},
		{"GET", "/health", "", "", false},
		{"POST", "/api/roles/", "roles", "write", true}, // trailing slash tolerated
	}

	for _, tc := range cases {
		perm, ok := handlers.RouteRequiredPermission(tc.method, tc.path)
		assert.Equal(t, tc.governed, ok, "%s %s governed", tc.method, tc.path)
		if tc.governed {
			assert.Equal(t, tc.resource, perm.Resource, "%s %s resource", tc.method, tc.path)
			assert.Equal(t, tc.action, perm.Action, "%s %s action", tc.method, tc.path)
		}
	}
}

func TestRequireRoutePermission(t *testing.T) {
	app := newTestApp(t)
	org := testutil.CreateTestOrganization(t, app.DB)

	limitedRole := testutil.CreateTestRoleWithKeys(t, app.DB, org.ID, "rp-limited", []string{"chat:read"})
	editorRole := testutil.CreateTestRoleWithKeys(t, app.DB, org.ID, "rp-editor", []string{"roles:write"})
	limitedUser := testutil.CreateTestUser(t, app.DB, org.ID, testutil.WithRoleID(&limitedRole.ID))
	editorUser := testutil.CreateTestUser(t, app.DB, org.ID, testutil.WithRoleID(&editorRole.ID))
	superAdmin := testutil.CreateTestUser(t, app.DB, org.ID, testutil.WithSuperAdmin())

	newReq := func(method, uri string) *fastglue.Request {
		req := testutil.NewRequest(t)
		req.RequestCtx.Request.Header.SetMethod(method)
		req.RequestCtx.Request.SetRequestURI(uri)
		return req
	}

	t.Run("denies a governed route without the permission", func(t *testing.T) {
		req := newReq("POST", "/api/roles")
		testutil.SetAuthContext(req, org.ID, limitedUser.ID)
		require.Nil(t, app.RequireRoutePermission(req))
		assert.Equal(t, fasthttp.StatusForbidden, testutil.GetResponseStatusCode(req))
	})

	t.Run("allows a governed route with the permission", func(t *testing.T) {
		req := newReq("POST", "/api/roles")
		testutil.SetAuthContext(req, org.ID, editorUser.ID)
		require.NotNil(t, app.RequireRoutePermission(req))
	})

	t.Run("super admins always pass", func(t *testing.T) {
		req := newReq("DELETE", "/api/roles/"+uuid.NewString())
		testutil.SetAuthContext(req, org.ID, superAdmin.ID)
		require.NotNil(t, app.RequireRoutePermission(req))
	})

	t.Run("ungoverned routes pass through", func(t *testing.T) {
		req := newReq("GET", "/api/templates")
		testutil.SetAuthContext(req, org.ID, limitedUser.ID)
		require.NotNil(t, app.RequireRoutePermission(req))
	})

	t.Run("governed route without auth context is unauthorized", func(t *testing.T) {
		req := newReq("PUT", "/api/org/settings")
		require.Nil(t, app.RequireRoutePermission(req))
		assert.Equal(t, fasthttp.StatusUnauthorized, testutil.GetResponseStatusCode(req))
	})

	t.Run("preflight requests are never blocked", func(t *testing.T) {
		req := newReq("OPTIONS", "/api/roles")
		require.NotNil(t, app.RequireRoutePermission(req))
	})
}
