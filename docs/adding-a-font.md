# Custom font

Default: system font stack.

1. In `frontend`:

   ```sh
   npm install @fontsource/your-font
   ```

2. `frontend/src/main.ts`:

   ```ts
   import '@fontsource/your-font/400.css'
   import '@fontsource/your-font/600.css'
   ```

3. `frontend/src/style.css`:

   ```css
   @theme {
     --font-sans: "Your Font", ui-sans-serif, system-ui, sans-serif;
     --font-mono: "Your Mono Font", ui-monospace, monospace;
   }
   ```

4. In `frontend`:

   ```sh
   npm run build
   ```
