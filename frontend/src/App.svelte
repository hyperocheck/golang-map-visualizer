<script>
  import { onMount, tick } from 'svelte'
  import JSONTree from 'svelte-json-tree'
  let selectedBucket = null
  let selectedKey = null
  let newValueJson = ''
  let hovered = null
  function formatPreview(val) {
    if (val === null || val === undefined) return ''
    if (typeof val === 'object') {
      const str = JSON.stringify(val)
      return str.length > 15 ? str.slice(0, 12) + '...' : str
    }
    return val.toString()
  }
  function withZeroValues(arr) {
    if (!Array.isArray(arr)) return arr
    return arr.map((v) => (v == null ? 'ZeroValue' : v))
  }
  function syncVbAspect() {
    const container = document.getElementById('canvas-container')
    const rect = container.getBoundingClientRect()
    const aspect = rect.height / rect.width
    vb.h = vb.w * aspect
  }

  let cameraInitialized = false
  function fitInitialBuckets(count = 4) {
    if (!svgBuckets.length) return

    const buckets = svgBuckets.slice(0, count)

    let minX = Infinity
    let minY = Infinity
    let maxX = -Infinity
    let maxY = -Infinity

    for (const b of buckets) {
      minX = Math.min(minX, b.x)
      minY = Math.min(minY, b.y)
      maxX = Math.max(maxX, b.x + b.width)
      maxY = Math.max(maxY, b.y + b.height)
    }

    const padding = 80 // визуальный воздух
    minX -= padding
    minY -= padding
    maxX += padding
    maxY += padding

    vb.x = minX
    vb.y = minY
    vb.w = maxX - minX

    syncVbAspect()
  }

  let socket = null
  let stats = null
  const bucketHeaderHeight = 18
  let buckets = []
  let oldBuckets = []
  let chains = []
  let oldChains = []
  let hmap = null
  const tophashHeight = 24
  const rowHeight = 22
  const gapX = 80
  const gapY = 40
  const arrowOffset = 4
  const bucketRadius = 12
  const padding = 12
  const bucketStrokeWidth = 1.5
  const fixedTophashCellWidth = 27
  let canvas
  let ctx
  let dpr = window.devicePixelRatio || 1
  let svgBuckets = [] // Теперь это данные для рендера на canvas
  let svgArrows = []
  let svgLabels = []
  let svgWidth = 2000
  let svgHeight = 2000
  // Состояние viewBox [x, y, width, height]
  let vb = { x: 0, y: 0, w: 1200, h: 900 }
  // Состояние панелей
  let sideWidth = 280
  let lastSideWidth = 280
  let isSideVisible = true
  let resizing = false
  let isPanning = false
  let rafId = null
  let lastMouseX = 0
  let lastMouseY = 0
  async function load() {
    try {
      const [vizRes, oldRes, hmapRes] = await Promise.all([
        fetch('/vizual').catch(() => null),
        fetch('/vizual_old').catch(() => null),
        fetch('/hmap').catch(() => null)
      ])
      if (vizRes?.ok) {
        const data = await vizRes.json()
        stats = data.stats ?? null
        if (Array.isArray(data.buckets) && data.buckets.length > 0) {
          buckets = data.buckets
        } else {
          buckets = []
          chains = []
          oldChains = []
          svgBuckets = []
          svgArrows = []
          svgLabels = []
        }
      }
      if (oldRes?.ok) {
        const data = await oldRes.json()
        if (Array.isArray(data.buckets)) {
          oldBuckets = data.buckets
        } else {
          oldBuckets = []
        }
      }
      if (hmapRes?.ok) {
        hmap = await hmapRes.json()
      }
      chains = []
      if (Array.isArray(buckets)) {
        let current = []
        for (const b of buckets) {
          if (b && b.type === 'main') {
            if (current.length) chains.push(current)
            current = [b]
          } else if (b) {
            current.push(b)
          }
        }
        if (current.length) chains.push(current)
      }
      oldChains = []
      if (Array.isArray(oldBuckets)) {
        let current = []
        for (const b of oldBuckets) {
          if (b && b.type === 'main') {
            if (current.length) oldChains.push(current)
            current = [b]
          } else if (b) {
            current.push(b)
          }
        }
        if (current.length) oldChains.push(current)
      }
      // Всегда сбрасываем viewBox на полный вид после загрузки данных
      buildCanvasData()

      if (!cameraInitialized) {
        fitInitialBuckets(4)
        cameraInitialized = true
        //vb = { x: 0, y: 0, w: svgWidth, h: 0 }
        //syncVbAspect()
        //cameraInitialized = true
      }

      drawCanvas()
    } catch (e) {
      console.error('Ошибка загрузки данных:', e)
    }
  }
  function connectWS() {
    if (socket) {
      socket.close()
    }
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    socket = new WebSocket(`${proto}://${location.host}/ws`)
    socket.onopen = () => {
      console.log('[ws] connected')
    }
    socket.onmessage = () => {
      load()
    }
    socket.onclose = () => {
      console.log('[ws] disconnected, retrying...')
      setTimeout(connectWS, 1000)
    }
    socket.onerror = () => {
      socket.close()
    }
  }
  function buildCanvasData() {
    svgBuckets = []
    svgArrows = []
    svgLabels = []
    const hasOldChains = oldChains && oldChains.length > 0
    const hasNewChains = chains && chains.length > 0
    const showLabels = hasOldChains && hasNewChains
    let mainCountOld = 0
    if (hasOldChains) {
      for (const chain of oldChains) {
        for (const b of chain) {
          if (!b) continue
          if (b.type === 'main') b.displayBid = mainCountOld++
        }
      }
    }
    let mainCountNew = 0
    if (hasNewChains) {
      for (const chain of chains) {
        for (const b of chain) {
          if (!b) continue
          if (b.type === 'main') b.displayBid = mainCountNew++
        }
      }
    }
    const chainWidths = []
    for (
      let idx = 0;
      idx <
      Math.max(
        hasOldChains ? oldChains.length : 0,
        hasNewChains ? chains.length : 0
      );
      idx++
    ) {
      let maxWidth = 260
      if (hasOldChains && idx < oldChains.length && oldChains[idx].length > 0) {
        const chain = oldChains[idx]
        const fixedTophashWidth =
          Math.max(
            ...chain.map((b) =>
              b.tophash && Array.isArray(b.tophash) ? b.tophash.length : 0
            )
          ) * fixedTophashCellWidth
        const widths = chain.map((b) => {
          const keys = b.keys && Array.isArray(b.keys) ? b.keys : []
          const values = b.values && Array.isArray(b.values) ? b.values : []
          const maxLen = Math.max(
            ...keys.map((k) => formatPreview(k).length),
            ...values.map((v) => formatPreview(v).length),
            b.overflow ? b.overflow.toString().length : 0,
            0
          )
          return maxLen * 8 + padding * 2
        })
        maxWidth = Math.max(maxWidth, ...widths, fixedTophashWidth)
      }
      if (hasNewChains && idx < chains.length && chains[idx].length > 0) {
        const chain = chains[idx]
        const fixedTophashWidth =
          Math.max(
            ...chain.map((b) =>
              b.tophash && Array.isArray(b.tophash) ? b.tophash.length : 0
            )
          ) * fixedTophashCellWidth
        const widths = chain.map((b) => {
          const keys = b.keys && Array.isArray(b.keys) ? b.keys : []
          const values = b.values && Array.isArray(b.values) ? b.values : []
          const maxLen = Math.max(
            ...keys.map((k) => formatPreview(k).length),
            ...values.map((v) => formatPreview(v).length),
            b.overflow ? b.overflow.toString().length : 0,
            0
          )
          return maxLen * 8 + padding * 2
        })
        maxWidth = Math.max(maxWidth, ...widths, fixedTophashWidth)
      }
      chainWidths.push(maxWidth)
    }
    let x = gapX
    let oldMaxY = 0
    let newMaxY = 0
    let oldStartY = gapY + (showLabels ? 30 : 0)
    let newStartY = gapY
    if (hasOldChains) {
      for (let idx = 0; idx < oldChains.length; idx++) {
        const chain = oldChains[idx]
        if (!chain || chain.length === 0) continue
        const width = chainWidths[idx] || 260
        let y = oldStartY
        for (let i = 0; i < chain.length; i++) {
          const b = chain[i]
          if (!b) continue
          const keys = b.keys && Array.isArray(b.keys) ? b.keys : []
          const values = b.values && Array.isArray(b.values) ? b.values : []
          const height =
            bucketHeaderHeight +
            tophashHeight +
            keys.length * rowHeight +
            values.length * rowHeight +
            rowHeight +
            padding * 2
          svgBuckets.push({
            id: `old-${b.id}`,
            x,
            y,
            width,
            height,
            bucket: b,
            padding,
            isOld: true
          })
          if (i < chain.length - 1) {
            svgArrows.push({
              x: x + width / 2,
              y1: y + height,
              y2: y + height + gapY - arrowOffset,
              isOld: true
            })
            y += height + gapY
          } else {
            y += height + gapY
          }
          if (y > oldMaxY) oldMaxY = y
        }
        x += width + gapX
      }
      if (showLabels) {
        svgLabels.push({ x: gapX, y: gapY + 20, text: 'OLD', isOld: true })
      }
      newStartY = oldMaxY + gapY * 2
    }
    if (hasNewChains) {
      x = gapX
      for (let idx = 0; idx < chains.length; idx++) {
        const chain = chains[idx]
        if (!chain || chain.length === 0) continue
        const width = chainWidths[idx] || 260
        let y = newStartY
        for (let i = 0; i < chain.length; i++) {
          const b = chain[i]
          if (!b) continue
          const keys = b.keys && Array.isArray(b.keys) ? b.keys : []
          const values = b.values && Array.isArray(b.values) ? b.values : []
          const height =
            bucketHeaderHeight +
            tophashHeight +
            keys.length * rowHeight +
            values.length * rowHeight +
            rowHeight +
            padding * 2
          svgBuckets.push({
            id: b.id,
            x,
            y,
            width,
            height,
            bucket: b,
            padding,
            isOld: false
          })
          if (i < chain.length - 1) {
            svgArrows.push({
              x: x + width / 2,
              y1: y + height,
              y2: y + height + gapY - arrowOffset,
              isOld: false
            })
            y += height + gapY
          } else {
            y += height + gapY
          }
          if (y > newMaxY) newMaxY = y
        }
        x += width + gapX
      }
      if (showLabels) {
        svgLabels.push({
          x: gapX,
          y: newStartY - gapY + 20,
          text: 'NEW',
          isOld: false
        })
      }
    }
    svgWidth = x + 200
    svgHeight = Math.max(oldMaxY, newMaxY) + 200
  }
  function drawCanvas() {
    if (!ctx) return
    const container = document.getElementById('canvas-container')
    const rect = container.getBoundingClientRect()

    canvas.width = rect.width * dpr
    canvas.height = rect.height * dpr
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, rect.width, rect.height)
    ctx.scale(dpr, dpr)
    const scaleX = rect.width / vb.w
    const scaleY = rect.height / vb.h
    const offsetX = -vb.x * scaleX
    const offsetY = -vb.y * scaleY
    ctx.translate(offsetX, offsetY)
    ctx.scale(scaleX, scaleY)
    // Рендерим только видимые элементы
    const visibleBuckets = svgBuckets.filter((b) => {
      const bx = b.x
      const by = b.y
      const bw = b.width
      const bh = b.height
      return (
        bx + bw > vb.x && bx < vb.x + vb.w && by + bh > vb.y && by < vb.y + vb.h
      )
    })
    const visibleArrows = svgArrows.filter((a) => {
      const ax = a.x
      const ay1 = a.y1
      const ay2 = a.y2
      return (
        ax > vb.x &&
        ax < vb.x + vb.w &&
        ((ay1 > vb.y && ay1 < vb.y + vb.h) || (ay2 > vb.y && ay2 < vb.y + vb.h))
      )
    })
    const visibleLabels = svgLabels.filter(
      (l) => l.x > vb.x && l.x < vb.x + vb.w && l.y > vb.y && l.y < vb.y + vb.h
    )
    // Рендерим labels
    ctx.font = 'bold 14px JetBrains Mono'
    visibleLabels.forEach((label) => {
      ctx.fillStyle = label.isOld ? '#ff6b6b' : '#51cf66'
      ctx.fillText(label.text, label.x, label.y)
    })
    // Рендерим buckets
    visibleBuckets.forEach((b) => {
      drawBucket(b, scaleX, scaleY)
    })
    // Рендерим arrows
    visibleArrows.forEach((a) => {
      ctx.strokeStyle = a.isOld ? '#ff6b6b' : '#000'
      ctx.lineWidth = 1.5 / Math.min(scaleX, scaleY)
      ctx.beginPath()
      ctx.moveTo(a.x, a.y1)
      ctx.lineTo(a.x, a.y2)
      ctx.stroke()
      // Arrow head
      ctx.fillStyle = a.isOld ? '#ff6b6b' : '#000'
      ctx.beginPath()
      ctx.moveTo(a.x, a.y2)
      ctx.lineTo(a.x - 4, a.y2 - 8)
      ctx.lineTo(a.x + 4, a.y2 - 8)
      ctx.closePath()
      ctx.fill()
    })
    ctx.resetTransform()
  }
  function drawBucket(b, scaleX, scaleY) {
    let bucketStroke = b.isOld ? '#ff6b6b' : '#000'
    const strokeW = b.isOld ? 2 : bucketStrokeWidth
    ctx.lineWidth = strokeW / Math.min(scaleX, scaleY)

    if (
      hovered &&
      hovered.bucket.id === b.bucket.id &&
      hovered.isOld === b.isOld
    ) {
      bucketStroke = b.isOld ? '#ff8787' : '#228be6'
    }
    ctx.fillStyle = '#fff'
    ctx.strokeStyle = bucketStroke
    ctx.lineWidth = bucketStrokeWidth / Math.min(scaleX, scaleY)
    ctx.beginPath()
    ctx.roundRect(b.x, b.y, b.width, b.height, bucketRadius)
    ctx.fill()
    ctx.stroke()
    // Header
    if (b.bucket?.type === 'main') {
      ctx.font = 'bold 11px JetBrains Mono'
      ctx.fillStyle = '#495057'
      ctx.textAlign = 'left'
      ctx.fillText(
        `bid ${b.bucket.displayBid}`,
        b.x + b.padding,
        b.y + b.padding + 12
      )
    }
    // Tophash
    const tophash = b.bucket?.tophash || []
    tophash.forEach((t, i) => {
      ctx.fillStyle = '#eee'
      ctx.strokeStyle = '#000'
      ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
      ctx.fillRect(
        b.x + b.padding + i * fixedTophashCellWidth,
        b.y + b.padding + bucketHeaderHeight,
        fixedTophashCellWidth,
        tophashHeight
      )
      ctx.strokeRect(
        b.x + b.padding + i * fixedTophashCellWidth,
        b.y + b.padding + bucketHeaderHeight,
        fixedTophashCellWidth,
        tophashHeight
      )
      ctx.font = '12px JetBrains Mono'
      ctx.fillStyle = '#000'
      ctx.textAlign = 'center'
      ctx.fillText(
        t,
        b.x + b.padding + i * fixedTophashCellWidth + fixedTophashCellWidth / 2,
        b.y + b.padding + bucketHeaderHeight + tophashHeight / 1.5
      )
    })
    // Keys
    const keys = b.bucket?.keys || []
    keys.forEach((k, i) => {
      let fill = k == null ? '#dbfdc9' : '#b2f2bb'
      if (
        hovered &&
        hovered.type === 'key' &&
        hovered.bucket.id === b.bucket.id &&
        hovered.isOld === b.isOld &&
        hovered.index === i
      ) {
        fill = '#9feaa4'
      }
      ctx.fillStyle = fill
      ctx.strokeStyle = '#12b886'
      ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
      ctx.fillRect(
        b.x + b.padding,
        b.y + b.padding + bucketHeaderHeight + tophashHeight + i * rowHeight,
        b.width - padding * 2,
        rowHeight
      )
      ctx.strokeRect(
        b.x + b.padding,
        b.y + b.padding + bucketHeaderHeight + tophashHeight + i * rowHeight,
        b.width - padding * 2,
        rowHeight
      )
      ctx.font = '13px JetBrains Mono'
      ctx.fillStyle = '#000'
      ctx.textAlign = 'left'
      ctx.fillText(
        formatPreview(k),
        b.x + b.padding + 6,
        b.y +
          b.padding +
          bucketHeaderHeight +
          tophashHeight +
          i * rowHeight +
          rowHeight / 1.5
      )
    })
    // Values
    const values = b.bucket?.values || []
    const keysLen = keys.length
    values.forEach((v, i) => {
      let fill = v == null ? '#fff2b8' : '#ffec99'
      if (
        hovered &&
        hovered.type === 'value' &&
        hovered.bucket.id === b.bucket.id &&
        hovered.isOld === b.isOld &&
        hovered.index === i
      ) {
        fill = '#ffe066'
      }
      ctx.fillStyle = fill
      ctx.strokeStyle = '#ffa94d'
      ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
      ctx.fillRect(
        b.x + b.padding,
        b.y +
          b.padding +
          bucketHeaderHeight +
          tophashHeight +
          keysLen * rowHeight +
          i * rowHeight,
        b.width - padding * 2,
        rowHeight
      )
      ctx.strokeRect(
        b.x + b.padding,
        b.y +
          b.padding +
          bucketHeaderHeight +
          tophashHeight +
          keysLen * rowHeight +
          i * rowHeight,
        b.width - padding * 2,
        rowHeight
      )
      ctx.font = '13px JetBrains Mono'
      ctx.fillStyle = '#000'
      ctx.fillText(
        formatPreview(v),
        b.x + b.padding + 6,
        b.y +
          b.padding +
          bucketHeaderHeight +
          tophashHeight +
          keysLen * rowHeight +
          i * rowHeight +
          rowHeight / 1.5
      )
    })
    // Overflow
    if (b.bucket) {
      const valuesLen = values.length
      ctx.fillStyle = '#ddd'
      ctx.strokeStyle = '#000'
      ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
      ctx.fillRect(
        b.x + b.padding,
        b.y +
          b.padding +
          bucketHeaderHeight +
          tophashHeight +
          keysLen * rowHeight +
          valuesLen * rowHeight,
        b.width - padding * 2,
        rowHeight
      )
      ctx.strokeRect(
        b.x + b.padding,
        b.y +
          b.padding +
          bucketHeaderHeight +
          tophashHeight +
          keysLen * rowHeight +
          valuesLen * rowHeight,
        b.width - padding * 2,
        rowHeight
      )
      ctx.font = '12px JetBrains Mono'
      ctx.fillStyle = '#000'
      ctx.fillText(
        b.bucket.overflow || '',
        b.x + b.padding + 6,
        b.y +
          b.padding +
          bucketHeaderHeight +
          tophashHeight +
          keysLen * rowHeight +
          valuesLen * rowHeight +
          rowHeight / 1.5
      )
    }
    // Selected highlight
    if (
      selectedKey &&
      selectedKey.bucket.id === b.bucket.id &&
      selectedKey.isOld === b.isOld
    ) {
      ctx.strokeStyle = '#ff0000'
      ctx.lineWidth = 3 / Math.min(scaleX, scaleY)
      ctx.strokeRect(
        b.x + b.padding,
        b.y +
          b.padding +
          bucketHeaderHeight +
          tophashHeight +
          selectedKey.index * rowHeight,
        b.width - padding * 2,
        rowHeight
      )
    }
  }
  // Навигация
  function handleWheel(e) {
    e.preventDefault()
    const rect = e.currentTarget.getBoundingClientRect()
    let newVb = { ...vb }
    if (e.ctrlKey || e.metaKey) {
      const mouseX = e.clientX - rect.left
      const mouseY = e.clientY - rect.top
      const svgMouseX = vb.x + (mouseX * vb.w) / rect.width
      const svgMouseY = vb.y + (mouseY * vb.h) / rect.height
      const zoomFactor = Math.pow(1.001, e.deltaY * -2)
      const newW = vb.w / zoomFactor
      const newH = vb.h / zoomFactor
      if (newW < 100 || newW > 40000) return
      newVb.x = svgMouseX - (mouseX / rect.width) * newW
      newVb.y = svgMouseY - (mouseY / rect.height) * newH
      newVb.w = newW
      newVb.h = newH
    } else {
      newVb.x += (e.deltaX * vb.w) / rect.width
      newVb.y += (e.deltaY * vb.h) / rect.height
    }
    scheduleVbUpdate(newVb)
  }
  function handleMouseDown(e) {
    if (e.button === 0) {
      isPanning = true
      lastMouseX = e.clientX
      lastMouseY = e.clientY
    }
  }
  function handleMouseUp() {
    isPanning = false
  }
  function handleMouseMove(e) {
    if (isPanning) {
      const dx = e.clientX - lastMouseX
      const dy = e.clientY - lastMouseY
      lastMouseX = e.clientX
      lastMouseY = e.clientY
      const rect = document
        .getElementById('canvas-container')
        .getBoundingClientRect()
      let newVb = { ...vb }
      newVb.x -= (dx * vb.w) / rect.width
      newVb.y -= (dy * vb.h) / rect.height
      scheduleVbUpdate(newVb)
    }
    handleHover(e)
  }
  function handleHover(e) {
    const rect = canvas.getBoundingClientRect()
    const hoverX = ((e.clientX - rect.left) / rect.width) * vb.w + vb.x
    const hoverY = ((e.clientY - rect.top) / rect.height) * vb.h + vb.y
    let newHovered = null
    for (const b of svgBuckets) {
      if (hoverX >= b.x + b.padding && hoverX <= b.x + b.width - b.padding) {
        // Check for keys
        const keyYStart = b.y + b.padding + bucketHeaderHeight + tophashHeight
        const keyYEnd = keyYStart + (b.bucket.keys || []).length * rowHeight
        if (hoverY >= keyYStart && hoverY <= keyYEnd) {
          const localY = hoverY - keyYStart
          const index = Math.floor(localY / rowHeight)
          newHovered = { type: 'key', bucket: b.bucket, isOld: b.isOld, index }
          break
        }
        // Check for values
        const valueYStart = keyYEnd
        const valueYEnd =
          valueYStart + (b.bucket.values || []).length * rowHeight
        if (hoverY >= valueYStart && hoverY <= valueYEnd) {
          const localY = hoverY - valueYStart
          const index = Math.floor(localY / rowHeight)
          newHovered = {
            type: 'value',
            bucket: b.bucket,
            isOld: b.isOld,
            index
          }
          break
        }
      }
      // Check for bucket hover
      if (
        hoverX >= b.x &&
        hoverX <= b.x + b.width &&
        hoverY >= b.y &&
        hoverY <= b.y + b.height
      ) {
        newHovered = { type: 'bucket', bucket: b.bucket, isOld: b.isOld }
        break
      }
    }
    if (JSON.stringify(newHovered) !== JSON.stringify(hovered)) {
      hovered = newHovered
      drawCanvas()
    }
  }
  function handleSingleClick(e) {
    const rect = canvas.getBoundingClientRect()
    const clickX = ((e.clientX - rect.left) / rect.width) * vb.w + vb.x
    const clickY = ((e.clientY - rect.top) / rect.height) * vb.h + vb.y
    for (const b of svgBuckets) {
      if (
        clickX >= b.x &&
        clickX <= b.x + b.width &&
        clickY >= b.y &&
        clickY <= b.y + b.height
      ) {
        selectedKey = null
        selectedBucket = b.bucket
        drawCanvas()
        return
      }
    }
  }
  function handleDblClick(e) {
    const rect = canvas.getBoundingClientRect()
    const clickX = ((e.clientX - rect.left) / rect.width) * vb.w + vb.x
    const clickY = ((e.clientY - rect.top) / rect.height) * vb.h + vb.y
    // Находим бакет и ячейку под кликом
    for (const b of svgBuckets) {
      if (clickX >= b.x + b.padding && clickX <= b.x + b.width - b.padding) {
        const keyYStart = b.y + b.padding + bucketHeaderHeight + tophashHeight
        const keyYEnd = keyYStart + (b.bucket.keys || []).length * rowHeight
        if (clickY >= keyYStart && clickY <= keyYEnd) {
          const localY = clickY - keyYStart
          const index = Math.floor(localY / rowHeight)
          if (
            index >= 0 &&
            index < b.bucket.keys.length &&
            b.bucket.keys[index] != null
          ) {
            selectKey(b.bucket, index, b.isOld)
            drawCanvas()
            return
          }
        }
      }
    }
  }
  function scheduleVbUpdate(newVb) {
    if (rafId) cancelAnimationFrame(rafId)
    rafId = requestAnimationFrame(() => {
      vb = newVb
      drawCanvas()
      rafId = null
    })
  }
  function startResize(e) {
    resizing = true
    const startMouseX = e.clientX
    const startSideWidth = sideWidth
    const container = document.getElementById('canvas-container')
    const initialRect = container.getBoundingClientRect()
    const unitsPerPixel = vb.w / initialRect.width
    const aspect = initialRect.height / initialRect.width
    const onMouseMove = (ev) => {
      if (!resizing) return
      const dx = startMouseX - ev.clientX
      const newSideWidth = Math.max(100, Math.min(800, startSideWidth + dx))
      const diffPx = newSideWidth - sideWidth
      sideWidth = newSideWidth
      let newVb = { ...vb }
      newVb.w -= diffPx * unitsPerPixel
      const currentContainerWidth =
        initialRect.width - (newSideWidth - startSideWidth)
      vb.w -= diffPx * unitsPerPixel
      syncVbAspect()
      scheduleVbUpdate(newVb)
    }
    const onMouseUp = () => {
      resizing = false
      window.removeEventListener('mousemove', onMouseMove)
      window.removeEventListener('mouseup', onMouseUp)
    }
    window.addEventListener('mousemove', onMouseMove)
    window.addEventListener('mouseup', onMouseUp)
  }
  function toggleSide() {
    const container = document.getElementById('canvas-container')
    const rect = container.getBoundingClientRect()
    const unitsPerPixel = vb.w / rect.width
    if (isSideVisible) {
      lastSideWidth = sideWidth
      sideWidth = 0
      isSideVisible = false
    } else {
      sideWidth = lastSideWidth
      isSideVisible = true
    }
    requestAnimationFrame(() => {
      const newRect = container.getBoundingClientRect()
      let newVb = { ...vb }
      newVb.w = unitsPerPixel * newRect.width
      newVb.h = newVb.w * (newRect.height / newRect.width)
      vb = newVb
      drawCanvas()
    })
  }
  function selectKey(bucket, index, isOld) {
    if (bucket.keys[index] == null) return
    selectedBucket = bucket
    selectedKey = {
      bucket,
      index,
      key: bucket.keys[index],
      value: bucket.values[index],
      isOld
    }
    newValueJson = JSON.stringify(selectedKey.value, null, 2)
  }
  function handleKeyDown(e) {
    if (
      e.target instanceof HTMLInputElement ||
      e.target instanceof HTMLTextAreaElement
    ) {
      return
    }
    if (!selectedKey) return
    if (e.key === 'Escape') {
      selectedKey = null
      drawCanvas()
      return
    }
    if (e.key === 'd') {
      fetch('/delete_key', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: selectedKey.key })
      })
        .then((res) => {
          if (res.ok) {
            load()
            selectedKey = null
          } else {
            alert('Error deleting key')
          }
        })
        .catch((err) => {
          console.error(err)
          alert('Error deleting key')
        })
    } else if (e.key === 'u') {
      let newVal = newValueJson
      try {
        newVal = JSON.parse(newValueJson)
      } catch {}
      fetch('/update_key', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: selectedKey.key, value: newVal })
      })
        .then((res) => {
          if (res.ok) {
            load()
            selectedKey = null
          } else {
            alert('Error updating key')
          }
        })
        .catch((err) => {
          console.error(err)
          alert('Error updating key')
        })
    }
  }
  onMount(() => {
    load()
    connectWS()
    ctx = canvas.getContext('2d')
    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener('resize', () => {
      syncVbAspect()
      drawCanvas()
    })
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      window.removeEventListener('resize', () => drawCanvas())
    }
  })
