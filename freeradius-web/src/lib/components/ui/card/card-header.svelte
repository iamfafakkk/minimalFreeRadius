<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<div class={cn('flex flex-col space-y-1.5 p-6', klass)} {...$$restProps}>
	<slot />
</div>
