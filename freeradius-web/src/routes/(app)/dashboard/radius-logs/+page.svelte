<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api, openRadiusLogStream, type RadiusLogLine, type RadiusLogType } from '$lib/api.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import {
		Select,
		SelectContent,
		SelectItem,
		SelectTrigger
	} from '$lib/components/ui/select/index.js';
	import { Pause, Play, Trash2, ArrowDownToLine, Eraser, LoaderCircle } from '@lucide/svelte';

	const MAX_LINES = 1000;

	let lines = $state<RadiusLogLine[]>([]);
	let connected = $state(false);
	let paused = $state(false);
	let autoscroll = $state(true);
	let search = $state('');
	let typeFilter = $state('all');
	let levelFilter = $state('all');
	let viewport: HTMLDivElement | undefined = $state();
	let es: EventSource | undefined;
	let pausedBuf: RadiusLogLine[] = [];

	const TYPE_OPTIONS: { value: string; label: string }[] = [
		{ value: 'all', label: 'All types' },
		{ value: 'AUTH_OK', label: 'Auth OK' },
		{ value: 'AUTH_FAIL', label: 'Auth Fail' },
		{ value: 'ACCT', label: 'Accounting' },
		{ value: 'SYS', label: 'System' }
	];
	const LEVEL_OPTIONS: { value: string; label: string }[] = [
		{ value: 'all', label: 'All levels' },
		{ value: 'Info', label: 'Info' },
		{ value: 'Warning', label: 'Warning' },
		{ value: 'Error', label: 'Error' },
		{ value: 'Debug', label: 'Debug' }
	];

	const filtered = $derived.by(() => {
		const q = search.trim().toLowerCase();
		return lines.filter((l) => {
			if (typeFilter !== 'all' && l.type !== typeFilter) return false;
			if (levelFilter !== 'all' && l.level !== levelFilter) return false;
			if (!q) return true;
			return (l.message + ' ' + l.user + ' ' + l.nas + ' ' + l.raw).toLowerCase().includes(q);
		});
	});

	const counts = $derived({
		authOk: lines.filter((l) => l.type === 'AUTH_OK').length,
		authFail: lines.filter((l) => l.type === 'AUTH_FAIL').length,
		acct: lines.filter((l) => l.type === 'ACCT').length,
		nas: new Set(lines.map((l) => l.nas).filter(Boolean)).size
	});

	function push(l: RadiusLogLine) {
		if (paused) {
			pausedBuf.push(l);
			if (pausedBuf.length > MAX_LINES) pausedBuf = pausedBuf.slice(-MAX_LINES);
			return;
		}
		lines.push(l);
		if (lines.length > MAX_LINES) lines = lines.slice(-MAX_LINES);
		if (autoscroll) scrollToBottom();
	}

	function scrollToBottom() {
		queueMicrotask(() => {
			if (viewport) viewport.scrollTop = viewport.scrollHeight;
		});
	}

	onMount(() => {
		es = openRadiusLogStream();
		es.addEventListener('open', () => (connected = true));
		es.addEventListener('log', (e) => {
			try {
				push(JSON.parse((e as MessageEvent).data) as RadiusLogLine);
			} catch {
				/* ignore malformed frame */
			}
		});
		es.addEventListener('error', () => (connected = false));
	});

	onDestroy(() => es?.close());

	function togglePause() {
		paused = !paused;
		if (!paused && pausedBuf.length) {
			lines = [...lines, ...pausedBuf].slice(-MAX_LINES);
			pausedBuf = [];
			scrollToBottom();
		}
	}

	function clear() {
		lines = [];
		pausedBuf = [];
	}

	// Confirm dialog: emptying the buffer is local, truncating the file is not.
	let confirmOpen = $state(false);
	let clearing = $state(false);
	let actionMsg = $state('');
	let actionErr = $state('');

	async function doClear() {
		confirmOpen = false;
		clearing = true;
		actionMsg = '';
		actionErr = '';
		const r = await api.del('/v1/radius/log');
		clearing = false;
		if (!r.ok) {
			actionErr = r.message ?? 'Failed to clear the RADIUS log.';
			return;
		}
		lines = [];
		pausedBuf = [];
		actionMsg = 'RADIUS log cleared.';
	}

	const typeClass: Record<RadiusLogType, string> = {
		AUTH_OK: 'bg-emerald-500/15 text-emerald-500 border-emerald-500/20',
		AUTH_FAIL: 'bg-red-500/15 text-red-500 border-red-500/20',
		ACCT: 'bg-sky-500/15 text-sky-500 border-sky-500/20',
		SYS: 'bg-muted text-muted-foreground border-border'
	};
	const statusClass: Record<string, string> = {
		ok: 'text-emerald-500',
		fail: 'text-red-500',
		acct: 'text-sky-500',
		sys: 'text-muted-foreground'
	};
	const levelClass: Record<string, string> = {
		Error: 'text-red-500',
		Warning: 'text-amber-500',
		Debug: 'text-sky-400',
		Info: 'text-emerald-500'
	};
