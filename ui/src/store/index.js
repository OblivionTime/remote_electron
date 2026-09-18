import { defineStore } from "pinia";
export const useDeviceStore = defineStore({
  id: "device",
  state: () => {
    return {
      deviceInfo: {
        device_id: "",
        identificationCode: "",
        verificationCode: "",
      },
      serverInfo: localStorage.getItem("serverInfo")
        ? JSON.parse(localStorage.getItem("serverInfo"))
        : {
            https: false,
            remote_url: "127.0.0.1:9998",
          },
      online: {
        status: false,
        message: "服务器连接失败",
      },
    };
  },
  actions: {
    resetDevice() {
      this.deviceInfo = {
        device_id: "",
        identificationCode: "",
        verificationCode: "",
      };
    },
    setOnline(status, message) {
      this.online = { status, message };
    },
  },
});
