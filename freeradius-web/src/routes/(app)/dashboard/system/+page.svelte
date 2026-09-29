<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api, type SystemHealth } from '$lib/api.js';
	import { toast } from 'svelte-sonner';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import {
		Activity,
		Server,
		Cpu,
		MemoryStick,
		HardDrive,
		Database,
		Clock,
		RefreshCw,
		LoaderCircle,
		RotateCw
	} from '@lucide/svelte';

	let health = $state<SystemHealth | null>(null);
	let loading = $state(true);
	let failing = $state(false);
	let updatedAt = $state<Date | null>(null);
	let timer: ReturnType<typeof setInterval> | undefined;

	// restart FreeRADIUS
	let restartOpen = $state(false);
	let restarting = $state(false);

	const REFRESH_MS = 5000;

	async function load() {
		const r = await api.get<SystemHealth>('/v1/system/health');
		loading = false;
		if (!r.ok) {
			// Only announce the transition into failure, not every 5s poll.
			if (!failing) {
				failing = true;
				toast.error(r.message ?? 'Failed to load system health.');
			}
			return;
		}
		if (failing) {
			failing = false;
			toast.success('System health restored.');
		}
		health = r.data as SystemHealth;
		updatedAt = new Date();
	}

	onMount(() => {
		load();
		timer = setInterval(load, REFRESH_MS);
		return () => clearInterval(timer);
	});
	onDestroy(() => clearInterval(timer));

	function pct(used: number, total: number): number {
		return total > 0 ? Math.min(100, Math.round((used / total) * 100)) : 0;
	}

	function bytes(n: number): string {
		if (!n) return '—';
		const units = ['B', 'KB', 'MB', 'GB', 'TB'];
		let i = 0;
		let v = n;
		while (v >= 1024 && i < units.length - 1) {
			v /= 1024;
			i++;
		}
		return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
	}

	function uptime(seconds: number): string {
		if (!seconds) return '—';
		const d = Math.floor(seconds / 86400);
		const h = Math.floor((seconds % 86400) / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		if (d) return `${d}d ${h}h ${m}m`;
		if (h) return `${h}h ${m}m`;
		return `${m}m ${Math.floor(seconds % 60)}s`;
	}

	const ramPct = $derived(health ? pct(health.host.ram_used_bytes, health.host.ram_total_bytes) : 0);
	const diskPct = $derived(health ? pct(health.host.disk_used_bytes, health.host.disk_total_bytes) : 0);
	const dbOk = $derived(health?.api.database === 'connected');
	const radiusOk = $derived(health?.radius.status === 'running');

	async function doRestart() {
		restartOpen = false;
		restarting = true;
		const r = await api.post('/v1/system/radius/restart');
		restarting = false;
		if (!r.ok) {
			const out = (r.data as { output?: string } | undefined)?.output;
			toast.error([r.message ?? 'Failed to restart FreeRADIUS.', out].filter(Boolean).join(' — '));
			return;
		}
		toast.success('FreeRADIUS restarted.');
		await load();
	}
</script>

<div class="flex flex-wrap items-end justify-between gap-3">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">System Health</h1>
		<p class="text-muted-foreground text-sm">
			Server metrics for the panel host and the FreeRADIUS service.
		</p>
	</div>
	<div class="flex items-center gap-2">
		{#if updatedAt}
			<span class="text-muted-foreground text-xs">
				Updated {updatedAt.toLocaleTimeString()} · every {REFRESH_MS / 1000}s
			</span>
		{/if}
		<Button variant="outline" size="sm" onclick={() => load()} disabled={loading}>
			<RefreshCw class={loading ? 'animate-spin' : ''} /> Refresh
		</Button>
	</div>
</div>

{#if loading && !health}
	<div class="text-muted-foreground flex items-center gap-2 text-sm">
		<LoaderCircle class="size-4 animate-spin" /> Loading system health…
	</div>
{:else if health}
	<!-- Services -->
	<div class="grid gap-4 md:grid-cols-3">
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2 text-sm"><Activity class="size-4" /> API</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-2">
				<Badge variant="default">running</Badge>
				<p class="text-muted-foreground text-xs">Uptime {uptime(health.api.uptime_seconds)}</p>
				<p class="text-muted-foreground text-xs">{health.api.go_version}</p>
			</Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2 text-sm"><Server class="size-4" /> FreeRADIUS</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-2">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<Badge variant={radiusOk ? 'default' : 'destructive'}>{health.radius.status}</Badge>
					<Button
						variant="outline"
						size="sm"
						onclick={() => (restartOpen = true)}
						disabled={restarting}
					>
						{#if restarting}<LoaderCircle class="animate-spin" />{:else}<RotateCw />{/if} Restart
					</Button>
				</div>
				<p class="text-muted-foreground text-xs">Auth {health.radius.test_addr}</p>
				<p class="text-muted-foreground truncate text-xs" title={health.radius.log_path}>{health.radius.log_path}</p>
			</Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2 text-sm"><Database class="size-4" /> Database</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-2">
				<Badge variant={dbOk ? 'default' : 'destructive'}>{health.api.database}</Badge>
				<p class="text-muted-foreground truncate text-xs" title={health.appdb.path}>App DB: {health.appdb.path}</p>
				<p class="text-muted-foreground text-xs">
					{health.appdb.counts.users ?? 0} admin users · {health.appdb.counts.logins ?? 0} logins
				</p>
			</Card.Content>
		</Card.Root>
	</div>

	<!-- Host -->
	<div class="grid gap-4 md:grid-cols-3">
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2 text-sm"><Cpu class="size-4" /> CPU Load</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-1">
				<p class="text-3xl font-bold">{health.host.load1.toFixed(2)}</p>
				<p class="text-muted-foreground text-xs">
					1 min · 5m {health.host.load5.toFixed(2)} · 15m {health.host.load15.toFixed(2)}
				</p>
				<p class="text-muted-foreground text-xs">{health.host.cpu_count} CPU cores</p>
			</Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2 text-sm"><MemoryStick class="size-4" /> Memory</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-2">
				<p class="text-3xl font-bold">{ramPct}%</p>
				<div class="bg-muted h-2 w-full overflow-hidden rounded-full">
					<div
						class="h-full rounded-full {ramPct > 90 ? 'bg-destructive' : 'bg-primary'}"
						style="width: {ramPct}%"
					></div>
				</div>
				<p class="text-muted-foreground text-xs">
					{bytes(health.host.ram_used_bytes)} / {bytes(health.host.ram_total_bytes)}
				</p>
			</Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2 text-sm"><HardDrive class="size-4" /> Disk (/)</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-2">
				<p class="text-3xl font-bold">{diskPct}%</p>
				<div class="bg-muted h-2 w-full overflow-hidden rounded-full">
					<div
						class="h-full rounded-full {diskPct > 90 ? 'bg-destructive' : 'bg-primary'}"
						style="width: {diskPct}%"
					></div>
				</div>
				<p class="text-muted-foreground text-xs">
					{bytes(health.host.disk_used_bytes)} / {bytes(health.host.disk_total_bytes)}
				</p>
			</Card.Content>
		</Card.Root>
	</div>

	<!-- Details -->
	<div class="grid gap-4 lg:grid-cols-2">
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2 text-sm"><Clock class="size-4" /> Host</Card.Title>
			</Card.Header>
			<Card.Content>
				<Table.Root>
					<Table.Body>
						<Table.Row>
							<Table.Cell class="text-muted-foreground">Hostname</Table.Cell>
							<Table.Cell class="font-medium">{health.host.hostname}</Table.Cell>
						</Table.Row>
						<Table.Row>
							<Table.Cell class="text-muted-foreground">Uptime</Table.Cell>
							<Table.Cell class="font-medium">{uptime(health.host.uptime_seconds)}</Table.Cell>
						</Table.Row>
						<Table.Row>
							<Table.Cell class="text-muted-foreground">CPU cores</Table.Cell>
							<Table.Cell class="font-medium">{health.host.cpu_count}</Table.Cell>
						</Table.Row>
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2 text-sm"><Activity class="size-4" /> RADIUS Auth Since Panel Start</Card.Title>
			</Card.Header>
			<Card.Content>
				<Table.Root>
					<Table.Body>
						<Table.Row>
							<Table.Cell class="text-muted-foreground">Accepted</Table.Cell>
							<Table.Cell class="font-medium">{health.radius.auth_ok}</Table.Cell>
						</Table.Row>
						<Table.Row>
							<Table.Cell class="text-muted-foreground">Rejected</Table.Cell>
							<Table.Cell class="font-medium">{health.radius.auth_fail}</Table.Cell>
						</Table.Row>
						<Table.Row>
							<Table.Cell class="text-muted-foreground">Failed logins (panel)</Table.Cell>
							<Table.Cell class="font-medium">{health.appdb.counts.failed_logins ?? 0}</Table.Cell>
						</Table.Row>
						<Table.Row>
							<Table.Cell class="text-muted-foreground">Activity entries</Table.Cell>
							<Table.Cell class="font-medium">{health.appdb.counts.activity ?? 0}</Table.Cell>
						</Table.Row>
					</Table.Body>
				</Table.Root>
				<p class="text-muted-foreground mt-3 text-xs">
					Auth counters are read from radius.log and only increase while FreeRADIUS logs auth results
					(<code>log auth = yes</code>).
				</p>
			</Card.Content>
		</Card.Root>
	</div>
{/if}

<!-- Restart FreeRADIUS -->
<AlertDialog.Root bind:open={restartOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Restart FreeRADIUS?</AlertDialog.Title>
			<AlertDialog.Description>
				The FreeRADIUS service is restarted via the configured reload command. Active sessions are dropped and the
				server re-reads its config (including the NAS client list).
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={doRestart}>Restart</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
