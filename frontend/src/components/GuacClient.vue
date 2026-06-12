<template>
  <div ref="viewport" class="viewport">
    <div ref="display" class="display" tabindex="0"/>
    <button class="fullscreen-btn"
            :style="{ left: btnPosition.x + 'px', top: btnPosition.y + 'px' }"
            @mousedown="startDrag"
            @click="handleBtnClick">
      <svg v-if="!isFullscreen" viewBox="0 0 24 24" width="20" height="20">
        <path fill="currentColor" d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z"/>
      </svg>
      <svg v-else viewBox="0 0 24 24" width="20" height="20">
        <path fill="currentColor" d="M5 16h3v3h2v-5H5v2zm3-8H5v2h5V5H8v3zm6 11h2v-3h3v-2h-5v5zm2-11V5h-2v5h5V8h-3z"/>
      </svg>
    </button>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import Guacamole from 'guacamole-common-js'
import GuacMouse from '@/libs/GuacMouse.js'

Guacamole.Mouse = GuacMouse.mouse

const viewport = ref(null)
const display = ref(null)
const isFullscreen = ref(false)
const btnPosition = ref({ x: 12, y: 12 })
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })

const query = ref({
  guacad_addr: '192.168.1.30:4822',
  asset_protocol: 'rdp',
  asset_host: '192.168.1.90',
  asset_port: '3389',
  asset_user: 'administrator',
  asset_password: '123456',
  screen_width: 1920,
  screen_height: 1080,
  screen_dpi: 96,
})

let client = null
let keyboard = null
let mouse = null
let displayElm = null
let resizeObserver = null
let audioContext = null
let audioEnabled = false

const wsUrl = computed(() => {
  // 只使用 wss:// 协议
  return `wss://${window.location.host}/ws`
})

function serialize(obj) {
  const str = []
  for (const p in obj) {
    if (obj[p]) {
      str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]))
    }
  }
  return str.join("&")
}

function handleMouseState(mouseState) {
  if (!client || !client.getDisplay()) return
  enableAudio()
  const scaledMouseState = {
    ...mouseState,
    x: mouseState.x / client.getDisplay().getScale(),
    y: mouseState.y / client.getDisplay().getScale(),
  }
  client.sendMouseState(scaledMouseState)
}

function setScreenSize() {
  const elm = viewport.value
  if (!elm) return
  const width = elm.clientWidth || window.innerWidth
  const height = elm.clientHeight || window.innerHeight
  const pixelDensity = window.devicePixelRatio || 1
  query.value.screen_width = width * pixelDensity
  query.value.screen_height = height * pixelDensity
}

function resize() {
  const elm = viewport.value
  if (!elm || !client || !client.getDisplay()) return

  const width = elm.clientWidth || window.innerWidth
  const height = elm.clientHeight || window.innerHeight
  const pixelDensity = window.devicePixelRatio || 1
  const pixelWidth = width * pixelDensity
  const pixelHeight = height * pixelDensity

  client.sendSize(pixelWidth, pixelHeight)

  const scale = Math.min(
    width / Math.max(client.getDisplay().getWidth(), 1),
    height / Math.max(client.getDisplay().getHeight(), 1)
  )
  client.getDisplay().scale(scale)
}

function toggleFullscreen() {
  const elm = viewport.value
  if (!elm) return

  if (!document.fullscreenElement) {
    elm.requestFullscreen().then(() => {
      isFullscreen.value = true
    }).catch(err => {
      console.error('全屏请求失败:', err)
    })
  } else {
    document.exitFullscreen().then(() => {
      isFullscreen.value = false
    }).catch(err => {
      console.error('退出全屏失败:', err)
    })
  }
}

function handleFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

