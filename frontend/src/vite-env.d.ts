/// <reference types="vite/client" />

declare const __APP_TITLE__: string
declare const __APP_DESCRIPTION__: string

declare module '*.vue' {
  import type {DefineComponent} from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}
