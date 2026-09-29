<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type ActivityRecord, type LoginRecord, type AppUser } from '$lib/api.js';
	import { toast } from 'svelte-sonner';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import {
		Select,
		SelectContent,
		SelectItem,
		SelectTrigger
	} from '$lib/components/ui/select/index.js';
	import { RefreshCw, LoaderCircle, CheckCircle2, XCircle, Users, Activity } from '@lucide/svelte';

	let logins = $state<LoginRecord[]>([]);
	let activity = $state<ActivityRecord[]>([]);
	let appUsers = $state<AppUser[]>([]);
	let loading = $state(true);
	let updatedAt = $state<Date | null>(null);

	let tab = $state('logins');
	let loginSearch = $state('');
	let loginStatus = $state('all');
	let actionSearch = $state('');
	let actionFilter = $state('all');
	let userSearch = $state('');

	const LOGIN_STATUS: { value: string; label: string }[] = [
		{ value: 'all', label: 'All statuses' },
		{ value: 'success', label: 'Success' },
		{ value: 'failed', label: 'Failed' }
	];
	const ACTION_FILTER: { value: string; label: string }[] = [
		{ value: 'all', label: 'All actions' },
		{ value: 'create', label: 'Create' },
		{ value: 'update', label: 'Update' },
		{ value: 'delete', label: 'Delete' }
	];

	async function load() {
		loading = true;
		const [l, a, u] = await Promise.all([
			api.get<LoginRecord[]>('/v1/system/login-history?limit=200'),
			api.get<ActivityRecord[]>('/v1/system/activity?limit=200'),
			api.get<AppUser[]>('/v1/system/users')
		]);
		loading = false;
		if (!l.ok || !a.ok || !u.ok) {
			toast.error(l.message ?? a.message ?? u.message ?? 'Failed to load logs.');
			return;
		}
		logins = (l.data as LoginRecord[]) ?? [];
		activity = (a.data as ActivityRecord[]) ?? [];
		appUsers = (u.data as AppUser[]) ?? [];
		updatedAt = new Date();
	}

	onMount(load);

	const failedLogins = $derived(logins.filter((l) => !l.success).length);

	const shownLogins = $derived.by(() => {
		const q = loginSearch.trim().toLowerCase();
		return logins.filter((l) => {
			if (loginStatus === 'success' && !l.success) return false;
			if (loginStatus === 'failed' && l.success) return false;
			if (!q) return true;
			return (l.username + ' ' + l.ip + ' ' + l.user_agent).toLowerCase().includes(q);
		});
	});

	const shownActivity = $derived.by(() => {
		const q = actionSearch.trim().toLowerCase();
		return activity.filter((a) => {
			if (actionFilter !== 'all' && a.action !== actionFilter) return false;
			if (!q) return true;
			return (a.username + ' ' + a.action + ' ' + a.method + ' ' + a.path + ' ' + a.ip)
				.toLowerCase()
				.includes(q);
		});
	});

	const shownUsers = $derived.by(() => {
		const q = userSearch.trim().toLowerCase();
		if (!q) return appUsers;
		return appUsers.filter((u) => (u.username + ' ' + u.role).toLowerCase().includes(q));
	});

	function actionVariant(action: string): 'default' | 'secondary' | 'destructive' {
		if (action === 'delete') return 'destructive';
		if (action === 'create') return 'default';
		return 'secondary';
	}

	function statusVariant(status: number): 'default' | 'secondary' | 'destructive' {
		if (status >= 400) return 'destructive';
		if (status >= 200 && status < 300) return 'default';
		return 'secondary';
	}
</script>

