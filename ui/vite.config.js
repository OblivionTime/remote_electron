import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { resolve } from "path";
const pathResolve = (p) => resolve(process.cwd(), p);
import electron from "vite-plugin-electron";
import AutoImport from "unplugin-auto-import/vite";

export default defineConfig(() => {
  return {
    plugins: [
      vue(),
      electron({ entry: "main/background.js" }),
      AutoImport({
        imports: ["vue", "vue-router"],
        dts: "./auto-imports.d.ts",
        eslintrc: { enabled: false },
      }),
    ],
    build: {
      cssCodeSplit: true,
      sourcemap: false,
      target: "es2020",
      chunkSizeWarningLimit: 550,
      assetsInlineLimit: 4096,
      terserOptions: { compress: { drop_console: true, drop_debugger: true } },
      outDir: "dist",
      assetsDir: "./static",
      reportCompressedSize: false,
    },
    resolve: { alias: { "@": pathResolve("src") } },
    server: { port: 3000 },
  };
});
