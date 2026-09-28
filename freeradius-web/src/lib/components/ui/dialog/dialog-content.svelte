<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<DialogPrimitive.Content
	class={cn(
		'fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg sm:rounded-lg',
		klass
	)}
	{...$$restProps}
>
	<slot />
</DialogPrimitive.Content>
