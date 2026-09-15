<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useContactsStore } from '@/stores/contacts'
import { usersService, chatbotService } from '@/services/api'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import {
  Popover,
  PopoverContent,
  PopoverTrigger
} from '@/components/ui/popover'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle
} from '@/components/ui/alert-dialog'
import { LogOut, User } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { getInitials } from '@/lib/utils'
import ThemeSwitcher from './ThemeSwitcher.vue'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'

const { t } = useI18n()

defineProps<{
  collapsed?: boolean
}>()

const emit = defineEmits<{
  logout: []
}>()

const authStore = useAuthStore()
const contactsStore = useContactsStore()
const isUserMenuOpen = ref(false)
const isUpdatingAvailability = ref(false)
const isCheckingTransfers = ref(false)
const showAwayWarning = ref(false)
const awayWarningTransferCount = ref(0)

const handleAvailabilityChange = async (checked: boolean) => {
  if (!checked) {
    isCheckingTransfers.value = true
    try {
      const response = await chatbotService.listTransfers({ status: 'active' })
      const data = response.data.data || response.data
      const transfers = data.transfers || []
      const userId = authStore.user?.id
      const myActiveTransfers = transfers.filter((t: any) => t.agent_id === userId)

      if (myActiveTransfers.length > 0) {
        awayWarningTransferCount.value = myActiveTransfers.length
        showAwayWarning.value = true
        return
      }
    } catch (error) {
      console.error('Failed to check transfers:', error)
    } finally {
      isCheckingTransfers.value = false
    }
  }

  await setAvailability(checked)
}

const confirmGoAway = async () => {
  showAwayWarning.value = false
  await setAvailability(false)
}

const setAvailability = async (checked: boolean) => {
  isUpdatingAvailability.value = true
  try {
    const response = await usersService.updateAvailability(checked)
    const data = response.data.data
    authStore.setAvailability(checked, data.break_started_at)

    if (checked) {
      toast.success(t('userMenu.available'), {
        description: t('userMenu.availableDesc')
      })
    } else {
      const transfersReturned = data.transfers_to_queue || 0
      toast.success(t('userMenu.away'), {
        description: transfersReturned > 0
          ? t('userMenu.transfersReturned', { count: transfersReturned })
          : t('userMenu.awayDesc')
      })

      if (transfersReturned > 0) {
        contactsStore.fetchContacts()
      }
    }
  } catch (error) {
    toast.error(t('common.error'), {
      description: t('userMenu.failedUpdateAvailability')
    })
  } finally {
    isUpdatingAvailability.value = false
  }
}

// Break duration tracking
const breakDuration = ref('')
let breakTimerInterval: ReturnType<typeof setInterval> | null = null

const updateBreakDuration = () => {
  if (!authStore.breakStartedAt) {
    breakDuration.value = ''
    return
  }
  const start = new Date(authStore.breakStartedAt)
  const now = new Date()
  const diffMs = now.getTime() - start.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const hours = Math.floor(diffMins / 60)
  const mins = diffMins % 60

  if (hours > 0) {
    breakDuration.value = `${hours}h ${mins}m`
  } else {
    breakDuration.value = `${mins}m`
  }
}

watch(() => authStore.isAvailable, (available) => {
  if (!available && authStore.breakStartedAt) {
    updateBreakDuration()
    breakTimerInterval = setInterval(updateBreakDuration, 60000)
  } else if (breakTimerInterval) {
    clearInterval(breakTimerInterval)
    breakTimerInterval = null
    breakDuration.value = ''
  }
}, { immediate: true })

onMounted(() => {
  authStore.restoreBreakTime()
  if (!authStore.isAvailable && authStore.breakStartedAt) {
    updateBreakDuration()
    breakTimerInterval = setInterval(updateBreakDuration, 60000)
  }
})

onUnmounted(() => {
  if (breakTimerInterval) {
    clearInterval(breakTimerInterval)
  }
})

const handleLogout = () => {
  emit('logout')
}
</script>

