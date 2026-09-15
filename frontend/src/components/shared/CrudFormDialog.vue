<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

const open = defineModel<boolean>('open', { default: false })

const props = withDefaults(defineProps<{
  title?: string
  editTitle?: string
  createTitle?: string
  description?: string
  editDescription?: string
  createDescription?: string
  isEditing?: boolean
  isSubmitting?: boolean
  submitLabel?: string
  editSubmitLabel?: string
  createSubmitLabel?: string
  cancelLabel?: string
  maxWidth?: string
}>(), {
  isEditing: false,
  isSubmitting: false,
  maxWidth: 'max-w-md',
})

const emit = defineEmits<{
  submit: []
  cancel: []
}>()

const { t } = useI18n()

function handleSubmit() {
  if (props.isSubmitting) return
  emit('submit')
}

function handleCancel() {
  if (props.isSubmitting) return
  open.value = false
  emit('cancel')
}

// While a submission is in flight the dialog must stay open: closing it via
// Escape or an outside click would reset the form under a pending request.
function preventCloseWhileSubmitting(event: Event) {
  if (props.isSubmitting) event.preventDefault()
}

const computedTitle = computed(() => {
  if (props.title) return props.title
  return props.isEditing
    ? (props.editTitle || t('common.editItem'))
    : (props.createTitle || t('common.createItem'))
})

const computedDescription = computed(() => {
  if (props.description) return props.description
  return props.isEditing
    ? (props.editDescription || t('common.editItemDesc'))
    : (props.createDescription || t('common.createItemDesc'))
})

const computedSubmitLabel = computed(() => {
  if (props.submitLabel) return props.submitLabel
  return props.isEditing
    ? (props.editSubmitLabel || t('common.update'))
    : (props.createSubmitLabel || t('common.create'))
})
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent
      :class="[maxWidth, 'max-h-[90vh] overflow-y-auto']"
      @escape-key-down="preventCloseWhileSubmitting"
      @pointer-down-outside="preventCloseWhileSubmitting"
    >
      <DialogHeader>
        <DialogTitle>{{ computedTitle }}</DialogTitle>
        <DialogDescription>{{ computedDescription }}</DialogDescription>
      </DialogHeader>

      <form class="py-2" @submit.prevent="handleSubmit">
        <slot />
        <!-- Lets Enter submit from any field without a visible extra button -->
        <button type="submit" class="hidden" tabindex="-1" aria-hidden="true" />
      </form>

      <DialogFooter>
        <Button variant="outline" size="sm" :disabled="isSubmitting" @click="handleCancel">
          {{ cancelLabel || t('common.cancel') }}
        </Button>
        <Button size="sm" :loading="isSubmitting" @click="handleSubmit">
          {{ computedSubmitLabel }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
