<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Application, Window } from '@wailsio/runtime'
import { Title } from '../bindings/boilerplate/app'

const title = ref('')
// The tall Mac bar has to be in the first paint. The user agent is enough for that.
const mac = ref(navigator.userAgent.includes('Macintosh'))
// Linux caption buttons are circles. Hover is a lighter disc, including close.
const linux = ref(!mac.value && navigator.userAgent.includes('Linux'))
const maximised = ref(false)

async function syncMaximised(): Promise<void> {
  maximised.value = await Window.IsMaximised()
}

async function toggleMaximise(): Promise<void> {
  if (await Window.IsMaximised()) {
    await Window.Restore()
  } else {
    await Window.Maximise()
  }
  window.setTimeout(syncMaximised, 50)
}

onMounted(async () => {
  title.value = await Title()
  await syncMaximised()
})
</script>

<template>
  <header class="titlebar" :class="{ 'titlebar--mac': mac, 'titlebar--linux': linux }" @dblclick="toggleMaximise">
    <div v-if="mac" class="titlebar-lights" />
    <div class="titlebar-title">{{ title }}</div>
    <div v-if="!mac" class="titlebar-controls" @dblclick.stop>
      <button type="button" class="titlebar-btn" aria-label="Minimise" @click="Window.Minimise">
        <svg viewBox="0 0 10 10" aria-hidden="true">
          <rect x="1" y="4.5" width="8" height="1" />
        </svg>
      </button>
      <button type="button" class="titlebar-btn" :aria-label="maximised ? 'Restore' : 'Maximise'" @click="toggleMaximise">
        <svg v-if="maximised" viewBox="0 0 10 10" aria-hidden="true">
          <!-- Rear square is only the edges the front square does not cover, at the same 1px weight. -->
          <path fill-rule="evenodd" d="M3 1h6v6H7V6h1V2H4V3H3zM1 3h6v6H1V3zm1 1v4h4V4H2z" />
        </svg>
        <svg v-else viewBox="0 0 10 10" aria-hidden="true">
          <path fill-rule="evenodd" d="M1 1h8v8H1V1zm1 1v6h6V2H2z" />
        </svg>
      </button>
      <button type="button" class="titlebar-btn close" aria-label="Close" @click="Application.Quit">
        <svg viewBox="0 0 10 10" aria-hidden="true">
          <path d="M1.6 1.1 5 4.5 8.4 1.1 8.9 1.6 5.5 5 8.9 8.4 8.4 8.9 5 5.5 1.6 8.9 1.1 8.4 4.5 5 1.1 1.6z" />
        </svg>
      </button>
    </div>
  </header>
</template>
