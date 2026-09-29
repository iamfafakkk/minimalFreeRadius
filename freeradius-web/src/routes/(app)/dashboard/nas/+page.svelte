<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type NAS } from '$lib/api.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { toast } from 'svelte-sonner';
	import {
		Plus,
		Search,
		Trash2,
		LoaderCircle,
		SquarePen,
		FlaskConical,
		Unplug,
		PlugZap
	} from '@lucide/svelte';

	const TYPE_OPTIONS = [
		'other',
		'cisco',
		'juniper',
		'mikrotik',
		'computone',
		'livingston',
		'max40xx',
		'multitech',
		'netserver',
		'pathras',
		'patton',
		'portslave',
		'tc',
		'usrhiper'
	];

	type NasForm = {
		name: string;
		ip: string;
		secret: string;
		type: string;
		ports: number;
		description: string;
	};

	interface TestResult {
		action: string;
		target: string;
		status: string;
		message: string;
		reply?: string;
		latency_ms: number;
	}

	let items = $state<NAS[]>([]);
	let search = $state('');
	let loading = $state(true);

	// create / edit dialog
	let formOpen = $state(false);
	let editing = $state<NAS | null>(null);
	let form = $state<NasForm>({ name: '', ip: '', secret: '', type: 'other', ports: 3799, description: '' });
	let saving = $state(false);

	// test dialog
	let testOpen = $state(false);
	let testNas = $state<NAS | null>(null);
	let testAction = $state<'disconnect' | 'coa'>('disconnect');
	let testUser = $state('');
	let coaResult = $state<TestResult | null>(null);
	let coaBusy = $state(false);
	let authUser = $state('');
	let authPass = $state('');
	let authResult = $state<TestResult | null>(null);
	let authBusy = $state(false);

	// delete dialog
	let confirmOpen = $state(false);
	let pending = $state<NAS | null>(null);

	async function load() {
		loading = true;
		const r = await api.get<NAS[]>(`/v1/nas/?page=1&limit=50&search=${encodeURIComponent(search)}`);
		loading = false;
		if (!r.ok) {
			toast.error(r.message ?? 'Failed to load NAS.');
			return;
		}
		items = (r.data as NAS[]) ?? [];
	}

	onMount(load);

	function openCreate() {
		editing = null;
		form = { name: '', ip: '', secret: '', type: 'other', ports: 3799, description: '' };
		formOpen = true;
	}

	function openEdit(n: NAS) {
		editing = n;
		form = {
			name: n.name,
			ip: n.ip,
			secret: '',
			type: n.type || 'other',
			ports: n.ports ?? 3799,
			description: n.description ?? ''
		};
		formOpen = true;
	}

	async function save() {
		saving = true;
		const payload = {
			name: form.name,
			ip: form.ip,
			secret: form.secret,
			type: form.type,
			ports: form.ports,
			description: form.description
		};
		const r = editing ? await api.put(`/v1/nas/${editing.id}`, payload) : await api.post('/v1/nas/', payload);
		saving = false;
		if (!r.ok) {
			toast.error(r.errors?.map((e) => e.message).join(', ') ?? r.message ?? 'Failed to save NAS.');
			return;
		}
		toast.success(r.message ?? (editing ? 'NAS updated.' : 'NAS added.'));
		formOpen = false;
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
		if (!r.ok) {
			toast.error(r.message ?? 'Failed to delete.');
			return;
		}
		toast.success(r.message ?? 'NAS deleted.');
		await load();
	}

	function openTest(n: NAS) {
		testNas = n;
		testAction = 'disconnect';
		testUser = '';
		coaResult = null;
		authUser = '';
		authPass = '';
		authResult = null;
		testOpen = true;
	}

	async function runCoa() {
		if (!testNas) return;
		coaBusy = true;
		coaResult = null;
		const r = await api.post<TestResult>(`/v1/nas/${testNas.id}/test`, {
			action: testAction,
			username: testUser
		});
		coaBusy = false;
		const res: TestResult = r.ok && r.data ? (r.data as TestResult) : { action: testAction, target: '', status: 'error', message: r.message ?? 'Request failed.', latency_ms: 0 };
		coaResult = res;
		if (res.status === 'ack' || res.status === 'accept') toast.success(res.message);
		else toast.error(res.message);
	}

	async function runAuth() {
		if (!testNas) return;
		authBusy = true;
		authResult = null;
		const r = await api.post<TestResult>('/v1/nas/test-auth', {
			username: authUser,
			password: authPass,
			nas_id: testNas.id
		});
		authBusy = false;
		const res: TestResult = r.ok && r.data ? (r.data as TestResult) : { action: 'auth', target: '', status: 'error', message: r.message ?? 'Request failed.', latency_ms: 0 };
		authResult = res;
		if (res.status === 'accept') toast.success('Authentication accepted.');
		else toast.error(res.message);
	}

	function statusClass(s: string): string {
		switch (s) {
			case 'ack':
			case 'accept':
				return 'border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400';
			case 'nak':
			case 'reject':
			case 'unexpected':
			case 'error':
				return 'border-destructive/40 bg-destructive/10 text-destructive';
			case 'timeout':
				return 'border-amber-500/40 bg-amber-500/10 text-amber-600 dark:text-amber-400';
			default:
				return '';
		}
	}
