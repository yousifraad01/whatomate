package handlers_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/shridarpatil/whatomate/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

// Sending a message needs chat:write; a role that can only read contacts
// must not be able to message them.
func TestApp_SendMessage_RequiresChatWrite(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	org := testutil.CreateTestOrganization(t, app.DB)
	readOnly := testutil.CreateTestRoleWithKeys(t, app.DB, org.ID, "contacts-reader", []string{"contacts:read", "chat:read"})
	user := testutil.CreateTestUser(t, app.DB, org.ID,
		testutil.WithEmail(testutil.UniqueEmail("chat-readonly")),
		testutil.WithRoleID(&readOnly.ID),
	)
	contact := testutil.CreateTestContact(t, app.DB, org.ID)

	for _, tc := range []struct {
		name    string
		handler func(*fastglue.Request) error
		body    map[string]any
	}{
		{"SendMessage", app.SendMessage, map[string]any{"content": "hi", "type": "text"}},
		{"SendMediaMessage", app.SendMediaMessage, map[string]any{"contact_id": contact.ID.String()}},
		{"SendReaction", app.SendReaction, map[string]any{"emoji": "+1"}},
		{"SendTemplateMessage", app.SendTemplateMessage, map[string]any{"contact_id": contact.ID.String(), "template_name": "x"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := testutil.NewJSONRequest(t, tc.body)
			testutil.SetAuthContext(req, org.ID, user.ID)
			testutil.SetPathParam(req, "id", contact.ID.String())
			testutil.SetPathParam(req, "message_id", uuid.New().String())
			require.NoError(t, tc.handler(req))
			assert.Equal(t, fasthttp.StatusForbidden, testutil.GetResponseStatusCode(req))
		})
	}
}

// users:write must not be enough to promote oneself: only super admins may
// change their own role.
func TestApp_UpdateUser_CannotChangeOwnRole(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	org := testutil.CreateTestOrganization(t, app.DB)
	adminRole := testutil.CreateAdminRole(t, app.DB, org.ID)
	limited := testutil.CreateTestRoleWithKeys(t, app.DB, org.ID, "user-manager", []string{"users:read", "users:write"})
	user := testutil.CreateTestUser(t, app.DB, org.ID,
		testutil.WithEmail(testutil.UniqueEmail("self-promote")),
		testutil.WithRoleID(&limited.ID),
	)

	req := testutil.NewJSONRequest(t, map[string]any{"role_id": adminRole.ID.String()})
	testutil.SetAuthContext(req, org.ID, user.ID)
	testutil.SetPathParam(req, "id", user.ID.String())
	require.NoError(t, app.UpdateUser(req))
	assert.Equal(t, fasthttp.StatusForbidden, testutil.GetResponseStatusCode(req))

	var reloaded models.User
	require.NoError(t, app.DB.Where("id = ?", user.ID).First(&reloaded).Error)
	require.NotNil(t, reloaded.RoleID)
	assert.Equal(t, limited.ID, *reloaded.RoleID)

	// Sending the same role is a no-op, not an error (profile forms echo it).
	req = testutil.NewJSONRequest(t, map[string]any{"role_id": limited.ID.String(), "full_name": "Still me"})
	testutil.SetAuthContext(req, org.ID, user.ID)
	testutil.SetPathParam(req, "id", user.ID.String())
	require.NoError(t, app.UpdateUser(req))
	assert.Equal(t, fasthttp.StatusOK, testutil.GetResponseStatusCode(req))
}

// A role editor may only grant permissions they hold themselves.
func TestApp_CreateRole_CannotGrantUnheldPermissions(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	org := testutil.CreateTestOrganization(t, app.DB)
	editor := testutil.CreateTestRoleWithKeys(t, app.DB, org.ID, "role-editor", []string{"roles:read", "roles:write", "contacts:read"})
	user := testutil.CreateTestUser(t, app.DB, org.ID,
		testutil.WithEmail(testutil.UniqueEmail("role-editor")),
		testutil.WithRoleID(&editor.ID),
	)

	req := testutil.NewJSONRequest(t, map[string]any{
		"name":        "escalated-" + uuid.NewString()[:8],
		"permissions": []string{"contacts:read", "organizations:write"},
	})
	testutil.SetAuthContext(req, org.ID, user.ID)
	require.NoError(t, app.CreateRole(req))
	assert.Equal(t, fasthttp.StatusForbidden, testutil.GetResponseStatusCode(req))

	req = testutil.NewJSONRequest(t, map[string]any{
		"name":        "subset-" + uuid.NewString()[:8],
		"permissions": []string{"contacts:read"},
	})
	testutil.SetAuthContext(req, org.ID, user.ID)
	require.NoError(t, app.CreateRole(req))
	assert.Equal(t, fasthttp.StatusOK, testutil.GetResponseStatusCode(req))

	// The same rule applies on update.
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(testutil.GetResponseBody(req), &created))
	req = testutil.NewJSONRequest(t, map[string]any{"permissions": []string{"users:write"}})
	testutil.SetAuthContext(req, org.ID, user.ID)
	testutil.SetPathParam(req, "id", created.Data.ID)
	require.NoError(t, app.UpdateRole(req))
	assert.Equal(t, fasthttp.StatusForbidden, testutil.GetResponseStatusCode(req))
}

