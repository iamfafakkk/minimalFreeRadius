<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type RadiusUser } from '$lib/api.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { Plus, Search, Trash2, LoaderCircle, SquarePen } from '@lucide/svelte';

	interface UserForm {
		user: string;
		password: string;
		profile: string;
	}

	let items = $state<RadiusUser[]>([]);
	let search = $state('');
	let profileFilter = $state('');
	let loading = $state(true);

	let msg = $state('');
	let err = $state('');

	// create / edit dialog
	let formOpen = $state(false);
	let editing = $state<RadiusUser | null>(null);
	let form = $state<UserForm>({ user: '', password: '', profile: '' });
	let saving = $state(false);

	// delete dialog
	let confirmOpen = $state(false);
	let pending = $state<RadiusUser | null>(null);

	async function load() {
		err = '';
		loading = true;
		const params = new URLSearchParams({ page: '1', limit: '50', search, profile: profileFilter });
		const r = await api.get<RadiusUser[]>(`/v1/users/?${params.toString()}`);
		loading = false;
		if (!r.ok) {
			err = r.message ?? 'Failed to load users.';
			return;
		}
		items = (r.data as RadiusUser[]) ?? [];
	}

	onMount(load);

	function openCreate() {
		editing = null;
		form = { user: '', password: '', profile: '' };
		msg = '';
		err = '';
		formOpen = true;
	}

	function openEdit(u: RadiusUser) {
		editing = u;
		form = { user: u.user, password: '', profile: u.profile ?? '' };
		msg = '';
		err = '';
		formOpen = true;
	}

	async function save() {
		msg = '';
		err = '';
		saving = true;
		let r;
		if (editing) {
			const payload: { password?: string; profile: string } = { profile: form.profile };
			if (form.password) payload.password = form.password;
			r = await api.put(`/v1/users/${encodeURIComponent(editing.user)}`, payload);
		} else {
			r = await api.post('/v1/users/', { user: form.user, password: form.password, profile: form.profile });
		}
		saving = false;
		if (!r.ok) {
			err = r.errors?.map((e) => e.message).join(', ') ?? r.message ?? 'Failed to save user.';
			return;
		}
		msg = editing ? 'User updated.' : 'User added.';
		formOpen = false;
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
		if (!r.ok) {
			err = r.message ?? 'Failed to delete.';
			return;
		}
		msg = 'User deleted.';
		await load();
	}
</script>

<div class="flex flex-wrap items-end justify-between gap-3">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">Users</h1>
		<p class="text-muted-foreground text-sm">
			PPPoE users — radcheck + Mikrotik-Group only. {items.length} entries.
		</p>
	</div>
	<Button onclick={openCreate}><Plus /> Add User</Button>
</div>

{#if msg}
	<p class="rounded-md border px-3 py-2 text-sm">{msg}</p>
{/if}
{#if err}
	<p class="border-destructive/30 bg-destructive/10 text-destructive rounded-md border px-3 py-2 text-sm">{err}</p>
{/if}

<Card.Root>
	<Card.Header>
		<div class="flex flex-wrap items-end justify-between gap-3">
			<Card.Title>User List</Card.Title>
			<form
				class="flex flex-wrap items-end gap-2"
				onsubmit={(e) => {
					e.preventDefault();
					load();
				}}
			>
				<Input bind:value={search} placeholder="Search username..." class="w-56" />
				<Input bind:value={profileFilter} placeholder="Filter by profile..." class="w-56" />
				<Button type="submit" variant="secondary"><Search /> Filter</Button>
			</form>
		</div>
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
						<Table.Cell class="font-medium break-all">{u.user}</Table.Cell>
						<Table.Cell>
							{#if u.profile}<Badge variant="secondary">{u.profile}</Badge>{:else}<span class="text-muted-foreground">—</span>{/if}
						</Table.Cell>
						<Table.Cell>
							<div class="flex justify-end gap-2">
								<Button size="sm" variant="outline" onclick={() => openEdit(u)}><SquarePen /> Edit</Button>
								<Button size="sm" variant="destructive" onclick={() => askRemove(u)}><Trash2 /> Delete</Button>
							</div>
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

<!-- Create / Edit -->
<Dialog.Root bind:open={formOpen}>
	<Dialog.Content class="max-w-lg">
		<Dialog.Header>
			<Dialog.Title>{editing ? 'Edit User' : 'Add User'}</Dialog.Title>
			<Dialog.Description>
				{editing
					? 'Change the password and/or the MikroTik PPP profile. Leave the password blank to keep it.'
					: 'Creates one radcheck row and one Mikrotik-Group row in radreply.'}
			</Dialog.Description>
		</Dialog.Header>
		<form
			class="grid gap-4"
			onsubmit={(e) => {
				e.preventDefault();
				save();
			}}
		>
			<div class="grid gap-2">
				<Label for="user">Username</Label>
				<Input
					id="user"
					bind:value={form.user}
					required
					disabled={!!editing}
					placeholder="e.g. olt-fajar.jb@dsnet"
				/>
				{#if !editing}
					<p class="text-muted-foreground text-xs">
						Full username, including "@" — it is not treated as a realm.
					</p>
				{/if}
			</div>
			<div class="grid gap-2">
				<Label for="password">Password</Label>
				<Input
					id="password"
					type="password"
					bind:value={form.password}
					required={!editing}
					placeholder={editing ? 'Leave blank to keep current password' : 'Password'}
				/>
			</div>
			<div class="grid gap-2">
				<Label for="profile">Profile</Label>
				<Input id="profile" bind:value={form.profile} placeholder="e.g. dsnet-10M" />
				<p class="text-muted-foreground text-xs">
					MikroTik PPP Profile name, sent as the Mikrotik-Group RADIUS attribute.
				</p>
			</div>
			<Dialog.Footer>
				<Button type="button" variant="outline" onclick={() => (formOpen = false)}>Cancel</Button>
				<Button type="submit" disabled={saving}>
					{#if saving}<LoaderCircle class="animate-spin" />{/if}
					{editing ? 'Save changes' : 'Add User'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- Delete -->
<AlertDialog.Root bind:open={confirmOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete User?</AlertDialog.Title>
			<AlertDialog.Description>
				User <span class="text-foreground font-medium break-all">{pending?.user}</span> and its Mikrotik-Group row will be
				permanently deleted.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={doRemove}>Delete</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