</script>

<div class="flex flex-wrap items-start justify-between gap-4">
	<div>
		<h1 class="flex items-center gap-3 text-2xl font-bold tracking-tight">
			Radius Logs
			<Badge variant={connected ? 'default' : 'destructive'} class="gap-1.5">
				<span class="size-2 animate-pulse rounded-full {connected
					? 'bg-emerald-300'
					: 'bg-red-300'}"></span>
				{connected ? 'Live' : 'Disconnected'}
			</Badge>
		</h1>
		<p class="text-muted-foreground text-sm">
			Realtime FreeRADIUS server log — {filtered.length} of {lines.length} events.
		</p>
	</div>
	<div class="flex items-center gap-2">
		<Button variant="outline" size="sm" onclick={togglePause}>
			{#if paused}<Play /> Resume{:else}<Pause /> Pause{/if}
		</Button>
		<Button
			variant="outline"
			size="sm"
			onclick={() => (autoscroll = !autoscroll)}
			class={autoscroll ? 'border-primary text-primary' : ''}
		>
			<ArrowDownToLine /> Autoscroll
		</Button>
		<Button variant="outline" size="sm" onclick={clear} title="Clear the panel buffer only"
			><Trash2 /> Clear</Button
		>
		<Button
			variant="destructive"
			size="sm"
			onclick={() => (confirmOpen = true)}
			disabled={clearing}
			title="Truncate the FreeRADIUS log file on the server"
		>
			{#if clearing}<LoaderCircle class="animate-spin" />{:else}<Eraser />{/if} Clear Log File
		</Button>
	</div>
</div>

{#if actionMsg}
	<p class="rounded-md border px-3 py-2 text-sm">{actionMsg}</p>
{/if}
{#if actionErr}
	<p class="border-destructive/30 bg-destructive/10 text-destructive rounded-md border px-3 py-2 text-sm">{actionErr}</p>
{/if}

<div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
	<Card.Root>
		<Card.Content class="p-4">
			<p class="text-muted-foreground text-xs font-medium tracking-wide uppercase">Auth OK</p>
			<p class="text-2xl font-bold text-emerald-500">{counts.authOk}</p>
		</Card.Content>
	</Card.Root>
	<Card.Root>
		<Card.Content class="p-4">
			<p class="text-muted-foreground text-xs font-medium tracking-wide uppercase">Auth Fail</p>
			<p class="text-2xl font-bold {counts.authFail > 0 ? 'text-red-500' : ''}">
				{counts.authFail}
			</p>
		</Card.Content>
	</Card.Root>
	<Card.Root>
		<Card.Content class="p-4">
			<p class="text-muted-foreground text-xs font-medium tracking-wide uppercase">Accounting</p>
			<p class="text-2xl font-bold text-sky-500">{counts.acct}</p>
		</Card.Content>
	</Card.Root>
	<Card.Root>
		<Card.Content class="p-4">
			<p class="text-muted-foreground text-xs font-medium tracking-wide uppercase">Active NAS</p>
			<p class="text-2xl font-bold">{counts.nas}</p>
		</Card.Content>
	</Card.Root>
</div>

<Tabs.Root value="table" class="gap-3">
	<div class="flex flex-wrap items-center justify-between gap-3">
		<Tabs.List>
			<Tabs.Trigger value="table">Table</Tabs.Trigger>
			<Tabs.Trigger value="raw">Raw</Tabs.Trigger>
		</Tabs.List>
		<div class="flex flex-wrap items-center gap-2">
			<Input bind:value={search} placeholder="Search user, NAS, message..." class="w-64" />
			<Select type="single" bind:value={typeFilter}>
				<SelectTrigger class="w-[160px]" aria-label="Filter by type">
					{TYPE_OPTIONS.find((o) => o.value === typeFilter)?.label}
				</SelectTrigger>
				<SelectContent>
					{#each TYPE_OPTIONS as o (o.value)}
						<SelectItem value={o.value} label={o.label}>{o.label}</SelectItem>
					{/each}
				</SelectContent>
			</Select>
			<Select type="single" bind:value={levelFilter}>
				<SelectTrigger class="w-[150px]" aria-label="Filter by level">
					{LEVEL_OPTIONS.find((o) => o.value === levelFilter)?.label}
				</SelectTrigger>
				<SelectContent>
					{#each LEVEL_OPTIONS as o (o.value)}
						<SelectItem value={o.value} label={o.label}>{o.label}</SelectItem>
					{/each}
				</SelectContent>
			</Select>
		</div>
	</div>

	<Tabs.Content value="table" class="outline-none">
		<Card.Root class="overflow-hidden">
			<Card.Content class="p-0">
				<div
					bind:this={viewport}
					class="relative max-h-[34rem] overflow-auto"
				>
					<Table.Root class="overflow-visible">
						<Table.Header class="bg-muted sticky top-0 z-10 [&_tr]:border-b">
							<Table.Row>
								<Table.Head class="w-[190px] text-xs tracking-wider uppercase">Time</Table.Head>
								<Table.Head class="w-[90px] text-xs tracking-wider uppercase">Type</Table.Head>
								<Table.Head class="w-[130px] text-xs tracking-wider uppercase">NAS</Table.Head>
								<Table.Head class="w-[150px] text-xs tracking-wider uppercase">User</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Message</Table.Head>
								<Table.Head class="w-[70px] text-xs tracking-wider uppercase">Status</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each filtered as l, i (i)}
								<Table.Row>
									<Table.Cell class="text-muted-foreground font-mono text-xs whitespace-nowrap"
										>{l.time}</Table.Cell
									>
									<Table.Cell>
										<Badge variant="outline" class="font-mono text-[10px] {typeClass[l.type]}">
											{l.type}
										</Badge>
									</Table.Cell>
									<Table.Cell class="font-mono text-xs">{l.nas || '—'}</Table.Cell>
									<Table.Cell class="font-mono text-xs">{l.user || '—'}</Table.Cell>
									<Table.Cell class="max-w-[28rem] text-sm">
										<span class="line-clamp-2">{l.message}</span>
									</Table.Cell>
									<Table.Cell class="font-mono text-xs uppercase {statusClass[l.status] ?? ''}"
										>{l.status}</Table.Cell
									>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={6} class="text-muted-foreground h-32 text-center text-sm">
										No logs match the filters — waiting for events...
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</Card.Content>
		</Card.Root>
	</Tabs.Content>

	<Tabs.Content value="raw" class="outline-none">
		<Card.Root>
			<Card.Content>
				<div
					class="bg-muted/30 h-[32rem] overflow-auto rounded-md border p-3 font-mono text-xs leading-relaxed"
				>
					{#if filtered.length === 0}
						<p class="text-muted-foreground">No log lines.</p>
					{:else}
						{#each filtered as l, i (i)}
							<div class="whitespace-pre-wrap">
								<span class="text-muted-foreground">{l.time}</span>
								{#if l.level}<span class={levelClass[l.level] ?? ''}> [{l.level}]</span>{/if}
								<span> {l.message}</span>
							</div>
						{/each}
					{/if}
				</div>
			</Card.Content>
		</Card.Root>
	</Tabs.Content>
</Tabs.Root>

<!-- Clear Log File -->
<AlertDialog.Root bind:open={confirmOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Clear the RADIUS log file?</AlertDialog.Title>
			<AlertDialog.Description>
				<span class="text-foreground font-mono text-xs">/var/log/freeradius/radius.log</span> will be truncated on the
				server, and all buffered lines in this panel are dropped. This cannot be undone.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={doClear}>Clear Log File</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
