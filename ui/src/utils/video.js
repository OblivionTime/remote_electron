// 屏幕共享的 WebRTC 封装。
// 之前 App.vue（采集端）和 Remote/index.vue（观看端）各自手写了一遍
// PeerConnection / ICE 配置 / 候选收发，这里统一。

let iceServers = []

// 服务端下发的是 { url, username, credential }，WebRTC 只认 urls，这里统一转换
export const setIceServers = (list = []) => {
  iceServers = (list || []).map((s) => ({
    urls: s.urls || s.url,
    username: s.username,
    credential: s.credential,
  }))
}

export const createPeer = () =>
  new RTCPeerConnection({ iceServers })

// 候选统一成信令里传的四个字段
const candidatePayload = (candidate) => ({
  id: candidate.sdpMid,
  label: candidate.sdpMLineIndex,
  sdpMLineIndex: candidate.sdpMLineIndex,
  candidate: candidate.candidate,
})

// 把本地候选交回给调用方（由调用方决定走哪条 socket 发出去）
export const pipeCandidates = (pc, send) => {
  if (!pc || !send) return
  pc.onicecandidate = (evt) => {
    if (evt.candidate) send(candidatePayload(evt.candidate))
  }
}

export const addRemoteCandidate = (pc, data) => {
  if (!pc || !data) return
  pc.addIceCandidate(new RTCIceCandidate(data)).catch((err) => {
    console.error("addIceCandidate 失败", err)
  })
}

export const createOffer = async (pc) => {
  const offer = await pc.createOffer()
  await pc.setLocalDescription(offer)
  return offer
}

// 观看端：设远端 offer，产出本地 answer
export const acceptOffer = async (pc, sdp) => {
  await pc.setRemoteDescription(new RTCSessionDescription(sdp))
  const answer = await pc.createAnswer()
  await pc.setLocalDescription(answer)
  return answer
}

export const acceptAnswer = (pc, sdp) => {
  if (!pc || !sdp) return Promise.resolve()
  return pc.setRemoteDescription(new RTCSessionDescription(sdp))
}

export const closePeer = (pc) => {
  if (!pc) return
  try {
    pc.close()
  } catch (err) {
    console.error("关闭 PeerConnection 失败", err)
  }
}

// 断开（failed/disconnected/closed）时回调，带 1s 延迟避免抖动
export const watchState = (pc, onBroken) => {
  if (!pc) return
  pc.oniceconnectionstatechange = (evt) => {
    setTimeout(() => {
      const state = evt.target.connectionState
      if (state === "failed" || state === "disconnected" || state === "closed") {
        onBroken()
      }
    }, 1000)
  }
  pc.onicecandidateerror = (evt) => {
    console.error("ICE 候选错误", evt.errorCode, evt.errorText, evt.url)
  }
}
