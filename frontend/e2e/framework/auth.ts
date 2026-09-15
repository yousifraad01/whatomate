/**
 * UI login helpers used by the new framework.
 *
 * These deliberately do not wrap the existing `loginAsAdmin` from
 * `helpers/auth.ts` — that one targets `admin@test.com`, which lacks
 * permissions like `analytics:read` and silently lands users on the
 * "not allowed" page (cost ~30 minutes of debugging earlier this month).
 *
 * For framework-level tests prefer `loginAsSuperAdmin` (admin@admin.com,
 * always succeeds) or `loginAs(page, customCreds)` for permission-scoped
 * users created via `createUserWithPermissions`.
 */

import type { Page } from '@playwright/test'

export const SUPER_ADMIN = {
  email: 'admin@admin.com',
  password: 'admin',
} as const

export interface Credentials {
  email: string
  password: string
}

export async function loginAs(page: Page, creds: Credentials): Promise<void> {
  // Clear any existing session to ensure the router doesn't intercept the /login
  // navigation and redirect us away if the browser already has an active session.
  await page.context().clearCookies()
  try {
    await page.evaluate(() => window.localStorage.clear())
  } catch {
    // Ignore if on about:blank
  }

  // Clearing the cookies can make an in-flight request of the previous page
  // fail with 401 and the app redirects to /login at the same moment, which
  // aborts our navigation (net::ERR_ABORTED). Retry once.
  // Leave the previous SPA document first: once its cookies are gone, any
  // in-flight request 401s and the app's interceptor redirects to /login on
  // its own, which interrupts our navigation.
  try {
    await page.goto('about:blank')
  } catch {
    // interrupted by the app's own redirect; either way the old document is gone
  }
  try {
    await page.goto('/login', { waitUntil: 'domcontentloaded' })
  } catch {
    await page.goto('/login', { waitUntil: 'domcontentloaded' })
  }
  await page.locator('input[type="email"], input[name="email"]').fill(creds.email)
  await page.locator('input[type="password"], input[name="password"]').fill(creds.password)
  await page.locator('button[type="submit"]').click()
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 10_000 })
}

export async function loginAsSuperAdmin(page: Page): Promise<void> {
  await loginAs(page, SUPER_ADMIN)
}
