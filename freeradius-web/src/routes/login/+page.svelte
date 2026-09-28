<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import { LogIn, Loader2, RadioTower } from 'lucide-svelte';

	let username = '';
	let password = '';
	let error = '';
	let loading = false;

	async function login() {
		if (loading) return;
		error = '';
		loading = true;
		try {
			// Backend Go menyet cookie sesi httpOnly fr_token (+ fr_user).
			const r = await api.post('/v1/auth/login', { username: username.trim(), password });
			if (!r.ok) {
				error = r.message ?? 'Login gagal.';
				return;
			}
			await goto('/dashboard');
		} catch {
			error = 'Backend tidak dapat dihubungi.';
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-muted/40 p-4">
	<Card class="w-full max-w-sm shadow-lg">
		<CardHeader class="space-y-3 text-center">
			<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground">
				<RadioTower class="h-6 w-6" />
			</div>
			<div class="space-y-1">
				<CardTitle class="text-xl">FreeRADIUS Panel</CardTitle>
				<CardDescription>Masuk untuk mengelola NAS dan user.</CardDescription>
			</div>
		</CardHeader>
		<CardContent>
			<form class="space-y-4" on:submit|preventDefault={login}>
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
						class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
					>
						{error}
					</p>
				{/if}
				<Button type="submit" class="w-full" disabled={loading}>
					{#if loading}
						<Loader2 class="h-4 w-4 animate-spin" />
						Memeriksa...
					{:else}
						<LogIn class="h-4 w-4" />
						Masuk
					{/if}
				</Button>
			</form>
		</CardContent>
	</Card>
</div>
