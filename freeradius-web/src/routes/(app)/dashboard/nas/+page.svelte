<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type NAS } from '$lib/api.js';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import Table from '$lib/components/ui/table/table.svelte';
	import TableHeader from '$lib/components/ui/table/table-header.svelte';
	import TableBody from '$lib/components/ui/table/table-body.svelte';
	import TableRow from '$lib/components/ui/table/table-row.svelte';
	import TableHead from '$lib/components/ui/table/table-head.svelte';
	import TableCell from '$lib/components/ui/table/table-cell.svelte';

	let items: NAS[] = [];
	let total = 0;
	let search = '';
	let form = { name: '', ip: '', secret: '', description: '' };
	let msg = '';
	let err = '';

	async function load() {
		err = '';
		const r = await api.get<NAS[]>(`/v1/nas/?page=1&limit=20&search=${encodeURIComponent(search)}`);
		if (!r.ok) {
			err = r.message ?? 'Gagal memuat NAS.';
			return;
		}
		items = (r.data as NAS[]) ?? [];
		total = items.length;
	}

	onMount(load);

	async function create() {
		msg = ''; err = '';
		const r = await api.post('/v1/nas/', { ...form, type: 'other', ports: 1812 });
		if (!r.ok) { err = r.message ?? 'Gagal menambah NAS.'; return; }
		msg = 'NAS ditambah.';
		form = { name: '', ip: '', secret: '', description: '' };
		await load();
	}

	async function remove(id: number) {
		if (!confirm(`Hapus NAS #${id}?`)) return;
		const r = await api.del(`/v1/nas/${id}`);
		if (!r.ok) { err = r.message ?? 'Gagal menghapus.'; return; }
		msg = 'NAS dihapus.';
		await load();
	}
</script>

<h1 class="mb-1 text-2xl font-bold">NAS</h1>
<p class="mb-6 text-sm text-muted-foreground">Network Access Server — {total} entri.</p>

{#if msg}<p class="mb-4 rounded-md border px-3 py-2 text-sm">{msg}</p>{/if}
{#if err}<p class="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">{err}</p>{/if}

<div class="grid gap-4 lg:grid-cols-3">
	<Card class="lg:col-span-1">
		<CardHeader><CardTitle>Tambah NAS</CardTitle><CardDescription>Via POST /api/v1/nas/.</CardDescription></CardHeader>
		<CardContent>
			<form class="space-y-3" on:submit|preventDefault={create}>
				<div class="space-y-1"><Label for="name">Name</Label><Input id="name" bind:value={form.name} required minlength={3} /></div>
				<div class="space-y-1"><Label for="ip">IP</Label><Input id="ip" bind:value={form.ip} required placeholder="192.168.1.1" /></div>
				<div class="space-y-1"><Label for="secret">Secret</Label><Input id="secret" type="password" bind:value={form.secret} required minlength={8} /></div>
				<div class="space-y-1"><Label for="desc">Description</Label><Input id="desc" bind:value={form.description} /></div>
				<Button type="submit" class="w-full">Tambah</Button>
			</form>
		</CardContent>
	</Card>

	<Card class="lg:col-span-2">
		<CardHeader>
			<CardTitle>Daftar NAS</CardTitle>
			<CardDescription>
				<form class="mt-2 flex gap-2" on:submit|preventDefault={load}>
					<Input bind:value={search} placeholder="Cari name / ip..." />
					<Button type="submit" variant="secondary">Cari</Button>
				</form>
			</CardDescription>
		</CardHeader>
		<CardContent>
			<Table>
				<TableHeader><TableRow><TableHead>ID</TableHead><TableHead>Name</TableHead><TableHead>IP</TableHead><TableHead>Type</TableHead><TableHead>Aksi</TableHead></TableRow></TableHeader>
				<TableBody>
					{#each items as n (n.id)}
						<TableRow>
							<TableCell>{n.id}</TableCell>
							<TableCell class="font-medium">{n.name}</TableCell>
							<TableCell>{n.ip}</TableCell>
							<TableCell>{n.type}</TableCell>
							<TableCell><Button size="sm" variant="destructive" on:click={() => remove(n.id)}>Hapus</Button></TableCell>
						</TableRow>
					{:else}
						<TableRow><TableCell colspan={5} class="text-center text-muted-foreground">Belum ada data.</TableCell></TableRow>
					{/each}
				</TableBody>
			</Table>
		</CardContent>
	</Card>
</div>