<template>
  <div class="border-t border-sidebar-border p-2">
    <Popover v-model:open="isUserMenuOpen">
      <PopoverTrigger as-child>
        <Button
          variant="ghost"
          :class="[
            'flex h-auto w-full items-center justify-start gap-2 px-2 py-1.5 text-sidebar-foreground hover:bg-sidebar-accent',
            collapsed && 'md:justify-center md:px-1'
          ]"
          :aria-label="t('nav.userMenu')"
          :aria-expanded="isUserMenuOpen"
        >
          <span
            :class="['avatar-status relative inline-flex', authStore.isAvailable ? '' : 'away']"
          >
            <Avatar class="h-7 w-7">
              <AvatarImage :src="undefined" />
              <AvatarFallback class="bg-primary text-xs font-medium text-primary-foreground">
                {{ getInitials(authStore.user?.full_name || 'U') }}
              </AvatarFallback>
            </Avatar>
          </span>
          <div v-if="!collapsed" class="flex min-w-0 flex-col items-start text-left">
            <span class="max-w-[150px] truncate text-[13px] font-medium">
              {{ authStore.user?.full_name }}
            </span>
            <span class="max-w-[150px] truncate text-[11px] text-sidebar-muted">
              {{ authStore.user?.email }}
            </span>
          </div>
        </Button>
      </PopoverTrigger>
      <PopoverContent side="top" align="start" class="w-56 p-1.5">
        <div class="px-2 py-1 text-xs font-medium text-muted-foreground">{{ t('userMenu.myAccount') }}</div>
        <Separator class="my-1" />
        <!-- Availability Toggle -->
        <div class="flex items-center justify-between px-2 py-1.5">
          <div class="flex items-center gap-2">
            <span id="availability-label" class="text-[13px] text-foreground">{{ t('userMenu.status') }}</span>
            <Badge :variant="authStore.isAvailable ? 'success' : 'secondary'" class="px-1.5 py-0 text-[10px]">
              {{ authStore.isAvailable ? t('userMenu.available') : t('userMenu.away') }}
            </Badge>
            <span v-if="!authStore.isAvailable && breakDuration" class="text-[10px] text-muted-foreground">
              {{ breakDuration }}
            </span>
          </div>
          <Switch
            :checked="authStore.isAvailable"
            :disabled="isUpdatingAvailability || isCheckingTransfers"
            :aria-label="t('userMenu.toggleAvailability')"
            @update:checked="handleAvailabilityChange"
          />
        </div>
        <Separator class="my-1" />
        <Button
          as-child
          variant="ghost"
          class="h-auto w-full justify-start px-2 py-1 text-[13px] font-normal text-foreground"
        >
          <RouterLink to="/profile" @click="isUserMenuOpen = false">
            <User class="mr-2 h-3.5 w-3.5" aria-hidden="true" />
            <span>{{ t('userMenu.profile') }}</span>
          </RouterLink>
        </Button>
        <Separator class="my-1" />
        <div class="px-2 py-1 text-xs font-medium text-muted-foreground">{{ t('userMenu.theme') }}</div>
        <ThemeSwitcher />
        <Separator class="my-1" />
        <div class="px-2 py-1 text-xs font-medium text-muted-foreground">{{ t('userMenu.language') }}</div>
        <div class="px-1.5 py-1">
          <LanguageSwitcher />
        </div>
        <Separator class="my-1" />
        <Button
          variant="ghost"
          class="h-auto w-full justify-start px-2 py-1 text-[13px] font-normal text-foreground"
          @click="handleLogout"
        >
          <LogOut class="mr-2 h-3.5 w-3.5" aria-hidden="true" />
          <span>{{ t('userMenu.logOut') }}</span>
        </Button>
      </PopoverContent>
    </Popover>
  </div>

  <!-- Away Warning Dialog -->
  <AlertDialog :open="showAwayWarning">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t('userMenu.awayWarningTitle') }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ t('userMenu.awayWarningDesc', { count: awayWarningTransferCount }) }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <Button variant="outline" @click="showAwayWarning = false">{{ t('common.cancel') }}</Button>
        <Button :loading="isUpdatingAvailability" @click="confirmGoAway">{{ t('userMenu.goAway') }}</Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
