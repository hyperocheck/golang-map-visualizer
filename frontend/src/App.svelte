<script>
	import { onMount } from 'svelte';

	let buckets = [];
	let oldBuckets = [];
	let chains = [];
	let oldChains = [];
	let hmap = null;

	const tophashHeight = 24;
	const rowHeight = 22;
	const gapX = 80;
	const gapY = 40;
	const arrowOffset = 4;
	const bucketRadius = 12;
	const padding = 12;
	const bucketStrokeWidth = 1.5;
	const fixedTophashCellWidth = 27;

	let svgBuckets = [];
	let svgArrows = [];
	let svgLabels = [];
	let svgWidth = 2000;
	let svgHeight = 2000;

	// Состояние viewBox [x, y, width, height]
	let vb = { x: 0, y: 0, w: 1200, h: 900 };

	// Состояние панелей
	let sideWidth = 280;
	let lastSideWidth = 280;
	let isSideVisible = true;
	let resizing = false;
	let isPanning = false;

	async function load() {
		try {
			const [vizRes, oldRes, hmapRes] = await Promise.all([
				fetch('/vizual').catch(() => null),
				fetch('/vizual_old').catch(() => null),
				fetch('/hmap').catch(() => null)
			]);
			
			if (vizRes?.ok) {
				buckets = await vizRes.json();
				if (!Array.isArray(buckets)) buckets = [];
			}
			
			if (oldRes?.ok) {
				oldBuckets = await oldRes.json();
				if (!Array.isArray(oldBuckets)) oldBuckets = [];
			}
			
			if (hmapRes?.ok) {
				hmap = await hmapRes.json();
			}
			
			chains = [];
			if (Array.isArray(buckets)) {
				let current = [];
				for (const b of buckets) {
					if (b && b.type === 'main') {
						if (current.length) chains.push(current);
						current = [b];
					} else if (b) {
						current.push(b);
					}
				}
				if (current.length) chains.push(current);
			}
			
			oldChains = [];
			if (Array.isArray(oldBuckets)) {
				let current = [];
				for (const b of oldBuckets) {
					if (b && b.type === 'main') {
						if (current.length) oldChains.push(current);
						current = [b];
					} else if (b) {
						current.push(b);
					}
				}
				if (current.length) oldChains.push(current);
			}
			
			buildSVG();
			
			// Начальная подстройка viewBox под контейнер
			const container = document.getElementById('svg-container');
			if (container) {
				const rect = container.getBoundingClientRect();
				const aspect = rect.height / rect.width;
				vb.w = 1200;
				vb.h = 1200 * aspect;
				vb = vb;
			}
		} catch (e) {
			console.error("Ошибка загрузки данных:", e);
		}
	}

	function buildSVG() {
		svgBuckets = [];
		svgArrows = [];
		svgLabels = [];
		
		const hasOldChains = oldChains && oldChains.length > 0;
		const hasNewChains = chains && chains.length > 0;
		const showLabels = hasOldChains && hasNewChains; // показываем метки только если есть обе цепочки
		
		// Собираем все цепочки и определяем ширины
		const allChains = [];
		if (hasOldChains) {
			oldChains.forEach(chain => {
				if (chain && chain.length > 0) {
					allChains.push({ chain, isOld: true });
				}
			});
		}
		if (hasNewChains) {
			chains.forEach(chain => {
				if (chain && chain.length > 0) {
					allChains.push({ chain, isOld: false });
				}
			});
		}
		
		// Вычисляем ширины для каждой позиции
		const chainWidths = [];
		for (let idx = 0; idx < Math.max(hasOldChains ? oldChains.length : 0, hasNewChains ? chains.length : 0); idx++) {
			let maxWidth = 260;
			
			// Проверяем old chain
			if (hasOldChains && idx < oldChains.length && oldChains[idx] && oldChains[idx].length > 0) {
				const chain = oldChains[idx];
				const fixedTophashWidth = Math.max(...chain.map(b => (b.tophash && Array.isArray(b.tophash)) ? b.tophash.length : 0)) * fixedTophashCellWidth;
				const widths = chain.map(b => {
					const keys = (b.keys && Array.isArray(b.keys)) ? b.keys : [];
					const values = (b.values && Array.isArray(b.values)) ? b.values : [];
					const maxLen = Math.max(
						...keys.map(k => k ? k.toString().length : 0),
						...values.map(v => v ? v.toString().length : 0),
						b.overflow ? b.overflow.toString().length : 0,
						0
					);
					return maxLen * 8 + padding * 2;
				});
				maxWidth = Math.max(maxWidth, ...widths, fixedTophashWidth);
			}
			
			// Проверяем new chain
			if (hasNewChains && idx < chains.length && chains[idx] && chains[idx].length > 0) {
				const chain = chains[idx];
				const fixedTophashWidth = Math.max(...chain.map(b => (b.tophash && Array.isArray(b.tophash)) ? b.tophash.length : 0)) * fixedTophashCellWidth;
				const widths = chain.map(b => {
					const keys = (b.keys && Array.isArray(b.keys)) ? b.keys : [];
					const values = (b.values && Array.isArray(b.values)) ? b.values : [];
					const maxLen = Math.max(
						...keys.map(k => k ? k.toString().length : 0),
						...values.map(v => v ? v.toString().length : 0),
						b.overflow ? b.overflow.toString().length : 0,
						0
					);
					return maxLen * 8 + padding * 2;
				});
				maxWidth = Math.max(maxWidth, ...widths, fixedTophashWidth);
			}
			
			chainWidths.push(maxWidth);
		}
		
		let x = gapX;
		let oldMaxY = 0;
		let newMaxY = 0;
		let oldStartY = gapY + (showLabels ? 30 : 0); // место для метки если нужно
		let newStartY = gapY;
		
		// Сначала рисуем oldBuckets сверху
		if (hasOldChains) {
			for (let idx = 0; idx < oldChains.length; idx++) {
				const chain = oldChains[idx];
				if (!chain || chain.length === 0) continue;
				
				const width = chainWidths[idx] || 260;
				let y = oldStartY;
				
				for (let i = 0; i < chain.length; i++) {
					const b = chain[i];
					if (!b) continue;
					const keys = (b.keys && Array.isArray(b.keys)) ? b.keys : [];
					const values = (b.values && Array.isArray(b.values)) ? b.values : [];
					const height = tophashHeight + keys.length * rowHeight + values.length * rowHeight + rowHeight + padding * 2;
					svgBuckets.push({ id: `old-${b.id}`, x, y, width, height, bucket: b, padding, isOld: true });
					
					if (i < chain.length - 1) {
						svgArrows.push({ x: x + width / 2, y1: y + height, y2: y + height + gapY - arrowOffset, isOld: true });
						y += height + gapY;
					} else {
						y += height + gapY;
					}
					if (y > oldMaxY) oldMaxY = y;
				}
				x += width + gapX;
			}
			
			// Добавляем метку для old если нужно
			if (showLabels) {
				svgLabels.push({
					x: gapX,
					y: gapY + 20,
					text: 'OLD',
					isOld: true
				});
			}
			
			// Вычисляем стартовую позицию для новых buckets
			newStartY = oldMaxY + gapY * 2;
		}
		
		// Затем рисуем обычные buckets снизу
		if (hasNewChains) {
			x = gapX; // начинаем с той же позиции X
			
			for (let idx = 0; idx < chains.length; idx++) {
				const chain = chains[idx];
				if (!chain || chain.length === 0) continue;
				
				const width = chainWidths[idx] || 260;
				let y = newStartY;
				
				for (let i = 0; i < chain.length; i++) {
					const b = chain[i];
					if (!b) continue;
					const keys = (b.keys && Array.isArray(b.keys)) ? b.keys : [];
					const values = (b.values && Array.isArray(b.values)) ? b.values : [];
					const height = tophashHeight + keys.length * rowHeight + values.length * rowHeight + rowHeight + padding * 2;
					svgBuckets.push({ id: b.id, x, y, width, height, bucket: b, padding, isOld: false });
					
					if (i < chain.length - 1) {
						svgArrows.push({ x: x + width / 2, y1: y + height, y2: y + height + gapY - arrowOffset, isOld: false });
						y += height + gapY;
					} else {
						y += height + gapY;
					}
					if (y > newMaxY) newMaxY = y;
				}
				x += width + gapX;
			}
			
			// Добавляем метку для new если нужно
			if (showLabels) {
				svgLabels.push({
					x: gapX,
					y: newStartY - gapY + 20,
					text: 'NEW',
					isOld: false
				});
			}
		}
		
		svgWidth = x + 200;
		svgHeight = Math.max(oldMaxY, newMaxY) + 200;
	}

	// Навигация
	function handleWheel(e) {
		e.preventDefault();
		const rect = e.currentTarget.getBoundingClientRect();
		
		if (e.ctrlKey || e.metaKey) {
			const mouseX = e.clientX - rect.left;
			const mouseY = e.clientY - rect.top;
			const svgMouseX = vb.x + (mouseX * vb.w) / rect.width;
			const svgMouseY = vb.y + (mouseY * vb.h) / rect.height;

			const zoomFactor = Math.pow(1.001, e.deltaY * 5); 
			const newW = vb.w * zoomFactor;
			const newH = vb.h * zoomFactor;

			if (newW < 100 || newW > 40000) return;

			vb.x = svgMouseX - (mouseX / rect.width) * newW;
			vb.y = svgMouseY - (mouseY / rect.height) * newH;
			vb.w = newW;
			vb.h = newH;
		} else {
			vb.x += (e.deltaX * vb.w) / rect.width;
			vb.y += (e.deltaY * vb.h) / rect.height;
		}
		vb = vb;
	}

	function handleMouseDown(e) { if (e.button === 0) isPanning = true; }
	function handleMouseUp() { isPanning = false; }
	function handleMouseMove(e) {
		if (isPanning) {
			const rect = document.getElementById('svg-container').getBoundingClientRect();
			vb.x -= (e.movementX * vb.w) / rect.width;
			vb.y -= (e.movementY * vb.h) / rect.height;
			vb = vb;
		}
	}

	// ФИКС РЕСАЙЗА: Линейный расчет без тряски
	function startResize(e) {
		resizing = true;
		const startMouseX = e.clientX;
		const startSideWidth = sideWidth;
		
		// Фиксируем масштаб и аспект один раз в момент клика
		const container = document.getElementById('svg-container');
		const initialRect = container.getBoundingClientRect();
		const unitsPerPixel = vb.w / initialRect.width;
		const aspect = initialRect.height / initialRect.width;

		const onMouseMove = (ev) => {
			if (!resizing) return;
			const dx = startMouseX - ev.clientX;
			const newSideWidth = Math.max(100, Math.min(800, startSideWidth + dx));
			const diffPx = newSideWidth - sideWidth;
			
			// Синхронное обновление CSS и viewBox
			sideWidth = newSideWidth;
			vb.w -= diffPx * unitsPerPixel; 
			
			// Используем математический аспект вместо опроса DOM для плавности
			const currentContainerWidth = initialRect.width - (newSideWidth - startSideWidth);
			vb.h = vb.w * (initialRect.height / currentContainerWidth);
			
			vb = vb;
		};

		const onMouseUp = () => {
			resizing = false;
			window.removeEventListener('mousemove', onMouseMove);
			window.removeEventListener('mouseup', onMouseUp);
		};

		window.addEventListener('mousemove', onMouseMove);
		window.addEventListener('mouseup', onMouseUp);
	}

	function toggleSide() {
		const container = document.getElementById('svg-container');
		const rect = container.getBoundingClientRect();
		const unitsPerPixel = vb.w / rect.width;

		if (isSideVisible) {
			lastSideWidth = sideWidth;
			vb.w += sideWidth * unitsPerPixel;
			sideWidth = 0;
			isSideVisible = false;
		} else {
			sideWidth = lastSideWidth;
			vb.w -= sideWidth * unitsPerPixel;
			isSideVisible = true;
		}
		
		setTimeout(() => {
			const newRect = container.getBoundingClientRect();
			vb.h = vb.w * (newRect.height / newRect.width);
			vb = vb;
		}, 0);
	}

	onMount(load);
