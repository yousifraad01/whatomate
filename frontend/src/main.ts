import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { i18n, getDefaultLocale, setLocale } from './i18n'

import './assets/fonts.css'
import './assets/index.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)

// ---------------------------------------------------------------------------
// Recover from a deployment that happened while this tab was open.
//
// Route views are lazy chunks with content hashes in their file names. After
// a new build the old names no longer exist, so navigating to a route that
// was not loaded yet fails with "Failed to fetch dynamically imported module"
// and the page looks broken. Reloading once fetches the new index.html and
// its chunk names. A timestamp in sessionStorage prevents a reload loop when
// the failure has another cause (for example the server being down).
// ---------------------------------------------------------------------------
const RELOAD_KEY = 'whatomate:chunk-reload-at'
const RELOAD_COOLDOWN_MS = 30_000

function isChunkLoadError(error: unknown): boolean {
  const message = error instanceof Error ? error.message : String(error ?? '')
  return /Failed to fetch dynamically imported module|Importing a module script failed|error loading dynamically imported module|Unable to preload CSS/i.test(message)
}

function reloadOnceForNewBuild(): boolean {
  let last = 0
  try {
    last = Number(sessionStorage.getItem(RELOAD_KEY) || 0)
  } catch {
    // sessionStorage unavailable: fall through and reload once
  }
  if (Date.now() - last < RELOAD_COOLDOWN_MS) return false
  try {
    sessionStorage.setItem(RELOAD_KEY, String(Date.now()))
  } catch {
    // ignore
  }
  window.location.reload()
  return true
}

window.addEventListener('vite:preloadError', (event) => {
  if (reloadOnceForNewBuild()) event.preventDefault()
})

router.onError((error) => {
  if (isChunkLoadError(error)) reloadOnceForNewBuild()
})

// Only the English bundle ships in the entry chunk; other locales are fetched
// on demand. Load the saved/browser locale before mounting so the first paint
// is already translated.
setLocale(getDefaultLocale()).finally(() => {
  app.mount('#app')
})
