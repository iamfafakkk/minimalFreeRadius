<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type RadiusUser } from '$lib/api.js';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import Table from '$lib/components/ui/table/table.svelte';
	import TableHeader from '$lib/components/ui/table/table-header.svelte';
	import TableBody from '$lib/components/ui/table/table-body.svelte';
	import TableRow from '$lib/components/ui/table/table-row.svelte';
	import TableHead from '$lib/components/ui/table/table-head.svelte';
	import TableCell from '$lib/components/ui/table/table-cell.svelte';

	let items: RadiusUser[] = [];
	let total = 0;
	let search = '';
	let form = { user: '', password: '', profile: 'PPP' };
	let msg = '';
	let err = '';

	async function load() {
		err = '';
		const r = await api.get<RadiusUser[]>(`/v1/users/?page=1&limit=20&search=${encodeURIComponent(search)}`);
		if (!r.ok) {
			err = r.message ?? 'Gagal memuat users.';
			return;
		}
		items = (r.data as RadiusUser[]) ?? [];
		total = items.length;
	}

	onMount(load);

	async function create() {
		msg = ''; err = '';
		const r = await api.post('/v1/users/', form);
		if (!r.ok) { err = r.errors?.map((e) => e.message).join(', ') ?? r.message ?? 'Gagal menambah user.'; return; }
		msg = 'User ditambah.';
		form = { user: '', password: '', profile: 'PPP' };
		await load();
	}

	async function remove(username: string) {
		if (!confirm(`Hapus user ${username}?`)) return;
		const r = await api.del(`/v1/users/${encodeURIComponent(username)}`);
		if (!r.ok) { err = r.message ?? 'Gagal menghapus.'; return; }
		msg = 'User dihapus.';
		await load();
	}
</script>

<h1 class="mb-1 text-2xl font-bold">Users</h1>
<p class="mb-6 text-sm text-muted-foreground">radcheck / radreply — {total} user.</p>

{#if msg}<p class="mb-4 rounded-md border px-3 py-2 text-sm">{msg}</p>{/if}
{#if err}<p class="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">{err}</p>{/if}

<div class="grid gap-4 lg:grid-cols-3">
	<Card class="lg:col-span-1">
		<CardHeader><CardTitle>Tambah User</CardTitle><CardDescription>Via POST /api/v1/users/.</CardDescription></CardHeader>
		<CardContent>
			<form class="space-y-3" on:submit|preventDefault={create}>
				<div class="space-y-1"><Label for="user">Username</Label><Input id="user" bind:value={form.user} required minlength={6} /></div>
				<div class="space-y-1"><Label for="password">Password</Label><Input id="password" type="password" bind:value={form.password} required minlength={6} /></div>
				<div class="space-y-1"><Label for="profile">Profile</Label><Input id="profile" bind:value={form.profile} /></div>
				<Button class="w-full">Tambah</Button>
			</form>
		</CardContent>
	</Card>

	<Card class="lg:col-span-2">
		<CardHeader>
			<CardTitle>Daftar User</CardTitle>
			<CardDescription>
				<form class="mt-2 flex gap-2" on:submit|preventDefault={load}>
					<Input bind:value={search} placeholder="Cari username..." />
					<Button variant="secondary">Cari</Button>
				</form>
			</CardDescription>
		</CardHeader>
		<CardContent>
			<Table>
				<TableHeader><TableRow><TableHead>ID</TableHead><TableHead>User</TableHead><TableHead>Profile</TableHead><TableHead>Aksi</TableHead></TableRow></TableHeader>
				<TableBody>
					{#each items as u (u.id)}
						<TableRow>
							<TableCell>{u.id}</TableCell>
							<TableCell class="font-medium">{u.user}</TableCell>
							<TableCell><Badge variant="secondary">{u.profile ?? '-'}</Badge></TableCell>
							<TableCell><Button size="sm" variant="destructive" on:click={() => remove(u.user)}>Hapus</Button></TableCell>
						</TableRow>
					{:else}
						<TableRow><TableCell colspan={4} class="text-center text-muted-foreground">Belum ada data.</TableCell></TableRow>
					{/each}
				</TableBody>
			</Table>
		</CardContent>
	</Card>
</div>
