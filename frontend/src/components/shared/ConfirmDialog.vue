<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'

const open = defineModel<boolean>('open', { default: false })

withDefaults(defineProps<{
  title: string
  description?: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'default' | 'destructive'
  isSubmitting?: boolean
}>(), {
  variant: 'default',
  isSubmitting: false,
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const { t } = useI18n()

function handleConfirm() {
  emit('confirm')
}

function handleCancel() {
  open.value = false
  emit('cancel')
}
</script>

<template>
  <AlertDialog v-model:open="open">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ title }}</AlertDialogTitle>
        <AlertDialogDescription>
          <slot name="description">
            {{ description }}
          </slot>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="isSubmitting" @click="handleCancel">
          {{ cancelLabel || t('common.cancel') }}
        </AlertDialogCancel>
        <Button
          :variant="variant === 'destructive' ? 'destructive' : 'default'"
          :loading="isSubmitting"
          @click="handleConfirm"
        >
          {{ confirmLabel || t('common.confirm') }}
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
