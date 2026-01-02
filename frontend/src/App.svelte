<script>
	import { onMount } from 'svelte';

	let buckets = [];
	let chains = [];
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
	let svgWidth = 2000;
	let svgHeight = 2000;

	let vb = { x: 0, y: 0, w: 1200, h: 900 };

	let sideWidth = 280;
	let lastSideWidth = 280;
	let isSideVisible = true;
	let resizing = false;
	let isPanning = false;

	let lastTouchX = 0;
	let lastTouchY = 0;
	let lastTouchDist = 0;

	async function load() {
		try {
			const [vizRes, hmapRes] = await Promise.all([
				fetch('/vizual').catch(() => null),
				fetch('/hmap').catch(() => null)
			]);
			
			if (vizRes && vizRes.ok) {
				const vizText = await vizRes.text();
				buckets = vizText ? JSON.parse(vizText) : [];
			}

			if (hmapRes && hmapRes.ok) {
				const hmapText = await hmapRes.text();
				hmap = hmapText ? JSON.parse(hmapText) : {};
			}
			
			chains = [];
			let current = [];
			if (Array.isArray(buckets)) {
				for (const b of buckets) {
					if (b.type === 'main') {
						if (current.length) chains.push(current);
						current = [b];
					} else {
						current.push(b);
					}
				}
				if (current.length) chains.push(current);
			}
			
			buildSVG();
			
			const container = document.getElementById('svg-container');
			if (container) {
				const rect = container.getBoundingClientRect();
				const aspect = rect.height / rect.width;
				vb.w = 1200;
				vb.h = 1200 * aspect;
				vb = vb;
			}
		} catch (e) {
			console.error("Ошибка загрузки:", e);
		}
	}

	function buildSVG() {
		svgBuckets = [];
		svgArrows = [];
		if (!chains.length) return;

		let x = gapX;
		let maxY = 0;

		for (const chain of chains) {
			let y = gapY;
			const fixedTophashWidth = Math.max(...chain.map(b => (b.tophash ? b.tophash.length : 0))) * fixedTophashCellWidth;
			const chainWidths = chain.map(b => {
				const keys = b.Keys || [];
				const vals = b.Values || [];
				const maxLen = Math.max(
					...keys.map(k => k ? k.toString().length : 0),
					...vals.map(v => v ? v.toString().length : 0),
					b.overflow ? b.overflow.toString().length : 0,
					0
				);
				return maxLen * 8 + padding * 2;
			});
			const maxBW = Math.max(...chainWidths, fixedTophashWidth, 260);

			for (let i = 0; i < chain.length; i++) {
				const b = chain[i];
				const kLen = b.Keys ? b.Keys.length : 0;
				const vLen = b.Values ? b.Values.length : 0;
				const height = tophashHeight + kLen * rowHeight + vLen * rowHeight + rowHeight + padding * 2;
				
				svgBuckets.push({ id: b.id || Math.random(), x, y, width: maxBW, height, bucket: b, padding });
				
				if (i < chain.length - 1) {
					svgArrows.push({ x: x + maxBW / 2, y1: y + height, y2: y + height + gapY - arrowOffset });
					y += height + gapY;
				} else { y += height + gapY; }
				if (y > maxY) maxY = y;
			}
			x += maxBW + gapX;
		}
		svgWidth = x + 200;
		svgHeight = maxY + 200;
	}

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
			vb.w = newW; vb.h = newH;
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

	function handleTouchStart(e) {
		if (e.touches.length === 1) {
			lastTouchX = e.touches[0].clientX;
			lastTouchY = e.touches[0].clientY;
		} else if (e.touches.length === 2) {
			lastTouchDist = Math.hypot(e.touches[0].clientX - e.touches[1].clientX, e.touches[0].clientY - e.touches[1].clientY);
		}
	}

	function handleTouchMove(e) {
		const rect = e.currentTarget.getBoundingClientRect();
		if (e.touches.length === 1) {
			const touchX = e.touches[0].clientX;
			const touchY = e.touches[0].clientY;
			vb.x -= ((touchX - lastTouchX) * vb.w) / rect.width;
			vb.y -= ((touchY - lastTouchY) * vb.h) / rect.height;
			lastTouchX = touchX; lastTouchY = touchY;
		} else if (e.touches.length === 2) {
			const dist = Math.hypot(e.touches[0].clientX - e.touches[1].clientX, e.touches[0].clientY - e.touches[1].clientY);
			const zoomFactor = lastTouchDist / dist;
			const midX = (e.touches[0].clientX + e.touches[1].clientX) / 2 - rect.left;
			const midY = (e.touches[0].clientY + e.touches[1].clientY) / 2 - rect.top;
			const svgMidX = vb.x + (midX * vb.w) / rect.width;
			const svgMidY = vb.y + (midY * vb.h) / rect.height;
			const newW = vb.w * zoomFactor; const newH = vb.h * zoomFactor;
			if (newW > 100 && newW < 40000) {
				vb.x = svgMidX - (midX / rect.width) * newW;
				vb.y = svgMidY - (midY / rect.height) * newH;
				vb.w = newW; vb.h = newH;
			}
			lastTouchDist = dist;
		}
		vb = vb;
	}

	// ИСПРАВЛЕННЫЙ РЕСАЙЗ: Без тряски
	function startResize(e) {
		resizing = true;
		const startMouseX = e.clientX;
		const startSideWidth = sideWidth;
		
		const container = document.getElementById('svg-container');
		const initialRect = container.getBoundingClientRect();
		// Фиксируем количество юнитов SVG на 1 пиксель экрана в момент начала ресайза
		const unitsPerPixel = vb.w / initialRect.width;
		// Фиксируем физическую высоту контейнера, чтобы аспект не плавал от микро-дрожаний
		const fixedHeight = initialRect.height;
		const initialContainerWidth = initialRect.width;

		const onMM = (ev) => {
			if (!resizing) return;
			// Считаем дельту мыши
			const mouseDeltaPx = startMouseX - ev.clientX;
			const newSideWidth = Math.max(100, Math.min(800, startSideWidth + mouseDeltaPx));
			
			// На сколько РЕАЛЬНО изменилась ширина сайдбара относительно предыдущего состояния
			const sideChangePx = newSideWidth - sideWidth;
			
			// Обновляем ширину сайдбара (CSS)
			sideWidth = newSideWidth;

			// Обновляем ширину viewBox (SVG) пропорционально изменению физического контейнера.
			// Если сайдбар увеличился (sideChangePx > 0), контейнер уменьшился -> vb.w должен уменьшиться.
			vb.w -= sideChangePx * unitsPerPixel; 

			// Вычисляем новую физическую ширину контейнера для пересчета аспекта высоты
			const currentContainerWidth = initialContainerWidth - (newSideWidth - startSideWidth);
			vb.h = vb.w * (fixedHeight / currentContainerWidth);
			
			vb = vb;
		};

		const onMU = () => { 
			resizing = false; 
			window.removeEventListener('mousemove', onMM); 
			window.removeEventListener('mouseup', onMU); 
		};

		window.addEventListener('mousemove', onMM);
		window.addEventListener('mouseup', onMU);
	}

	function toggleSide() {
		const container = document.getElementById('svg-container');
		const rect = container.getBoundingClientRect();
		const unitsPerPixel = vb.w / rect.width;
		if (isSideVisible) {
			lastSideWidth = sideWidth;
			vb.w += sideWidth * unitsPerPixel;
			sideWidth = 0; isSideVisible = false;
		} else {
			sideWidth = lastSideWidth;
			vb.w -= sideWidth * unitsPerPixel;
			isSideVisible = true;
		}
		setTimeout(() => {
			const newRect = container.getBoundingClientRect();
			if (newRect.width > 0) {
				vb.h = vb.w * (newRect.height / newRect.width);
				vb = vb;
			}
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
		on:touchstart|nonpassive={handleTouchStart}
		on:touchmove|nonpassive={handleTouchMove}
	>
		<button class="toggle-btn" on:click={toggleSide}>
			{isSideVisible ? '→' : '← hmap'}
		</button>

		<svg viewBox="{vb.x} {vb.y} {vb.w} {vb.h}" width="100%" height="100%" preserveAspectRatio="xMinYMin meet">
			<defs>
				<marker id="arrow" markerWidth="6" markerHeight="6" refX="3" refY="3" orient="auto">
					<path d="M0,0 L6,3 L0,6 Z" fill="#000" />
				</marker>
			</defs>

			{#each svgBuckets as b (b.id)}
				{#if b.x + b.width > vb.x && b.x < vb.x + vb.w && b.y + b.height > vb.y && b.y < vb.y + vb.h}
					<g transform={`translate(${b.x}, ${b.y})`} class="bucket-group">
						<rect class="bucket-rect" width={b.width} height={b.height} fill="#fff" stroke="#000" stroke-width={bucketStrokeWidth} rx={bucketRadius} ry={bucketRadius} />
						{#if b.bucket.tophash}
							{#each b.bucket.tophash as t, i}
								<rect x={b.padding + i * fixedTophashCellWidth} y={b.padding} width={fixedTophashCellWidth} height={tophashHeight} fill="#eee" stroke="#000" />
								<text x={b.padding + i * fixedTophashCellWidth + fixedTophashCellWidth / 2} y={b.padding + tophashHeight / 1.5} font-size="12" text-anchor="middle">{t}</text>
							{/each}
						{/if}
						{#if b.bucket.Keys}
							{#each b.bucket.Keys as k, i}
								<rect x={b.padding} y={b.padding + tophashHeight + i * rowHeight} width={b.width - padding * 2} height={rowHeight} fill={k == null ? '#dbfdc9' : '#b2f2bb'} stroke="#12b886" class="cell-key" />
								<text x={b.padding + 6} y={b.padding + tophashHeight + i * rowHeight + rowHeight / 1.5} font-size="13">{k ?? ''}</text>
							{/each}
						{/if}
						{#if b.bucket.Values}
							{#each b.bucket.Values as v, i}
								<rect x={b.padding} y={b.padding + tophashHeight + (b.bucket.Keys ? b.bucket.Keys.length : 0) * rowHeight + i * rowHeight} width={b.width - padding * 2} height={rowHeight} fill={v == null ? '#fff2b8' : '#ffec99'} stroke="#ffa94d" class="cell-value" />
								<text x={b.padding + 6} y={b.padding + tophashHeight + (b.bucket.Keys ? b.bucket.Keys.length : 0) * rowHeight + i * rowHeight + rowHeight / 1.5} font-size="13">{v ?? ''}</text>
							{/each}
						{/if}
						<rect x={b.padding} y={b.padding + tophashHeight + ((b.bucket.Keys ? b.bucket.Keys.length : 0) + (b.bucket.Values ? b.bucket.Values.length : 0)) * rowHeight} width={b.width - padding * 2} height={rowHeight} fill="#ddd" stroke="#000" />
						<text x={b.padding + 6} y={b.padding + tophashHeight + ((b.bucket.Keys ? b.bucket.Keys.length : 0) + (b.bucket.Values ? b.bucket.Values.length : 0)) * rowHeight + rowHeight / 1.5} font-size="12">{b.bucket.overflow || ''}</text>
					</g>
				{/if}
			{/each}

			{#each svgArrows as a}
				{#if a.x > vb.x && a.x < vb.x + vb.w}
					<line x1={a.x} y1={a.y1} x2={a.x} y2={a.y2} stroke="#000" stroke-width="2" marker-end="url(#arrow)" />
				{/if}
			{/each}
		</svg>
	</div>

	{#if isSideVisible}<div class="splitter" on:mousedown={startResize}></div>{/if}

	<div class="side" style="width: {sideWidth}px; display: {isSideVisible ? 'block' : 'none'};">
		<h3>hmap structure</h3>
		{#if hmap && Object.keys(hmap).length > 0}
			{#each Object.entries(hmap) as [key, value]}
				<div class="row">
					<span class="key-label">{key}</span>
					<b class="val-label">
						{#if value !== null && typeof value === 'object'}
							<pre class="json-block">{JSON.stringify(value, null, 1)}</pre>
						{:else}
							{value ?? 'nil'}
						{/if}
					</b>
				</div>
			{/each}
		{:else}
			<div class="row">No hmap data loaded</div>
		{/if}
	</div>
</div>

<style>
	:global(::-webkit-scrollbar) {
		width: 0 !important;
		height: 0 !important;
		display: none !important;
	}
	:global(*) {
		-ms-overflow-style: none !important;
		scrollbar-width: none !important;
	}
	.root {
		display: flex;
		width: 100vw;
		height: 100vh;
		overflow: hidden;
		background: #ebfbee;
		position: fixed;
	}
	#svg-container {
		flex: 1;
		position: relative;
		overflow: hidden;
		cursor: grab;
		touch-action: none;
		user-select: none;
	}
	#svg-container:active {
		cursor: grabbing;
	}
	.toggle-btn {
		position: absolute;
		right: 15px;
		top: 15px;
		z-index: 100;
		background: #fff;
		border: 1px solid #ccc;
		padding: 10px 15px;
		border-radius: 8px;
		cursor: pointer;
		font-weight: 500;
		box-shadow: 0 2px 8px rgba(0,0,0,0.2);
	}
	.splitter {
		width: 8px;
		background: #ccc;
		cursor: col-resize;
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
	.key-label {
		color: #666;
		font-size: 11px;
		word-break: break-all;
		padding-right: 8px;
	}
	.val-label {
		color: #000;
	}
	.json-block {
		margin: 0;
		font-size: 10px;
		background: #eee;
		padding: 4px;
		border-radius: 4px;
		width: 100%;
		overflow-x: auto;
	}
	svg {
		display: block;
	}
	.cell-key:hover {
		fill: #82c91e !important;
		filter: brightness(0.9);
	}
	.cell-value:hover {
		fill: #fab005 !important;
		filter: brightness(0.9);
	}
	.bucket-group:hover .bucket-rect {
		stroke: #228be6;
		stroke-width: 2.5;
	}
	.row {
		display: flex;
		justify-content: space-between;
		padding: 4px 0;
		border-bottom: 1px solid #eee;
		flex-wrap: wrap;
	}
</style>

