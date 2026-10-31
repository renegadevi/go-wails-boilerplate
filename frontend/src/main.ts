import {createApp} from 'vue'
import App from './App.vue'
import { applyContextMenuMode } from './contextMenu'
import './style.css'

applyContextMenuMode().catch((error: unknown) => {
  console.error('context menu mode not applied:', error)
})

createApp(App).mount('#app')
