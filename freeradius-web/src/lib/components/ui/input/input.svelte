<script lang="ts">
	import { cn } from '$lib/utils.js';
	export let value: string | number | undefined = undefined;
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<input
	class={cn(
		'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
		klass
	)}
	bind:value
	on:input
	on:change
	on:keydown
	on:blur
	on:focus
	{...$$restProps}
/>
