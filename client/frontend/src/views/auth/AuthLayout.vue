<script setup lang="ts">
// AuthLayout 认证壳：左品牌区（logo+标语+声波装饰）+ 右表单区（对齐 Auth.html）。
import { AudioWaveform } from 'lucide-vue-next'
</script>

<template>
  <div class="auth-wrap">
    <div class="shell">
      <aside class="brand">
        <div class="brand-top">
          <div class="logo-mark"><AudioWaveform :size="26" /></div>
          <div class="brand-name">轻音<small>LITETALK</small></div>
        </div>
        <div class="brand-mid">
          <p class="brand-tag">轻量、清晰的桌面语音。<br>打开频道，即刻开聊。</p>
          <div class="waves" aria-hidden="true">
            <span v-for="i in 8" :key="i" :style="{ '--i': i - 1 }" />
          </div>
        </div>
        <div class="brand-foot">© 2026 LiteTalk</div>
      </aside>
      <main class="panel">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.auth-wrap {
  height: 100vh;
  display: flex; align-items: center; justify-content: center;
  padding: 32px;
  background: var(--brand-50);
}
.shell {
  width: 960px; max-width: 100%; height: 600px; max-height: 100%;
  display: grid; grid-template-columns: 1fr 460px;
  background: var(--surface); border: 1px solid var(--line);
  border-radius: 24px; box-shadow: var(--shadow-3); overflow: hidden;
}
.brand {
  position: relative; padding: 56px 48px;
  display: flex; flex-direction: column;
  background: linear-gradient(160deg, var(--brand-50), #fff 70%);
}
.brand-top { display: flex; align-items: center; gap: 14px; }
.logo-mark {
  width: 48px; height: 48px; border-radius: 14px;
  background: var(--brand-500);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 6px 16px rgba(79, 164, 243, 0.35);
}
.logo-mark :deep(svg) { color: #fff; }
.brand-name { font-size: 20px; font-weight: 700; letter-spacing: 0.5px; }
.brand-name small {
  display: block; font-size: 11px; font-weight: 500;
  color: var(--ink-400); letter-spacing: 2.5px; margin-top: 2px;
}
.brand-mid { flex: 1; display: flex; flex-direction: column; justify-content: center; gap: 28px; }
.brand-tag { font-size: 15px; color: var(--ink-600); line-height: 1.8; text-wrap: pretty; }
.waves { display: flex; align-items: center; gap: 10px; height: 64px; }
.waves span {
  width: 5px; border-radius: 3px; background: var(--brand-200);
  animation: wave 2.4s var(--ease) infinite;
  animation-delay: calc(var(--i) * 0.14s);
}
.waves span:nth-child(3n) { background: var(--brand-500); }
.waves span:nth-child(1) { height: 22px; }
.waves span:nth-child(2) { height: 38px; }
.waves span:nth-child(3) { height: 56px; }
.waves span:nth-child(4) { height: 34px; }
.waves span:nth-child(5) { height: 48px; }
.waves span:nth-child(6) { height: 26px; }
.waves span:nth-child(7) { height: 40px; }
.waves span:nth-child(8) { height: 18px; }
@keyframes wave { 0%, 100% { transform: scaleY(0.45); } 50% { transform: scaleY(1); } }
.brand-foot { font-size: 12px; color: var(--ink-300); }
.panel { padding: 44px 48px; display: flex; flex-direction: column; background: var(--surface); }
@media (max-width: 880px) {
  .auth-wrap { padding: 0; }
  .shell { grid-template-columns: 1fr; height: 100vh; border: none; border-radius: 0; }
  .brand { padding: 20px 24px; flex-direction: row; align-items: center; gap: 12px; }
  .brand-mid, .brand-foot { display: none; }
}
</style>
