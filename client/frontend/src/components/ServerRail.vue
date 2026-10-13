<script setup lang="ts">
// ServerRail 服务器轨道（对齐 Main.html 左栏：图标按钮+未读角标+悬浮提示）。
// 图标为预设 Lucide（禁头像/上传图——硬性约束）。
import { Plus } from 'lucide-vue-next'
import { computed } from 'vue'
import { useServersStore } from '../stores/data'

const store = useServersStore()

// Lucide 图标名 → 组件（静态注册，避免动态导入复杂度）。
import {
  MessagesSquare, Gamepad2, Music, Palette, BookOpen,
} from 'lucide-vue-next'
const icons: Record<string, unknown> = {
  MessagesSquare, Gamepad2, Music, Palette, BookOpen,
}
const currentIcons = computed(() =>
  store.servers.map((s) => ({ ...s, iconComp: icons[s.icon] ?? MessagesSquare })))
const badges: Record<string, number> = { game: 3 }
</script>

<template>
  <nav class="rail">
    <div v-for="s in currentIcons" :key="s.id" class="rail-item"
         :class="{ on: s.id === store.currentServerId }" :data-tip="s.name">
      <button class="ri-btn" @click="store.selectServer(s.id)">
        <component :is="s.iconComp" :size="21" />
      </button>
      <span v-if="badges[s.id]" class="rail-badge">{{ badges[s.id] }}</span>
    </div>
    <div class="spacer" />
    <button class="rail-add" title="创建或加入服务器"><Plus :size="20" /></button>
  </nav>
</template>

<style scoped>
.rail {
  width: 72px; flex: none;
  display: flex; flex-direction: column; align-items: center;
  padding: 12px 0; gap: 10px;
  background: linear-gradient(180deg, var(--brand-600), var(--brand-700));
}
.rail-item { position: relative; width: 46px; height: 46px; flex: none; }
.ri-btn {
  width: 46px; height: 46px; border-radius: 16px;
  display: flex; align-items: center; justify-content: center;
  background: rgba(255, 255, 255, 0.16); color: #fff;
  transition: border-radius 0.25s var(--ease), background 0.25s var(--ease);
}
.rail-item:hover .ri-btn {
  border-radius: 13px; background: rgba(255, 255, 255, 0.28);
}
.rail-item.on .ri-btn {
  border-radius: 13px; background: #fff; color: var(--brand-600);
}
.rail-item::before {
  content: ""; position: absolute; left: -13px; top: 50%;
  transform: translateY(-50%) scaleY(0);
  width: 4px; height: 22px; border-radius: 2px; background: #fff;
  transition: transform 0.25s var(--ease);
}
.rail-item.on::before { transform: translateY(-50%) scaleY(1); }
.rail-badge {
  position: absolute; right: -2px; bottom: -2px;
  min-width: 17px; height: 17px; padding: 0 4px;
  border-radius: 9px; background: var(--red); color: #fff;
  font-size: 10px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
  border: 2px solid var(--brand-700);
}
.rail-item[data-tip]::after {
  content: attr(data-tip);
  position: absolute; left: 60px; top: 50%;
  transform: translateY(-50%) translateX(-4px);
  padding: 6px 11px; border-radius: 8px; white-space: nowrap;
  background: var(--ink-900); color: #fff;
  font-size: 12px; font-weight: 500;
  opacity: 0; pointer-events: none;
  transition: opacity 0.2s var(--ease), transform 0.2s var(--ease);
  z-index: 40; box-shadow: var(--shadow-2);
}
.rail-item[data-tip]:hover::after {
  opacity: 1; transform: translateY(-50%) translateX(0);
}
.spacer { flex: 1; }
.rail-add {
  width: 46px; height: 46px; border-radius: 16px;
  border: 1.5px dashed rgba(255, 255, 255, 0.45);
  display: flex; align-items: center; justify-content: center;
  color: rgba(255, 255, 255, 0.8);
  transition: border-radius 0.25s var(--ease), background 0.2s, color 0.2s;
}
.rail-add:hover {
  border-radius: 13px; background: rgba(255, 255, 255, 0.14); color: #fff;
}
</style>
