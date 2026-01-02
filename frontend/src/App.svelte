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
	const fixedTophashCellWidth = 27; // фиксированная ширина каждой ячейки tophash
	let svgBuckets = [];
	let svgArrows = [];
	let svgWidth = 2000;
	let svgHeight = 2000;
	let sideWidth = 280;
	let resizing = false;
	let startX;
	let startWidth;
	async function load() {
		const [vizRes, hmapRes] = await Promise.all([
			fetch('/vizual'),
			fetch('/hmap')
		]);
		buckets = await vizRes.json();
		hmap = await hmapRes.json();
		chains = [];
		let current = [];
		for (const b of buckets) {
			if (b.type === 'main') {
				if (current.length) chains.push(current);
				current = [b];
			} else {
				current.push(b);
			}
		}
		if (current.length) chains.push(current);
		buildSVG();
		setTimeout(() => {
			const svgContainer = document.getElementById('svg-container');
			if (svgContainer) {
				svgContainer.scrollLeft = 0;
				svgContainer.scrollTop = 0;
			}
		}, 0);
	}
	function buildSVG() {
		svgBuckets = [];
		svgArrows = [];
		let x = gapX;
		let maxY = 0;
		for (const chain of chains) {
			let y = gapY;
			// минимальная ширина по сумме tophash
			const fixedTophashWidth = Math.max(...chain.map(b => b.tophash.length)) * fixedTophashCellWidth;
			// ищем максимальную ширину всех ключей, значений и overflow в цепочке
			const chainWidths = chain.map(b => {
				const allKeys = b.Keys.map(k => k ? k.toString().length : 0);
				const allVals = b.Values.map(v => v ? v.toString().length : 0);
				const overflowLen = b.overflow ? b.overflow.toString().length : 0;
				const maxLen = Math.max(...allKeys, ...allVals, overflowLen, 0);
				return maxLen * 8 + padding * 2; // 8px на символ + padding
			});
			const maxBucketWidthInChain = Math.max(...chainWidths, fixedTophashWidth, 260);
			for (let i = 0; i < chain.length; i++) {
				const b = chain[i];
				const width = maxBucketWidthInChain;
				const innerHeight =
					tophashHeight +
					b.Keys.length * rowHeight +
					b.Values.length * rowHeight +
					rowHeight; // overflow
				const height = innerHeight + padding * 2;
				svgBuckets.push({
					id: b.id,
					x,
					y,
					width,
					height,
					bucket: b,
					padding
				});
				if (i < chain.length - 1) {
					const nextY = y + height + gapY;
					svgArrows.push({
						x: x + width / 2,
						y1: y + height,
						y2: nextY - arrowOffset
					});
					y = nextY;
				} else {
					y += height + gapY;
				}
				if (y > maxY) maxY = y;
			}
			x += maxBucketWidthInChain + gapX;
		}
		svgWidth = x + 200;
		svgHeight = maxY + 200;
	}
	function startResize(e) {
		resizing = true;
		startX = e.clientX;
		startWidth = sideWidth;
		document.addEventListener('mousemove', resize);
		document.addEventListener('mouseup', stopResize);
	}
	function resize(e) {
		if (!resizing) return;
		const dx = startX - e.clientX;
		sideWidth = Math.max(100, Math.min(800, startWidth + dx));
	}
	function stopResize() {
		resizing = false;
		document.removeEventListener('mousemove', resize);
		document.removeEventListener('mouseup', stopResize);
	}
	onMount(load);
</script>

