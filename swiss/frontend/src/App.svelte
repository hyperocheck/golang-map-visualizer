<script>
  import { onMount } from 'svelte'

  let canvas
  let ctx
  let dpr = window.devicePixelRatio || 1

  let data = null

  // Состояние viewBox для навигации
  let vb = { x: 0, y: 0, w: 2000, h: 1500 }
  
  // Состояние панорамирования
  let isPanning = false
  let lastMouseX = 0
  let lastMouseY = 0

  // Константы для отрисовки
  const ctrlCellSize = 27
  const slotHeight = 22
  const padding = 12
  const groupGapX = 40
  const tableGapY = 60
  const groupRadius = 8
  const addrBlockWidth = 80 // ширина блока адреса
  const addrBlockGap = 100 // отступ между адресом и таблицей

  function base64ToBytes(base64) {
    const binary = atob(base64)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    return Array.from(bytes)
  }

  function isEmptySlot(ctrlByte) {
    return ctrlByte === 0x80 || ctrlByte === 128
  }

  // Генерация случайного пастельного цвета для таблицы
  function getRandomTableColor(index) {
    const colors = [
      { bg: 'rgba(255, 200, 200, 0.15)', border: 'rgba(255, 100, 100, 0.6)' }, // красный
      { bg: 'rgba(200, 220, 255, 0.15)', border: 'rgba(100, 150, 255, 0.6)' }, // синий
      { bg: 'rgba(200, 255, 200, 0.15)', border: 'rgba(100, 200, 100, 0.6)' }, // зеленый
      { bg: 'rgba(255, 220, 200, 0.15)', border: 'rgba(255, 150, 100, 0.6)' }, // оранжевый
      { bg: 'rgba(230, 200, 255, 0.15)', border: 'rgba(180, 100, 255, 0.6)' }, // фиолетовый
      { bg: 'rgba(255, 255, 200, 0.15)', border: 'rgba(200, 200, 100, 0.6)' }, // желтый
      { bg: 'rgba(200, 255, 255, 0.15)', border: 'rgba(100, 200, 200, 0.6)' }, // циан
    ]
    return colors[index % colors.length]
  }

  async function loadData() {
    try {
      const response = await fetch('/data')
      if (!response.ok) {
        console.error('Failed to load data:', response.statusText)
        return
      }
      data = await response.json()
      drawVisualization()
    } catch (err) {
      console.error('Error loading data:', err)
    }
  }

  function drawVisualization() {
    if (!ctx || !data || !data.tables) return

    const container = canvas.parentElement
    const rect = container.getBoundingClientRect()

    canvas.width = rect.width * dpr
    canvas.height = rect.height * dpr
    
    // Сбрасываем трансформацию
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.scale(dpr, dpr)

    // Применяем трансформацию viewBox
    const scaleX = rect.width / vb.w
    const scaleY = rect.height / vb.h
    const offsetX = -vb.x * scaleX
    const offsetY = -vb.y * scaleY
    
    ctx.translate(offsetX, offsetY)
    ctx.scale(scaleX, scaleY)

    let currentY = 40

    // Группируем таблицы по адресам
    const tablesByAddr = new Map()
    for (let tableIdx = 0; tableIdx < data.tables.length; tableIdx++) {
      const table = data.tables[tableIdx]
      const addr = table.addr
      
      if (!tablesByAddr.has(addr)) {
        tablesByAddr.set(addr, { table, indices: [tableIdx] })
      } else {
        tablesByAddr.get(addr).indices.push(tableIdx)
      }
    }

    // Рисуем уникальные таблицы
    let uniqueTableIdx = 0
    for (const [addr, { table, indices }] of tablesByAddr) {
      const tableColor = getRandomTableColor(uniqueTableIdx)
      
      let currentX = 40 + addrBlockWidth + addrBlockGap
      const tableStartX = currentX - 20
      const tableStartY = currentY - 20
      
      // Вычисляем размеры таблицы
      let tableWidth = 0
      let tableHeight = 0
      
      for (const group of table.groups) {
        const ctrls = base64ToBytes(group.ctrls)
        const groupWidth = ctrls.length * ctrlCellSize + padding * 2
        const groupHeight = padding + ctrlCellSize + 16 * slotHeight + padding
        
        tableWidth += groupWidth + groupGapX
        tableHeight = Math.max(tableHeight, groupHeight)
      }
      
      tableWidth = tableWidth - groupGapX + 40
      tableHeight = tableHeight + 40

      // Рисуем рамку таблицы
      ctx.fillStyle = tableColor.bg
      ctx.strokeStyle = tableColor.border
      ctx.lineWidth = 3 / Math.min(scaleX, scaleY)
      ctx.beginPath()
      ctx.roundRect(tableStartX, tableStartY, tableWidth, tableHeight, 16)
      ctx.fill()
      ctx.stroke()

      // Рисуем адресные блоки для всех индексов, указывающих на эту таблицу
      for (let i = 0; i < indices.length; i++) {
        const addrX = 40
        const addrY = tableStartY + (i * (tableHeight / indices.length))
        const addrHeight = indices.length > 1 ? tableHeight / indices.length - 10 : tableHeight
        
        // Блок адреса
        ctx.fillStyle = tableColor.bg
        ctx.strokeStyle = tableColor.border
        ctx.lineWidth = 2 / Math.min(scaleX, scaleY)
        ctx.beginPath()
        ctx.roundRect(addrX, addrY, addrBlockWidth, addrHeight, 8)
        ctx.fill()
        ctx.stroke()
        
        // Текст адреса (вертикально)
        ctx.save()
        ctx.translate(addrX + addrBlockWidth / 2, addrY + addrHeight / 2)
        ctx.rotate(-Math.PI / 2)
        ctx.font = 'bold 12px monospace'
        ctx.fillStyle = tableColor.border
        ctx.textAlign = 'center'
        ctx.textBaseline = 'middle'
        ctx.fillText(`0x${addr.toString(16)}`, 0, 0)
        ctx.restore()
        
        // Стрелка от адреса к таблице
        const arrowStartX = addrX + addrBlockWidth
        const arrowStartY = addrY + addrHeight / 2
        const arrowEndX = tableStartX - 5
        const arrowEndY = arrowStartY
        
        ctx.strokeStyle = tableColor.border
        ctx.lineWidth = 2 / Math.min(scaleX, scaleY)
        ctx.beginPath()
        ctx.moveTo(arrowStartX, arrowStartY)
        ctx.lineTo(arrowEndX, arrowEndY)
        ctx.stroke()
        
        // Наконечник стрелки
        ctx.fillStyle = tableColor.border
        ctx.beginPath()
        ctx.moveTo(arrowEndX, arrowEndY)
        ctx.lineTo(arrowEndX - 8, arrowEndY - 5)
        ctx.lineTo(arrowEndX - 8, arrowEndY + 5)
        ctx.closePath()
        ctx.fill()
      }

      // Рисуем группы
      currentX = 40 + addrBlockWidth + addrBlockGap

      for (const group of table.groups) {
        const ctrls = base64ToBytes(group.ctrls)
        
        // Ширина группы = ширина контрольных байтов + отступы
        const groupWidth = ctrls.length * ctrlCellSize + padding * 2
        
        const groupHeight = padding + ctrlCellSize + 16 * slotHeight + padding

        // Рисуем прямоугольник группы
        ctx.fillStyle = 'rgba(255, 255, 255, 0.8)'
        ctx.strokeStyle = '#000'
        ctx.lineWidth = 1.5 / Math.min(scaleX, scaleY)
        ctx.beginPath()
        ctx.roundRect(currentX, currentY, groupWidth, groupHeight, groupRadius)
        ctx.fill()
        ctx.stroke()

        // Рисуем контрольные байты
        let ctrlX = currentX + padding
        let ctrlY = currentY + padding

        for (let i = 0; i < ctrls.length; i++) {
          const ctrl = ctrls[i]
          const isEmpty = isEmptySlot(ctrl)

          ctx.fillStyle = isEmpty ? '#ddd' : '#a8dadc'
          ctx.strokeStyle = '#000'
          ctx.lineWidth = 1 / Math.min(scaleX, scaleY)

          ctx.fillRect(ctrlX, ctrlY, ctrlCellSize, ctrlCellSize)
          ctx.strokeRect(ctrlX, ctrlY, ctrlCellSize, ctrlCellSize)

          ctx.font = '11px monospace'
          ctx.fillStyle = '#000'
          ctx.textAlign = 'center'
          ctx.textBaseline = 'middle'
          ctx.fillText(ctrl.toString(), ctrlX + ctrlCellSize / 2, ctrlY + ctrlCellSize / 2)

          ctrlX += ctrlCellSize
        }

        // Рисуем слоты (k, v пары)
        let slotY = currentY + padding + ctrlCellSize

        for (let i = 0; i < group.slots.length; i++) {
          const slot = group.slots[i]
          const ctrl = ctrls[i]
          const isEmpty = isEmptySlot(ctrl)

          // Key
          ctx.fillStyle = isEmpty ? '#dbfdc9' : '#b2f2bb'
          ctx.strokeStyle = '#12b886'
          ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
          ctx.fillRect(currentX + padding, slotY, groupWidth - padding * 2, slotHeight)
          ctx.strokeRect(currentX + padding, slotY, groupWidth - padding * 2, slotHeight)

          ctx.font = '12px monospace'
          ctx.fillStyle = '#000'
          ctx.textAlign = 'left'
          ctx.textBaseline = 'middle'
          ctx.fillText(`${slot.k}`, currentX + padding + 6, slotY + slotHeight / 2)

          slotY += slotHeight

          // Value
          ctx.fillStyle = isEmpty ? '#fff2b8' : '#ffec99'
          ctx.strokeStyle = '#ffa94d'
          ctx.lineWidth = 1 / Math.min(scaleX, scaleY)
          ctx.fillRect(currentX + padding, slotY, groupWidth - padding * 2, slotHeight)
          ctx.strokeRect(currentX + padding, slotY, groupWidth - padding * 2, slotHeight)

          ctx.fillStyle = '#000'
          ctx.fillText(`${slot.v}`, currentX + padding + 6, slotY + slotHeight / 2)

          slotY += slotHeight
        }

        currentX += groupWidth + groupGapX
      }

      // Переходим к следующей таблице по вертикали
      currentY += tableHeight + tableGapY
      uniqueTableIdx++
    }
  }

  // Обработка колеса мыши (зум и прокрутка)
  function handleWheel(e) {
    e.preventDefault()
    const rect = canvas.getBoundingClientRect()
    
    if (e.ctrlKey || e.metaKey) {
      // Zoom
      const mouseX = e.clientX - rect.left
      const mouseY = e.clientY - rect.top
      
      const svgMouseX = vb.x + (mouseX * vb.w) / rect.width
      const svgMouseY = vb.y + (mouseY * vb.h) / rect.height
      
      const zoomFactor = Math.pow(1.001, e.deltaY * -2)
      const newW = vb.w / zoomFactor
      const newH = vb.h / zoomFactor
      
      if (newW < 100 || newW > 40000) return
      
      vb.x = svgMouseX - (mouseX / rect.width) * newW
      vb.y = svgMouseY - (mouseY / rect.height) * newH
      vb.w = newW
      vb.h = newH
    } else {
      // Pan
      vb.x += (e.deltaX * vb.w) / rect.width
      vb.y += (e.deltaY * vb.h) / rect.height
    }
    
    drawVisualization()
  }

  // Обработка нажатия мыши
  function handleMouseDown(e) {
    if (e.button === 0) {
      isPanning = true
      lastMouseX = e.clientX
      lastMouseY = e.clientY
      canvas.style.cursor = 'grabbing'
    }
  }

  // Обработка отпускания мыши
  function handleMouseUp() {
    isPanning = false
    canvas.style.cursor = 'grab'
  }

  // Обработка движения мыши
  function handleMouseMove(e) {
    if (isPanning) {
      const dx = e.clientX - lastMouseX
      const dy = e.clientY - lastMouseY
      lastMouseX = e.clientX
      lastMouseY = e.clientY
      
      const rect = canvas.getBoundingClientRect()
      vb.x -= (dx * vb.w) / rect.width
      vb.y -= (dy * vb.h) / rect.height
      
      drawVisualization()
    }
  }

  onMount(() => {
    ctx = canvas.getContext('2d')
    loadData()

    // Синхронизация viewBox с размером контейнера
    const syncViewBox = () => {
      const container = canvas.parentElement
      const rect = container.getBoundingClientRect()
      const aspect = rect.height / rect.width
      vb.h = vb.w * aspect
      drawVisualization()
    }

    syncViewBox()

    window.addEventListener('resize', syncViewBox)
    return () => {
      window.removeEventListener('resize', syncViewBox)
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
    <canvas
      bind:this={canvas}
      style="width: 100%; height: 100%;"
    ></canvas>
  </div>
</div>

<style>
  .root {
    width: 100%;
    height: 100vh;
    overflow: hidden;
    background: #ebfbee;
  }

  .canvas-container {
    width: 100%;
    height: 100%;
    overflow: hidden;
    cursor: grab;
    touch-action: none;
  }

  canvas {
    display: block;
  }
</style>