</script>

<div class="root">
	<div 
		id="svg-container" 
		on:wheel|nonpassive={handleWheel}
		on:mousedown={handleMouseDown}
		on:mousemove={handleMouseMove}
		on:mouseup={handleMouseUp}
		on:mouseleave={handleMouseUp}
	>
		<button class="toggle-btn" on:click={toggleSide}>
			{isSideVisible ? '→' : '← hmap'}
		</button>

		<svg 
			viewBox="{vb.x} {vb.y} {vb.w} {vb.h}" 
			width="100%" 
			height="100%"
			preserveAspectRatio="xMinYMin meet"
		>
			<defs>
				<marker id="arrow" markerWidth="6" markerHeight="6" refX="3" refY="3" orient="auto">
					<path d="M0,0 L6,3 L0,6 Z" fill="#000" />
				</marker>
				<marker id="arrow-old" markerWidth="6" markerHeight="6" refX="3" refY="3" orient="auto">
					<path d="M0,0 L6,3 L0,6 Z" fill="#ff6b6b" />
				</marker>
				<filter id="bucket-shadow" x="-30%" y="-30%" width="140%" height="140%">
					<feDropShadow dx="4" dy="4" stdDeviation="1" flood-color="#51cf66" flood-opacity="0.5"/>
				</filter>
			</defs>
			
			<!-- Метки old/new -->
			{#each svgLabels as label}
				<text
					x={label.x}
					y={label.y}
					class={label.isOld ? 'label-old' : 'label-new'}
					font-size="14"
					font-weight="bold"
					font-family="Arial, sans-serif"
				>
					{label.text}
				</text>
			{/each}

			{#each svgBuckets as b (b.id)}
				{#if b.x + b.width > vb.x && b.x < vb.x + vb.w && b.y + b.height > vb.y && b.y < vb.y + vb.h}
					<g transform={`translate(${b.x}, ${b.y})`} class="bucket-group">
						<rect 
							class="bucket-rect {b.isOld ? 'bucket-old' : ''}" 
							width={b.width} 
							height={b.height} 
							fill="#fff" 
							stroke={b.isOld ? '#ff6b6b' : '#000'} 
							stroke-width={b.isOld ? 2 : bucketStrokeWidth} 
							rx={bucketRadius} 
							ry={bucketRadius}
							filter="url(#bucket-shadow)"
						/>
						{#each (b.bucket?.tophash || []) as t, i}
							<rect x={b.padding + i * fixedTophashCellWidth} y={b.padding} width={fixedTophashCellWidth} height={tophashHeight} fill="#eee" stroke="#000" />
							<text x={b.padding + i * fixedTophashCellWidth + fixedTophashCellWidth / 2} y={b.padding + tophashHeight / 1.5} font-size="12" text-anchor="middle">{t}</text>
						{/each}
						{#each (b.bucket?.keys || []) as k, i}
							<rect x={b.padding} y={b.padding + tophashHeight + i * rowHeight} width={b.width - padding * 2} height={rowHeight} fill={k == null ? '#dbfdc9' : '#b2f2bb'} stroke="#12b886" class="cell-key" />
							<text x={b.padding + 6} y={b.padding + tophashHeight + i * rowHeight + rowHeight / 1.5} font-size="13">{k ?? ''}</text>
						{/each}
						{#each (b.bucket?.values || []) as v, i}
							{@const keysLen = (b.bucket?.keys || []).length}
							<rect x={b.padding} y={b.padding + tophashHeight + keysLen * rowHeight + i * rowHeight} width={b.width - padding * 2} height={rowHeight} fill={v == null ? '#fff2b8' : '#ffec99'} stroke="#ffa94d" class="cell-value" />
							<text x={b.padding + 6} y={b.padding + tophashHeight + keysLen * rowHeight + i * rowHeight + rowHeight / 1.5} font-size="13">{v ?? ''}</text>
						{/each}
						{#if b.bucket}
							{@const keysLen = (b.bucket.keys || []).length}
							{@const valuesLen = (b.bucket.values || []).length}
							<rect x={b.padding} y={b.padding + tophashHeight + keysLen * rowHeight + valuesLen * rowHeight} width={b.width - padding * 2} height={rowHeight} fill="#ddd" stroke="#000" />
							<text x={b.padding + 6} y={b.padding + tophashHeight + keysLen * rowHeight + valuesLen * rowHeight + rowHeight / 1.5} font-size="12">{b.bucket.overflow || ''}</text>
						{/if}
					</g>
				{/if}
			{/each}

			{#each svgArrows as a}
				{#if a.x > vb.x && a.x < vb.x + vb.w}
					<line 
						x1={a.x} 
						y1={a.y1} 
						x2={a.x} 
						y2={a.y2} 
						stroke={a.isOld ? '#ff6b6b' : '#000'} 
						stroke-width="2" 
						marker-end={a.isOld ? 'url(#arrow-old)' : 'url(#arrow)'} 
					/>
				{/if}
			{/each}
		</svg>
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
	</div>
</div>

<style>
	:global(::-webkit-scrollbar) { width: 0 !important; height: 0 !important; display: none !important; }
	:global(*) { -ms-overflow-style: none !important; scrollbar-width: none !important; }

	.root {
		display: flex;
		width: 100vw;
		height: 100vh;
		overflow: hidden;
		background: #ebfbee;
	}

	#svg-container {
		flex: 1;
		position: relative;
		overflow: hidden;
		cursor: grab;
		touch-action: none;
		transition: none !important;
	}
	#svg-container:active { cursor: grabbing; }

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
		box-shadow: 0 2px 8px rgba(0,0,0,0.1);
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
		transition: none !important;
	}

	svg {
		display: block;
		user-select: none;
	}

	.row { display: flex; justify-content: space-between; padding: 2px 0; border-bottom: 1px solid #eee; }
	.bucket-group:hover .bucket-rect { stroke-width: 2.5; }
	.bucket-group:hover .bucket-old { stroke: #ff8787 !important; }
	.bucket-group:hover .bucket-rect:not(.bucket-old) { stroke: #228be6; }
	.label-old { fill: #ff6b6b; }
	.label-new { fill: #51cf66; }
	.cell-key:hover { fill: #9feaa4; }
	.cell-value:hover { fill: #ffe066; }
</style>

