<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { verifySession } from '$lib/api.js';

	onMount(async () => {
		const s = await verifySession();
		await goto(s.valid ? '/dashboard' : '/login', { replaceState: true });
	});
</script>

<div class="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
	<p>Opening FreeRADIUS panel...</p>
</div>
