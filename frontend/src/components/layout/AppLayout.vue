<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  MessageSquare,
  PanelLeftClose,
  PanelLeftOpen,
  Menu,
  X
} from 'lucide-vue-next'
import { wsService } from '@/services/websocket'
import { authService } from '@/services/api'
import OrganizationSwitcher from './OrganizationSwitcher.vue'
import UserMenu from './UserMenu.vue'
import ActiveCallPanel from '@/components/calling/ActiveCallPanel.vue'
// Direct import: the shared barrel would pull DataTable, the calendar picker
// and every dialog into the always-loaded layout chunk.
import ScrollToTop from '@/components/shared/ScrollToTop.vue'
import { navigationSections, type NavSection } from './navigation'

const { t } = useI18n()

const route = useRoute()
const authStore = useAuthStore()

const SIDEBAR_STATE_KEY = 'sidebar-collapsed'

function readCollapsed(): boolean {
  try {
    return localStorage.getItem(SIDEBAR_STATE_KEY) === '1'
  } catch {
    return false
  }
}

const isCollapsed = ref(readCollapsed())
const isMobileMenuOpen = ref(false)
const mobileToggleRef = ref<InstanceType<typeof Button> | null>(null)
const navRef = ref<HTMLElement | null>(null)

// Refresh user data and connect WebSocket on mount
onMounted(() => {
  if (authStore.isAuthenticated) {
    // Fetch fresh permissions in background (non-destructive — interceptor handles 401)
    authStore.refreshUserData()

    wsService.connect(async () => {
      try {
        const resp = await authService.getWSToken()
        return resp.data.data.token
      } catch {
        return null
      }
    })
  }
  document.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
})

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && isMobileMenuOpen.value) {
    closeMobileMenu()
  }
}

// The drawer traps neither focus nor scrolling; it does move focus into the
// navigation when opened and gives it back to the toggle when closed, so a
// keyboard user does not lose their place.
async function openMobileMenu() {
  isMobileMenuOpen.value = true
  await nextTick()
  navRef.value?.querySelector<HTMLElement>('a, button')?.focus()
}

function closeMobileMenu() {
  if (!isMobileMenuOpen.value) return
  isMobileMenuOpen.value = false
  const toggle = mobileToggleRef.value?.$el as HTMLElement | undefined
  toggle?.focus?.()
}

function toggleMobileMenu() {
  if (isMobileMenuOpen.value) closeMobileMenu()
  else openMobileMenu()
}

// Close the drawer whenever navigation happens (including browser back)
watch(() => route.fullPath, () => {
  isMobileMenuOpen.value = false
})

function filterItems(items: NavSection['items']) {
  return items
    .filter(item => {
      if (item.childPermissions) {
        return item.childPermissions.some(p => authStore.hasPermission(p, 'read'))
      }
      return !item.permission || authStore.hasPermission(item.permission, 'read')
    })
    .map(item => {
      const filteredChildren = item.children?.filter(
        child => !child.permission || authStore.hasPermission(child.permission, 'read')
      )

      let effectivePath = item.path
      if (item.childPermissions && item.permission && !authStore.hasPermission(item.permission, 'read') && filteredChildren?.length) {
        effectivePath = filteredChildren[0].path
      }

      const originalPath = item.path
      const isActive = originalPath === '/'
        ? route.name === 'dashboard'
        : originalPath === '/chat'
          ? route.name === 'chat' || route.name === 'chat-conversation'
          : route.path.startsWith(originalPath)

      return {
        ...item,
        path: effectivePath,
        active: isActive,
        children: filteredChildren
      }
    })
}

// Filter navigation sections based on user permissions
const navSections = computed(() => {
  return navigationSections
    .map(section => ({
      ...section,
      items: filterItems(section.items)
    }))
    .filter(section => section.items.length > 0)
})

const mainSections = computed(() => navSections.value.filter(s => !s.pinBottom))
const bottomSections = computed(() => navSections.value.filter(s => s.pinBottom))

const toggleSidebar = () => {
  isCollapsed.value = !isCollapsed.value
  try {
    localStorage.setItem(SIDEBAR_STATE_KEY, isCollapsed.value ? '1' : '0')
  } catch {
    // Persistence is a convenience only
  }
}

