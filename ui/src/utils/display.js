// 远程画面相关的预设与偏好。
// 画质与黑白会被下发给被控端：画质落到 RTCRtpSender 编码参数（真实降带宽），
// 黑白落在被控端的采集侧（改的是"目标的画面"，同时也省掉色度编码）。

// setParameters 是合并语义：字段不写就保持原值，所以"取消限制"只能给一个很大的值
export const UNLIMITED_BITRATE = 100000000

export const QUALITY_PRESETS = {
  original: { maxBitrate: 0, maxFramerate: 30, scale: 1 },
  high: { maxBitrate: 3000000, maxFramerate: 30, scale: 1 },
  smooth: { maxBitrate: 700000, maxFramerate: 15, scale: 2 },
}

export const QUALITY_OPTIONS = [
  { value: "original", label: "原画质" },
  { value: "high", label: "高清" },
  { value: "smooth", label: "流畅" },
]

export const ASPECT_OPTIONS = [
  { value: "fit", label: "自动适应窗口" },
  { value: "original", label: "保持原始比例" },
]

const KEY_QUALITY = "remoteQuality"
const KEY_ASPECT = "remoteAspect"

export const qualityLabel = (v) =>
  (QUALITY_OPTIONS.find((o) => o.value === v) || QUALITY_OPTIONS[1]).label

export const aspectLabel = (v) => (v === "original" ? "原始比例" : "适应窗口")

// 远控会话启动时读取的默认值
export const loadRemotePrefs = () => ({
  quality: QUALITY_PRESETS[localStorage.getItem(KEY_QUALITY)] ? localStorage.getItem(KEY_QUALITY) : "high",
  aspect: localStorage.getItem(KEY_ASPECT) === "original" ? "original" : "fit",
})

export const saveDefaultQuality = (v) => {
  if (QUALITY_PRESETS[v]) localStorage.setItem(KEY_QUALITY, v)
}
export const saveDefaultAspect = (v) => {
  if (v === "fit" || v === "original") localStorage.setItem(KEY_ASPECT, v)
}