<div class="flex flex-wrap items-end justify-between gap-3">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">Logs</h1>
		<p class="text-muted-foreground text-sm">
			Login history and activity audit · {logins.length} logins, {activity.length} actions,
			{appUsers.length} admins.
		</p>
	</div>
	<div class="flex items-center gap-2">
		{#if updatedAt}
			<span class="text-muted-foreground text-xs">Updated {updatedAt.toLocaleTimeString()}</span>
		{/if}
		<Button variant="outline" size="sm" onclick={load} disabled={loading}>
			{#if loading}<LoaderCircle class="animate-spin" />{:else}<RefreshCw />{/if}
			Refresh
		</Button>
	</div>
</div>

<div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
	<Card.Root>
		<Card.Content class="p-4">
			<p class="text-muted-foreground text-xs font-medium tracking-wide uppercase">Logins</p>
			<p class="text-2xl font-bold">{logins.length}</p>
		</Card.Content>
	</Card.Root>
	<Card.Root>
		<Card.Content class="p-4">
			<p class="text-muted-foreground text-xs font-medium tracking-wide uppercase">Failed logins</p>
			<p class="text-2xl font-bold {failedLogins > 0 ? 'text-destructive' : ''}">
				{failedLogins}
			</p>
		</Card.Content>
	</Card.Root>
	<Card.Root>
		<Card.Content class="p-4">
			<p class="text-muted-foreground text-xs font-medium tracking-wide uppercase">Actions</p>
			<p class="text-2xl font-bold">{activity.length}</p>
		</Card.Content>
	</Card.Root>
	<Card.Root>
		<Card.Content class="p-4">
			<p class="text-muted-foreground text-xs font-medium tracking-wide uppercase">Admins</p>
			<p class="text-2xl font-bold">{appUsers.length}</p>
		</Card.Content>
	</Card.Root>
</div>

<Tabs.Root bind:value={tab} class="gap-3">
	<Tabs.List>
		<Tabs.Trigger value="logins">
			<CheckCircle2 class="mr-1.5 size-4" /> Login History
			<Badge variant="secondary" class="ml-2">{logins.length}</Badge>
		</Tabs.Trigger>
		<Tabs.Trigger value="activity">
			<Activity class="mr-1.5 size-4" /> Activity Audit
			<Badge variant="secondary" class="ml-2">{activity.length}</Badge>
		</Tabs.Trigger>
		<Tabs.Trigger value="users">
			<Users class="mr-1.5 size-4" /> App Users
			<Badge variant="secondary" class="ml-2">{appUsers.length}</Badge>
		</Tabs.Trigger>
	</Tabs.List>

	<Tabs.Content value="logins" class="outline-none">
		<Card.Root class="overflow-hidden">
			<Card.Header>
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<Card.Title class="text-sm">Login History</Card.Title>
						<Card.Description>Recent panel sign-in attempts.</Card.Description>
					</div>
					<div class="flex flex-wrap items-center gap-2">
						<Input
							bind:value={loginSearch}
							placeholder="Search user, IP, agent..."
							class="w-60"
						/>
						<Select type="single" bind:value={loginStatus}>
							<SelectTrigger class="w-[150px]" aria-label="Filter by status">
								{LOGIN_STATUS.find((o) => o.value === loginStatus)?.label}
							</SelectTrigger>
							<SelectContent>
								{#each LOGIN_STATUS as o (o.value)}
									<SelectItem value={o.value} label={o.label}>{o.label}</SelectItem>
								{/each}
							</SelectContent>
						</Select>
					</div>
				</div>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="relative max-h-[30rem] overflow-auto">
					<Table.Root class="overflow-visible">
						<Table.Header class="bg-muted sticky top-0 z-10 [&_tr]:border-b">
							<Table.Row>
								<Table.Head class="text-xs tracking-wider uppercase">User</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Status</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">IP</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">User Agent</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Time</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each shownLogins as row (row.id)}
								<Table.Row>
									<Table.Cell class="font-medium">{row.username}</Table.Cell>
									<Table.Cell>
										{#if row.success}
											<Badge variant="outline" class="gap-1 border-emerald-500/20 bg-emerald-500/15 text-emerald-500">
												<CheckCircle2 class="size-3" /> Success
											</Badge>
										{:else}
											<Badge variant="outline" class="gap-1 border-red-500/20 bg-red-500/15 text-red-500">
												<XCircle class="size-3" /> Failed
											</Badge>
										{/if}
									</Table.Cell>
									<Table.Cell class="text-muted-foreground font-mono text-xs">{row.ip || '—'}</Table.Cell>
									<Table.Cell class="text-muted-foreground max-w-[22rem] truncate text-xs" title={row.user_agent}>
										{row.user_agent || '—'}
									</Table.Cell>
									<Table.Cell class="text-muted-foreground font-mono text-xs whitespace-nowrap">
										{row.created_at}
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="text-muted-foreground h-32 text-center text-sm">
										{#if loading}<LoaderCircle class="mx-auto size-4 animate-spin" />{:else}No logins match the filters.{/if}
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</Card.Content>
		</Card.Root>
	</Tabs.Content>

	<Tabs.Content value="activity" class="outline-none">
		<Card.Root class="overflow-hidden">
			<Card.Header>
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<Card.Title class="text-sm">Activity Audit</Card.Title>
						<Card.Description>Mutating actions (create / update / delete).</Card.Description>
					</div>
					<div class="flex flex-wrap items-center gap-2">
						<Input
							bind:value={actionSearch}
							placeholder="Search user, method, path..."
							class="w-60"
						/>
						<Select type="single" bind:value={actionFilter}>
							<SelectTrigger class="w-[150px]" aria-label="Filter by action">
								{ACTION_FILTER.find((o) => o.value === actionFilter)?.label}
							</SelectTrigger>
							<SelectContent>
								{#each ACTION_FILTER as o (o.value)}
									<SelectItem value={o.value} label={o.label}>{o.label}</SelectItem>
								{/each}
							</SelectContent>
						</Select>
					</div>
				</div>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="relative max-h-[30rem] overflow-auto">
					<Table.Root class="overflow-visible">
						<Table.Header class="bg-muted sticky top-0 z-10 [&_tr]:border-b">
							<Table.Row>
								<Table.Head class="text-xs tracking-wider uppercase">User</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Action</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Method</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Path</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Status</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">IP</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Time</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each shownActivity as row (row.id)}
								<Table.Row>
									<Table.Cell class="font-medium">{row.username || '—'}</Table.Cell>
									<Table.Cell><Badge variant={actionVariant(row.action)}>{row.action}</Badge></Table.Cell>
									<Table.Cell class="text-muted-foreground font-mono text-xs">{row.method}</Table.Cell>
									<Table.Cell class="text-muted-foreground max-w-[18rem] truncate font-mono text-xs" title={row.path}>
										{row.path}
									</Table.Cell>
									<Table.Cell><Badge variant={statusVariant(row.status)}>{row.status}</Badge></Table.Cell>
									<Table.Cell class="text-muted-foreground font-mono text-xs">{row.ip || '—'}</Table.Cell>
									<Table.Cell class="text-muted-foreground font-mono text-xs whitespace-nowrap">
										{row.created_at}
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={7} class="text-muted-foreground h-32 text-center text-sm">
										{#if loading}<LoaderCircle class="mx-auto size-4 animate-spin" />{:else}No activity matches the filters.{/if}
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</Card.Content>
		</Card.Root>
	</Tabs.Content>

	<Tabs.Content value="users" class="outline-none">
		<Card.Root class="overflow-hidden">
			<Card.Header>
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<Card.Title class="text-sm">App Users</Card.Title>
						<Card.Description>Admin accounts stored in the SQLite app database.</Card.Description>
					</div>
					<Input bind:value={userSearch} placeholder="Search username, role..." class="w-60" />
				</div>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="relative max-h-[30rem] overflow-auto">
					<Table.Root class="overflow-visible">
						<Table.Header class="bg-muted sticky top-0 z-10 [&_tr]:border-b">
							<Table.Row>
								<Table.Head class="text-xs tracking-wider uppercase">ID</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Username</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Role</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Created</Table.Head>
								<Table.Head class="text-xs tracking-wider uppercase">Last Login</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each shownUsers as u (u.id)}
								<Table.Row>
									<Table.Cell class="text-muted-foreground font-mono text-xs">{u.id}</Table.Cell>
									<Table.Cell class="font-medium">{u.username}</Table.Cell>
									<Table.Cell><Badge variant="secondary">{u.role}</Badge></Table.Cell>
									<Table.Cell class="text-muted-foreground font-mono text-xs">{u.created_at}</Table.Cell>
									<Table.Cell class="text-muted-foreground font-mono text-xs">
										{u.last_login_at || '—'}
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="text-muted-foreground h-32 text-center text-sm">
										{#if loading}<LoaderCircle class="mx-auto size-4 animate-spin" />{:else}No users match the filters.{/if}
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</Card.Content>
		</Card.Root>
	</Tabs.Content>
</Tabs.Root>
