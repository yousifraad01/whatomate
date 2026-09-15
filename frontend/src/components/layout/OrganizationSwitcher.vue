<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useOrganizationsStore } from '@/stores/organizations'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { organizationsService } from '@/services/api'
import { toast } from 'vue-sonner'
import { Building2, Plus } from 'lucide-vue-next'

defineProps<{
  collapsed?: boolean
}>()

const { t } = useI18n()
const organizationsStore = useOrganizationsStore()
const authStore = useAuthStore()

const isSuperAdmin = computed(() => authStore.user?.is_super_admin || false)
const canCreateOrg = computed(() => authStore.hasPermission('organizations', 'write'))

const shouldShowSwitcher = computed(() =>
  isSuperAdmin.value || organizationsStore.isMultiOrg
)

// Build the org list depending on user type
const orgList = computed(() => {
  if (isSuperAdmin.value) {
    return organizationsStore.organizations.map(org => ({ id: org.id, name: org.name }))
  }
  return organizationsStore.myOrganizations.map(org => ({ id: org.organization_id, name: org.name }))
})

const currentOrgId = computed(() => {
  if (isSuperAdmin.value) {
    return organizationsStore.selectedOrgId || ''
  }
  return authStore.user?.organization_id || ''
})

const currentOrgName = computed(() => orgList.value.find(o => o.id === currentOrgId.value)?.name || '')

onMounted(async () => {
  // Fetch user's org memberships for all authenticated users
  await organizationsStore.fetchMyOrganizations()

  if (isSuperAdmin.value) {
    organizationsStore.init()
    await organizationsStore.fetchOrganizations()

    // If no org selected, default to user's own org
    if (!organizationsStore.selectedOrgId && authStore.user?.organization_id) {
      organizationsStore.selectOrganization(authStore.user.organization_id)
    }
  }
})

// Watch for auth changes
watch(() => authStore.user?.is_super_admin, async (superAdmin) => {
  if (superAdmin) {
    organizationsStore.init()
    await organizationsStore.fetchOrganizations()
  }
})

const isSwitching = ref(false)

const handleOrgChange = async (value: string | number | bigint | Record<string, any> | null) => {
  if (!value || typeof value !== 'string' || value === currentOrgId.value) return

  // Everyone goes through switch-org, super admins included: the X-Organization-ID
  // header alone only scopes axios calls. The WebSocket token, <img src> media and
  // window.open previews carry no headers, so the JWT itself has to name the target
  // org or those keep talking to the previous one. The full reload afterwards
  // discards every store that still holds the previous organization's data.
  isSwitching.value = true
  try {
    await authStore.switchOrg(value)
    organizationsStore.selectOrganization(value)
    window.location.reload()
  } catch {
    isSwitching.value = false
    toast.error(t('common.error'))
  }
}

// Create org dialog
const isCreateDialogOpen = ref(false)
const newOrgName = ref('')
const isCreating = ref(false)

async function submitCreateOrg() {
  if (!newOrgName.value.trim() || isCreating.value) return
  isCreating.value = true
  try {
    await organizationsService.create({ name: newOrgName.value.trim() })
    toast.success(t('organizations.created'))
    isCreateDialogOpen.value = false
    newOrgName.value = ''
    await refreshOrgs()
  } catch {
    toast.error(t('organizations.createFailed'))
  } finally {
    isCreating.value = false
  }
}

const refreshOrgs = async () => {
  if (isSuperAdmin.value) {
    await organizationsStore.fetchOrganizations()
  } else {
    await organizationsStore.fetchMyOrganizations()
  }
}
</script>

<template>
  <div v-if="shouldShowSwitcher" class="border-b border-sidebar-border px-2 py-2">
    <div v-if="!collapsed" class="space-y-1">
      <div class="flex items-center justify-between">
        <Label for="org-switcher" class="px-1 text-[11px] font-semibold uppercase tracking-wide text-sidebar-muted">
          {{ t('nav.organization') }}
        </Label>
        <Button
          v-if="canCreateOrg"
          variant="ghost"
          size="icon"
          class="h-6 w-6 text-sidebar-muted hover:text-sidebar-foreground"
          :aria-label="t('nav.createOrganization')"
          :title="t('nav.createOrganization')"
          @click="isCreateDialogOpen = true"
        >
          <Plus class="h-3.5 w-3.5" aria-hidden="true" />
        </Button>
      </div>
      <Select
        v-if="orgList.length > 0"
        :model-value="currentOrgId"
        :disabled="isSwitching"
        @update:model-value="handleOrgChange"
      >
        <SelectTrigger id="org-switcher" class="h-8 text-[13px]" :aria-label="t('nav.selectOrganization')">
          <SelectValue :placeholder="t('nav.selectOrganization')" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem
            v-for="org in orgList"
            :key="org.id"
            :value="org.id"
          >
            <div class="flex items-center gap-2">
              <Building2 class="h-3.5 w-3.5 text-muted-foreground" aria-hidden="true" />
              <span>{{ org.name }}</span>
            </div>
          </SelectItem>
        </SelectContent>
      </Select>
      <div v-else-if="organizationsStore.loading" class="px-1 text-xs text-sidebar-muted" role="status">
        {{ t('common.loading') }}…
      </div>
      <div v-else-if="organizationsStore.error" class="px-1 text-xs text-destructive" role="alert">
        {{ organizationsStore.error }}
      </div>
      <div v-else class="px-1 text-xs text-sidebar-muted">
        {{ t('nav.noOrganizations') }}
      </div>
    </div>

    <!-- Collapsed view: icon with the current organization as its name -->
    <div v-else class="flex justify-center">
      <span
        class="flex h-8 w-8 items-center justify-center rounded-md text-sidebar-muted"
        :title="currentOrgName || t('nav.organization')"
        :aria-label="currentOrgName || t('nav.organization')"
        role="img"
      >
        <Building2 class="h-4 w-4" aria-hidden="true" />
      </span>
    </div>
  </div>

  <!-- Create Org Dialog -->
  <Dialog v-model:open="isCreateDialogOpen">
    <DialogContent class="max-w-sm">
      <DialogHeader>
        <DialogTitle>{{ t('organizations.createTitle') }}</DialogTitle>
        <DialogDescription>{{ t('organizations.createDesc') }}</DialogDescription>
      </DialogHeader>
      <form class="py-2" @submit.prevent="submitCreateOrg">
        <Label for="new-org-name" class="sr-only">{{ t('organizations.namePlaceholder') }}</Label>
        <Input
          id="new-org-name"
          v-model="newOrgName"
          :placeholder="t('organizations.namePlaceholder')"
          :disabled="isCreating"
        />
      </form>
      <DialogFooter>
        <Button variant="outline" :disabled="isCreating" @click="isCreateDialogOpen = false">{{ t('common.cancel') }}</Button>
        <Button :loading="isCreating" :disabled="!newOrgName.trim()" @click="submitCreateOrg">
          {{ t('common.create') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
