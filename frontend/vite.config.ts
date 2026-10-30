import {readFileSync} from 'node:fs'
import {dirname, join} from 'node:path'
import {fileURLToPath} from 'node:url'
import {defineConfig, type Plugin} from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import wails from '@wailsio/runtime/plugins/vite'

// info.productName and info.description in build/config.yml. Values stay
// double-quoted, same as scripts/build.sh.
function readAppInfo(): {title: string; description: string} {
  const configPath = join(dirname(fileURLToPath(import.meta.url)), '../build/config.yml')
  const text = readFileSync(configPath, 'utf8')
  return {
    title: readQuoted(text, 'productName'),
    description: readQuoted(text, 'description'),
  }
}

function readQuoted(text: string, key: string): string {
  const match = text.match(new RegExp(`^  ${key}: "([^"]*)"`, 'm'))
  if (!match?.[1]) {
    throw new Error(`info.${key} must be a double-quoted string in build/config.yml`)
  }
  return match[1]
}

function escapeHtml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
}

function appMeta(title: string): Plugin {
  return {
    name: 'app-meta',
    transformIndexHtml(html) {
      return html.replaceAll('%APP_TITLE%', escapeHtml(title))
    },
  }
}

const appInfo = readAppInfo()

export default defineConfig({
  define: {
    __APP_TITLE__: JSON.stringify(appInfo.title),
    __APP_DESCRIPTION__: JSON.stringify(appInfo.description),
  },
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [appMeta(appInfo.title), vue(), tailwindcss(), wails('./bindings')],
})