<div class="root">
	<div id="svg-container">
		<svg width={svgWidth} height={svgHeight}>
			<defs>
				<marker id="arrow" markerWidth="6" markerHeight="6" refX="3" refY="3" orient="auto">
					<path d="M0,0 L6,3 L0,6 Z" fill="#000" />
				</marker>
			</defs>
			{#each svgBuckets as b}
				<g transform={`translate(${b.x}, ${b.y})`} class="bucket-group">
					<rect
						class="bucket-rect"
						x={0}
						y={0}
						width={b.width}
						height={b.height}
						fill="#fff"
						stroke="#000"
						stroke-width={bucketStrokeWidth}
						rx={bucketRadius}
						ry={bucketRadius}
					/>
					<!-- TOPHASH FIXED -->
					{#each b.bucket.tophash as t, i}
						<rect
							x={b.padding + i * fixedTophashCellWidth}
							y={b.padding}
							width={fixedTophashCellWidth}
							height={tophashHeight}
							fill="#eee"
							stroke="#000"
						/>
						<text
							x={b.padding + i * fixedTophashCellWidth + fixedTophashCellWidth / 2}
							y={b.padding + tophashHeight / 1.5}
							font-size="12"
							text-anchor="middle"
						>{t}</text>
					{/each}
					{#each b.bucket.Keys as k, i}
						<rect
							x={b.padding}
							y={b.padding + tophashHeight + i * rowHeight}
							width={b.width - padding * 2}
							height={rowHeight}
							fill={k == null ? '#dbfdc9' : '#b2f2bb'}
							stroke="#12b886"
							class="cell-key"
						/>
						<text
							x={b.padding + 6}
							y={b.padding + tophashHeight + i * rowHeight + rowHeight / 1.5}
							font-size="13"
						>{k ?? ''}</text>
					{/each}
					{#each b.bucket.Values as v, i}
						<rect
							x={b.padding}
							y={b.padding + tophashHeight + b.bucket.Keys.length * rowHeight + i * rowHeight}
							width={b.width - padding * 2}
							height={rowHeight}
							fill={v == null ? '#fff2b8' : '#ffec99'}
							stroke="#ffa94d"
							class="cell-value"
						/>
						<text
							x={b.padding + 6}
							y={b.padding + tophashHeight + b.bucket.Keys.length * rowHeight + i * rowHeight + rowHeight / 1.5}
							font-size="13"
						>{v ?? ''}</text>
					{/each}
					<rect
						x={b.padding}
						y={b.padding + tophashHeight + b.bucket.Keys.length * rowHeight + b.bucket.Values.length * rowHeight}
						width={b.width - padding * 2}
						height={rowHeight}
						fill="#ddd"
						stroke="#000"
					/>
					<text
						x={b.padding + 6}
						y={b.padding + tophashHeight + b.bucket.Keys.length * rowHeight + b.bucket.Values.length * rowHeight + rowHeight / 1.5}
						font-size="12"
					>{b.bucket.overflow}</text>
				</g>
			{/each}
			{#each svgArrows as a}
				<line
					x1={a.x}
					y1={a.y1}
					x2={a.x}
					y2={a.y2}
					stroke="#000"
					stroke-width="2"
					marker-end="url(#arrow)"
				/>
			{/each}
		</svg>
	</div>
	<div class="splitter" on:mousedown={startResize}></div>
	<div class="side" style="width: {sideWidth}px;">
		<h3>hmap</h3>
		{#if hmap}
			{#each Object.entries(hmap) as [k, v]}
				<div class="row">
					<span>{k}</span>
					<b>{v}</b>
				</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.root {
		display: flex;
		width: 100vw;
		height: 100vh;
		overflow: hidden;
	}
	#svg-container {
		flex: 1;
		overflow: auto;
		background: #ebfbee;
	}
	.splitter {
		width: 5px;
		background: #ccc;
		cursor: col-resize;
		user-select: none;
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
	.side h3 {
		margin: 0 0 10px;
	}
	.row {
		display: flex;
		justify-content: space-between;
		padding: 2px 0;
	}
	svg {
		display: block;
	}
	.bucket-group:hover .bucket-rect {
		stroke: red;
	}
	.cell-key:hover {
		fill: #9feaa4;
	}
	.cell-value:hover {
		fill: #ffe066;
	}
	/* 1. Скрываем стандартные полосы прокрутки браузера во всем приложении */
	:global(::-webkit-scrollbar) {
	    width: 0px;
	    height: 0px;
	    background: transparent; /* делает полосу полностью прозрачной */
	}
	
	/* 2. Для Firefox (отдельное свойство) */
	:global(*) {
	    scrollbar-width: none;
	}
	
	/* 3. Убеждаемся, что у корневых элементов не вылезают лишние полосы */
	:global(html, body) {
	    overflow: hidden; 
	    margin: 0;
	    padding: 0;
	}
	
	/* Ваши контейнеры остаются с overflow, но без видимых полос */
	#svg-container {
	    flex: 1;
	    overflow: auto; /* функционал прокрутки остается */
	    background: #ebfbee;
	    -ms-overflow-style: none; /* для старого Edge */
	}
	
	.side {
	    overflow-y: auto; /* функционал прокрутки остается */
	    -ms-overflow-style: none;
	}
</style>
