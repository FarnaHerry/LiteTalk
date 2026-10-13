<script setup lang="ts">
// VoiceRoom 语音房间网格（对齐 Main.html：卡片网格 + 声波球 + 状态胶囊 + 邀请卡）。
// 声波球四态见 SoundOrb.vue；双关态胶囊在此（.chip.red.dual）。
import { AudioLines, MicOff, HeadphoneOff, MoreHorizontal, UserPlus } from 'lucide-vue-next'
import SoundOrb from '../components/SoundOrb.vue'
import { useVoiceStore } from '../stores/voice'

const voice = useVoiceStore()

// 状态胶囊（双关态优先——用户要求的合并样式）。
function chipFor(m: { mute: boolean; deafen: boolean; speaking: boolean }) {
  if (m.mute && m.deafen) return { cls: 'red dual', text: '已静音 已耳聋', dual: true }
  if (m.mute) return { cls: 'red', text: '已静音', icon: MicOff }
  if (m.deafen) return { cls: 'red', text: '已耳聋', icon: HeadphoneOff }
  if (m.speaking) return { cls: 'brand', text: '正在说话', icon: AudioLines }
  return null
}
</script>

<template>
  <div class="voice-grid">
    <div v-for="m in voice.members" :key="m.user_id" class="vcard">
      <button class="more" title="更多"><MoreHorizontal :size="16" /></button>
      <SoundOrb :user-id="m.user_id" :speaking="m.speaking"
                :muted="m.mute" :deafen="m.deafen" />
      <div class="vname">
        {{ m.nickname }}
        <span v-if="m.user_id === voice.selfId()" class="me-tag">我</span>
      </div>
      <span v-if="chipFor(m)" class="chip" :class="chipFor(m)!.cls">
        <template v-if="chipFor(m)!.dual">
          <MicOff :size="12" />已静音<HeadphoneOff :size="12" />已耳聋
        </template>
        <template v-else>
          <component :is="chipFor(m)!.icon" :size="12" />{{ chipFor(m)!.text }}
        </template>
      </span>
    </div>

    <div class="vcard invite">
      <div class="i-circle"><UserPlus :size="20" /></div>
      <span>邀请朋友加入频道</span>
    </div>
  </div>
</template>

<style scoped>
.voice-grid {
  flex: 1; overflow-y: auto; padding: 22px;
  display: grid; gap: 16px; align-content: start;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.vcard {
  position: relative; min-height: 172px;
  background: var(--surface); border: 1px solid var(--line);
  border-radius: 16px;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 10px; padding: 20px;
  transition: transform 0.25s var(--ease), box-shadow 0.25s var(--ease), border-color 0.25s var(--ease);
}
.vcard:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-2); border-color: var(--brand-200);
}
.vcard .more {
  position: absolute; top: 10px; right: 10px;
  width: 28px; height: 28px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  color: var(--ink-300); opacity: 0;
  transition: opacity 0.2s, background 0.15s, color 0.15s;
}
.vcard:hover .more { opacity: 1; }
.vcard .more:hover { background: var(--brand-50); color: var(--ink-600); }
.vname { font-size: 15px; font-weight: 600; }
.me-tag {
  font-size: 11px; font-weight: 600;
  color: var(--brand-600); background: rgba(79, 164, 243, 0.12);
  padding: 1px 6px; border-radius: 6px; margin-left: 5px; vertical-align: 1px;
}
.chip {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 3px 9px; border-radius: 99px;
  font-size: 11.5px; font-weight: 500;
}
.chip.red { background: rgba(232, 91, 91, 0.1); color: var(--red); }
.chip.brand { background: rgba(79, 164, 243, 0.1); color: var(--brand-600); }
/* 双关态胶囊：红底加深加重（区别于单关） */
.chip.red.dual { background: rgba(232, 91, 91, 0.16); font-weight: 600; }
.vcard.invite {
  border: 1.5px dashed var(--brand-200); background: transparent;
  cursor: pointer; color: var(--ink-400);
}
.vcard.invite:hover {
  border-color: var(--brand-500); color: var(--brand-600);
  box-shadow: none; background: rgba(79, 164, 243, 0.03);
}
.vcard.invite .i-circle {
  width: 44px; height: 44px; border-radius: 50%;
  background: rgba(79, 164, 243, 0.1);
  display: flex; align-items: center; justify-content: center;
  color: var(--brand-500);
  transition: transform 0.25s var(--ease);
}
.vcard.invite:hover .i-circle { transform: scale(1.08); }
.vcard.invite span { font-size: 13px; font-weight: 500; }
</style>
