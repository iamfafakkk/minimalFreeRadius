<script lang="ts">
	import type { ActionData } from './$types';
	import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import { LogIn } from 'lucide-svelte';

	export let form: ActionData;
	let loading = false;
</script>

<div class="flex min-h-screen items-center justify-center bg-muted/40 p-4">
	<Card class="w-full max-w-sm">
		<CardHeader>
			<CardTitle>FreeRADIUS Panel</CardTitle>
			<CardDescription>Masuk untuk mengelola NAS dan user.</CardDescription>
		</CardHeader>
		<CardContent>
			<form method="POST" action="?/login" class="space-y-4" on:submit={() => (loading = true)}>
				<div class="space-y-2">
					<Label for="username">Username</Label>
					<Input id="username" name="username" autocomplete="username" required />
				</div>
				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input id="password" name="password" type="password" autocomplete="current-password" required />
				</div>
				{#if form?.error}
					<p class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
						{form.error}
					</p>
				{/if}
				<Button class="w-full" disabled={loading}>
					<LogIn class="mr-2 h-4 w-4" />
					{loading ? 'Memeriksa...' : 'Masuk'}
				</Button>
			</form>
		</CardContent>
	</Card>
</div>
