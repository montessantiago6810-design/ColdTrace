<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Overview · ColdTrace</title>
</svelte:head>

<h1>Overview</h1>

<section class="card" aria-labelledby="api-status">
	<h2 id="api-status">API</h2>
	{#if data.health.ok}
		<p class="status-ok">
			<strong>Online</strong> — {data.health.health.service}, version {data.health.health.version}
		</p>
		{#if data.health.requestId}
			<p><small>Request ID {data.health.requestId}</small></p>
		{/if}
	{:else}
		<p class="status-down"><strong>Offline</strong> — {data.health.reason}</p>
		<p><small>Start it with <code>go run ./cmd/api</code>.</small></p>
	{/if}
</section>

<section class="card" aria-labelledby="fleet">
	<h2 id="fleet">Live shipments</h2>
	<p>No sensors are connected yet. Live shipments appear here once ingestion is running.</p>
</section>
