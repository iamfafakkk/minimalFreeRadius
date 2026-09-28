<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, sessionUsername, verifySession } from '$lib/api.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import ThemeSwitcher from '$lib/components/theme-switcher.svelte';
	import { Separator } from '$lib/components/ui/separator/index.js';

	let { children } = $props();
	let username = $state('');
	let ready = $state(false);

	onMount(async () => {
		const s = await verifySession();
		if (!s.valid) {
			await goto('/login', { replaceState: true });
			return;
		}
		username = s.username ?? sessionUsername();
		ready = true;
	});

	async function logout() {
		try {
			await api.post('/v1/auth/logout');
		} finally {
			await goto('/login');
		}
	}
</script>

{#if ready}
	<Sidebar.Provider>
		<AppSidebar {username} onlogout={logout} />
		<Sidebar.Inset>
			<header class="flex h-16 shrink-0 items-center gap-2 border-b px-4">
				<Sidebar.Trigger class="-ml-1" />
				<Separator orientation="vertical" class="mr-2 h-4" />
				<h1 class="text-sm font-medium">FreeRADIUS Panel</h1>
				<div class="ml-auto"><ThemeSwitcher /></div>
			</header>
			<div class="flex flex-1 flex-col gap-4 p-4">
				{@render children()}
			</div>
		</Sidebar.Inset>
	</Sidebar.Provider>
{:else}
	<div class="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
		<p>Checking session...</p>
	</div>
{/if}
