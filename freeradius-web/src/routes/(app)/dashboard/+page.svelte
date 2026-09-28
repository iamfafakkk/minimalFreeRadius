<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card/index.js';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import { Server, Users, Activity } from 'lucide-svelte';

	let nas = 0;
	let users = 0;
	let db = 'unknown';
	let healthy = false;
	let error = '';

	onMount(async () => {
		const [n, u, h] = await Promise.all([
			api.get<{ total_nas?: number }>('/v1/nas/stats'),
			api.get<{ total_users?: number }>('/v1/users/stats'),
			api.get<{ database?: string }>('/v1/auth/health')
		]);
		if (!n.ok && n.status === 401) {
			await goto('/login');
			return;
		}
		if (!n.ok || !u.ok || !h.ok) {
			error = 'Gagal memuat statistik dari backend.';
			return;
		}
		nas = (n.data as { total_nas?: number })?.total_nas ?? 0;
		users = (u.data as { total_users?: number })?.total_users ?? 0;
		db = (h.data as { database?: string })?.database ?? 'unknown';
		healthy = db === 'connected';
	});
</script>

<h1 class="mb-1 text-2xl font-bold">Dashboard</h1>
<p class="mb-6 text-sm text-muted-foreground">Ringkasan FreeRADIUS.</p>

{#if error}
	<p class="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</p>
{/if}

<div class="grid gap-4 md:grid-cols-3">
	<Card>
		<CardHeader><CardTitle class="flex items-center gap-2 text-sm"><Server class="h-4 w-4" /> Total NAS</CardTitle></CardHeader>
		<CardContent><p class="text-3xl font-bold">{nas}</p></CardContent>
	</Card>
	<Card>
		<CardHeader><CardTitle class="flex items-center gap-2 text-sm"><Users class="h-4 w-4" /> Total Users</CardTitle></CardHeader>
		<CardContent><p class="text-3xl font-bold">{users}</p></CardContent>
	</Card>
	<Card>
		<CardHeader><CardTitle class="flex items-center gap-2 text-sm"><Activity class="h-4 w-4" /> Database</CardTitle></CardHeader>
		<CardContent>
			<Badge variant={healthy ? 'default' : 'destructive'}>{db}</Badge>
		</CardContent>
	</Card>
</div>

<Card class="mt-6">
	<CardHeader>
		<CardTitle>Kelola</CardTitle>
		<CardDescription>CRUD via /api/* same-origin dengan cookie sesi.</CardDescription>
	</CardHeader>
	<CardContent class="flex gap-3">
		<a href="/dashboard/nas" class="rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground">Kelola NAS</a>
		<a href="/dashboard/users" class="rounded-md border px-4 py-2 text-sm">Kelola Users</a>
	</CardContent>
</Card>
