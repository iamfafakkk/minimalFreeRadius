<script lang="ts">
	import { LayoutDashboard, Server, Users, LogOut, Radio } from 'lucide-svelte';
	import Button from '$lib/components/ui/button/button.svelte';

	export let data: { username: string };
</script>

<div class="flex min-h-screen bg-muted/20">
	<aside class="hidden w-60 flex-col border-r bg-background p-4 md:flex">
		<a href="/dashboard" class="mb-6 flex items-center gap-2 px-2 text-lg font-semibold">
			<Radio class="h-5 w-5" /> FreeRADIUS
		</a>
		<nav class="flex flex-col gap-1 text-sm">
			<a href="/dashboard" class="rounded-md px-3 py-2 hover:bg-accent hover:text-accent-foreground">
				<span class="inline-flex items-center gap-2"><LayoutDashboard class="h-4 w-4" /> Dashboard</span>
			</a>
			<a href="/dashboard/nas" class="rounded-md px-3 py-2 hover:bg-accent hover:text-accent-foreground">
				<span class="inline-flex items-center gap-2"><Server class="h-4 w-4" /> NAS</span>
			</a>
			<a href="/dashboard/users" class="rounded-md px-3 py-2 hover:bg-accent hover:text-accent-foreground">
				<span class="inline-flex items-center gap-2"><Users class="h-4 w-4" /> Users</span>
			</a>
		</nav>
		<div class="mt-auto border-t pt-4 text-sm text-muted-foreground">
			<p class="px-2">Login sebagai <span class="font-medium text-foreground">{data.username}</span></p>
			<form method="POST" action="/logout" class="mt-2 px-2">
				<Button variant="outline" size="sm"><LogOut class="mr-2 h-4 w-4" /> Keluar</Button>
			</form>
		</div>
	</aside>
	<div class="flex-1">
		<header class="flex items-center justify-between border-b bg-background px-4 py-3 md:hidden">
			<span class="font-semibold">FreeRADIUS</span>
			<nav class="flex gap-3 text-sm">
				<a href="/dashboard">Dashboard</a>
				<a href="/dashboard/nas">NAS</a>
				<a href="/dashboard/users">Users</a>
			</nav>
		</header>
		<main class="mx-auto w-full max-w-6xl p-4 md:p-6">
			<slot />
		</main>
	</div>
</div>
