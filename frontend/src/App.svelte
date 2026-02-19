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
    const container = document.getElementById('canvas-container')
    if (!container) return
    const rect = container.getBoundingClientRect()
    const aspect = rect.height / rect.width
    vb.h = vb.w * aspect
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
  let svgWidth = 2000
  let svgHeight = 2000
  let vb = { x: 0, y: 0, w: 1200, h: 900 }
  let sideWidth = 280
  let lastSideWidth = 280
  let isSideVisible = true
  let resizing = false
  let isPanning = false
  let rafId = null
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
          const height = bucketHeaderHeight + tophashHeight + keys.length * rowHeight + values.length * rowHeight + rowHeight + padding * 2
          
          svgBuckets.push({ x, y, width: fixedBucketWidth, height, bucket: b, padding, isOld: true, chainIdx, bucketIdx, isMain: bucketIdx === 0 })
          
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
          const height = bucketHeaderHeight + tophashHeight + keys.length * rowHeight + values.length * rowHeight + rowHeight + padding * 2
          
          svgBuckets.push({ x, y, width: fixedBucketWidth, height, bucket: b, padding, isOld: false, chainIdx, bucketIdx, isMain: bucketIdx === 0 })
          
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
    if (!ctx || !canvas) return
    const container = document.getElementById('canvas-container')
    if (!container) return
    const rect = container.getBoundingClientRect()

    canvas.width = rect.width * dpr
    canvas.height = rect.height * dpr
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.scale(dpr, dpr)
    
    const scaleX = rect.width / vb.w
    const scaleY = rect.height / vb.h
    const offsetX = -vb.x * scaleX
    const offsetY = -vb.y * scaleY
    ctx.translate(offsetX, offsetY)
    ctx.scale(scaleX, scaleY)
    
    const visibleBuckets = svgBuckets.filter((b) => {
      return b.x + b.width > vb.x && b.x < vb.x + vb.w && b.y + b.height > vb.y && b.y < vb.y + vb.h
    })
    
    const visibleArrows = svgArrows.filter((a) => {
      return a.x > vb.x && a.x < vb.x + vb.w && ((a.y1 > vb.y && a.y1 < vb.y + vb.h) || (a.y2 > vb.y && a.y2 < vb.y + vb.h))
    })
    
    const visibleLabels = svgLabels.filter((l) => l.x > vb.x && l.x < vb.x + vb.w && l.y > vb.y && l.y < vb.y + vb.h)
    
    ctx.font = 'bold 14px "JetBrains Mono", monospace'
    visibleLabels.forEach((label) => {
      ctx.fillStyle = label.isOld ? '#ff6b6b' : '#51cf66'
      ctx.fillText(label.text, label.x, label.y)
    })
    
    visibleBuckets.forEach((b) => drawBucket(b, scaleX, scaleY))
    
    visibleArrows.forEach((a) => {
      ctx.strokeStyle = a.isOld ? '#ff6b6b' : '#000'
      ctx.lineWidth = 1.5 / Math.min(scaleX, scaleY)
      ctx.beginPath()
      ctx.moveTo(a.x, a.y1)
      ctx.lineTo(a.x, a.y2)
      ctx.stroke()
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

    if (hovered && hovered.chainIdx === b.chainIdx && hovered.bucketIdx === b.bucketIdx && hovered.isOld === b.isOld) {
      bucketStroke = b.isOld ? '#ff8787' : '#228be6'
    }
    
    ctx.fillStyle = '#fff'
    ctx.strokeStyle = bucketStroke
    ctx.lineWidth = strokeW / Math.min(scaleX, scaleY)
    ctx.beginPath()
    ctx.roundRect(b.x, b.y, b.width, b.height, bucketRadius)
    ctx.fill()
    ctx.stroke()
    
    // Отображаем displayBid только для main бакетов
    if (b.isMain && b.bucket.displayBid !== undefined) {
      ctx.font = 'bold 11px "JetBrains Mono", monospace'
      ctx.fillStyle = '#495057'
      ctx.textAlign = 'left'
      ctx.fillText(`bid ${b.bucket.displayBid}`, b.x + b.padding, b.y + b.padding + 12)
    }
    if (!b.isMain) {
      ctx.font = 'bold 11px "JetBrains Mono", monospace'
      ctx.fillStyle = '#495057'
      ctx.textAlign = 'left'
      ctx.fillText(`overflow`, b.x + b.padding, b.y + b.padding + 12)
    }
    
    const tophash = b.bucket?.tophash || []
    const tophashWidth = b.width - b.padding * 2
    const cellWidth = tophash.length > 0 ? tophashWidth / tophash.length : fixedTophashCellWidth
    
    tophash.forEach((t, i) => {
      let tophashColor = '#eee'
const tVal = parseInt(t)
if (!isNaN(tVal)) {
  if (tVal === 0) tophashColor = '#F1F2F0'
  else if (tVal === 1) tophashColor = '#f5f29d'
  else if (tVal === 2) tophashColor = '#F5DF58'
  else if (tVal === 3) tophashColor = '#f5a662'
  else if (tVal === 4) tophashColor = '#F06559'
  else if (tVal >= 5) tophashColor = '#BBF059'
}
ctx.fillStyle = tophashColor
      ctx.strokeStyle = '#000'
      ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
      ctx.fillRect(b.x + b.padding + i * cellWidth, b.y + b.padding + bucketHeaderHeight, cellWidth, tophashHeight)
      ctx.strokeRect(b.x + b.padding + i * cellWidth, b.y + b.padding + bucketHeaderHeight, cellWidth, tophashHeight)
ctx.font = '12px "JetBrains Mono", monospace'
      ctx.fillStyle = '#000'
      ctx.textAlign = 'center'
      ctx.fillText(t, b.x + b.padding + i * cellWidth + cellWidth / 2, b.y + b.padding + bucketHeaderHeight + tophashHeight / 1.5)
    })
    
    const keys = b.bucket?.keys || []
    const bucketTophash = b.bucket?.tophash || []
    keys.forEach((k, i) => {
      const isEmpty = bucketTophash[i] < 5
      let fill = isEmpty ? '#dbfdc9' : '#b2f2bb'
      if (hovered && hovered.type === 'key' && hovered.chainIdx === b.chainIdx && hovered.bucketIdx === b.bucketIdx && hovered.isOld === b.isOld && hovered.index === i) {
        fill = '#9feaa4'
      }
      ctx.fillStyle = fill
      ctx.strokeStyle = '#12b886'
      ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
      ctx.fillRect(b.x + b.padding, b.y + b.padding + bucketHeaderHeight + tophashHeight + i * rowHeight, b.width - padding * 2, rowHeight)
      ctx.strokeRect(b.x + b.padding, b.y + b.padding + bucketHeaderHeight + tophashHeight + i * rowHeight, b.width - padding * 2, rowHeight)
      ctx.font = '13px "JetBrains Mono", monospace'
      ctx.fillStyle = '#000'
      ctx.textAlign = 'left'
      ctx.fillText(formatPreview(k, isEmpty), b.x + b.padding + 6, b.y + b.padding + bucketHeaderHeight + tophashHeight + i * rowHeight + rowHeight / 1.5)
    })
    
    const values = b.bucket?.values || []
    const keysLen = keys.length
    values.forEach((v, i) => {
      const isEmpty = bucketTophash[i] < 5
      let fill = isEmpty ? '#fff2b8' : '#ffec99'
      if (hovered && hovered.type === 'value' && hovered.chainIdx === b.chainIdx && hovered.bucketIdx === b.bucketIdx && hovered.isOld === b.isOld && hovered.index === i) {
        fill = '#ffe066'
      }
      ctx.fillStyle = fill
      ctx.strokeStyle = '#ffa94d'
      ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
      ctx.fillRect(b.x + b.padding, b.y + b.padding + bucketHeaderHeight + tophashHeight + keysLen * rowHeight + i * rowHeight, b.width - padding * 2, rowHeight)
      ctx.strokeRect(b.x + b.padding, b.y + b.padding + bucketHeaderHeight + tophashHeight + keysLen * rowHeight + i * rowHeight, b.width - padding * 2, rowHeight)
      ctx.font = '13px "JetBrains Mono", monospace'
      ctx.fillStyle = '#000'
      ctx.fillText(formatPreview(v, isEmpty), b.x + b.padding + 6, b.y + b.padding + bucketHeaderHeight + tophashHeight + keysLen * rowHeight + i * rowHeight + rowHeight / 1.5)
    })
    
    if (b.bucket) {
      const valuesLen = values.length
      ctx.fillStyle = '#ddd'
      ctx.strokeStyle = '#000'
      ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
      ctx.fillRect(b.x + b.padding, b.y + b.padding + bucketHeaderHeight + tophashHeight + keysLen * rowHeight + valuesLen * rowHeight, b.width - padding * 2, rowHeight)
      ctx.strokeRect(b.x + b.padding, b.y + b.padding + bucketHeaderHeight + tophashHeight + keysLen * rowHeight + valuesLen * rowHeight, b.width - padding * 2, rowHeight)
      ctx.font = '12px "JetBrains Mono", monospace'
      ctx.fillStyle = '#000'
      ctx.fillText(b.bucket.overflow || '', b.x + b.padding + 6, b.y + b.padding + bucketHeaderHeight + tophashHeight + keysLen * rowHeight + valuesLen * rowHeight + rowHeight / 1.5)
    }
    
    if (selectedKey && selectedKey.chainIdx === b.chainIdx && selectedKey.bucketIdx === b.bucketIdx && selectedKey.isOld === b.isOld) {
      ctx.strokeStyle = '#ff0000'
      ctx.lineWidth = 3 / Math.min(scaleX, scaleY)
      ctx.strokeRect(b.x + b.padding, b.y + b.padding + bucketHeaderHeight + tophashHeight + selectedKey.index * rowHeight, b.width - padding * 2, rowHeight)
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
      const rect = document.getElementById('canvas-container')?.getBoundingClientRect()
      if (!rect) return
      let newVb = { ...vb }
      newVb.x -= (dx * vb.w) / rect.width
      newVb.y -= (dy * vb.h) / rect.height
      scheduleVbUpdate(newVb)
    }
    handleHover(e)
  }
  
  function handleHover(e) {
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    const hoverX = ((e.clientX - rect.left) / rect.width) * vb.w + vb.x
    const hoverY = ((e.clientY - rect.top) / rect.height) * vb.h + vb.y
    let newHovered = null
    
    for (const b of svgBuckets) {
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
    
    if (JSON.stringify(newHovered) !== JSON.stringify(hovered)) {
      hovered = newHovered
      drawCanvas()
    }
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
    resizing = true
    const startMouseX = e.clientX
    const startSideWidth = sideWidth
    const container = document.getElementById('canvas-container')
    if (!container) return
    const initialRect = container.getBoundingClientRect()
    const unitsPerPixel = vb.w / initialRect.width
    
    const onMouseMove = (ev) => {
      if (!resizing) return
      const dx = startMouseX - ev.clientX
      const newSideWidth = Math.max(100, Math.min(800, startSideWidth + dx))
      const diffPx = newSideWidth - sideWidth
      sideWidth = newSideWidth
      let newVb = { ...vb }
      newVb.w -= diffPx * unitsPerPixel
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
    if (!container) return
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
    
    const handleResize = () => {
      syncVbAspect()
      drawCanvas()
    }
    
    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener('resize', handleResize)
    
    load()
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
  <div class="side" style="width: {sideWidth}px; display: {isSideVisible ? 'block' : 'none'};">
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
        <div style="margin-bottom: 10px; color: #555;">
          <small>Key at index {selectedKey.index}</small>
        </div>
        <div class="tree-label">Selected Key:</div>
        <div class="tree-container"><JSONTree value={selectedKey.key} shouldShowPreview={false}/></div>
        <div class="tree-label" style="margin-top: 10px;">Current Value:</div>
        <div class="tree-container"><JSONTree value={selectedKey.value} /></div>
        <div class="tree-label" style="margin-top: 10px;">New Value (JSON or raw string):</div>
        <textarea bind:value={newValueJson} style="width: 100%; height: 100px; font-family: monospace; font-size: 12px;"></textarea>
        <p style="font-size: 11px; color: #666; margin-top: 5px;">
          Enter JSON for objects or raw string/number. Press 'd' to delete, 'u' to update, 'Esc' to deselect.
        </p>
      {:else if selectedBucket}
        <div class="tree-label">Keys:</div>
        <div class="tree-container"><JSONTree value={withZeroValues(selectedBucket.keys, selectedBucket.tophash)} /></div>
        <div class="tree-label" style="margin-top: 10px;">Values:</div>
        <div class="tree-container"><JSONTree value={withZeroValues(selectedBucket.values, selectedBucket.tophash)} /></div>
        <div class="row" style="margin-top: 10px; font-size: 11px;"><span>Overflow</span><code>{selectedBucket.overflow}</code></div>
      {:else}
        <p style="color: #999; font-size: 12px; font-style: italic; text-align: center;">Click on the bucket to analyze</p>
      {/if}
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
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: url('/public/1.png') center/cover fixed;
  filter: blur(50px); /* меняй значение для большего/меньшего размытия */
  z-index: -1;
}
  #canvas-container:active { cursor: grabbing;}
  #canvas-container { 
  flex: 1; 
  position: relative; 
  overflow: hidden; 
  cursor: grab; 
  touch-action: none; 
}

#canvas-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: url('/public/mikuiru.gif') center/cover fixed;
  filter: blur(5px); /* увеличь значение для большего размытия */
  opacity: 0.9; /* добавь прозрачность */
  z-index: 0;
}

canvas {
  position: relative;
  z-index: 1; /* чтобы canvas был поверх фона */
}
  .toggle-btn { position: absolute; right: 15px; top: 15px; z-index: 100; background: #ffffff; border: 1px solid #ccc; padding: 6px 12px; border-radius: 6px; cursor: pointer; font-family: sans-serif; font-weight: 500; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); }
  .splitter { width: 6px; background: #ccc; cursor: col-resize; user-select: none; z-index: 10; }
  .side { padding: 12px; border-left: 1px solid #ccc; background: #fafafa; overflow-y: auto; font-family: monospace; font-size: 13px; flex-shrink: 0; }
  .row { display: flex; justify-content: space-between; padding: 2px 0; border-bottom: 1px solid #eee; }
  .tree-label { font-size: 11px; font-weight: bold; color: #666; text-transform: uppercase; margin-bottom: 4px; }
  .tree-container { background: #fff; border: 1px solid #eee; border-radius: 4px; padding: 4px; max-height: 250px; overflow: auto; }
  code { background: #eee; padding: 1px 4px; border-radius: 3px; }
  .inspector-section { margin-top: 20px; border-top: 2px solid #ccc; padding-top: 10px; }
</style>

