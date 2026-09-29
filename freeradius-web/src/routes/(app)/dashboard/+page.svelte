<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';
	import { toast } from 'svelte-sonner';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Server, Users, Activity } from '@lucide/svelte';

	let nas = $state(0);
	let users = $state(0);
	let db = $state('unknown');
	let loading = $state(true);

	const healthy = $derived(db === 'connected');

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
			toast.error('Failed to load statistics from the backend.');
			loading = false;
			return;
		}
		nas = (n.data as { total_nas?: number })?.total_nas ?? 0;
		users = (u.data as { total_users?: number })?.total_users ?? 0;
		db = (h.data as { database?: string })?.database ?? 'unknown';
		loading = false;
	});
</script>

<div>
	<h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
	<p class="text-muted-foreground text-sm">FreeRADIUS overview.</p>
</div>

<div class="grid gap-4 md:grid-cols-3">
	<Card.Root>
		<Card.Header>
			<Card.Title class="flex items-center gap-2 text-sm">
				<Server class="size-4" /> Total NAS
			</Card.Title>
		</Card.Header>
		<Card.Content>
			<p class="text-3xl font-bold">{loading ? '—' : nas}</p>
		</Card.Content>
	</Card.Root>
	<Card.Root>
		<Card.Header>
			<Card.Title class="flex items-center gap-2 text-sm">
				<Users class="size-4" /> Total Users
			</Card.Title>
		</Card.Header>
		<Card.Content>
			<p class="text-3xl font-bold">{loading ? '—' : users}</p>
		</Card.Content>
	</Card.Root>
	<Card.Root>
		<Card.Header>
			<Card.Title class="flex items-center gap-2 text-sm">
				<Activity class="size-4" /> Database
			</Card.Title>
		</Card.Header>
		<Card.Content>
			<Badge variant={healthy ? 'default' : 'destructive'}>{db}</Badge>
		</Card.Content>
	</Card.Root>
</div>

<Card.Root>
	<Card.Header>
		<Card.Title>Manage</Card.Title>
		<Card.Description>CRUD via /api/* same-origin with session cookie.</Card.Description>
	</Card.Header>
	<Card.Content class="flex gap-3">
		<Button href="/dashboard/nas">Manage NAS</Button>
		<Button href="/dashboard/users" variant="outline">Manage Users</Button>
	</Card.Content>
</Card.Root>
