import { createRouter, createWebHashHistory } from "vue-router";
const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/",
      name: "first",
      component: () => import("@/views/First/index.vue"),
    },
    {
      path: "/index",
      name: "index",
      component: () => import("@/views/Index/index.vue"),
    },{
      path: "/remote",
      name: "remote",
      component: () => import("@/views/Remote/index.vue"),
    },
    {
      path: "/controlledEnd",
      name: "controlledEnd",
      component: () => import("@/views/controlledEnd/index.vue"),
    },
    {
      path: "/fileManage",
      name: "fileManage",
      component: () => import("@/views/fileManage/index.vue"),
    },
  ],
});

export default router;
