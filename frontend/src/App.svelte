<script>
  import { onMount } from 'svelte'
  import JSONTree from 'svelte-json-tree'
  
  let selectedBucket = null
  let selectedKey = null
  let newValueJson = ''
  let hovered = null
  
  function formatPreview(val, isEmpty = false) {
    if (isEmpty) return ''
    if (val === null || val === undefined) return ''
    if (typeof val === 'object') {
      const str = JSON.stringify(val)
      return str.length > 25 ? str.slice(0, 22) + '...' : str
    }
    const str = val.toString()
    return str.length > 25 ? str.slice(0, 22) + '...' : str
  }

  function withZeroValues(arr, tophash) {
    if (!Array.isArray(arr)) return arr
    if (!tophash) return arr
    return arr.map((v, i) => (tophash[i] < 5 ? null : v))
  }
  
  function syncVbAspect() {
    const rect = containerRect
    if (!rect || rect.width === 0) return
    vb.h = vb.w * (rect.height / rect.width)
  }

  let cameraInitialized = false
  
  function fitInitialBuckets(count = 4) {
    if (!svgBuckets.length) return
    const buckets = svgBuckets.slice(0, count)
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity
    for (const b of buckets) {
      minX = Math.min(minX, b.x)
      minY = Math.min(minY, b.y)
      maxX = Math.max(maxX, b.x + b.width)
      maxY = Math.max(maxY, b.y + b.height)
    }
    const padding = 80
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
  const fixedBucketWidth = 260
  let canvas
  let ctx
  let dpr = window.devicePixelRatio || 1
  let svgBuckets = []
  let svgArrows = []
  let svgLabels = []
  let visibleBuckets = []
  let svgWidth = 2000
  let svgHeight = 2000
  let vb = { x: 0, y: 0, w: 1200, h: 900 }
  let sideWidth = 280
  let lastSideWidth = 280
  let isSideVisible = true
  let lastContainerWidth = 0
  let containerEl = null
  let containerRect = { width: 0, height: 0 }
  let resizing = false
  let isPanning = false
  let rafId = null
  let hoverRafId = null
  let lastMouseX = 0
  let lastMouseY = 0
  
// Функция для построения chains на основе overflow
function buildChains(bucketArray) {
  const chains = []
  let currentChain = []
  
  console.log('Building chains from buckets:', bucketArray.length)
  
  for (let i = 0; i < bucketArray.length; i++) {
    const b = bucketArray[i]
    if (!b) continue
    
    currentChain.push(b)
    
    console.log(`Bucket ${i}: overflow = "${b.overflow}"`)
    
    // Если overflow не указывает на следующий бакет, цепочка заканчивается
    // overflow может быть: "nil", "0x0" (нулевой указатель), null, undefined
    const hasNextOverflow = b.overflow && 
                           b.overflow !== 'nil' && 
                           b.overflow !== 'null' &&
                           b.overflow !== '0x0'
    
    if (!hasNextOverflow) {
      console.log(`Chain ended at bucket ${i}, chain length: ${currentChain.length}`)
      chains.push(currentChain)
      currentChain = []
    }
  }
  
  // Если осталась незакрытая цепочка
  if (currentChain.length > 0) {
    console.log(`Remaining chain length: ${currentChain.length}`)
    chains.push(currentChain)
  }
  
  console.log(`Total chains built: ${chains.length}`)
  chains.forEach((chain, idx) => {
    console.log(`Chain ${idx}: ${chain.length} buckets`)
  })
  
  return chains
}
  
  const theme = {
    bucketFill: '#fff',
    bucketStrokeNew: '#000',
    bucketHoverStrokeNew: '#228be6',
    text: '#000',
    textMuted: '#495057',
    keyFill: '#b2f2bb',
    keyFillEmpty: '#dbfdc9',
    keyStroke: '#12b886',
    keyHoverFill: '#9feaa4',
    valueFill: '#ffec99',
    valueFillEmpty: '#fff2b8',
    valueStroke: '#ffa94d',
    valueHoverFill: '#ffe066',
    overflowFill: '#ddd',
    overflowStroke: '#000',
    arrowNew: '#000',
    arrowOld: '#ff6b6b',
    tophashStroke: '#000',
    tophashText: '#000',
    tophash: ['#F1F2F0', '#f5f29d', '#F5DF58', '#f5a662', '#F06559'],
    tophashGte5: '#BBF059',
  }
  function getTheme() { return theme }

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
          console.log('Loaded buckets:', buckets)
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
      
      chains = buildChains(buckets)
      oldChains = buildChains(oldBuckets)
      
      buildCanvasData()
      if (!cameraInitialized) {
        fitInitialBuckets(4)
        cameraInitialized = true
      }
      drawCanvas()
    } catch (e) {
      console.error('Ошибка загрузки данных:', e)
    }
  }
  
  function connectWS() {
    if (socket) socket.close()
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    socket = new WebSocket(`${proto}://${location.host}/ws`)
    socket.onopen = () => console.log('[ws] connected')
    socket.onmessage = () => load()
    socket.onclose = () => {
      console.log('[ws] disconnected, retrying...')
      setTimeout(connectWS, 1000)
    }
    socket.onerror = () => socket.close()
  }
  
  function buildCanvasData() {
    svgBuckets = []
    svgArrows = []
    svgLabels = []
    
    const hasOldChains = oldChains && oldChains.length > 0
    const hasNewChains = chains && chains.length > 0
    const showLabels = hasOldChains && hasNewChains
    
    console.log('Building canvas data:', { hasOldChains, hasNewChains, oldChainsCount: oldChains.length, newChainsCount: chains.length })
    
    let mainCountOld = 0
    if (hasOldChains) {
      for (const chain of oldChains) {
        if (chain.length > 0) {
          chain[0].displayBid = mainCountOld++
        }
      }
    }
    
    let mainCountNew = 0
    if (hasNewChains) {
      for (const chain of chains) {
        if (chain.length > 0) {
          chain[0].displayBid = mainCountNew++
        }
      }
    }
    
    let x = gapX
    let oldMaxY = 0
    let newMaxY = 0
    let oldStartY = gapY + (showLabels ? 30 : 0)
    let newStartY = gapY
    
    if (hasOldChains) {
      for (let chainIdx = 0; chainIdx < oldChains.length; chainIdx++) {
        const chain = oldChains[chainIdx]
        if (!chain || chain.length === 0) continue
        let y = oldStartY
        
        console.log(`Drawing old chain ${chainIdx} with ${chain.length} buckets at x=${x}`)
        
        for (let bucketIdx = 0; bucketIdx < chain.length; bucketIdx++) {
          const b = chain[bucketIdx]
          if (!b) continue
          const keys = b.keys && Array.isArray(b.keys) ? b.keys : []
          const values = b.values && Array.isArray(b.values) ? b.values : []
          const tophash = b.tophash || []
          const height = bucketHeaderHeight + tophashHeight + keys.length * rowHeight + values.length * rowHeight + rowHeight + padding * 2
          const keyPreviews = keys.map((k, i) => formatPreview(k, tophash[i] < 5))
          const valuePreviews = values.map((v, i) => formatPreview(v, tophash[i] < 5))
          const tophashParsed = tophash.map(tv => { const n = parseInt(tv); return isNaN(n) ? -1 : n })
          const tophashHex = tophashParsed.map(n => n < 0 ? '?' : n.toString(16))

          svgBuckets.push({ x, y, width: fixedBucketWidth, height, bucket: b, padding, isOld: true, chainIdx, bucketIdx, isMain: bucketIdx === 0, keyPreviews, valuePreviews, tophashParsed, tophashHex })

          if (bucketIdx < chain.length - 1) {
            svgArrows.push({ x: x + fixedBucketWidth / 2, y1: y + height, y2: y + height + gapY - arrowOffset, isOld: true })
            y += height + gapY
          } else {
            y += height + gapY
          }
          if (y > oldMaxY) oldMaxY = y
        }
        x += fixedBucketWidth + gapX
      }

      if (showLabels) svgLabels.push({ x: gapX, y: gapY + 20, text: 'OLD', isOld: true })
      newStartY = oldMaxY + gapY * 2
    }

    if (hasNewChains) {
      x = gapX
      for (let chainIdx = 0; chainIdx < chains.length; chainIdx++) {
        const chain = chains[chainIdx]
        if (!chain || chain.length === 0) continue
        let y = newStartY

        console.log(`Drawing new chain ${chainIdx} with ${chain.length} buckets at x=${x}`)

        for (let bucketIdx = 0; bucketIdx < chain.length; bucketIdx++) {
          const b = chain[bucketIdx]
          if (!b) continue
          const keys = b.keys && Array.isArray(b.keys) ? b.keys : []
          const values = b.values && Array.isArray(b.values) ? b.values : []
          const tophash = b.tophash || []
          const height = bucketHeaderHeight + tophashHeight + keys.length * rowHeight + values.length * rowHeight + rowHeight + padding * 2
          const keyPreviews = keys.map((k, i) => formatPreview(k, tophash[i] < 5))
          const valuePreviews = values.map((v, i) => formatPreview(v, tophash[i] < 5))
          const tophashParsed = tophash.map(tv => { const n = parseInt(tv); return isNaN(n) ? -1 : n })
          const tophashHex = tophashParsed.map(n => n < 0 ? '?' : n.toString(16))

          svgBuckets.push({ x, y, width: fixedBucketWidth, height, bucket: b, padding, isOld: false, chainIdx, bucketIdx, isMain: bucketIdx === 0, keyPreviews, valuePreviews, tophashParsed, tophashHex })
          
          if (bucketIdx < chain.length - 1) {
            svgArrows.push({ x: x + fixedBucketWidth / 2, y1: y + height, y2: y + height + gapY - arrowOffset, isOld: false })
            y += height + gapY
          } else {
            y += height + gapY
          }
          if (y > newMaxY) newMaxY = y
        }
        x += fixedBucketWidth + gapX
      }
      
      if (showLabels) svgLabels.push({ x: gapX, y: newStartY - gapY + 20, text: 'NEW', isOld: false })
    }
    
    svgWidth = x + 200
    svgHeight = Math.max(oldMaxY, newMaxY) + 200
  }
  
  function drawCanvas() {
    if (!ctx || !canvas || !containerEl) return
    const rect = containerRect
    if (rect.width === 0 || rect.height === 0) return

    vb.h = vb.w * (rect.height / rect.width)

    const targetW = Math.round(rect.width * dpr)
    const targetH = Math.round(rect.height * dpr)
    if (canvas.width !== targetW || canvas.height !== targetH) {
      canvas.width = targetW
      canvas.height = targetH
    }
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.scale(dpr, dpr)

    const scaleX = rect.width / vb.w
    const scaleY = rect.height / vb.h
    ctx.translate(-vb.x * scaleX, -vb.y * scaleY)
    ctx.scale(scaleX, scaleY)

    const vbR = vb.x + vb.w
    const vbB = vb.y + vb.h
    visibleBuckets = svgBuckets.filter((b) =>
      b.x + b.width > vb.x && b.x < vbR && b.y + b.height > vb.y && b.y < vbB
    )

    const visibleArrows = svgArrows.filter((a) =>
      a.x > vb.x && a.x < vbR && ((a.y1 > vb.y && a.y1 < vbB) || (a.y2 > vb.y && a.y2 < vbB))
    )

    const visibleLabels = svgLabels.filter((l) => l.x > vb.x && l.x < vbR && l.y > vb.y && l.y < vbB)

    ctx.font = 'bold 14px "JetBrains Mono", monospace'
    for (const label of visibleLabels) {
      ctx.fillStyle = label.isOld ? '#ff6b6b' : '#51cf66'
      ctx.fillText(label.text, label.x, label.y)
    }

    const t = theme
    for (const b of visibleBuckets) drawBucket(b, scaleX, scaleY, t)

    const invScale = 1 / Math.min(scaleX, scaleY)
    // Batch arrows by color
    if (visibleArrows.length > 0) {
      const lineW = 1.5 * invScale
      ctx.lineWidth = lineW
      // lines
      ctx.strokeStyle = t.arrowNew
      ctx.beginPath()
      for (const a of visibleArrows) { if (!a.isOld) { ctx.moveTo(a.x, a.y1); ctx.lineTo(a.x, a.y2) } }
      ctx.stroke()
      ctx.strokeStyle = t.arrowOld
      ctx.beginPath()
      for (const a of visibleArrows) { if (a.isOld) { ctx.moveTo(a.x, a.y1); ctx.lineTo(a.x, a.y2) } }
      ctx.stroke()
      // arrowheads
      ctx.fillStyle = t.arrowNew
      ctx.beginPath()
      for (const a of visibleArrows) { if (!a.isOld) { ctx.moveTo(a.x, a.y2); ctx.lineTo(a.x - 4, a.y2 - 8); ctx.lineTo(a.x + 4, a.y2 - 8); ctx.closePath() } }
      ctx.fill()
      ctx.fillStyle = t.arrowOld
      ctx.beginPath()
      for (const a of visibleArrows) { if (a.isOld) { ctx.moveTo(a.x, a.y2); ctx.lineTo(a.x - 4, a.y2 - 8); ctx.lineTo(a.x + 4, a.y2 - 8); ctx.closePath() } }
      ctx.fill()
    }

    ctx.resetTransform()
  }
  
  function drawBucket(b, scaleX, scaleY, th) {
    const invScale = 1 / Math.min(scaleX, scaleY)
    const bx = b.x, by = b.y, bp = b.padding
    const rowW = b.width - padding * 2
    const bxp = bx + bp

    const isBucketHovered = hovered !== null && hovered.chainIdx === b.chainIdx && hovered.bucketIdx === b.bucketIdx && hovered.isOld === b.isOld

    // Bucket outline
    ctx.fillStyle = th.bucketFill
    ctx.strokeStyle = isBucketHovered
      ? (b.isOld ? '#ff8787' : th.bucketHoverStrokeNew)
      : (b.isOld ? '#ff6b6b' : th.bucketStrokeNew)
    ctx.lineWidth = (b.isOld ? 2 : bucketStrokeWidth) * invScale
    ctx.beginPath()
    ctx.roundRect(bx, by, b.width, b.height, bucketRadius)
    ctx.fill()
    ctx.stroke()

    // Header label
    ctx.font = 'bold 11px "JetBrains Mono", monospace'
    ctx.fillStyle = th.textMuted
    ctx.textAlign = 'left'
    if (b.isMain && b.bucket.displayBid !== undefined) {
      ctx.fillText(`bid ${b.bucket.displayBid}`, bxp, by + bp + 12)
    } else if (!b.isMain) {
      ctx.fillText('overflow', bxp, by + bp + 12)
    }

    // Tophash
    const tophash = b.bucket?.tophash
    const tophashParsed = b.tophashParsed
    const n = tophash ? tophash.length : 0
    if (n > 0) {
      const cellW = (b.width - bp * 2) / n
      const thY = by + bp + bucketHeaderHeight
      const thTextY = thY + tophashHeight / 1.5

      // Batch fills by color group (0-4 = tophash[], 5 = gte5)
      const g0=[],g1=[],g2=[],g3=[],g4=[],g5=[]
      const gs=[g0,g1,g2,g3,g4,g5]
      for (let i = 0; i < n; i++) {
        const v = tophashParsed[i]
        gs[v >= 5 ? 5 : v < 0 ? 0 : v].push(i)
      }
      const colors = [th.tophash[0],th.tophash[1],th.tophash[2],th.tophash[3],th.tophash[4],th.tophashGte5]
      for (let g = 0; g < 6; g++) {
        const group = gs[g]
        if (group.length === 0) continue
        ctx.fillStyle = colors[g] ?? th.tophash[0]
        ctx.beginPath()
        for (let j = 0; j < group.length; j++) ctx.rect(bxp + group[j] * cellW, thY, cellW, tophashHeight)
        ctx.fill()
      }
      // Single stroke for all cells
      ctx.strokeStyle = th.tophashStroke
      ctx.lineWidth = invScale
      ctx.beginPath()
      for (let i = 0; i < n; i++) ctx.rect(bxp + i * cellW, thY, cellW, tophashHeight)
      ctx.stroke()
      // Text
      ctx.font = '12px "JetBrains Mono", monospace'
      ctx.fillStyle = th.tophashText
      ctx.textAlign = 'center'
      const thex = b.tophashHex
      for (let i = 0; i < n; i++) ctx.fillText(thex[i], bxp + i * cellW + cellW / 2, thTextY)
    }

    // Keys
    const tp = b.tophashParsed || []
    const keysLen = b.bucket?.keys ? b.bucket.keys.length : 0
    const keyBaseY = by + bp + bucketHeaderHeight + tophashHeight
    if (keysLen > 0) {
      let hovIdx = -1
      const norm = [], empty = []
      for (let i = 0; i < keysLen; i++) {
        if (isBucketHovered && hovered.type === 'key' && hovered.index === i) hovIdx = i
        else if (tp[i] < 5) empty.push(i)
        else norm.push(i)
      }
      if (norm.length > 0) { ctx.fillStyle = th.keyFill; ctx.beginPath(); for (const i of norm) ctx.rect(bxp, keyBaseY + i * rowHeight, rowW, rowHeight); ctx.fill() }
      if (empty.length > 0) { ctx.fillStyle = th.keyFillEmpty; ctx.beginPath(); for (const i of empty) ctx.rect(bxp, keyBaseY + i * rowHeight, rowW, rowHeight); ctx.fill() }
      if (hovIdx >= 0) { ctx.fillStyle = th.keyHoverFill; ctx.beginPath(); ctx.rect(bxp, keyBaseY + hovIdx * rowHeight, rowW, rowHeight); ctx.fill() }
      ctx.strokeStyle = th.keyStroke; ctx.lineWidth = invScale
      ctx.beginPath(); for (let i = 0; i < keysLen; i++) ctx.rect(bxp, keyBaseY + i * rowHeight, rowW, rowHeight); ctx.stroke()
      ctx.font = '13px "JetBrains Mono", monospace'; ctx.fillStyle = th.text; ctx.textAlign = 'left'
      const kp = b.keyPreviews
      for (let i = 0; i < keysLen; i++) ctx.fillText(kp[i] ?? '', bxp + 6, keyBaseY + i * rowHeight + rowHeight / 1.5)
    }

    // Values
    const valLen = b.bucket?.values ? b.bucket.values.length : 0
    const valBaseY = keyBaseY + keysLen * rowHeight
    if (valLen > 0) {
      let hovIdx = -1
      const norm = [], empty = []
      for (let i = 0; i < valLen; i++) {
        if (isBucketHovered && hovered.type === 'value' && hovered.index === i) hovIdx = i
        else if (tp[i] < 5) empty.push(i)
        else norm.push(i)
      }
      if (norm.length > 0) { ctx.fillStyle = th.valueFill; ctx.beginPath(); for (const i of norm) ctx.rect(bxp, valBaseY + i * rowHeight, rowW, rowHeight); ctx.fill() }
      if (empty.length > 0) { ctx.fillStyle = th.valueFillEmpty; ctx.beginPath(); for (const i of empty) ctx.rect(bxp, valBaseY + i * rowHeight, rowW, rowHeight); ctx.fill() }
      if (hovIdx >= 0) { ctx.fillStyle = th.valueHoverFill; ctx.beginPath(); ctx.rect(bxp, valBaseY + hovIdx * rowHeight, rowW, rowHeight); ctx.fill() }
      ctx.strokeStyle = th.valueStroke; ctx.lineWidth = invScale
      ctx.beginPath(); for (let i = 0; i < valLen; i++) ctx.rect(bxp, valBaseY + i * rowHeight, rowW, rowHeight); ctx.stroke()
      ctx.font = '13px "JetBrains Mono", monospace'; ctx.fillStyle = th.text; ctx.textAlign = 'left'
      const vp = b.valuePreviews
      for (let i = 0; i < valLen; i++) ctx.fillText(vp[i] ?? '', bxp + 6, valBaseY + i * rowHeight + rowHeight / 1.5)
    }

    // Overflow row
    if (b.bucket) {
      const ovY = valBaseY + valLen * rowHeight
      ctx.fillStyle = th.overflowFill; ctx.strokeStyle = th.overflowStroke; ctx.lineWidth = invScale
      ctx.beginPath(); ctx.rect(bxp, ovY, rowW, rowHeight); ctx.fill(); ctx.stroke()
      ctx.font = '12px "JetBrains Mono", monospace'; ctx.fillStyle = th.text; ctx.textAlign = 'left'
      ctx.fillText(b.bucket.overflow || '', bxp + 6, ovY + rowHeight / 1.5)
    }

    // Selected key highlight
    if (selectedKey && selectedKey.chainIdx === b.chainIdx && selectedKey.bucketIdx === b.bucketIdx && selectedKey.isOld === b.isOld) {
      ctx.strokeStyle = '#ff0000'; ctx.lineWidth = 3 * invScale
      ctx.strokeRect(bxp, keyBaseY + selectedKey.index * rowHeight, rowW, rowHeight)
    }
  }
  
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
      const rect = containerRect
      if (!rect || rect.width === 0) return
      let newVb = { ...vb }
      newVb.x -= (dx * vb.w) / rect.width
      newVb.y -= (dy * vb.h) / rect.height
      scheduleVbUpdate(newVb)
      return
    }
    handleHover(e)
  }
  
  function hoveredEqual(a, b) {
    if (a === b) return true
    if (!a || !b) return false
    return a.type === b.type && a.chainIdx === b.chainIdx && a.bucketIdx === b.bucketIdx && a.isOld === b.isOld && a.index === b.index
  }

  function handleHover(e) {
    if (!canvas) return
    if (hoverRafId) return
    hoverRafId = requestAnimationFrame(() => {
      hoverRafId = null
      const rect = containerRect
      if (!rect || rect.width === 0) return
      const hoverX = ((e.clientX - rect.left) / rect.width) * vb.w + vb.x
      const hoverY = ((e.clientY - rect.top) / rect.height) * vb.h + vb.y
      let newHovered = null

      for (const b of visibleBuckets) {
        if (hoverX >= b.x + b.padding && hoverX <= b.x + b.width - b.padding) {
          const keyYStart = b.y + b.padding + bucketHeaderHeight + tophashHeight
          const keyYEnd = keyYStart + (b.bucket.keys || []).length * rowHeight
          if (hoverY >= keyYStart && hoverY <= keyYEnd) {
            const localY = hoverY - keyYStart
            const index = Math.floor(localY / rowHeight)
            newHovered = { type: 'key', chainIdx: b.chainIdx, bucketIdx: b.bucketIdx, isOld: b.isOld, index }
            break
          }
          const valueYStart = keyYEnd
          const valueYEnd = valueYStart + (b.bucket.values || []).length * rowHeight
          if (hoverY >= valueYStart && hoverY <= valueYEnd) {
            const localY = hoverY - valueYStart
            const index = Math.floor(localY / rowHeight)
            newHovered = { type: 'value', chainIdx: b.chainIdx, bucketIdx: b.bucketIdx, isOld: b.isOld, index }
            break
          }
        }
        if (hoverX >= b.x && hoverX <= b.x + b.width && hoverY >= b.y && hoverY <= b.y + b.height) {
          newHovered = { type: 'bucket', chainIdx: b.chainIdx, bucketIdx: b.bucketIdx, isOld: b.isOld }
          break
        }
      }

      if (!hoveredEqual(newHovered, hovered)) {
        hovered = newHovered
        drawCanvas()
      }
    })
  }
  
  function handleSingleClick(e) {
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    const clickX = ((e.clientX - rect.left) / rect.width) * vb.w + vb.x
    const clickY = ((e.clientY - rect.top) / rect.height) * vb.h + vb.y
    
    for (const b of svgBuckets) {
      if (clickX >= b.x && clickX <= b.x + b.width && clickY >= b.y && clickY <= b.y + b.height) {
        selectedKey = null
        selectedBucket = b.bucket
        drawCanvas()
        return
      }
    }
  }
  
  function handleDblClick(e) {
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    const clickX = ((e.clientX - rect.left) / rect.width) * vb.w + vb.x
    const clickY = ((e.clientY - rect.top) / rect.height) * vb.h + vb.y
    
    for (const b of svgBuckets) {
      if (clickX >= b.x + b.padding && clickX <= b.x + b.width - b.padding) {
        const keyYStart = b.y + b.padding + bucketHeaderHeight + tophashHeight
        const keyYEnd = keyYStart + (b.bucket.keys || []).length * rowHeight
        if (clickY >= keyYStart && clickY <= keyYEnd) {
          const localY = clickY - keyYStart
          const index = Math.floor(localY / rowHeight)
          const tophash = b.bucket?.tophash || []
          if (index >= 0 && index < b.bucket.keys.length && tophash[index] >= 5) {
            selectKey(b.bucket, index, b.isOld, b.chainIdx, b.bucketIdx)
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
    if (!containerEl) return
    resizing = true
    const startMouseX = e.clientX
    const startSideWidth = sideWidth
    const unitsPerPixel = vb.w / containerRect.width

    const onMouseMove = (ev) => {
      if (!resizing) return
      const dx = startMouseX - ev.clientX
      const newSideWidth = Math.max(100, Math.min(800, startSideWidth + dx))
      const diffPx = newSideWidth - sideWidth
      sideWidth = newSideWidth
      containerRect = { left: containerRect.left, top: containerRect.top, width: containerRect.width - diffPx, height: containerRect.height }
      let newVb = { ...vb }
      newVb.w -= diffPx * unitsPerPixel
      vb.w -= diffPx * unitsPerPixel
      syncVbAspect()
      scheduleVbUpdate(newVb)
    }

    const onMouseUp = () => {
      resizing = false
      containerRect = containerEl.getBoundingClientRect()
      window.removeEventListener('mousemove', onMouseMove)
      window.removeEventListener('mouseup', onMouseUp)
    }
    
    window.addEventListener('mousemove', onMouseMove)
    window.addEventListener('mouseup', onMouseUp)
  }
  
  function toggleSide() {
    if (!containerEl) return
    const unitsPerPixel = vb.w / containerRect.width
    
    if (isSideVisible) {
      lastSideWidth = sideWidth
      sideWidth = 0
      isSideVisible = false
    } else {
      sideWidth = lastSideWidth
      isSideVisible = true
    }
    
    requestAnimationFrame(() => {
      containerRect = containerEl.getBoundingClientRect()
      let newVb = { ...vb }
      newVb.w = unitsPerPixel * containerRect.width
      newVb.h = newVb.w * (containerRect.height / containerRect.width)
      vb = newVb
      drawCanvas()
    })
  }
  
  function selectKey(bucket, index, isOld, chainIdx, bucketIdx) {
    const tophash = bucket?.tophash || []
    if (tophash[index] < 5) return
    selectedBucket = bucket
    selectedKey = {
      bucket,
      index,
      key: bucket.keys[index],
      value: bucket.values[index],
      isOld,
      chainIdx,
      bucketIdx
    }
    newValueJson = JSON.stringify(selectedKey.value, null, 2)
  }
  
  function handleKeyDown(e) {
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
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
      }).then((res) => {
        if (res.ok) {
          load()
          selectedKey = null
        } else {
          alert('Error deleting key')
        }
      }).catch((err) => {
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
      }).then((res) => {
        if (res.ok) {
          load()
          selectedKey = null
        } else {
          alert('Error updating key')
        }
      }).catch((err) => {
        console.error(err)
        alert('Error updating key')
      })
    }
  }
  
  onMount(() => {
    ctx = canvas?.getContext('2d')
    containerEl = document.getElementById('canvas-container')

    const handleResize = () => {
      if (!containerEl) return
      containerRect = containerEl.getBoundingClientRect()
      if (lastContainerWidth > 0 && containerRect.width > 0) {
        vb = { ...vb, w: vb.w * (containerRect.width / lastContainerWidth) }
      }
      lastContainerWidth = containerRect.width
      syncVbAspect()
      drawCanvas()
    }

    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener('resize', handleResize)

    requestAnimationFrame(() => {
      if (containerEl) {
        containerRect = containerEl.getBoundingClientRect()
        lastContainerWidth = containerRect.width
      }
      load()
    })
    connectWS()
    
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      window.removeEventListener('resize', handleResize)
      if (socket) socket.close()
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
    <canvas bind:this={canvas} style="width: 100%; height: 100%;"></canvas>
  </div>
  {#if isSideVisible}
    <div class="splitter" on:mousedown={startResize}></div>
  {/if}
  <div class="side" style="width: {sideWidth}px; display: {isSideVisible ? 'flex' : 'none'}; flex-direction: column;">
    <div class="side-content">
      <h3>hmap</h3>
      {#if hmap}
        {#each Object.entries(hmap) as [k, v]}
          <div class="row"><span>{k}</span><b>{v}</b></div>
        {/each}
      {/if}
      {#if stats}
        <h3>stats</h3>
        <div class="row"><span>Load factor</span><b>{stats.loadFactor.toFixed(2)}</b></div>
        <div class="row"><span>Max chain len</span><b>{stats.maxChainLen}</b></div>
        <div class="row"><span>Chains count</span><b>{stats.numChains}</b></div>
        <div class="row"><span>Empty buckets</span><b>{stats.numEmptyBuckets}</b></div>
        <div class="row"><span>Key type</span><b>{stats.keytype}</b></div>
        <div class="row"><span>Value type</span><b>{stats.valuetype}</b></div>
      {/if}
      <div class="inspector-section">
        <h3>inspector</h3>
        {#if selectedKey}
          <div style="margin-bottom: 10px;">
            <small>Key at index {selectedKey.index}</small>
          </div>
          <div class="tree-label">Selected Key:</div>
          <div class="tree-container"><JSONTree value={selectedKey.key} shouldShowPreview={false}/></div>
          <div class="tree-label" style="margin-top: 10px;">Current Value:</div>
          <div class="tree-container"><JSONTree value={selectedKey.value} /></div>
          <div class="tree-label" style="margin-top: 10px;">New Value (JSON or raw string):</div>
          <textarea bind:value={newValueJson} class="side-textarea"></textarea>
          <p class="hint-text">
            Enter JSON for objects or raw string/number. Press 'd' to delete, 'u' to update, 'Esc' to deselect.
          </p>
        {:else if selectedBucket}
          <div class="tree-label">Keys:</div>
          <div class="tree-container"><JSONTree value={withZeroValues(selectedBucket.keys, selectedBucket.tophash)} /></div>
          <div class="tree-label" style="margin-top: 10px;">Values:</div>
          <div class="tree-container"><JSONTree value={withZeroValues(selectedBucket.values, selectedBucket.tophash)} /></div>
          <div class="row" style="margin-top: 10px; font-size: 11px;"><span>Overflow</span><code>{selectedBucket.overflow}</code></div>
        {:else}
          <p class="empty-hint">Click on the bucket to analyze</p>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  :global(*) { box-sizing: border-box; }
  :global(html, body) { margin: 0; padding: 0; width: 100%; height: 100%; overflow: hidden; }
  .root { display: flex; width: 100%; height: 100dvh; overflow: hidden; background: #ebfbee; }
  #canvas-container { flex: 1; position: relative; overflow: hidden; cursor: grab; touch-action: none; }
  .root::before {
    content: '';
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: url('/public/1.png') center/cover fixed;
    filter: blur(50px);
    z-index: -1;
  }
  #canvas-container:active { cursor: grabbing; }
  #canvas-container::before {
    content: '';
    position: absolute;
    top: 0; left: 0; right: 0; bottom: 0;
    background: url('/public/mikuiru.gif') center/cover fixed;
    filter: blur(5px);
    opacity: 0.9;
    z-index: 0;
  }
  canvas { position: relative; z-index: 1; }

  .toggle-btn { position: absolute; right: 15px; top: 15px; z-index: 100; background: #ffffff; border: 1px solid #ccc; padding: 6px 12px; border-radius: 6px; cursor: pointer; font-family: sans-serif; font-weight: 500; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); transition: background 0.2s, color 0.2s, border-color 0.2s; }
  .splitter { width: 6px; background: #ccc; cursor: col-resize; user-select: none; z-index: 10; transition: background 0.2s; }

  /* Sidebar */
  .side { border-left: 1px solid #ccc; background: #fafafa; font-family: monospace; font-size: 13px; flex-shrink: 0; transition: background 0.2s, color 0.2s, border-color 0.2s; }
  .side-content { padding: 12px; flex: 1; overflow-y: auto; }
  .row { display: flex; justify-content: space-between; padding: 2px 0; border-bottom: 1px solid #eee; }
  .tree-label { font-size: 11px; font-weight: bold; color: #666; text-transform: uppercase; margin-bottom: 4px; }
  .tree-container { background: #fff; border: 1px solid #eee; border-radius: 4px; padding: 4px; max-height: 250px; overflow: auto; }
  code { background: #eee; padding: 1px 4px; border-radius: 3px; }
  .inspector-section { margin-top: 20px; border-top: 2px solid #ccc; padding-top: 10px; }
  .side-textarea { width: 100%; height: 100px; font-family: monospace; font-size: 12px; background: #fff; color: #000; border: 1px solid #ccc; border-radius: 4px; padding: 4px; resize: vertical; }
  .hint-text { font-size: 11px; color: #666; margin-top: 5px; }
  .empty-hint { color: #999; font-size: 12px; font-style: italic; text-align: center; }

</style>

