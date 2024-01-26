import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import { createPinia } from "pinia"
const pinia = createPinia()
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { debounce } from 'lodash';
const app = createApp(App)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
app.use(router).use(ElementPlus).use(pinia).mount('#app')
