<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type RadiusUser } from '$lib/api.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { Plus, Search, Trash2, LoaderCircle } from '@lucide/svelte';

	let items = $state<RadiusUser[]>([]);
	let total = $state(0);
	let search = $state('');
	let loading = $state(true);
	let form = $state({ user: '', password: '', profile: 'PPP' });
	let msg = $state('');
	let err = $state('');
	let confirmOpen = $state(false);
	let pending = $state<RadiusUser | null>(null);

	async function load() {
		err = '';
		loading = true;
		const r = await api.get<RadiusUser[]>(`/v1/users/?page=1&limit=20&search=${encodeURIComponent(search)}`);
		loading = false;
		if (!r.ok) {
			err = r.message ?? 'Failed to load users.';
			return;
		}
		items = (r.data as RadiusUser[]) ?? [];
		total = items.length;
	}

	onMount(load);

	async function create() {
		msg = ''; err = '';
		const r = await api.post('/v1/users/', form);
		if (!r.ok) {
			err = r.errors?.map((e) => e.message).join(', ') ?? r.message ?? 'Failed to add user.';
			return;
		}
		msg = 'User added.';
		form = { user: '', password: '', profile: 'PPP' };
		await load();
	}

	function askRemove(u: RadiusUser) {
		pending = u;
		confirmOpen = true;
	}

	async function doRemove() {
		if (!pending) return;
		const username = pending.user;
		confirmOpen = false;
		const r = await api.del(`/v1/users/${encodeURIComponent(username)}`);
		if (!r.ok) { err = r.message ?? 'Failed to delete.'; return; }
		msg = 'User deleted.';
		await load();
	}
</script>

<div>
	<h1 class="text-2xl font-bold tracking-tight">Users</h1>
	<p class="text-muted-foreground text-sm">radcheck / radreply — {total} users.</p>
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
			<Card.Title>Add User</Card.Title>
			<Card.Description>Via POST /api/v1/users/.</Card.Description>
		</Card.Header>
		<Card.Content>
			<form class="space-y-3" onsubmit={(e) => { e.preventDefault(); create(); }}>
				<div class="space-y-2">
					<Label for="user">Username</Label>
					<Input id="user" bind:value={form.user} required minlength={6} placeholder="e.g. johndoe (min. 6 chars)" />
				</div>
				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input id="password" type="password" bind:value={form.password} required minlength={6} placeholder="Min. 6 characters" />
				</div>
				<div class="space-y-2">
					<Label for="profile">Profile</Label>
					<Input id="profile" bind:value={form.profile} placeholder="e.g. PPP" />
				</div>
				<Button type="submit" class="w-full"><Plus /> Add</Button>
			</form>
		</Card.Content>
	</Card.Root>

	<Card.Root class="lg:col-span-2">
		<Card.Header>
			<Card.Title>User List</Card.Title>
			<Card.Description>
				<form
					class="mt-2 flex gap-2"
					onsubmit={(e) => { e.preventDefault(); load(); }}
				>
					<Input bind:value={search} placeholder="Search username..." />
					<Button type="submit" variant="secondary"><Search /> Search</Button>
				</form>
			</Card.Description>
		</Card.Header>
		<Card.Content>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>ID</Table.Head>
						<Table.Head>User</Table.Head>
						<Table.Head>Profile</Table.Head>
						<Table.Head class="text-right">Actions</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each items as u (u.id)}
						<Table.Row>
							<Table.Cell>{u.id}</Table.Cell>
							<Table.Cell class="font-medium">{u.user}</Table.Cell>
							<Table.Cell><Badge variant="secondary">{u.profile ?? '-'}</Badge></Table.Cell>
							<Table.Cell class="text-right">
								<Button size="sm" variant="destructive" onclick={() => askRemove(u)}>
									<Trash2 /> Delete
								</Button>
							</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row>
							<Table.Cell colspan={4} class="text-muted-foreground h-24 text-center">
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
			<AlertDialog.Title>Delete User?</AlertDialog.Title>
			<AlertDialog.Description>
				User <span class="text-foreground font-medium">{pending?.user}</span> will be permanently deleted.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={doRemove}>Delete</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