function startDrag(e) {
  isDragging.value = false
  dragStart.value = {
    x: e.clientX - btnPosition.value.x,
    y: e.clientY - btnPosition.value.y
  }

  const onMouseMove = (e) => {
    isDragging.value = true
    let newX = e.clientX - dragStart.value.x
    let newY = e.clientY - dragStart.value.y

    // 边界限制
    const btnSize = 36
    const maxX = window.innerWidth - btnSize
    const maxY = window.innerHeight - btnSize
    newX = Math.max(0, Math.min(newX, maxX))
    newY = Math.max(0, Math.min(newY, maxY))

    btnPosition.value = { x: newX, y: newY }
  }

  const onMouseUp = () => {
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }

  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

function handleBtnClick(e) {
  // 如果发生了拖拽，不触发点击
  if (isDragging.value) {
    return
  }
  toggleFullscreen()
}

function installKeyboard() {
  keyboard.onkeydown = (keysym) => {
    client.sendKeyEvent(1, keysym)
  }
  keyboard.onkeyup = (keysym) => {
    client.sendKeyEvent(0, keysym)
  }
}

function uninstallKeyboard() {
  if (keyboard) {
    keyboard.onkeydown = () => {}
    keyboard.onkeyup = () => {}
  }
}

function enableAudio() {
  if (audioEnabled) return

  // 创建或恢复 AudioContext
  if (!audioContext) {
    const AudioContextClass = window.AudioContext || window.webkitAudioContext
    audioContext = new AudioContextClass()
  }

  if (audioContext.state === 'suspended') {
    audioContext.resume().then(() => {
      audioEnabled = true
      console.log('Audio enabled')
    }).catch(err => {
      console.error('Failed to enable audio:', err)
    })
  } else {
    audioEnabled = true
  }
}

function startGuacamole() {
  // 确保viewport有尺寸
  setScreenSize()

  const tunnel = new Guacamole.WebSocketTunnel(wsUrl.value)

  if (client) {
    uninstallKeyboard()
    try {
      client.getDisplay().scale(0)
    } catch (e) {}
  }

  client = new Guacamole.Client(tunnel)

  tunnel.onerror = (status) => {
    console.error(`Tunnel failed ${JSON.stringify(status)}`)
  }

  tunnel.onstatechange = (state) => {
    if (state === Guacamole.Tunnel.State.CLOSED) {
      console.log('Connection closed')
    } else if (state === Guacamole.Tunnel.State.OPEN) {
      // 连接建立后立即resize
      nextTick(() => resize())
    }
  }

  client.onerror = (error) => {
    try {
      client.disconnect()
    } catch (e) {}
    console.error(`Client error ${JSON.stringify(error)}`)
  }

  client.onclipboard = (stream, mimetype) => {
    // Clipboard handling simplified
  }

  const disp = client.getDisplay()
  displayElm = display.value
  if (displayElm) {
    displayElm.appendChild(disp.getElement())

    displayElm.addEventListener('contextmenu', (e) => {
      e.stopPropagation()
      if (e.preventDefault) e.preventDefault()
      e.returnValue = false
    })

    displayElm.onclick = () => {
      displayElm.focus()
      enableAudio()
    }
    displayElm.onfocus = () => displayElm.className = 'focus'
    displayElm.onblur = () => displayElm.className = ''
  }

  const param = serialize(query.value)
  client.connect(param)
  window.onunload = () => {
    try {
      client.disconnect()
    } catch (e) {}
  }

  mouse = new Guacamole.Mouse(displayElm)
  mouse.onmouseout = () => {
    if (!client) return
    try {
      client.getDisplay().showCursor(false)
    } catch (e) {}
  }

  keyboard = new Guacamole.Keyboard(displayElm)
  installKeyboard()
  mouse.onmousedown = mouse.onmouseup = mouse.onmousemove = handleMouseState

  // 连接后多次快速调用resize确保显示，然后减慢频率
  let resizeCount = 0
  let resizeInterval = null

  const doResize = () => {
    if (!client) return
    resize()
    resizeCount++

    // 前10次快速调用（每50ms），之后减慢到每500ms
    if (resizeCount < 10) {
      // 保持当前interval
    } else if (resizeCount === 10) {
      clearInterval(resizeInterval)
      resizeInterval = setInterval(resize, 500)
    } else if (resizeCount >= 60) {
      // 60次后停止自动调用，由 ResizeObserver 处理
      clearInterval(resizeInterval)
    }
  }

  resizeInterval = setInterval(doResize, 50)
}

onMounted(() => {
  // 设置按钮初始位置为右上角
  const btnSize = 36
  btnPosition.value = {
    x: window.innerWidth - btnSize - 12,
    y: 12
  }

  // 等待DOM渲染后启动连接
  nextTick(() => {
    startGuacamole()
  })

  // 监听窗口大小变化
  window.addEventListener('resize', resize)

  // 监听全屏状态变化
  document.addEventListener('fullscreenchange', handleFullscreenChange)

  // 使用 ResizeObserver 监听容器大小变化
  if (viewport.value) {
    resizeObserver = new ResizeObserver(() => {
      resize()
    })
    resizeObserver.observe(viewport.value)
  }
})

onUnmounted(() => {
  // 组件卸载时清理
  if (client) {
    try {
      client.disconnect()
    } catch (e) {}
    client = null
  }
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  window.removeEventListener('resize', resize)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
}
html, body {
  width: 100%;
  height: 100%;
  overflow: hidden;
}
#app {
  width: 100%;
  height: 100%;
}
.display {
  overflow: hidden;
  width: 100%;
  height: 100%;
}
.viewport {
  background-color: #000;
  position: relative;
  width: 100vw;
  height: 100vh;
}
.fullscreen-btn {
  position: absolute;
  width: 36px;
  height: 36px;
  padding: 8px;
  background-color: rgba(0, 0, 0, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 6px;
  color: #fff;
  cursor: grab;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s, border-color 0.2s;
  z-index: 1000;
  user-select: none;
}
.fullscreen-btn:active {
  cursor: grabbing;
  background-color: rgba(0, 0, 0, 0.8);
}
.fullscreen-btn:hover {
  background-color: rgba(0, 0, 0, 0.7);
  border-color: rgba(255, 255, 255, 0.5);
}
</style>