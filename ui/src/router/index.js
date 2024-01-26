import { createRouter, createWebHashHistory } from "vue-router";
const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/",
      name: "index",
      component: () => import("@/views/Index/index.vue"),
    },{
      path: "/remote",
      name: "remote",
      component: () => import("@/views/Remote/index.vue"),
    },
  ],
});

export default router;