</script>

<div class="root">
  <div
    id="canvas-container"
    on:wheel|nonpassive={handleWheel}
    on:mousedown={handleMouseDown}
    on:mousemove={handleMouseMove}
    on:mouseup={handleMouseUp}
    on:mouseleave={handleMouseUp}
    on:dblclick={handleDblClick}
    on:click={handleSingleClick}
  >
    <button class="toggle-btn" on:click={toggleSide}>
      {isSideVisible ? '→' : '← hmap'}
    </button>
    <canvas
      bind:this={canvas}
      width="100%"
      height="100%"
      style="width: 100%; height: 100%;"
    ></canvas>
  </div>
  {#if isSideVisible}
    <div class="splitter" on:mousedown={startResize}></div>
  {/if}
  <div
    class="side"
    style="width: {sideWidth}px; display: {isSideVisible ? 'block' : 'none'};"
  >
    <h3>hmap</h3>
    {#if hmap}
      {#each Object.entries(hmap) as [k, v]}
        <div class="row"><span>{k}</span><b>{v}</b></div>
      {/each}
    {/if}
    {#if stats}
      <h3>stats</h3>
      <div class="row">
        <span>Load factor</span>
        <b>{stats.loadFactor.toFixed(2)}</b>
      </div>
      <div class="row">
        <span>Max chain length</span>
        <b>{stats.maxChainLen}</b>
      </div>
      <div class="row">
        <span>First BID with max chain</span>
        <b>
          {stats.maxChainBucketID >= 0 ? stats.maxChainBucketID : '—'}
        </b>
      </div>
      <div class="row">
        <span>Chains count</span>
        <b>{stats.numChains}</b>
      </div>
      <div class="row">
        <span>Empty buckets</span>
        <b>{stats.numEmptyBuckets}</b>
      </div>
      <div class="row">
        <span>Key type</span>
        <b>{stats.keytype}</b>
      </div>
      <div class="row">
        <span>Value type</span>
        <b>{stats.valuetype}</b>
      </div>
    {/if}
    <!-- НОВАЯ СЕКЦИЯ: ИНСПЕКТОР БАКЕТА -->
    <div
      class="inspector-section"
      style="margin-top: 20px; border-top: 2px solid #ccc; padding-top: 10px;"
    >
      <h3>inspector</h3>
      {#if selectedKey}
        <div style="margin-bottom: 10px; color: #555;">
          <small
            >({selectedKey.bucket.type}) - Key at index {selectedKey.index}</small
          >
        </div>
        <div class="tree-label">Selected Key:</div>
        <div class="tree-container">
          <JSONTree value={selectedKey.key} />
        </div>
        <div class="tree-label" style="margin-top: 10px;">Current Value:</div>
        <div class="tree-container">
          <JSONTree value={selectedKey.value} />
        </div>
        <div class="tree-label" style="margin-top: 10px;">
          New Value (JSON or raw string):
        </div>
        <textarea
          bind:value={newValueJson}
          style="width: 100%; height: 100px; font-family: monospace; font-size: 12px;"
        />
        <p style="font-size: 11px; color: #666; margin-top: 5px;">
          Enter JSON for objects or raw string/number. Press 'd' to delete,
          'u' to update, 'Esc' to deselect.
        </p>
      {:else if selectedBucket}
        <div style="margin-bottom: 10px; color: #555;">
          <small>({selectedBucket.type})</small>
        </div>
        <div class="tree-label">Keys:</div>
        <div class="tree-container">
          <JSONTree value={withZeroValues(selectedBucket.keys)} />
        </div>
        <div class="tree-label" style="margin-top: 10px;">Values:</div>
        <div class="tree-container">
          <JSONTree value={withZeroValues(selectedBucket.values)} />
        </div>
        <div class="row" style="margin-top: 10px; font-size: 11px;">
          <span>Overflow</span>
          <code>{selectedBucket.overflow}</code>
        </div>
      {:else}
        <p
          style="color: #999; font-size: 12px; font-style: italic; text-align: center;"
        >
		  Click on the bucket to analyze
        </p>
      {/if}
    </div>
  </div>
</div>

<style>
  html,
  body {
    margin: 0;
    padding: 0;
    width: 100%;
    height: 100%;
  }

  * {
    box-sizing: border-box;
  }

  .root {
    display: flex;
    width: 100%;
    height: 100dvh;
    overflow: hidden;
    background: #ebfbee;
  }
  #canvas-container {
    flex: 1;
    position: relative;
    overflow: hidden;
    cursor: grab;
    touch-action: none;
  }
  #canvas-container:active {
    cursor: grabbing;
  }
  .toggle-btn {
    position: absolute;
    right: 15px;
    top: 15px;
    z-index: 100;
    background: #ffffff;
    border: 1px solid #ccc;
    padding: 6px 12px;
    border-radius: 6px;
    cursor: pointer;
    font-family: sans-serif;
    font-weight: 500;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
  .splitter {
    width: 6px;
    background: #ccc;
    cursor: col-resize;
    user-select: none;
    z-index: 10;
  }
  .side {
    padding: 12px;
    border-left: 1px solid #ccc;
    background: #fafafa;
    overflow-y: auto;
    font-family: monospace;
    font-size: 13px;
    flex-shrink: 0;
  }
  .row {
    display: flex;
    justify-content: space-between;
    padding: 2px 0;
    border-bottom: 1px solid #eee;
  }
  .tree-label {
    font-size: 11px;
    font-weight: bold;
    color: #666;
    text-transform: uppercase;
    margin-bottom: 4px;
  }
  .tree-container {
    background: #fff;
    border: 1px solid #eee;
    border-radius: 4px;
    padding: 4px;
    max-height: 250px;
    overflow: auto;
  }
  code {
    background: #eee;
    padding: 1px 4px;
    border-radius: 3px;
  }
  :global(html, body) {
    margin: 0;
    padding: 0;
    width: 100%;
    height: 100%;
    overflow: hidden; /* ← ВАЖНО */
  }
</style>
