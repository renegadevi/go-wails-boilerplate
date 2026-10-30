<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Frameless, Greet, Version } from '../bindings/boilerplate/app'
import Toolbar from './Toolbar.vue'

const title = __APP_TITLE__
const description = __APP_DESCRIPTION__
const version = ref('')
const name = ref('')
const greeting = ref('')
const frameless = ref(false)
const canGreet = computed(() => name.value.trim() !== '')

onMounted(async () => {
  version.value = await Version()
  frameless.value = await Frameless()
})

async function greet(): Promise<void> {
  greeting.value = await Greet(name.value)
}
</script>

<template>
  <div class="flex h-full flex-col bg-bg text-ink">
    <Toolbar v-if="frameless" />
    <div class="relative flex min-h-0 flex-1 items-center justify-center px-6 pb-16">
      <main class="relative flex w-full max-w-[26rem] flex-col items-center text-center">
        <h1 class="m-0 text-[22px] font-medium tracking-tight">{{ title }}</h1>
        <p class="m-0 mt-2 text-[13px] text-muted">{{ description }}</p>
        <form class="composer" @submit.prevent="greet">
          <input v-model="name" placeholder="Your name" autocomplete="off" spellcheck="false" />
          <button type="submit" :class="{ ready: canGreet }">Greet</button>
        </form>
        <p v-if="greeting" class="reply">{{ greeting }}</p>
      </main>
      <footer v-if="version" class="absolute inset-x-0 bottom-6 text-center text-[11px] text-muted">{{ version }}</footer>
    </div>
  </div>
</template>

<style scoped>
.composer {
  display: flex;
  align-items: center;
  width: 100%;
  height: 44px;
  margin-top: 28px;
  padding: 0 6px 0 14px;
  border: 1px solid var(--color-line);
  border-radius: 12px;
  background: var(--color-panel);
  transition: border-color 150ms ease;
}

.composer:hover {
  border-color: #363a40;
}

.composer:focus-within {
  border-color: #4a5058;
}

.composer input {
  min-width: 0;
  flex: 1;
  height: 100%;
  margin: 0;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  font-size: 13px;
  text-align: left;
}

.composer input::placeholder {
  color: var(--color-muted);
}

.composer button {
  flex: none;
  height: 28px;
  padding: 0 10px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--color-muted);
  font-size: 12px;
  font-weight: 500;
}

.composer button:hover {
  color: var(--color-ink);
}

.composer button.ready {
  background: #e8e9eb;
  color: #16181a;
}

.composer button.ready:hover {
  background: #fff;
  color: #16181a;
}

.reply {
  position: absolute;
  top: 100%;
  right: 0;
  left: 0;
  margin: 16px 0 0;
  font-size: 13px;
}
</style>
