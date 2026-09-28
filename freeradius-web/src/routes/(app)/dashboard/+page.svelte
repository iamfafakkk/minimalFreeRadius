<script lang="ts">
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card/index.js';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import { Server, Users, Activity } from 'lucide-svelte';

	export let data: { nas: number; users: number; db: string; healthy: boolean };
</script>

<h1 class="mb-1 text-2xl font-bold">Dashboard</h1>
<p class="mb-6 text-sm text-muted-foreground">Ringkasan FreeRADIUS.</p>

<div class="grid gap-4 md:grid-cols-3">
	<Card>
		<CardHeader><CardTitle class="flex items-center gap-2 text-sm"><Server class="h-4 w-4" /> Total NAS</CardTitle></CardHeader>
		<CardContent><p class="text-3xl font-bold">{data.nas}</p></CardContent>
	</Card>
	<Card>
		<CardHeader><CardTitle class="flex items-center gap-2 text-sm"><Users class="h-4 w-4" /> Total Users</CardTitle></CardHeader>
		<CardContent><p class="text-3xl font-bold">{data.users}</p></CardContent>
	</Card>
	<Card>
		<CardHeader><CardTitle class="flex items-center gap-2 text-sm"><Activity class="h-4 w-4" /> Database</CardTitle></CardHeader>
		<CardContent>
			<Badge variant={data.healthy ? 'default' : 'destructive'}>{data.db}</Badge>
		</CardContent>
	</Card>
</div>

<Card class="mt-6">
	<CardHeader>
		<CardTitle>Kelola</CardTitle>
		<CardDescription>CRUD lewat proxy /api/* dengan cookie httpOnly.</CardDescription>
	</CardHeader>
	<CardContent class="flex gap-3">
		<a href="/dashboard/nas" class="rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground">Kelola NAS</a>
		<a href="/dashboard/users" class="rounded-md border px-4 py-2 text-sm">Kelola Users</a>
	</CardContent>
</Card>
