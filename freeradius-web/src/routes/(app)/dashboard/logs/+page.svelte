<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type ActivityRecord, type LoginRecord, type AppUser } from '$lib/api.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { RefreshCw, LoaderCircle, CheckCircle2, XCircle } from '@lucide/svelte';

	let logins = $state<LoginRecord[]>([]);
	let activity = $state<ActivityRecord[]>([]);
	let appUsers = $state<AppUser[]>([]);
	let loading = $state(true);
	let err = $state('');

	async function load() {
		err = '';
		loading = true;
		const [l, a, u] = await Promise.all([
			api.get<LoginRecord[]>('/v1/system/login-history?limit=50'),
			api.get<ActivityRecord[]>('/v1/system/activity?limit=100'),
			api.get<AppUser[]>('/v1/system/users')
		]);
		loading = false;
		if (!l.ok || !a.ok || !u.ok) {
			err = l.message ?? a.message ?? u.message ?? 'Failed to load logs.';
			return;
		}
		logins = (l.data as LoginRecord[]) ?? [];
		activity = (a.data as ActivityRecord[]) ?? [];
		appUsers = (u.data as AppUser[]) ?? [];
	}

	onMount(load);

	function actionVariant(action: string): 'default' | 'secondary' | 'destructive' {
		if (action === 'delete') return 'destructive';
		if (action === 'create') return 'default';
		return 'secondary';
	}

	function statusVariant(status: number): 'default' | 'secondary' | 'destructive' {
		if (status >= 500) return 'destructive';
		if (status >= 400) return 'destructive';
		if (status >= 200 && status < 300) return 'default';
		return 'secondary';
	}
</script>

<div class="flex items-start justify-between gap-4">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">Logs</h1>
		<p class="text-muted-foreground text-sm">Login history and activity audit — {logins.length} logins, {activity.length} actions.</p>
	</div>
	<Button variant="outline" size="sm" onclick={load} disabled={loading}>
		{#if loading}<LoaderCircle class="animate-spin" />{:else}<RefreshCw />{/if}
		Refresh
	</Button>
</div>

{#if err}
	<p class="border-destructive/30 bg-destructive/10 text-destructive rounded-md border px-3 py-2 text-sm">{err}</p>
{/if}

<div class="grid gap-4 md:grid-cols-3">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-sm">Login History</Card.Title>
			<Card.Description>Recent sign-in attempts.</Card.Description>
		</Card.Header>
		<Card.Content>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>User</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head>IP</Table.Head>
						<Table.Head>Time</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each logins as row (row.id)}
						<Table.Row>
							<Table.Cell class="font-medium">{row.username}</Table.Cell>
							<Table.Cell>
								{#if row.success}
									<Badge variant="default" class="gap-1"><CheckCircle2 class="size-3" /> Success</Badge>
								{:else}
									<Badge variant="destructive" class="gap-1"><XCircle class="size-3" /> Failed</Badge>
								{/if}
							</Table.Cell>
							<Table.Cell class="text-muted-foreground text-xs">{row.ip}</Table.Cell>
							<Table.Cell class="text-muted-foreground text-xs">{row.created_at}</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row>
							<Table.Cell colspan={4} class="text-muted-foreground h-24 text-center">
								{#if loading}<LoaderCircle class="mx-auto size-4 animate-spin" />{:else}No logins yet.{/if}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Content>
	</Card.Root>

	<Card.Root class="md:col-span-2">
		<Card.Header>
			<Card.Title class="text-sm">Activity Audit</Card.Title>
			<Card.Description>Mutating actions (create / update / delete).</Card.Description>
		</Card.Header>
		<Card.Content>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>User</Table.Head>
						<Table.Head>Action</Table.Head>
						<Table.Head>Method</Table.Head>
						<Table.Head>Path</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head>IP</Table.Head>
						<Table.Head>Time</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each activity as row (row.id)}
						<Table.Row>
							<Table.Cell class="font-medium">{row.username || '—'}</Table.Cell>
							<Table.Cell><Badge variant={actionVariant(row.action)}>{row.action}</Badge></Table.Cell>
							<Table.Cell class="text-muted-foreground text-xs">{row.method}</Table.Cell>
							<Table.Cell class="text-muted-foreground max-w-[16rem] truncate text-xs">{row.path}</Table.Cell>
							<Table.Cell><Badge variant={statusVariant(row.status)}>{row.status}</Badge></Table.Cell>
							<Table.Cell class="text-muted-foreground text-xs">{row.ip}</Table.Cell>
							<Table.Cell class="text-muted-foreground text-xs">{row.created_at}</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row>
							<Table.Cell colspan={7} class="text-muted-foreground h-24 text-center">
								{#if loading}<LoaderCircle class="mx-auto size-4 animate-spin" />{:else}No activity yet.{/if}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Content>
	</Card.Root>
</div>

<Card.Root>
	<Card.Header>
		<Card.Title class="text-sm">App Users</Card.Title>
		<Card.Description>Admin accounts stored in the SQLite app database.</Card.Description>
	</Card.Header>
	<Card.Content>
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>ID</Table.Head>
					<Table.Head>Username</Table.Head>
					<Table.Head>Role</Table.Head>
					<Table.Head>Created</Table.Head>
					<Table.Head>Last Login</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each appUsers as u (u.id)}
					<Table.Row>
						<Table.Cell>{u.id}</Table.Cell>
						<Table.Cell class="font-medium">{u.username}</Table.Cell>
						<Table.Cell><Badge variant="secondary">{u.role}</Badge></Table.Cell>
						<Table.Cell class="text-muted-foreground text-xs">{u.created_at}</Table.Cell>
						<Table.Cell class="text-muted-foreground text-xs">{u.last_login_at || '—'}</Table.Cell>
					</Table.Row>
				{:else}
					<Table.Row>
						<Table.Cell colspan={5} class="text-muted-foreground h-24 text-center">
							{#if loading}<LoaderCircle class="mx-auto size-4 animate-spin" />{:else}No users yet.{/if}
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</Card.Content>
</Card.Root>
