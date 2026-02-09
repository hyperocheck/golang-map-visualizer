<script>
  import { onMount } from 'svelte'

  let canvas, ctx
  let dpr = window.devicePixelRatio || 1
  let data = null

  let vb = { x: 0, y: 0, w: 2000, h: 1500 }

  let isPanning = false
  let lastMouseX = 0
  let lastMouseY = 0

  // Состояние наведения
  let hoveredTable = null
  let hoveredGroup = null
  let hoveredSlot = null
  let mouseCanvasX = 0
  let mouseCanvasY = 0

  const ctrlCellSize = 27
  const slotHeight = 22
  const padding = 12
  const groupGapX = 40
  const tableGapY = 60
  const groupRadius = 8
  const addrBlockWidth = 80
  const addrBlockGap = 100

  function base64ToBytes(b64) {
    return Array.from(atob(b64), c => c.charCodeAt(0))
  }

  function isEmptySlot(b) {
    return b === 0x80 || b === 128
  }

  function tableColor(i) {
    const c = [
      { bg: 'rgba(255,200,200,0.15)', border: 'rgba(255,100,100,0.6)' },
      { bg: 'rgba(200,220,255,0.15)', border: 'rgba(100,150,255,0.6)' },
      { bg: 'rgba(200,255,200,0.15)', border: 'rgba(100,200,100,0.6)' }
    ]
    return c[i % c.length]
  }

  function isVisible(x, y, w, h) {
    return !(
      x + w < vb.x ||
      x > vb.x + vb.w ||
      y + h < vb.y ||
      y > vb.y + vb.h
    )
  }

  function isPointInRect(px, py, x, y, w, h) {
    return px >= x && px <= x + w && py >= y && py <= y + h
  }

  async function loadData() {
    data = await (await fetch('/data')).json()
    draw()
  }

  function draw() {
    if (!ctx || !data) return

    const r = canvas.parentElement.getBoundingClientRect()
    canvas.width = r.width * dpr
    canvas.height = r.height * dpr

    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.scale(dpr, dpr)

    const sx = r.width / vb.w
    const sy = r.height / vb.h
    ctx.translate(-vb.x * sx, -vb.y * sy)
    ctx.scale(sx, sy)

    let currentY = 40

    const map = new Map()
    data.tables.forEach((t, i) => {
      if (!map.has(t.addr)) map.set(t.addr, { table: t, indices: [] })
      map.get(t.addr).indices.push(i)
    })

    let idx = 0
    let newHoveredTable = null
    let newHoveredGroup = null
    let newHoveredSlot = null

    for (const [addr, { table, indices }] of map) {
      const col = tableColor(idx)

      let tableW = 0
      let tableH = 0

      for (const g of table.groups) {
        const ctrls = base64ToBytes(g.ctrls)
        const gw = ctrls.length * ctrlCellSize + padding * 2
        const gh = padding + ctrlCellSize + 16 * slotHeight + padding
        tableW += gw + groupGapX
        tableH = Math.max(tableH, gh)
      }

      tableW = tableW - groupGapX + 40
      tableH += 40

      const addrGapY = 20
      const addrBlockH = tableH
      const addrColH =
        indices.length * addrBlockH +
        (indices.length - 1) * addrGapY

      const blockH = Math.max(tableH, addrColH)
      const tableYOffset =
        addrColH > tableH ? (addrColH - tableH) / 2 : 0

      const tableX = 40 + addrBlockWidth + addrBlockGap - 20
      const tableY = currentY - 20 + tableYOffset
      const groupsY = tableY + 20

      const addrColY =
        currentY - 20 + blockH / 2 - addrColH / 2

      if (!isVisible(40, currentY - 20, tableX + tableW, blockH)) {
        currentY += blockH + tableGapY
        idx++
        continue
      }

      // Проверка наведения на таблицу
      const isTableHovered = isPointInRect(mouseCanvasX, mouseCanvasY, tableX, tableY, tableW, tableH)
      if (isTableHovered) {
        newHoveredTable = idx
      }

      // TABLE FRAME
      ctx.fillStyle = col.bg
      ctx.strokeStyle = isTableHovered ? '#4169E1' : col.border
      ctx.lineWidth = isTableHovered ? 4 / Math.min(sx, sy) : 3 / Math.min(sx, sy)
      ctx.beginPath()
      ctx.roundRect(tableX, tableY, tableW, tableH, 16)
      ctx.fill()
      ctx.stroke()

      // ADDR BLOCKS + ARROWS
      indices.forEach((_, i) => {
        const x = 40
        const y = addrColY + i * (addrBlockH + addrGapY)

        ctx.fillStyle = col.bg
        ctx.strokeStyle = col.border
        ctx.lineWidth = 2 / Math.min(sx, sy)
        ctx.beginPath()
        ctx.roundRect(x, y, addrBlockWidth, addrBlockH, 8)
        ctx.fill()
        ctx.stroke()

        ctx.save()
        ctx.translate(x + addrBlockWidth / 2, y + addrBlockH / 2)
        ctx.rotate(-Math.PI / 2)
        ctx.font = 'bold 28px monospace'
        ctx.fillStyle = col.border
        ctx.textAlign = 'center'
        ctx.textBaseline = 'middle'
        ctx.fillText(`0x${addr.toString(16)}`, 0, 0)
        ctx.restore()

        const ax = x + addrBlockWidth
        const ay = y + addrBlockH / 2
        const bx = tableX - 6
        const by = tableY + tableH / 2

        ctx.strokeStyle = col.border
        ctx.lineWidth = 2 / Math.min(sx, sy)
        ctx.beginPath()
        ctx.moveTo(ax, ay)
        ctx.lineTo(bx, by)
        ctx.stroke()

        const ang = Math.atan2(by - ay, bx - ax)
        const arrowSize = 12
        ctx.fillStyle = col.border
        ctx.beginPath()
        ctx.moveTo(bx, by)
        ctx.lineTo(
          bx - arrowSize * Math.cos(ang - 0.4),
          by - arrowSize * Math.sin(ang - 0.4)
        )
        ctx.lineTo(
          bx - arrowSize * Math.cos(ang + 0.4),
          by - arrowSize * Math.sin(ang + 0.4)
        )
        ctx.closePath()
        ctx.fill()
      })

      // GROUPS
      let gx = 40 + addrBlockWidth + addrBlockGap
      let groupIdx = 0

      for (const g of table.groups) {
        const ctrls = base64ToBytes(g.ctrls)
        const gw = ctrls.length * ctrlCellSize + padding * 2
        const gh = padding + ctrlCellSize + 16 * slotHeight + padding

        if (!isVisible(gx, groupsY, gw, gh)) {
          gx += gw + groupGapX
          groupIdx++
          continue
        }

        // Проверка наведения на группу
        const isGroupHovered = isPointInRect(mouseCanvasX, mouseCanvasY, gx, groupsY, gw, gh)
        if (isGroupHovered) {
          newHoveredGroup = `${idx}-${groupIdx}`
        }

        ctx.fillStyle = 'rgba(255,255,255,0.85)'
        ctx.strokeStyle = isGroupHovered ? '#4169E1' : '#000'
        ctx.lineWidth = isGroupHovered ? 2.5 / Math.min(sx, sy) : 1.5 / Math.min(sx, sy)
        ctx.beginPath()
        ctx.roundRect(gx, groupsY, gw, gh, groupRadius)
        ctx.fill()
        ctx.stroke()

        let cx = gx + padding
        let cy = groupsY + padding

        const ctrlsLen = ctrls.length
        for (let i = 0; i < ctrlsLen; i++) {
          const c = ctrls[i]
          ctx.fillStyle = isEmptySlot(c) ? '#ddd' : '#a8dadc'
          ctx.strokeStyle = '#000'
          ctx.lineWidth = 1 / Math.min(sx, sy)
          ctx.fillRect(cx, cy, ctrlCellSize, ctrlCellSize)
          ctx.strokeRect(cx, cy, ctrlCellSize, ctrlCellSize)

          ctx.font = '11px monospace'
          ctx.fillStyle = '#000'
          ctx.textAlign = 'center'
          ctx.textBaseline = 'middle'
          ctx.fillText(c.toString(), cx + ctrlCellSize / 2, cy + ctrlCellSize / 2)

          cx += ctrlCellSize
        }

        let slotY = cy + ctrlCellSize
        const scale = Math.min(sx, sy)
        const showDetails = scale > 0.3

        for (let i = 0; i < ctrlsLen; i++) {
          const empty = isEmptySlot(ctrls[i])
          const s = g.slots[i]

          const slotId = `${idx}-${groupIdx}-${i}`

          // Проверка наведения на key слот
          const isKeyHovered = isPointInRect(mouseCanvasX, mouseCanvasY, gx + padding, slotY, gw - padding * 2, slotHeight)
          if (isKeyHovered) {
            newHoveredSlot = `${slotId}-key`
          }

          // Key row
          let keyBg = empty ? '#e8f5e9' : '#b2f2bb'
          let keyBorder = empty ? '#c8e6c9' : '#12b886'
          
          if (isKeyHovered) {
            keyBg = empty ? '#c8e6c9' : '#69f0ae'
            keyBorder = '#00897b'
          }

          ctx.fillStyle = keyBg
          ctx.strokeStyle = keyBorder
          ctx.lineWidth = isKeyHovered ? 2 / Math.min(sx, sy) : 1 / Math.min(sx, sy)
          ctx.fillRect(gx + padding, slotY, gw - padding * 2, slotHeight)
          ctx.strokeRect(gx + padding, slotY, gw - padding * 2, slotHeight)

          if (showDetails && !empty && s) {
            ctx.font = 'bold 12px monospace'
            ctx.fillStyle = '#000'
            ctx.textAlign = 'left'
            ctx.textBaseline = 'middle'
            ctx.fillText(String(s.k || ''), gx + padding + 6, slotY + slotHeight / 2)
          }

          slotY += slotHeight

          // Проверка наведения на value слот
          const isValueHovered = isPointInRect(mouseCanvasX, mouseCanvasY, gx + padding, slotY, gw - padding * 2, slotHeight)
          if (isValueHovered) {
            newHoveredSlot = `${slotId}-value`
          }

          // Value row
          let valueBg = empty ? '#fff9e6' : '#ffec99'
          let valueBorder = empty ? '#ffe0b2' : '#ffa94d'
          
          if (isValueHovered) {
            valueBg = empty ? '#ffe0b2' : '#ffd54f'
            valueBorder = '#ff8f00'
          }

          ctx.fillStyle = valueBg
          ctx.strokeStyle = valueBorder
          ctx.lineWidth = isValueHovered ? 2 / Math.min(sx, sy) : 1 / Math.min(sx, sy)
          ctx.fillRect(gx + padding, slotY, gw - padding * 2, slotHeight)
          ctx.strokeRect(gx + padding, slotY, gw - padding * 2, slotHeight)

          if (showDetails && !empty && s) {
            ctx.font = 'bold 12px monospace'
            ctx.fillStyle = '#000'
            ctx.textAlign = 'left'
            ctx.textBaseline = 'middle'
            ctx.fillText(String(s.v || ''), gx + padding + 6, slotY + slotHeight / 2)
          }
          
          slotY += slotHeight
        }

        gx += gw + groupGapX
        groupIdx++
      }

      currentY += blockH + tableGapY
      idx++
    }

    // Обновляем состояние наведения
    if (hoveredTable !== newHoveredTable || hoveredGroup !== newHoveredGroup || hoveredSlot !== newHoveredSlot) {
      hoveredTable = newHoveredTable
      hoveredGroup = newHoveredGroup
      hoveredSlot = newHoveredSlot
    }
  }

  let rafId = null
  let needsRedraw = false

  function scheduleRedraw() {
    if (needsRedraw) return
    needsRedraw = true
    rafId = requestAnimationFrame(() => {
      draw()
      needsRedraw = false
    })
  }

  // Throttle для обновления позиции мыши
  let mouseUpdateTimeout = null
  function updateMousePosition(e) {
    const r = canvas.getBoundingClientRect()
    const sx = r.width / vb.w
    const sy = r.height / vb.h
    
    const mouseX = e.clientX - r.left
    const mouseY = e.clientY - r.top
    
    mouseCanvasX = vb.x + mouseX / sx
    mouseCanvasY = vb.y + mouseY / sy
    
    // Используем throttle только если не двигаем
    if (!isPanning) {
      if (mouseUpdateTimeout) return
      mouseUpdateTimeout = setTimeout(() => {
        mouseUpdateTimeout = null
        scheduleRedraw()
      }, 16) // ~60fps
    } else {
      scheduleRedraw()
    }
  }

  // Debounce для wheel событий
  let wheelTimeout = null
  let accumulatedDeltaX = 0
  let accumulatedDeltaY = 0
  let lastWheelEvent = null

  function handleWheel(e) {
    e.preventDefault()
    
    lastWheelEvent = e
    accumulatedDeltaX += e.deltaX
    accumulatedDeltaY += e.deltaY
    
    if (wheelTimeout) {
      clearTimeout(wheelTimeout)
    }
    
    wheelTimeout = setTimeout(() => {
      processWheel(lastWheelEvent, accumulatedDeltaX, accumulatedDeltaY)
      accumulatedDeltaX = 0
      accumulatedDeltaY = 0
      wheelTimeout = null
    }, 10)
  }

  function processWheel(e, deltaX, deltaY) {
    const r = canvas.getBoundingClientRect()

    if (e.ctrlKey) {
      const mx = e.clientX - r.left
      const my = e.clientY - r.top
      const z = Math.pow(1.001, -deltaY * 2)

      const nw = vb.w / z
      if (nw < 100 || nw > 40000) return

      vb.x += (mx / r.width) * (vb.w - nw)
      vb.y += (my / r.height) * (vb.h - vb.h / z)
      vb.w = nw
      vb.h /= z
    } else {
      vb.x += (deltaX * vb.w) / r.width
      vb.y += (deltaY * vb.h) / r.height
    }

    updateMousePosition(e)
  }

  function handleMouseDown(e) {
    isPanning = true
    lastMouseX = e.clientX
    lastMouseY = e.clientY
    canvas.style.cursor = 'grabbing'
  }

  function handleMouseUp() {
    isPanning = false
    canvas.style.cursor = 'grab'
  }

  function handleMouseMove(e) {
    if (isPanning) {
      const dx = e.clientX - lastMouseX
      const dy = e.clientY - lastMouseY
      lastMouseX = e.clientX
      lastMouseY = e.clientY

      const r = canvas.getBoundingClientRect()
      vb.x -= (dx * vb.w) / r.width
      vb.y -= (dy * vb.h) / r.height
    }
    
    updateMousePosition(e)
  }

  onMount(() => {
    ctx = canvas.getContext('2d')
    loadData()

    const sync = () => {
      const r = canvas.parentElement.getBoundingClientRect()
      vb.h = vb.w * (r.height / r.width)
      scheduleRedraw()
    }

    sync()
    window.addEventListener('resize', sync)

    return () => {
      if (rafId) cancelAnimationFrame(rafId)
      if (wheelTimeout) clearTimeout(wheelTimeout)
      if (mouseUpdateTimeout) clearTimeout(mouseUpdateTimeout)
    }
  })
</script>

<div class="root">
  <div
    class="canvas-container"
    on:wheel|nonpassive={handleWheel}
    on:mousedown={handleMouseDown}
    on:mousemove={handleMouseMove}
    on:mouseup={handleMouseUp}
    on:mouseleave={handleMouseUp}
  >
    <canvas bind:this={canvas}></canvas>
  </div>
</div>

<style>
  .root {
    width: 100%;
    height: 100vh;
    background: #ebfbee;
    overflow: hidden;
  }
  .canvas-container {
    width: 100%;
    height: 100%;
    cursor: grab;
  }
  canvas {
    width: 100%;
    height: 100%;
    display: block;
  }
</style>
