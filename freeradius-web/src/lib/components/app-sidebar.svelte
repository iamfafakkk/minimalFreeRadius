<script lang="ts">
	import { page } from '$app/stores';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { LayoutDashboard, Server, Users, RadioTower, LogOut, ShieldCheck, ScrollText, Activity, FileCode, HeartPulse } from '@lucide/svelte';

	let { username, onlogout }: { username: string; onlogout: () => void } = $props();

	const items = [
		{ title: 'Dashboard', url: '/dashboard', icon: LayoutDashboard },
		{ title: 'NAS', url: '/dashboard/nas', icon: Server },
		{ title: 'Users', url: '/dashboard/users', icon: Users },
		{ title: 'Radius Logs', url: '/dashboard/radius-logs', icon: Activity },
		{ title: 'System Health', url: '/dashboard/system', icon: HeartPulse },
		{ title: 'Logs', url: '/dashboard/logs', icon: ScrollText }
	];
</script>

<Sidebar.Root collapsible="icon">
	<Sidebar.Header>
		<Sidebar.Menu>
			<Sidebar.MenuItem>
				<Sidebar.MenuButton size="lg" class="data-[active=true]:bg-transparent">
					{#snippet child({ props })}
						<a href="/dashboard" {...props}>
							<div
								class="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg"
							>
								<RadioTower class="size-4" />
							</div>
							<div class="grid flex-1 text-left text-sm leading-tight">
								<span class="truncate font-semibold">FreeRADIUS</span>
								<span class="text-muted-foreground truncate text-xs">Admin Panel</span>
							</div>
						</a>
					{/snippet}
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
	</Sidebar.Header>

	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel>Menu</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					{#each items as item (item.url)}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton isActive={$page.url.pathname === item.url}>
								{#snippet child({ props })}
									<a href={item.url} {...props}>
										<item.icon />
										<span>{item.title}</span>
									</a>
								{/snippet}
								{#snippet tooltipContent()}
									<span>{item.title}</span>
								{/snippet}
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
					<Sidebar.MenuItem>
						<Sidebar.MenuButton>
							{#snippet child({ props })}
								<a href="/api-docs" target="_blank" rel="noopener" {...props}>
									<FileCode />
									<span>API Docs</span>
								</a>
							{/snippet}
							{#snippet tooltipContent()}
								<span>API Docs</span>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>

	<Sidebar.Footer>
		<Sidebar.Menu>
			<Sidebar.MenuItem>
				<Sidebar.MenuButton size="lg">
					{#snippet child({ props })}
						<div {...props}>
							<div
								class="bg-sidebar-accent text-sidebar-accent-foreground flex aspect-square size-8 items-center justify-center rounded-lg"
							>
								<ShieldCheck class="size-4" />
							</div>
							<div class="grid flex-1 text-left text-sm leading-tight">
								<span class="truncate font-semibold">{username}</span>
								<span class="text-muted-foreground truncate text-xs">Administrator</span>
							</div>
						</div>
					{/snippet}
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
		<Button
			variant="outline"
			size="sm"
			onclick={onlogout}
			class="group-data-[collapsible=icon]:hidden"
		>
			<LogOut />
			Sign out
		</Button>
	</Sidebar.Footer>

	<Sidebar.Rail />
</Sidebar.Root>
