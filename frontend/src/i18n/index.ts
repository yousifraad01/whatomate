import { createI18n } from 'vue-i18n'
import en from './locales/en.json'

export type MessageSchema = typeof en

// English ships in the entry bundle (it is the fallback for every other
// locale). The remaining locale files are code-split and fetched only when
// selected: bundling all of them eagerly put ~690 KB of JSON into the initial
// chunk regardless of the user's language.
const localeLoaders = import.meta.glob('./locales/*.json', { import: 'default' }) as Record<
  string,
  () => Promise<MessageSchema>
>

const localeNames: Record<string, { name: string; nativeName: string }> = {
  en: { name: 'English', nativeName: 'English' },
  es: { name: 'Spanish', nativeName: 'Español' },
  fr: { name: 'French', nativeName: 'Français' },
  de: { name: 'German', nativeName: 'Deutsch' },
  hi: { name: 'Hindi', nativeName: 'हिंदी' },
  pt: { name: 'Portuguese', nativeName: 'Português' },
  zh: { name: 'Chinese', nativeName: '中文' },
  ja: { name: 'Japanese', nativeName: '日本語' },
  ko: { name: 'Korean', nativeName: '한국어' },
  ar: { name: 'Arabic', nativeName: 'العربية' },
  ru: { name: 'Russian', nativeName: 'Русский' },
  it: { name: 'Italian', nativeName: 'Italiano' },
  nl: { name: 'Dutch', nativeName: 'Nederlands' },
  tr: { name: 'Turkish', nativeName: 'Türkçe' },
  vi: { name: 'Vietnamese', nativeName: 'Tiếng Việt' },
  th: { name: 'Thai', nativeName: 'ไทย' },
  id: { name: 'Indonesian', nativeName: 'Bahasa Indonesia' },
  ms: { name: 'Malay', nativeName: 'Bahasa Melayu' },
  pl: { name: 'Polish', nativeName: 'Polski' },
  uk: { name: 'Ukrainian', nativeName: 'Українська' },
  ta: { name: 'Tamil', nativeName: 'தமிழ்' },
}

function codeFromPath(path: string): string {
  return path.replace('./locales/', '').replace('.json', '')
}

// Auto-generate SUPPORTED_LOCALES from available files
export const SUPPORTED_LOCALES = Object.keys(localeLoaders)
  .map(codeFromPath)
  .sort()
  .map(code => {
    const names = localeNames[code] || { name: code, nativeName: code }
    return { code, ...names }
  })

export type SupportedLocale = string

const availableCodes = new Set(SUPPORTED_LOCALES.map(l => l.code))
const loadedLocales = new Set<string>(['en'])

export function isSupportedLocale(code: string): boolean {
  return availableCodes.has(code)
}

// Get saved locale or detect from browser
export function getDefaultLocale(): string {
  // Check localStorage first
  let saved: string | null = null
  try {
    saved = localStorage.getItem('locale')
  } catch {
    saved = null
  }
  if (saved && isSupportedLocale(saved)) {
    return saved
  }

  // Detect from browser
  const browserLang = navigator.language.split('-')[0]
  if (isSupportedLocale(browserLang)) {
    return browserLang
  }

  return 'en'
}

export const i18n = createI18n({
  legacy: false, // Use Composition API
  locale: 'en',
  fallbackLocale: 'en',
  messages: { en } as Record<string, MessageSchema>,
})

/**
 * Ensure a locale's messages are registered, fetching the chunk on first use.
 * Resolves to false when the locale is unknown.
 */
export async function loadLocale(code: string): Promise<boolean> {
  if (loadedLocales.has(code)) return true
  const loader = localeLoaders[`./locales/${code}.json`]
  if (!loader) return false
  const messages = await loader()
  i18n.global.setLocaleMessage(code, messages)
  loadedLocales.add(code)
  return true
}

// Helper to change locale
export async function setLocale(locale: string): Promise<void> {
  if (!(await loadLocale(locale))) {
    console.warn(`Locale '${locale}' not available`)
    return
  }
  i18n.global.locale.value = locale
  try {
    localStorage.setItem('locale', locale)
  } catch {
    // Storage may be unavailable; the locale still applies for this session.
  }
  document.documentElement.setAttribute('lang', locale)
}

// Get current locale
export function getLocale(): string {
  return i18n.global.locale.value
}