const handleLogout = async () => {
  await authStore.logout()
  // Full navigation instead of a client-side route: it discards every Pinia
  // store (contacts, transfers, users, ...) so a different account signing
  // in on this browser never sees the previous session's data.
  const basePath = ((window as any).__BASE_PATH__ ?? '').replace(/\/$/, '')
  window.location.assign(`${basePath}/login`)
}

const navLinkClass = (active: boolean) => [
  'nav-active-indicator flex items-center gap-2.5 rounded-md px-2.5 py-2 text-[13px] font-medium transition-colors',
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
  active
    ? 'bg-sidebar-accent text-sidebar-foreground'
    : 'text-sidebar-muted hover:bg-sidebar-accent/70 hover:text-sidebar-foreground',
  isCollapsed.value && 'md:justify-center md:px-2'
]

const childLinkClass = (active: boolean) => [
  'ml-4 flex items-center gap-2.5 rounded-md px-2.5 py-1.5 text-[13px] transition-colors',
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
  active
    ? 'bg-sidebar-accent font-medium text-sidebar-foreground'
    : 'text-sidebar-muted hover:bg-sidebar-accent/70 hover:text-sidebar-foreground'
]
</script>

<template>
  <div class="flex h-screen bg-background">
    <!-- Skip link for accessibility -->
    <a href="#main-content" class="skip-link">{{ t('nav.skipToMain') }}</a>

    <!-- Mobile header -->
    <header class="fixed left-0 right-0 top-0 z-50 flex h-12 items-center justify-between border-b border-sidebar-border bg-sidebar px-3 md:hidden">
      <RouterLink to="/" class="flex items-center gap-2 rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
        <span class="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground" aria-hidden="true">
          <MessageSquare class="h-4 w-4" />
        </span>
        <span class="text-sm font-semibold text-sidebar-foreground">Whatomate</span>
      </RouterLink>
      <Button
        ref="mobileToggleRef"
        variant="ghost"
        size="icon"
        class="h-8 w-8 text-sidebar-foreground"
        :aria-label="isMobileMenuOpen ? t('nav.closeMenu') : t('nav.toggleMenu')"
        :aria-expanded="isMobileMenuOpen"
        aria-controls="app-sidebar"
        @click="toggleMobileMenu"
      >
        <X v-if="isMobileMenuOpen" class="h-5 w-5" aria-hidden="true" />
        <Menu v-else class="h-5 w-5" aria-hidden="true" />
      </Button>
    </header>

    <!-- Mobile menu overlay -->
    <div
      v-if="isMobileMenuOpen"
      class="fixed inset-0 z-40 bg-black/50 md:hidden"
      aria-hidden="true"
      @click="closeMobileMenu"
    />

    <!-- Sidebar -->
    <aside
      id="app-sidebar"
      :class="[
        'flex flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-[transform,width] duration-200',
        'fixed inset-y-0 left-0 z-40 md:relative',
        isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0',
        isCollapsed ? 'w-64 md:w-14' : 'w-64'
      ]"
    >
      <!-- Brand + collapse control (hidden on mobile, shown in header instead) -->
      <div class="hidden h-12 items-center justify-between border-b border-sidebar-border px-3 md:flex">
        <RouterLink to="/" class="flex items-center gap-2 rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" :aria-label="t('nav.dashboard')">
          <span class="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground" aria-hidden="true">
            <MessageSquare class="h-4 w-4" />
          </span>
          <span v-if="!isCollapsed" class="text-sm font-semibold">Whatomate</span>
        </RouterLink>
        <Button
          v-if="!isCollapsed"
          variant="ghost"
          size="icon"
          class="h-7 w-7 text-sidebar-muted hover:text-sidebar-foreground"
          :aria-label="t('nav.collapseSidebar')"
          :aria-expanded="true"
          aria-controls="app-sidebar"
          @click="toggleSidebar"
        >
          <PanelLeftClose class="h-4 w-4" aria-hidden="true" />
        </Button>
      </div>
      <!-- Expand control when collapsed -->
      <div v-if="isCollapsed" class="hidden justify-center py-1 md:flex">
        <Button
          variant="ghost"
          size="icon"
          class="h-7 w-7 text-sidebar-muted hover:text-sidebar-foreground"
          :aria-label="t('nav.expandSidebar')"
          :aria-expanded="false"
          aria-controls="app-sidebar"
          @click="toggleSidebar"
        >
          <PanelLeftOpen class="h-4 w-4" aria-hidden="true" />
        </Button>
      </div>
      <!-- Mobile logo spacer -->
      <div class="h-12 md:hidden" />

      <!-- Organization Switcher (super admins and multi-org members) -->
      <OrganizationSwitcher :collapsed="isCollapsed" />

      <!-- Navigation -->
      <ScrollArea class="flex-1 py-2">
        <nav ref="navRef" class="px-2" :aria-label="t('nav.mainNavigation')">
          <template v-for="(section, sIdx) in mainSections" :key="section.label">
            <!-- Section header -->
            <h2
              v-if="section.label && !isCollapsed"
              :class="['px-2.5 pb-1 pt-4 text-[11px] font-semibold uppercase tracking-wider text-sidebar-muted', sIdx === 0 && 'pt-1']"
            >
              {{ t(section.label) }}
            </h2>
            <div v-else-if="sIdx > 0" :class="['mx-2.5 my-2 border-t border-sidebar-border', isCollapsed && 'mx-1']" role="presentation" />

            <!-- Section items -->
            <ul class="space-y-0.5">
              <li v-for="item in section.items" :key="item.path">
                <RouterLink
                  :to="item.path"
                  :class="navLinkClass(item.active)"
                  :data-active="item.active"
                  :aria-current="item.active ? 'page' : undefined"
                  :title="isCollapsed ? t(item.name) : undefined"
                >
                  <component :is="item.icon" class="h-4 w-4 shrink-0" aria-hidden="true" />
                  <span :class="isCollapsed && 'md:sr-only'">{{ t(item.name) }}</span>
                </RouterLink>

                <!-- Submenu items -->
                <ul v-if="item.children && item.active && !isCollapsed" class="mt-0.5 space-y-0.5">
                  <li v-for="child in item.children" :key="child.path">
                    <RouterLink
                      :to="child.path"
                      :class="childLinkClass(route.path === child.path)"
                      :aria-current="route.path === child.path ? 'page' : undefined"
                    >
                      <component :is="child.icon" class="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
                      <span>{{ t(child.name) }}</span>
                    </RouterLink>
                  </li>
                </ul>
              </li>
            </ul>
          </template>
        </nav>
      </ScrollArea>

      <!-- Bottom-pinned navigation (Settings) -->
      <nav v-if="bottomSections.length > 0" class="border-t border-sidebar-border px-2 py-2" :aria-label="t('nav.secondaryNavigation')">
        <ul v-for="section in bottomSections" :key="section.label" class="space-y-0.5">
          <li v-for="item in section.items" :key="item.path">
            <RouterLink
              :to="item.path"
              :class="navLinkClass(item.active)"
              :data-active="item.active"
              :aria-current="item.active ? 'page' : undefined"
              :title="isCollapsed ? t(item.name) : undefined"
            >
              <component :is="item.icon" class="h-4 w-4 shrink-0" aria-hidden="true" />
              <span :class="isCollapsed && 'md:sr-only'">{{ t(item.name) }}</span>
            </RouterLink>

            <ul v-if="item.children && item.active && !isCollapsed" class="mt-0.5 max-h-[40vh] space-y-0.5 overflow-y-auto">
              <li v-for="child in item.children" :key="child.path">
                <RouterLink
                  :to="child.path"
                  :class="childLinkClass(route.path === child.path)"
                  :aria-current="route.path === child.path ? 'page' : undefined"
                >
                  <component :is="child.icon" class="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
                  <span>{{ t(child.name) }}</span>
                </RouterLink>
              </li>
            </ul>
          </li>
        </ul>
      </nav>

      <!-- User Menu -->
      <UserMenu :collapsed="isCollapsed" @logout="handleLogout" />
    </aside>

    <!-- Main content -->
    <main id="main-content" class="flex-1 overflow-hidden bg-background pt-12 md:pt-0">
      <RouterView v-slot="{ Component, route: viewRoute }">
        <Transition name="page" mode="out-in">
          <component :is="Component" :key="viewRoute.meta.stableKey ? String(viewRoute.name) : viewRoute.path" />
        </Transition>
      </RouterView>
      <ActiveCallPanel />
      <ScrollToTop />
    </main>
  </div>
</template>
