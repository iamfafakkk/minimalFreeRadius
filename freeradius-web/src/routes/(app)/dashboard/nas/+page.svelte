<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type NAS } from '$lib/api.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { Plus, Search, Trash2, LoaderCircle } from '@lucide/svelte';

	let items = $state<NAS[]>([]);
	let total = $state(0);
	let search = $state('');
	let loading = $state(true);
	let form = $state({ name: '', ip: '', secret: '', description: '' });
	let msg = $state('');
	let err = $state('');
	let confirmOpen = $state(false);
	let pending = $state<NAS | null>(null);

	async function load() {
		err = '';
		loading = true;
		const r = await api.get<NAS[]>(`/v1/nas/?page=1&limit=20&search=${encodeURIComponent(search)}`);
		loading = false;
		if (!r.ok) {
			err = r.message ?? 'Failed to load NAS.';
			return;
		}
		items = (r.data as NAS[]) ?? [];
		total = items.length;
	}

	onMount(load);

	async function create() {
		msg = ''; err = '';
		const r = await api.post('/v1/nas/', { ...form, type: 'other', ports: 1812 });
		if (!r.ok) {
			err = r.errors?.map((e) => e.message).join(', ') ?? r.message ?? 'Failed to add NAS.';
			return;
		}
		msg = 'NAS added.';
		form = { name: '', ip: '', secret: '', description: '' };
		await load();
	}

	function askRemove(n: NAS) {
		pending = n;
		confirmOpen = true;
	}

	async function doRemove() {
		if (!pending) return;
		const id = pending.id;
		confirmOpen = false;
		const r = await api.del(`/v1/nas/${id}`);
		if (!r.ok) { err = r.message ?? 'Failed to delete.'; return; }
		msg = 'NAS deleted.';
		await load();
	}
</script>

<div>
	<h1 class="text-2xl font-bold tracking-tight">NAS</h1>
	<p class="text-muted-foreground text-sm">Network Access Server — {total} entries.</p>
</div>

{#if msg}
	<p class="rounded-md border px-3 py-2 text-sm">{msg}</p>
{/if}
{#if err}
	<p class="border-destructive/30 bg-destructive/10 text-destructive rounded-md border px-3 py-2 text-sm">{err}</p>
{/if}

<div class="grid gap-4 lg:grid-cols-3">
	<Card.Root class="lg:col-span-1">
		<Card.Header>
			<Card.Title>Add NAS</Card.Title>
			<Card.Description>Via POST /api/v1/nas/.</Card.Description>
		</Card.Header>
		<Card.Content>
			<form class="space-y-3" onsubmit={(e) => { e.preventDefault(); create(); }}>
				<div class="space-y-2">
					<Label for="name">Name</Label>
					<Input id="name" bind:value={form.name} required minlength={3} placeholder="e.g. core-router" />
				</div>
				<div class="space-y-2">
					<Label for="ip">IP</Label>
					<Input id="ip" bind:value={form.ip} required placeholder="192.168.1.1" />
				</div>
				<div class="space-y-2">
					<Label for="secret">Secret</Label>
					<Input id="secret" type="password" bind:value={form.secret} required minlength={8} placeholder="Shared secret (min. 8 chars)" />
				</div>
				<div class="space-y-2">
					<Label for="desc">Description</Label>
					<Input id="desc" bind:value={form.description} placeholder="e.g. Main router" />
				</div>
				<Button type="submit" class="w-full"><Plus /> Add</Button>
			</form>
		</Card.Content>
	</Card.Root>

	<Card.Root class="lg:col-span-2">
		<Card.Header>
			<Card.Title>NAS List</Card.Title>
			<Card.Description>
				<form
					class="mt-2 flex gap-2"
					onsubmit={(e) => { e.preventDefault(); load(); }}
				>
					<Input bind:value={search} placeholder="Search name / IP..." />
					<Button type="submit" variant="secondary"><Search /> Search</Button>
				</form>
			</Card.Description>
		</Card.Header>
		<Card.Content>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>ID</Table.Head>
						<Table.Head>Name</Table.Head>
						<Table.Head>IP</Table.Head>
						<Table.Head>Type</Table.Head>
						<Table.Head class="text-right">Actions</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each items as n (n.id)}
						<Table.Row>
							<Table.Cell>{n.id}</Table.Cell>
							<Table.Cell class="font-medium">{n.name}</Table.Cell>
							<Table.Cell>{n.ip}</Table.Cell>
							<Table.Cell><Badge variant="secondary">{n.type}</Badge></Table.Cell>
							<Table.Cell class="text-right">
								<Button size="sm" variant="destructive" onclick={() => askRemove(n)}>
									<Trash2 /> Delete
								</Button>
							</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row>
							<Table.Cell colspan={5} class="text-muted-foreground h-24 text-center">
								{#if loading}<LoaderCircle class="mx-auto size-4 animate-spin" />{:else}No data yet.{/if}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Content>
	</Card.Root>
</div>

<AlertDialog.Root bind:open={confirmOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete NAS?</AlertDialog.Title>
			<AlertDialog.Description>
				NAS <span class="text-foreground font-medium">{pending?.name}</span> ({pending?.ip}) will be permanently deleted.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={doRemove}>Delete</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
