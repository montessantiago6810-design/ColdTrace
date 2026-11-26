<script lang="ts">
	import { enhance } from '$app/forms';
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
</script>

<svelte:head>
	<title>Sign in · ColdTrace</title>
</svelte:head>

<h1>Sign in</h1>

<form method="POST" class="card" use:enhance>
	{#if form?.message}
		<p role="alert">{form.message}</p>
	{/if}
	<label>
		Email
		<input name="email" type="email" autocomplete="username" required value={form?.email ?? ''} />
	</label>
	<label>
		Password
		<input name="password" type="password" autocomplete="current-password" required />
	</label>
	<button type="submit">Sign in</button>
</form>
