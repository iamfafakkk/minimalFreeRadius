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
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select/index.js';
	import { toast } from 'svelte-sonner';
	import { Plus, Search, Trash2, LoaderCircle, SquarePen } from '@lucide/svelte';

	interface UserForm {
		user: string;
		password: string;
		profile: string;
	}

	let items = $state<RadiusUser[]>([]);
	let profiles = $state<string[]>([]);
	let search = $state('');
	let profileFilter = $state('');
	let loading = $state(true);

	// create / edit dialog
	let formOpen = $state(false);
	let editing = $state<RadiusUser | null>(null);
	let form = $state<UserForm>({ user: '', password: '', profile: '' });
	let saving = $state(false);

	// profile picker: dropdown when profiles exist, free text when they don't
	let profileMode = $state<'select' | 'manual'>('manual');

	// delete dialog
	let confirmOpen = $state(false);
	let pending = $state<RadiusUser | null>(null);

	async function load() {
		loading = true;
		const params = new URLSearchParams({ page: '1', limit: '50', search, profile: profileFilter });
		const [r, pr] = await Promise.all([
			api.get<RadiusUser[]>(`/v1/users/?${params.toString()}`),
			api.get<string[]>('/v1/users/profiles')
		]);
		loading = false;
		if (!r.ok) {
			toast.error(r.message ?? 'Failed to load users.');
			return;
		}
		items = (r.data as RadiusUser[]) ?? [];
		profiles = pr.ok ? ((pr.data as string[]) ?? []) : [];
	}

	onMount(load);

	function openCreate() {
		editing = null;
		form = { user: '', password: '', profile: '' };
		profileMode = profiles.length ? 'select' : 'manual';
		formOpen = true;
	}

	function openEdit(u: RadiusUser) {
		editing = u;
		form = { user: u.user, password: '', profile: u.profile ?? '' };
		profileMode = profiles.length ? 'select' : 'manual';
		formOpen = true;
	}

	// Keeps a stored profile selectable even when it is not in the list.
	function profileOptions(extra: string) {
		return extra && !profiles.includes(extra) ? [...profiles, extra] : profiles;
	}

	async function save() {
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
			toast.error(r.errors?.map((e) => e.message).join(', ') ?? r.message ?? 'Failed to save user.');
			return;
		}
		toast.success(editing ? 'User updated.' : 'User added.');
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
			toast.error(r.message ?? 'Failed to delete.');
			return;
		}
		toast.success('User deleted.');
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
				{#if profileMode === 'select'}
					<Select type="single" value={form.profile} onValueChange={(v: string) => (form.profile = v)}>
						<SelectTrigger aria-label="Select a profile">{form.profile || 'Select a profile...'}</SelectTrigger>
						<SelectContent>
							{#each profileOptions(form.profile) as name (name)}
								<SelectItem value={name} label={name}>{name}</SelectItem>
							{/each}
						</SelectContent>
					</Select>
					<p class="text-muted-foreground text-xs">
						MikroTik PPP profile, sent as the Mikrotik-Group RADIUS attribute.
					</p>
					<Button
						type="button"
						variant="link"
						class="h-auto justify-start p-0 text-xs"
						onclick={() => (profileMode = 'manual')}
					>
						Type a profile name manually
					</Button>
				{:else}
					<Input id="profile" bind:value={form.profile} placeholder="e.g. dsnet-10M" />
					{#if profiles.length}
						<p class="text-muted-foreground text-xs">
							MikroTik PPP profile, sent as the Mikrotik-Group RADIUS attribute.
						</p>
						<Button
							type="button"
							variant="link"
							class="h-auto justify-start p-0 text-xs"
							onclick={() => (profileMode = 'select')}
						>
							Select from existing profiles
						</Button>
					{:else}
						<p class="text-muted-foreground text-xs">
							No profiles yet — type a name manually. Profiles appear here once a user has been given one.
						</p>
					{/if}
				{/if}
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
