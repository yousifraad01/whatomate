import { ref, watch } from 'vue'

export type ColorMode = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'color-mode'

// Module-level state: the theme is global, so every component that calls
// useColorMode() shares one source of truth and one media-query listener
// (previously each mount registered a new listener that was never removed).
const colorMode = ref<ColorMode>('dark')
const isDark = ref(true)
let initialized = false

function readSavedMode(): ColorMode | null {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    return saved === 'light' || saved === 'dark' || saved === 'system' ? saved : null
  } catch {
    return null
  }
}

function systemPrefersDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyTheme() {
  isDark.value = colorMode.value === 'system' ? systemPrefersDark() : colorMode.value === 'dark'

  // Dark-first stylesheet: .light switches the palette, .dark is kept for
  // legacy selectors.
  const root = document.documentElement
  root.classList.toggle('dark', isDark.value)
  root.classList.toggle('light', !isDark.value)
  root.style.colorScheme = isDark.value ? 'dark' : 'light'
}

function init() {
  if (initialized || typeof window === 'undefined') return
  initialized = true

  colorMode.value = readSavedMode() ?? 'dark'
  applyTheme()

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (colorMode.value === 'system') applyTheme()
  })

  watch(colorMode, (mode) => {
    try {
      localStorage.setItem(STORAGE_KEY, mode)
    } catch {
      // Storage may be unavailable (private mode); the theme still applies.
    }
    applyTheme()
  })
}

export function useColorMode() {
  init()

  function setColorMode(mode: ColorMode) {
    colorMode.value = mode
  }

  return {
    colorMode,
    isDark,
    setColorMode
  }
}
