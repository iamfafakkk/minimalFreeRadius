<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	export let open = false;
	export let onOpenChange: ((open: boolean) => void) | undefined = undefined;
</script>

<DialogPrimitive.Root bind:open {onOpenChange}>
	<slot />
</DialogPrimitive.Root>
