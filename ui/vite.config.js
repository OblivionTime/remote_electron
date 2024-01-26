import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { resolve } from "path"; // path 模块提供了一些工具函数，用于处理文件与目录的路径
const pathResolve = (path) => resolve(process.cwd(), path);
import electron from "vite-plugin-electron";
import AutoImport from "unplugin-auto-import/vite";

export default defineConfig({
  plugins: [
    vue(),
    electron({
      // 主进程入口文件
      entry: "./main/background.js",
    }),
    AutoImport({
      // 自动导入 Vue 相关函数，如：ref, reactive, toRef 等
      imports: ["vue", "vue-router"],
      dts: "./auto-imports.d.ts",
      eslintrc: {
        enabled: false, // 1、true时生成eslint配置文件，2、生成后改为false，避免重复消耗
      },
    }),
  ],
  build: {
    cssCodeSplit: true, // 如果设置为false，整个项目中的所有 CSS 将被提取到一个 CSS 文件中
    sourcemap: false, // 构建后是否生成 source map 文件。如果为 true，将会创建一个独立的 source map 文件
    target: "edge88", // 设置最终构建的浏览器兼容目标。默认值是一个 Vite 特有的值——'modules'  还可设置为 'es2015' 'es2016'等
    chunkSizeWarningLimit: 550, // 单位kb  打包后文件大小警告的限制 (文件大于此此值会出现警告)
    assetsInlineLimit: 4096, // 单位字节（1024等于1kb） 小于此阈值的导入或引用资源将内联为 base64 编码，以避免额外的 http 请求。设置为 0 可以完全禁用此项。
    terserOptions: {
      compress: {
        drop_console: true, // 生产环境去除console
        drop_debugger: true, // 生产环境去除debugger
      },
    },
    outDir: "dist",
    assetsDir: "./static",
    reportCompressedSize: false,
  },
  resolve: {
    alias: {
      "@": pathResolve("src"), // 设置 `@` 指向 `src` 目录
    },
  },
  /*开发服务器选项*/
  server: {
    // 端口
    port: 3000,
  },
});