// Non-super-admins only see the organizations they belong to.
func TestApp_ListOrganizations_MembersOnlySeeTheirOrgs(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	orgA := testutil.CreateTestOrganization(t, app.DB)
	orgB := testutil.CreateTestOrganization(t, app.DB)
	reader := testutil.CreateTestRoleWithKeys(t, app.DB, orgA.ID, "org-reader", []string{"organizations:read"})
	user := testutil.CreateTestUser(t, app.DB, orgA.ID,
		testutil.WithEmail(testutil.UniqueEmail("org-reader")),
		testutil.WithRoleID(&reader.ID),
	)

	listIDs := func(userID uuid.UUID) []string {
		req := testutil.NewGETRequest(t)
		testutil.SetAuthContext(req, orgA.ID, userID)
		require.NoError(t, app.ListOrganizations(req))
		require.Equal(t, fasthttp.StatusOK, testutil.GetResponseStatusCode(req))
		var resp struct {
			Data struct {
				Organizations []struct {
					ID string `json:"id"`
				} `json:"organizations"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(testutil.GetResponseBody(req), &resp))
		ids := make([]string, 0, len(resp.Data.Organizations))
		for _, o := range resp.Data.Organizations {
			ids = append(ids, o.ID)
		}
		return ids
	}

	ids := listIDs(user.ID)
	assert.Contains(t, ids, orgA.ID.String())
	assert.NotContains(t, ids, orgB.ID.String(), "another tenant must not be listed")

	// A super admin still sees everything.
	super := testutil.CreateTestUser(t, app.DB, orgA.ID,
		testutil.WithEmail(testutil.UniqueEmail("org-super")), testutil.WithSuperAdmin())
	assert.Contains(t, listIDs(super.ID), orgB.ID.String())
}

// Agents without contacts:read only see chatbot sessions of contacts assigned
// to them; the endpoint needs chat:read at all.
func TestApp_ListChatbotSessions_AgentSeesAssignedContactsOnly(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	org := testutil.CreateTestOrganization(t, app.DB)
	agentRole := testutil.CreateTestRoleWithKeys(t, app.DB, org.ID, "session-agent", []string{"chat:read", "chat:write"})
	agent := testutil.CreateTestUser(t, app.DB, org.ID,
		testutil.WithEmail(testutil.UniqueEmail("session-agent")),
		testutil.WithRoleID(&agentRole.ID),
	)
	noChat := testutil.CreateTestRoleWithKeys(t, app.DB, org.ID, "no-chat", []string{"analytics:read"})
	outsider := testutil.CreateTestUser(t, app.DB, org.ID,
		testutil.WithEmail(testutil.UniqueEmail("session-outsider")),
		testutil.WithRoleID(&noChat.ID),
	)

	mine := testutil.CreateTestContact(t, app.DB, org.ID)
	require.NoError(t, app.DB.Model(mine).Update("assigned_user_id", agent.ID).Error)
	other := testutil.CreateTestContact(t, app.DB, org.ID)
	for _, c := range []*models.Contact{mine, other} {
		require.NoError(t, app.DB.Create(&models.ChatbotSession{
			OrganizationID: org.ID,
			ContactID:      c.ID,
			PhoneNumber:    c.PhoneNumber,
			Status:         "active",
			SessionData:    models.JSONB{},
		}).Error)
	}

	req := testutil.NewGETRequest(t)
	testutil.SetAuthContext(req, org.ID, agent.ID)
	require.NoError(t, app.ListChatbotSessions(req))
	require.Equal(t, fasthttp.StatusOK, testutil.GetResponseStatusCode(req))
	var resp struct {
		Data struct {
			Sessions []struct {
				ContactID string `json:"contact_id"`
			} `json:"sessions"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(testutil.GetResponseBody(req), &resp))
	require.Len(t, resp.Data.Sessions, 1)
	assert.Equal(t, mine.ID.String(), resp.Data.Sessions[0].ContactID)

	req = testutil.NewGETRequest(t)
	testutil.SetAuthContext(req, org.ID, outsider.ID)
	require.NoError(t, app.ListChatbotSessions(req))
	assert.Equal(t, fasthttp.StatusForbidden, testutil.GetResponseStatusCode(req))
}

// Only the handling agent or a user with transfers:write may resume a
// transferred conversation.
func TestApp_ResumeFromTransfer_RequiresOwnershipOrWrite(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	org := testutil.CreateTestOrganization(t, app.DB)
	agentRole := testutil.CreateAgentRole(t, app.DB, org.ID)
	owner := testutil.CreateTestUser(t, app.DB, org.ID, testutil.WithEmail(testutil.UniqueEmail("resume-owner")), testutil.WithRoleID(&agentRole.ID))
	bystander := testutil.CreateTestUser(t, app.DB, org.ID, testutil.WithEmail(testutil.UniqueEmail("resume-bystander")), testutil.WithRoleID(&agentRole.ID))
	account := testutil.CreateTestWhatsAppAccount(t, app.DB, org.ID)
	contact := testutil.CreateTestContact(t, app.DB, org.ID)
	transfer := createTestTransfer(t, app, org.ID, contact.ID, account.Name, models.TransferStatusActive, &owner.ID)

	req := testutil.NewJSONRequest(t, nil)
	testutil.SetAuthContext(req, org.ID, bystander.ID)
	testutil.SetPathParam(req, "id", transfer.ID.String())
	require.NoError(t, app.ResumeFromTransfer(req))
	assert.Equal(t, fasthttp.StatusForbidden, testutil.GetResponseStatusCode(req))

	req = testutil.NewJSONRequest(t, nil)
	testutil.SetAuthContext(req, org.ID, owner.ID)
	testutil.SetPathParam(req, "id", transfer.ID.String())
	require.NoError(t, app.ResumeFromTransfer(req))
	assert.Equal(t, fasthttp.StatusOK, testutil.GetResponseStatusCode(req))
}

// A campaign must not be re-pointed at another organization's template or
// account.
func TestApp_UpdateCampaign_RejectsForeignTemplate(t *testing.T) {
	t.Parallel()
	mockQueue := testutil.NewMockQueue()
	app := newTestApp(t, withQueue(mockQueue))
	org := testutil.CreateTestOrganization(t, app.DB)
	otherOrg := testutil.CreateTestOrganization(t, app.DB)
	user := testutil.CreateTestUser(t, app.DB, org.ID, testutil.WithEmail(testutil.UniqueEmail("campaign-foreign")))
	account := testutil.CreateTestWhatsAppAccountWith(t, app.DB, org.ID, testutil.WithAccountName("cf-acc-"+uuid.NewString()[:8]))
	template := testutil.CreateTestTemplate(t, app.DB, org.ID, account.Name)
	foreignAccount := testutil.CreateTestWhatsAppAccountWith(t, app.DB, otherOrg.ID, testutil.WithAccountName("cf-foreign-"+uuid.NewString()[:8]))
	foreignTemplate := testutil.CreateTestTemplate(t, app.DB, otherOrg.ID, foreignAccount.Name)
	campaign := createTestCampaign(t, app, org.ID, template.ID, user.ID, account.Name, models.CampaignStatusDraft)

	req := testutil.NewJSONRequest(t, map[string]any{"name": "x", "template_id": foreignTemplate.ID.String()})
	testutil.SetAuthContext(req, org.ID, user.ID)
	testutil.SetPathParam(req, "id", campaign.ID.String())
	require.NoError(t, app.UpdateCampaign(req))
	assert.Equal(t, fasthttp.StatusNotFound, testutil.GetResponseStatusCode(req))

	req = testutil.NewJSONRequest(t, map[string]any{"name": "x", "whatsapp_account": foreignAccount.Name})
	testutil.SetAuthContext(req, org.ID, user.ID)
	testutil.SetPathParam(req, "id", campaign.ID.String())
	require.NoError(t, app.UpdateCampaign(req))
	assert.Equal(t, fasthttp.StatusBadRequest, testutil.GetResponseStatusCode(req))

	var reloaded models.BulkMessageCampaign
	require.NoError(t, app.DB.Where("id = ?", campaign.ID).First(&reloaded).Error)
	assert.Equal(t, template.ID, reloaded.TemplateID)
	assert.Equal(t, account.Name, reloaded.WhatsAppAccount)
}
