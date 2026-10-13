<script setup lang="ts">
import { AudioWaveform } from 'lucide-vue-next'

// TODO(M0 后续)：改为调用 Wails 绑定 window.go.main.App.GetVersion()，
// 骨架阶段先写死 dev。
const version = 'dev'
</script>

<template>
  <main class="home">
    <div class="logo">
      <AudioWaveform :size="28" :stroke-width="2.2" />
    </div>
    <h1 class="title">轻音 LiteTalk</h1>
    <p class="version">{{ version }}</p>
  </main>
</template>

<style scoped>
.home {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.logo {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: var(--brand-500);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-2);
}

.title {
  font-size: 16px;
  font-weight: 600;
  color: var(--ink-900);
}

.version {
  font-size: 12px;
  color: var(--ink-400);
}
</style>
