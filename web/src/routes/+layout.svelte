<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import { visibleNav } from '#lib/nav.ts';
	import type { LayoutData } from './$types';
	import '../app.css';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();

	const items = $derived(visibleNav(data.user));
</script>

<svelte:head>
	<title>ColdTrace</title>
</svelte:head>

<a class="skip" href="#main">Skip to content</a>

<header>
	<a class="brand" href="/">ColdTrace</a>
	<nav aria-label="Main">
		<ul>
			{#each items as item (item.href)}
				<li>
					<a href={item.href} aria-current={page.url.pathname === item.href ? 'page' : undefined}>{item.label}</a>
				</li>
			{/each}
		</ul>
	</nav>
	<div class="account">
		{#if data.user}
			<span>{data.user.name}</span>
		{:else}
			<a href="/login">Sign in</a>
		{/if}
	</div>
</header>

<main id="main">
	{@render children()}
</main>

<footer>All sensors and shipments are simulated. Not a validated GxP system.</footer>
