<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import ThemeSwitcher from '$lib/components/theme-switcher.svelte';
	import { LogIn, LoaderCircle, RadioTower } from '@lucide/svelte';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function login() {
		if (loading) return;
		error = '';
		loading = true;
		try {
			// Backend Go sets the httpOnly session cookie fr_token (+ fr_user).
			const r = await api.post('/v1/auth/login', { username: username.trim(), password });
			if (!r.ok) {
				error = r.message ?? 'Login failed.';
				return;
			}
			await goto('/dashboard');
		} catch {
			error = 'Cannot reach the backend.';
		} finally {
			loading = false;
		}
	}
</script>

<div class="bg-muted/40 relative flex min-h-screen items-center justify-center p-4">
	<div class="absolute right-4 top-4"><ThemeSwitcher /></div>
	<Card.Root class="w-full max-w-sm shadow-lg">
		<Card.Header class="space-y-3 text-center">
			<div class="bg-primary text-primary-foreground mx-auto flex size-12 items-center justify-center rounded-xl">
				<RadioTower class="size-6" />
			</div>
			<div class="space-y-1">
				<Card.Title class="text-xl">FreeRADIUS Panel</Card.Title>
				<Card.Description>Sign in to manage NAS and users.</Card.Description>
			</div>
		</Card.Header>
		<Card.Content>
			<form class="space-y-4" onsubmit={(e) => { e.preventDefault(); login(); }}>
				<div class="space-y-2">
					<Label for="username">Username</Label>
					<Input
						id="username"
						bind:value={username}
						autocomplete="username"
						placeholder="admin"
						disabled={loading}
						required
					/>
				</div>
				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input
						id="password"
						type="password"
						bind:value={password}
						autocomplete="current-password"
						placeholder="••••••••"
						disabled={loading}
						required
					/>
				</div>
				{#if error}
					<p
						role="alert"
						class="border-destructive/30 bg-destructive/10 text-destructive rounded-md border px-3 py-2 text-sm"
					>
						{error}
					</p>
				{/if}
				<Button type="submit" class="w-full" disabled={loading}>
					{#if loading}
						<LoaderCircle class="animate-spin" />
						Checking...
					{:else}
						<LogIn />
						Sign in
					{/if}
				</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>
