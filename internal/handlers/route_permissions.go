package handlers

import (
	"strings"

	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

// RoutePermission is the resource permission a route requires.
type RoutePermission struct {
	Resource string
	Action   string
}

type routeRule struct {
	method   string
	segments []string
	perm     RoutePermission
}

func rule(method, pattern, resource, action string) routeRule {
	return routeRule{
		method:   method,
		segments: splitRoutePath(pattern),
		perm:     RoutePermission{Resource: resource, Action: action},
	}
}

// routeRules maps routes to the permission they require. It is enforced by
// App.RequireRoutePermission as a safety net in addition to the handler-level
// requireAuth calls, so a handler that omits its own check can still not be
// reached without the permission.
//
// Read routes that the chat workspace depends on (templates, custom actions,
// canned responses, WhatsApp flows, chatbot settings, org settings) are
// deliberately not listed: the built-in agent role has no read grant on those
// resources but legitimately uses them while chatting.
var routeRules = []routeRule{
	// Roles
	rule("POST", "/api/roles", models.ResourceRoles, models.ActionWrite),
	rule("PUT", "/api/roles/{id}", models.ResourceRoles, models.ActionWrite),
	rule("DELETE", "/api/roles/{id}", models.ResourceRoles, models.ActionDelete),

	// Users
	rule("GET", "/api/users/{id}", models.ResourceUsers, models.ActionRead),

	// Campaigns
	rule("GET", "/api/campaigns", models.ResourceCampaigns, models.ActionRead),
	rule("POST", "/api/campaigns", models.ResourceCampaigns, models.ActionWrite),
	rule("GET", "/api/campaigns/{id}", models.ResourceCampaigns, models.ActionRead),
	rule("PUT", "/api/campaigns/{id}", models.ResourceCampaigns, models.ActionWrite),
	rule("DELETE", "/api/campaigns/{id}", models.ResourceCampaigns, models.ActionDelete),
	rule("POST", "/api/campaigns/{id}/start", models.ResourceCampaigns, models.ActionExecute),
	rule("POST", "/api/campaigns/{id}/pause", models.ResourceCampaigns, models.ActionExecute),
	rule("POST", "/api/campaigns/{id}/cancel", models.ResourceCampaigns, models.ActionExecute),
	rule("POST", "/api/campaigns/{id}/retry-failed", models.ResourceCampaigns, models.ActionExecute),
	rule("GET", "/api/campaigns/{id}/progress", models.ResourceCampaigns, models.ActionRead),
	rule("POST", "/api/campaigns/{id}/recipients/import", models.ResourceCampaigns, models.ActionWrite),
	rule("GET", "/api/campaigns/{id}/recipients", models.ResourceCampaigns, models.ActionRead),
	rule("DELETE", "/api/campaigns/{id}/recipients/{recipientId}", models.ResourceCampaigns, models.ActionWrite),
	rule("POST", "/api/campaigns/{id}/media", models.ResourceCampaigns, models.ActionWrite),
	rule("GET", "/api/campaigns/{id}/media", models.ResourceCampaigns, models.ActionRead),

	// Templates (list/get stay open: agents send templates from chat)
	rule("POST", "/api/templates", models.ResourceTemplates, models.ActionWrite),
	rule("POST", "/api/templates/sync", models.ResourceTemplates, models.ActionSync),
	rule("POST", "/api/templates/upload-media", models.ResourceTemplates, models.ActionWrite),
	rule("PUT", "/api/templates/{id}", models.ResourceTemplates, models.ActionWrite),
	rule("DELETE", "/api/templates/{id}", models.ResourceTemplates, models.ActionDelete),
	rule("POST", "/api/templates/{id}/publish", models.ResourceTemplates, models.ActionWrite),

	// WhatsApp flows
	rule("POST", "/api/flows", models.ResourceFlowsWhatsApp, models.ActionWrite),
	rule("POST", "/api/flows/sync", models.ResourceFlowsWhatsApp, models.ActionWrite),
	rule("PUT", "/api/flows/{id}", models.ResourceFlowsWhatsApp, models.ActionWrite),
	rule("DELETE", "/api/flows/{id}", models.ResourceFlowsWhatsApp, models.ActionDelete),
	rule("POST", "/api/flows/{id}/save-to-meta", models.ResourceFlowsWhatsApp, models.ActionWrite),
	rule("POST", "/api/flows/{id}/publish", models.ResourceFlowsWhatsApp, models.ActionWrite),
	rule("POST", "/api/flows/{id}/deprecate", models.ResourceFlowsWhatsApp, models.ActionWrite),
	rule("POST", "/api/flows/{id}/duplicate", models.ResourceFlowsWhatsApp, models.ActionWrite),

	// Chatbot configuration
	rule("PUT", "/api/chatbot/settings", models.ResourceSettingsChatbot, models.ActionWrite),
	rule("POST", "/api/chatbot/keywords", models.ResourceChatbotKeywords, models.ActionWrite),
	rule("PUT", "/api/chatbot/keywords/{id}", models.ResourceChatbotKeywords, models.ActionWrite),
	rule("DELETE", "/api/chatbot/keywords/{id}", models.ResourceChatbotKeywords, models.ActionDelete),
	rule("POST", "/api/chatbot/flows", models.ResourceFlowsChatbot, models.ActionWrite),
	rule("PUT", "/api/chatbot/flows/{id}", models.ResourceFlowsChatbot, models.ActionWrite),
	rule("DELETE", "/api/chatbot/flows/{id}", models.ResourceFlowsChatbot, models.ActionDelete),
	rule("POST", "/api/chatbot/ai-contexts", models.ResourceChatbotAI, models.ActionWrite),
	rule("PUT", "/api/chatbot/ai-contexts/{id}", models.ResourceChatbotAI, models.ActionWrite),
	rule("DELETE", "/api/chatbot/ai-contexts/{id}", models.ResourceChatbotAI, models.ActionDelete),
	rule("POST", "/api/chatbot/transfers", models.ResourceTransfers, models.ActionWrite),

	// Outbound webhooks (subscribing a URL to message events exposes conversations)
	rule("GET", "/api/webhooks", models.ResourceWebhooks, models.ActionRead),
	rule("POST", "/api/webhooks", models.ResourceWebhooks, models.ActionWrite),
	rule("GET", "/api/webhooks/{id}", models.ResourceWebhooks, models.ActionRead),
	rule("PUT", "/api/webhooks/{id}", models.ResourceWebhooks, models.ActionWrite),
	rule("DELETE", "/api/webhooks/{id}", models.ResourceWebhooks, models.ActionDelete),
	rule("POST", "/api/webhooks/{id}/test", models.ResourceWebhooks, models.ActionWrite),

	// Custom actions (list/execute stay open for chat)
	rule("POST", "/api/custom-actions", models.ResourceCustomActions, models.ActionWrite),
	rule("PUT", "/api/custom-actions/{id}", models.ResourceCustomActions, models.ActionWrite),
	rule("DELETE", "/api/custom-actions/{id}", models.ResourceCustomActions, models.ActionDelete),

	// Canned responses (list/use stay open for chat)
	rule("POST", "/api/canned-responses", models.ResourceCannedResponses, models.ActionWrite),
	rule("PUT", "/api/canned-responses/{id}", models.ResourceCannedResponses, models.ActionWrite),
	rule("DELETE", "/api/canned-responses/{id}", models.ResourceCannedResponses, models.ActionDelete),

	// SSO settings
	rule("GET", "/api/settings/sso", models.ResourceSettingsSSO, models.ActionRead),
	rule("PUT", "/api/settings/sso/{provider}", models.ResourceSettingsSSO, models.ActionWrite),
	rule("DELETE", "/api/settings/sso/{provider}", models.ResourceSettingsSSO, models.ActionWrite),

	// Organization settings
	rule("PUT", "/api/org/settings", models.ResourceSettingsGeneral, models.ActionWrite),
	rule("POST", "/api/org/audio", models.ResourceSettingsGeneral, models.ActionWrite),

	// WhatsApp accounts
	rule("POST", "/api/accounts/exchange-token", models.ResourceAccounts, models.ActionWrite),
	rule("POST", "/api/accounts/{id}/register", models.ResourceAccounts, models.ActionWrite),
	rule("POST", "/api/accounts/{id}/test", models.ResourceAccounts, models.ActionRead),
	rule("POST", "/api/accounts/{id}/subscribe", models.ResourceAccounts, models.ActionWrite),
	rule("GET", "/api/accounts/{id}/business_profile", models.ResourceAccounts, models.ActionRead),
	rule("PUT", "/api/accounts/{id}/business_profile", models.ResourceAccounts, models.ActionWrite),
	rule("POST", "/api/accounts/{id}/business_profile/photo", models.ResourceAccounts, models.ActionWrite),

	// Chat: sending on behalf of the organization needs chat:write; the
	// handlers additionally restrict agents to their assigned contacts.
	rule("POST", "/api/contacts/{id}/messages", models.ResourceChat, models.ActionWrite),
	rule("POST", "/api/contacts/{id}/messages/{message_id}/reaction", models.ResourceChat, models.ActionWrite),
	rule("POST", "/api/contacts/{id}/mark-read", models.ResourceChat, models.ActionRead),
	rule("POST", "/api/messages", models.ResourceChat, models.ActionWrite),
	rule("POST", "/api/messages/template", models.ResourceChat, models.ActionWrite),
	rule("POST", "/api/messages/media", models.ResourceChat, models.ActionWrite),

	// Chatbot sessions expose collected answers and contact details
	rule("GET", "/api/chatbot/sessions", models.ResourceChat, models.ActionRead),
	rule("GET", "/api/chatbot/sessions/{id}", models.ResourceChat, models.ActionRead),

	// AI contexts carry API configuration (headers with credentials)
	rule("GET", "/api/chatbot/ai-contexts", models.ResourceChatbotAI, models.ActionRead),
	rule("GET", "/api/chatbot/ai-contexts/{id}", models.ResourceChatbotAI, models.ActionRead),

	// Commerce catalogs are managed with the WhatsApp account's credentials,
	// so they follow the accounts permissions (there is no catalog resource).
	rule("GET", "/api/catalogs", models.ResourceAccounts, models.ActionRead),
	rule("POST", "/api/catalogs", models.ResourceAccounts, models.ActionWrite),
	rule("POST", "/api/catalogs/sync", models.ResourceAccounts, models.ActionWrite),
	rule("GET", "/api/catalogs/{id}", models.ResourceAccounts, models.ActionRead),
	rule("DELETE", "/api/catalogs/{id}", models.ResourceAccounts, models.ActionWrite),
	rule("GET", "/api/catalogs/{id}/products", models.ResourceAccounts, models.ActionRead),
	rule("POST", "/api/catalogs/{id}/products", models.ResourceAccounts, models.ActionWrite),
	rule("GET", "/api/products/{id}", models.ResourceAccounts, models.ActionRead),
	rule("PUT", "/api/products/{id}", models.ResourceAccounts, models.ActionWrite),
	rule("DELETE", "/api/products/{id}", models.ResourceAccounts, models.ActionWrite),

	// Analytics & dashboard
	rule("GET", "/api/analytics/dashboard", models.ResourceAnalytics, models.ActionRead),
	rule("GET", "/api/widgets/data", models.ResourceAnalytics, models.ActionRead),
	rule("GET", "/api/widgets/{id}/data", models.ResourceAnalytics, models.ActionRead),
	rule("POST", "/api/widgets/layout", models.ResourceAnalytics, models.ActionWrite),
}

// RouteRequiredPermission returns the permission required to call method+path,
// or ok=false when the route is not governed by routeRules. When several
// patterns match, the one with the most literal segments wins so a literal
// route such as /api/widgets/data is never shadowed by /api/widgets/{id}.
func RouteRequiredPermission(method, path string) (perm RoutePermission, ok bool) {
	segs := splitRoutePath(path)
	best := -1
	for _, rr := range routeRules {
		if rr.method != method {
			continue
		}
		score, matched := matchSegments(rr.segments, segs)
		if matched && score > best {
			best = score
			perm = rr.perm
		}
	}
	return perm, best >= 0
}

func splitRoutePath(p string) []string {
	p = strings.TrimSuffix(strings.TrimPrefix(p, "/"), "/")
	return strings.Split(p, "/")
}

// matchSegments reports whether the request segments match the pattern
// segments ({name} matches any non-empty segment) and how many literal
// segments matched.
func matchSegments(pattern, segs []string) (int, bool) {
	if len(pattern) != len(segs) {
		return 0, false
	}
	literal := 0
	for i, ps := range pattern {
		if strings.HasPrefix(ps, "{") && strings.HasSuffix(ps, "}") {
			if segs[i] == "" {
				return 0, false
			}
			continue
		}
		if ps != segs[i] {
			return 0, false
		}
		literal++
	}
	return literal, true
}

// RequireRoutePermission is a fastglue "before" middleware that enforces
// routeRules for the current request. Routes not present in the table pass
// through untouched; listed routes require an authenticated user holding the
// permission in the active organization (super admins always pass).
func (a *App) RequireRoutePermission(r *fastglue.Request) *fastglue.Request {
	method := string(r.RequestCtx.Method())
	if method == fasthttp.MethodOptions {
		return r
	}

	perm, ok := RouteRequiredPermission(method, string(r.RequestCtx.Path()))
	if !ok {
		return r
	}

	orgID, userID, err := a.getOrgAndUserID(r)
	if err != nil {
		_ = r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
		return nil
	}
	if !a.HasPermission(userID, perm.Resource, perm.Action, orgID) {
		_ = r.SendErrorEnvelope(fasthttp.StatusForbidden, "Insufficient permissions", nil, "")
		return nil
	}
	return r
}
