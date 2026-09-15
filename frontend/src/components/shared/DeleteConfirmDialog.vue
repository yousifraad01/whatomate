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
  title?: string
  itemName?: string
  description?: string
  confirmLabel?: string
  cancelLabel?: string
  isSubmitting?: boolean
}>(), {
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
        <AlertDialogTitle>{{ title || t('common.deleteItem') }}</AlertDialogTitle>
        <AlertDialogDescription>
          <slot name="description">
            <template v-if="description">{{ description }}</template>
            <template v-else-if="itemName">{{ t('common.deleteConfirmNamed', { name: itemName }) }}</template>
            <template v-else>{{ t('common.deleteConfirmGeneric') }}</template>
          </slot>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="isSubmitting" @click="handleCancel">{{ cancelLabel || t('common.cancel') }}</AlertDialogCancel>
        <Button
          variant="destructive"
          :loading="isSubmitting"
          @click="handleConfirm"
        >
          {{ confirmLabel || t('common.delete') }}
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