</script>

<div class="flex flex-wrap items-end justify-between gap-3">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">NAS</h1>
		<p class="text-muted-foreground text-sm">MikroTik / RADIUS clients — {items.length} entries.</p>
	</div>
	<Button onclick={openCreate}><Plus /> Add NAS</Button>
</div>

<Card.Root>
	<Card.Header>
		<div class="flex flex-wrap items-center justify-between gap-3">
			<Card.Title>NAS List</Card.Title>
			<form class="flex gap-2" onsubmit={(e) => { e.preventDefault(); load(); }}>
				<Input bind:value={search} placeholder="Search short name / NAS IP..." class="w-64" />
				<Button type="submit" variant="secondary"><Search /> Search</Button>
			</form>
		</div>
	</Card.Header>
	<Card.Content>
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Short name</Table.Head>
					<Table.Head>NAS IP</Table.Head>
					<Table.Head>Type</Table.Head>
					<Table.Head>CoA Port</Table.Head>
					<Table.Head>Description</Table.Head>
					<Table.Head class="text-right">Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each items as n (n.id)}
					<Table.Row>
						<Table.Cell class="font-medium">{n.name}</Table.Cell>
						<Table.Cell class="font-mono text-xs">{n.ip}</Table.Cell>
						<Table.Cell><Badge variant="secondary">{n.type}</Badge></Table.Cell>
						<Table.Cell>{n.ports ?? 3799}</Table.Cell>
						<Table.Cell class="text-muted-foreground max-w-[240px] truncate">{n.description || '—'}</Table.Cell>
						<Table.Cell>
							<div class="flex justify-end gap-2">
								<Button size="sm" variant="secondary" onclick={() => openTest(n)}><FlaskConical /> Test</Button>
								<Button size="sm" variant="outline" onclick={() => openEdit(n)}><SquarePen /> Edit</Button>
								<Button size="sm" variant="destructive" onclick={() => askRemove(n)}><Trash2 /> Delete</Button>
							</div>
						</Table.Cell>
					</Table.Row>
				{:else}
					<Table.Row>
						<Table.Cell colspan={6} class="text-muted-foreground h-24 text-center">
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
			<Dialog.Title>{editing ? 'Edit NAS' : 'Add NAS'}</Dialog.Title>
			<Dialog.Description>
				{editing ? 'Leave the secret blank to keep the current one.' : 'Register a MikroTik / RADIUS client.'}
			</Dialog.Description>
		</Dialog.Header>
		<form class="grid gap-4" onsubmit={(e) => { e.preventDefault(); save(); }}>
			<div class="grid gap-2">
				<Label for="nas-shortname">Short name</Label>
				<Input id="nas-shortname" bind:value={form.name} required minlength={3} maxlength={30} placeholder="e.g. office-gw" />
			</div>
			<div class="grid gap-2">
				<Label for="nas-ip">NAS IP</Label>
				<Input id="nas-ip" bind:value={form.ip} required placeholder="e.g. 192.168.88.1" />
			</div>
			<div class="grid gap-2">
				<Label for="nas-secret">Secret</Label>
				<Input id="nas-secret" type="password" bind:value={form.secret} required={!editing} minlength={8} placeholder={editing ? 'Leave blank to keep current secret' : 'Shared secret (min. 8 chars)'} />
			</div>
			<div class="grid gap-4 sm:grid-cols-2">
				<div class="grid gap-2">
					<Label for="nas-type">Type</Label>
					<select
						id="nas-type"
						bind:value={form.type}
						class="border-input bg-background ring-offset-background focus:ring-ring flex h-10 w-full rounded-md border px-3 text-sm focus:outline-none focus:ring-2 focus:ring-offset-2"
					>
						{#each TYPE_OPTIONS as t (t)}<option value={t}>{t}</option>{/each}
					</select>
				</div>
				<div class="grid gap-2">
					<Label for="nas-port">CoA Port</Label>
					<Input id="nas-port" type="number" min={1} max={65535} bind:value={form.ports} placeholder="3799" />
				</div>
			</div>
			<div class="grid gap-2">
				<Label for="nas-desc">Description</Label>
				<Input id="nas-desc" bind:value={form.description} maxlength={200} placeholder="e.g. Main office router" />
			</div>
			<Dialog.Footer>
				<Button type="button" variant="outline" onclick={() => (formOpen = false)}>Cancel</Button>
				<Button type="submit" disabled={saving}>
					{#if saving}<LoaderCircle class="animate-spin" />{/if}
					{editing ? 'Save changes' : 'Add NAS'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- Test -->
<Dialog.Root bind:open={testOpen}>
	<Dialog.Content class="max-w-xl">
		<Dialog.Header>
			<Dialog.Title>Test NAS</Dialog.Title>
			<Dialog.Description>
				{testNas?.name} — <span class="font-mono text-xs">{testNas?.ip}</span>
			</Dialog.Description>
		</Dialog.Header>

		<div class="grid gap-3">
			<div class="flex items-center gap-2">
				<Badge variant="outline">CoA / Disconnect</Badge>
				<span class="text-muted-foreground text-xs">Sent to the NAS on its CoA port.</span>
			</div>
			<div class="grid gap-4 sm:grid-cols-[160px_1fr]">
				<select
					bind:value={testAction}
					aria-label="Request type"
					class="border-input bg-background ring-offset-background focus:ring-ring flex h-10 w-full rounded-md border px-3 text-sm focus:outline-none focus:ring-2 focus:ring-offset-2"
				>
					<option value="disconnect">Disconnect</option>
					<option value="coa">CoA</option>
				</select>
				<Input bind:value={testUser} placeholder="Username (optional)" />
			</div>
			<Button onclick={runCoa} disabled={coaBusy}>
				{#if coaBusy}<LoaderCircle class="animate-spin" />{:else}<Unplug />{/if}
				Run {testAction === 'coa' ? 'CoA' : 'Disconnect'} test
			</Button>
			{#if coaResult}
				<div class="rounded-md border px-3 py-2 text-sm {statusClass(coaResult.status)}">
					<p class="font-medium">{coaResult.message}</p>
					<p class="mt-1 text-xs opacity-80">
						status: {coaResult.status} · target: {coaResult.target} · {coaResult.latency_ms} ms
						{#if coaResult.reply}· reply: {coaResult.reply}{/if}
					</p>
				</div>
			{/if}
		</div>

		<Separator />

		<div class="grid gap-3">
			<div class="flex items-center gap-2">
				<Badge variant="outline">Auth test (radtest)</Badge>
				<span class="text-muted-foreground text-xs">Access-Request to the local RADIUS server.</span>
			</div>
			<div class="grid gap-2">
				<Input bind:value={authUser} placeholder="Username" />
				<Input type="password" bind:value={authPass} placeholder="Password" />
			</div>
			<Button variant="secondary" onclick={runAuth} disabled={authBusy}>
				{#if authBusy}<LoaderCircle class="animate-spin" />{:else}<PlugZap />{/if}
				Run auth test
			</Button>
			{#if authResult}
				<div class="rounded-md border px-3 py-2 text-sm {statusClass(authResult.status)}">
					<p class="font-medium">{authResult.message}</p>
					<p class="mt-1 text-xs opacity-80">
						status: {authResult.status} · target: {authResult.target} · {authResult.latency_ms} ms
						{#if authResult.reply}· reply: {authResult.reply}{/if}
					</p>
				</div>
			{/if}
		</div>

		<Dialog.Footer>
			<Button type="button" variant="outline" onclick={() => (testOpen = false)}>Close</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Delete -->
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
