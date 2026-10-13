<script setup lang="ts">
// Toast 轻提示（对齐 Auth.html：深底圆角、成功/错误态、自动消失）。
import { ref } from 'vue'
import { CircleCheck, CircleAlert } from 'lucide-vue-next'

const visible = ref(false)
const text = ref('')
const isError = ref(false)
let timer: ReturnType<typeof setTimeout> | null = null

function show(msg: string, err = false) {
  text.value = msg
  isError.value = err
  visible.value = true
  if (timer) clearTimeout(timer)
  timer = setTimeout(() => (visible.value = false), 2200)
}

defineExpose({ show })
</script>

<template>
  <Transition name="toast">
    <div v-if="visible" class="toast" :class="{ err: isError }">
      <CircleAlert v-if="isError" :size="16" />
      <CircleCheck v-else :size="16" />
      <span>{{ text }}</span>
    </div>
  </Transition>
</template>

<style scoped>
.toast {
  position: fixed; top: 52px; left: 50%; z-index: 60;
  display: flex; align-items: center; gap: 8px;
  padding: 10px 18px; border-radius: 11px;
  background: var(--ink-900); color: #fff;
  font-size: 13px; font-weight: 500;
  box-shadow: var(--shadow-2);
}
.toast :deep(svg) { color: var(--green); }
.toast.err :deep(svg) { color: var(--red); }
.toast-enter-active, .toast-leave-active {
  transition: opacity 0.3s var(--ease), transform 0.3s var(--ease);
}
.toast-enter-from, .toast-leave-to {
  opacity: 0; transform: translate(-50%, -16px);
}
.toast { transform: translate(-50%, 0); }
</style>
